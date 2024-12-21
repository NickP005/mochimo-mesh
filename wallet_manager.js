import fetch from 'node-fetch';
import CryptoJS from 'crypto-js';
export class WOTS {
    constructor() {
        this.PARAMSN = 32;
        this.WOTSW = 16;
        this.WOTSLOGW = 4;
        this.WOTSLEN2 = 3;
        this.TXSIGLEN = 2144;
        this.TXADDRLEN = 2208;
        this.XMSS_HASH_PADDING_F = 0;
        this.XMSS_HASH_PADDING_PRF = 3;
        this.WOTSLEN1 = (8 * this.PARAMSN / this.WOTSLOGW);
        this.WOTSLEN = this.WOTSLEN1 + this.WOTSLEN2;
        this.WOTSSIGBYTES = this.WOTSLEN * this.PARAMSN;
        this.validate_params();
    }
    generateKeyPairFrom(wots_seed, tag) {
        // if (!wots_seed) {
        //   throw new Error('Seed is required')
        // }
        // Add tag validation
        if (tag !== undefined) {
            if (tag.length !== 24) {
                // Use default tag for invalid length
                tag = undefined;
            }
            else {
                // Check if tag contains only valid hex characters (0-9, A-F)
                const validHex = /^[0-9A-F]{24}$/i;
                if (!validHex.test(tag)) {
                    throw new Error('Invalid tag format');
                }
            }
        }
        const private_seed = this.sha256(wots_seed + "seed");
        const public_seed = this.sha256(wots_seed + "publ");
        const addr_seed = this.sha256(wots_seed + "addr");
        let wots_public = this.public_key_gen(private_seed, public_seed, addr_seed);
        // Create a single array with all components
        const totalLength = wots_public.length + public_seed.length + 20 + 12;
        const result = new Uint8Array(totalLength);
        let offset = 0;
        result.set(wots_public, offset);
        offset += wots_public.length;
        result.set(public_seed, offset);
        offset += public_seed.length;
        result.set(addr_seed.slice(0, 20), offset);
        offset += 20;
        // Add tag
        const tagBytes = !tag || tag.length !== 24
            ? new Uint8Array([66, 0, 0, 0, 14, 0, 0, 0, 1, 0, 0, 0])
            : this.hexToBytes(tag);
        result.set(tagBytes, offset);
        return result;
    }
    generateSignatureFrom(wots_seed, payload) {
        const private_seed = this.sha256(wots_seed + "seed");
        const public_seed = this.sha256(wots_seed + "publ");
        const addr_seed = this.sha256(wots_seed + "addr");
        const to_sign = this.sha256(payload);
        return this.wots_sign(to_sign, private_seed, public_seed, addr_seed);
    }
    sha256(input) {
        if (typeof input === 'string') {
            const hash = CryptoJS.SHA256(input);
            return new Uint8Array(this.hexToBytes(hash.toString()));
        }
        else {
            const hash = CryptoJS.SHA256(this.bytesToHex(input));
            return new Uint8Array(this.hexToBytes(hash.toString()));
        }
    }
    hexToBytes(hex) {
        const bytes = [];
        for (let i = 0; i < hex.length; i += 2) {
            bytes.push(parseInt(hex.substr(i, 2), 16));
        }
        return bytes;
    }
    bytesToHex(bytes) {
        return Array.from(bytes)
            .map(b => b.toString(16).padStart(2, '0'))
            .join('');
    }
    /**
     * Generates WOTS public key from private key
     */
    public_key_gen(seed, pub_seed, addr_bytes) {
        const private_key = this.expand_seed(seed);
        const public_key = new Uint8Array(this.WOTSSIGBYTES);
        let addr = this.bytes_to_addr(addr_bytes);
        for (let i = 0; i < this.WOTSLEN; i++) {
            this.set_chain_addr(i, addr);
            const private_key_portion = private_key.slice(i * this.PARAMSN, (i + 1) * this.PARAMSN);
            const chain = this.gen_chain(private_key_portion, 0, this.WOTSW - 1, pub_seed, addr);
            public_key.set(chain, i * this.PARAMSN);
        }
        return public_key;
    }
    /**
     * Signs a message using WOTS
     */
    wots_sign(msg, seed, pub_seed, addr_bytes) {
        const private_key = this.expand_seed(seed);
        const signature = new Uint8Array(this.WOTSSIGBYTES);
        const lengths = this.chain_lengths(msg);
        let addr = this.bytes_to_addr(addr_bytes);
        for (let i = 0; i < this.WOTSLEN; i++) {
            this.set_chain_addr(i, addr);
            const private_key_portion = private_key.slice(i * this.PARAMSN, (i + 1) * this.PARAMSN);
            const chain = this.gen_chain(private_key_portion, 0, lengths[i], pub_seed, addr);
            signature.set(chain, i * this.PARAMSN);
        }
        return signature;
    }
    /**
     * Verifies a WOTS signature
     */
    wots_publickey_from_sig(sig, msg, pub_seed, addr_bytes) {
        let addr = this.bytes_to_addr(addr_bytes);
        const lengths = this.chain_lengths(msg);
        const public_key = new Uint8Array(this.WOTSSIGBYTES);
        for (let i = 0; i < this.WOTSLEN; i++) {
            this.set_chain_addr(i, addr);
            const sig_portion = sig.slice(i * this.PARAMSN, (i + 1) * this.PARAMSN);
            const chain = this.gen_chain(sig_portion, lengths[i], this.WOTSW - 1 - lengths[i], pub_seed, addr);
            public_key.set(chain, i * this.PARAMSN);
        }
        return public_key;
    }
    /**
     * Expands seed into private key
     */
    expand_seed(seed) {
        const private_key = new Uint8Array(this.WOTSSIGBYTES);
        for (let i = 0; i < this.WOTSLEN; i++) {
            const ctr = this.ull_to_bytes(this.PARAMSN, [i]);
            const portion = this.prf(ctr, seed);
            private_key.set(portion, i * this.PARAMSN);
        }
        return private_key;
    }
    /**
     * Generates hash chain
     */
    gen_chain(input, start, steps, pub_seed, addr) {
        let out = new Uint8Array(input);
        for (let i = start; i < start + steps && i < this.WOTSW; i++) {
            this.set_hash_addr(i, addr);
            out = this.t_hash(out, pub_seed, addr);
        }
        return out;
    }
    /**
     * Computes PRF using SHA-256
     */
    prf(input, key) {
        const buf = new Uint8Array(32 * 3);
        // Add padding
        buf.set(this.ull_to_bytes(this.PARAMSN, [this.XMSS_HASH_PADDING_PRF]));
        // Add key and input
        const byte_copied_key = this.byte_copy(key, this.PARAMSN);
        buf.set(byte_copied_key, this.PARAMSN);
        const byte_copied_input = this.byte_copy(input, 32);
        buf.set(byte_copied_input, this.PARAMSN * 2);
        return this.sha256(buf);
    }
    /**
     * Computes t_hash for WOTS chain
     */
    t_hash(input, pub_seed, addr) {
        const buf = new Uint8Array(32 * 3);
        let addr_bytes;
        // Add padding
        buf.set(this.ull_to_bytes(this.PARAMSN, [this.XMSS_HASH_PADDING_F]));
        // Get key mask
        this.set_key_and_mask(0, addr);
        addr_bytes = this.addr_to_bytes(addr);
        buf.set(this.prf(addr_bytes, pub_seed), this.PARAMSN);
        // Get bitmask
        this.set_key_and_mask(1, addr);
        addr_bytes = this.addr_to_bytes(addr);
        const bitmask = this.prf(addr_bytes, pub_seed);
        // XOR input with bitmask
        const XOR_bitmask_input = new Uint8Array(input.length);
        for (let i = 0; i < this.PARAMSN; i++) {
            XOR_bitmask_input[i] = input[i] ^ bitmask[i];
        }
        buf.set(XOR_bitmask_input, this.PARAMSN * 2);
        return this.sha256(buf);
    }
    /**
     * Converts number array to bytes with specified length
     */
    ull_to_bytes(outlen, input) {
        const out = new Uint8Array(outlen);
        for (let i = outlen - 1; i >= 0; i--) {
            out[i] = input[i] || 0;
        }
        return out;
    }
    /**
     * Copies bytes with specified length
     */
    byte_copy(source, num_bytes) {
        const output = new Uint8Array(num_bytes);
        for (let i = 0; i < num_bytes; i++) {
            output[i] = source[i] || 0;
        }
        return output;
    }
    /**
     * Converts address to bytes
     */
    addr_to_bytes(addr) {
        const out_bytes = new Uint8Array(32);
        for (let i = 0; i < 8; i++) {
            const chunk = addr[i.toString()] || new Uint8Array(4);
            out_bytes.set(chunk, i * 4);
        }
        return out_bytes;
    }
    /**
     * Converts bytes to address
     */
    bytes_to_addr(addr_bytes) {
        const out_addr = {};
        for (let i = 0; i < 8; i++) {
            out_addr[i.toString()] = this.ull_to_bytes(4, Array.from(addr_bytes.slice(i * 4, (i + 1) * 4)));
        }
        return out_addr;
    }
    /**
     * Sets chain address
     */
    set_chain_addr(chain_address, addr) {
        addr['5'] = new Uint8Array([0, 0, 0, chain_address]);
    }
    /**
     * Sets hash address
     */
    set_hash_addr(hash, addr) {
        addr['6'] = new Uint8Array([0, 0, 0, hash]);
    }
    /**
     * Sets key and mask
     */
    set_key_and_mask(key_and_mask, addr) {
        addr['7'] = new Uint8Array([0, 0, 0, key_and_mask]);
    }
    /**
     * Calculates chain lengths from message
     */
    chain_lengths(msg) {
        const msg_base_w = this.base_w(this.WOTSLEN1, msg);
        const csum_base_w = this.wots_checksum(msg_base_w);
        // Combine message and checksum base-w values
        const lengths = new Uint8Array(this.WOTSLEN);
        lengths.set(msg_base_w);
        lengths.set(csum_base_w, this.WOTSLEN1);
        return lengths;
    }
    /**
     * Converts bytes to base-w representation
     */
    base_w(outlen, input) {
        const output = new Uint8Array(outlen);
        let in_ = 0;
        let total = 0;
        let bits = 0;
        for (let i = 0; i < outlen; i++) {
            if (bits === 0) {
                total = input[in_];
                in_++;
                bits += 8;
            }
            bits -= this.WOTSLOGW;
            output[i] = (total >> bits) & (this.WOTSW - 1);
        }
        return output;
    }
    /**
     * Computes WOTS checksum
     */
    wots_checksum(msg_base_w) {
        let csum = 0;
        // Calculate checksum
        for (let i = 0; i < this.WOTSLEN1; i++) {
            csum += this.WOTSW - 1 - msg_base_w[i];
        }
        // Convert checksum to base_w
        csum = csum << (8 - ((this.WOTSLEN2 * this.WOTSLOGW) % 8));
        const csum_bytes = this.int_to_bytes(csum);
        const csum_base_w = this.base_w(this.WOTSLEN2, this.byte_copy(csum_bytes, Math.floor((this.WOTSLEN2 * this.WOTSLOGW + 7) / 8)));
        return csum_base_w;
    }
    /**
     * Converts integer to bytes
     */
    int_to_bytes(value) {
        const bytes = new Uint8Array(8);
        for (let i = 7; i >= 0; i--) {
            bytes[i] = value & 0xff;
            value = value >> 8;
        }
        return bytes;
    }
    /**
     * Validates input parameters
     */
    validate_params() {
        if (this.PARAMSN !== 32) {
            throw new Error('PARAMSN must be 32');
        }
        if (this.WOTSW !== 16) {
            throw new Error('WOTSW must be 16');
        }
        if (this.WOTSLOGW !== 4) {
            throw new Error('WOTSLOGW must be 4');
        }
    }
    /**
     * Initializes WOTS instance
     */
    init() {
        this.validate_params();
    }
    // Add array extension functionality
    concatUint8Arrays(...arrays) {
        const totalLength = arrays.reduce((acc, arr) => acc + arr.length, 0);
        const result = new Uint8Array(totalLength);
        let offset = 0;
        for (const arr of arrays) {
            result.set(arr, offset);
            offset += arr.length;
        }
        return result;
    }
    /**
     * Verifies a signature
     */
    verifySignature(signature, message, pubSeed, addrSeed) {
        const messageHash = this.sha256(message);
        return this.wots_publickey_from_sig(signature, messageHash, pubSeed, addrSeed);
    }
}
export class MochimoRosettaClient {
    constructor(baseUrl = 'http://ip.leonapp.it:8081') {
        this.baseUrl = baseUrl;
        this.networkIdentifier = {
            blockchain: 'mochimo',
            network: 'mainnet'
        };
    }
    async post(endpoint, data) {
        //console.log(`Sending request to ${this.baseUrl}${endpoint}`);
        console.log('Request data:', JSON.stringify(data, null, 2));
        const response = await fetch(`${this.baseUrl}${endpoint}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });
        const responseData = await response.json();
        console.log('Response:', JSON.stringify(responseData, null, 2));
        if (!response.ok) {
            throw new Error(`API Error: ${JSON.stringify(responseData)}`);
        }
        return responseData;
    }
    async initialize() {
        const [status, options] = await Promise.all([
            this.getNetworkStatus(),
            this.getNetworkOptions()
        ]);
        return { status, options };
    }
    async getNetworkStatus() {
        return this.post('/network/status', {
            network_identifier: this.networkIdentifier
        });
    }
    async getNetworkOptions() {
        return this.post('/network/options', {
            network_identifier: this.networkIdentifier
        });
    }
    async getBlock(identifier) {
        return this.post('/block', {
            network_identifier: this.networkIdentifier,
            block_identifier: identifier
        });
    }
    async getAccountBalance(address) {
        return this.post('/account/balance', {
            network_identifier: this.networkIdentifier,
            account_identifier: { address }
        });
    }
    async constructionDerive(publicKey, curveType = 'wotsp') {
        const request = {
            network_identifier: this.networkIdentifier,
            public_key: {
                hex_bytes: publicKey,
                curve_type: curveType
            }
        };
        return this.post('/construction/derive', request);
    }
    async constructionPreprocess(operations, metadata) {
        const request = {
            network_identifier: this.networkIdentifier,
            operations,
            metadata
        };
        return this.post('/construction/preprocess', request);
    }
    async constructionMetadata(options, publicKeys) {
        const request = {
            network_identifier: this.networkIdentifier,
            options,
            public_keys: publicKeys
        };
        return this.post('/construction/metadata', request);
    }
    async constructionPayloads(operations, metadata, publicKeys) {
        const request = {
            network_identifier: this.networkIdentifier,
            operations,
            metadata,
            public_keys: publicKeys
        };
        return this.post('/construction/payloads', request);
    }
    async constructionParse(transaction, signed) {
        const request = {
            network_identifier: this.networkIdentifier,
            signed,
            transaction
        };
        return this.post('/construction/parse', request);
    }
    async constructionCombine(unsignedTransaction, signatures) {
        const request = {
            network_identifier: this.networkIdentifier,
            unsigned_transaction: unsignedTransaction,
            signatures
        };
        return this.post('/construction/combine', request);
    }
    async constructionHash(signedTransaction) {
        const request = {
            network_identifier: this.networkIdentifier,
            signed_transaction: signedTransaction
        };
        return this.post('/construction/hash', request);
    }
    async constructionSubmit(signedTransaction) {
        const request = {
            network_identifier: this.networkIdentifier,
            signed_transaction: signedTransaction
        };
        return this.post('/construction/submit', request);
    }
}
async function test() {
    // Create a WOTS instance
    const wots = new WOTS();
    wots.init();
    // Generate a key pair
    const seed = "your_seed_here";
    const tag = "OPTIONAL24CHARACTERHEXTAG"; // optional
    const keyPair = wots.generateKeyPairFrom(seed, tag);
    // Create a Rosetta client instance
    const client = new MochimoRosettaClient();
    // Initialize the client
    const networkInfo = await client.initialize();
    // Get account balance
    const balance = await client.getAccountBalance("your_address_here");
    // Create and send a transaction
    const operations = [
    // Define your operations here
    ];
    // Construct and send transaction
    const preprocessResponse = await client.constructionPreprocess(operations);
    const metadataResponse = await client.constructionMetadata(preprocessResponse.options);
    const payloadsResponse = await client.constructionPayloads(operations, metadataResponse.metadata);
    // Sign the transaction using WOTS
    const signature = wots.generateSignatureFrom(seed, new TextEncoder().encode(payloadsResponse.unsigned_transaction));
    // Submit the signed transaction
    const submitResponse = await client.constructionSubmit(signature.toString());
}
test();
