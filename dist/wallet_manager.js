"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.TransactionManager = exports.MochimoRosettaClient = exports.WOTS = void 0;
const node_fetch_1 = __importDefault(require("node-fetch"));
const crypto_1 = require("crypto");
class WOTS {
    constructor(params) {
        this.WOTSLEN2 = 3;
        this.XMSS_HASH_PADDING_F = 0;
        this.XMSS_HASH_PADDING_PRF = 3;
        this.PARAMSN = params?.n ?? 32;
        this.WOTSW = params?.w ?? 16;
        this.WOTSLOGW = params?.logW ?? 4;
        this.WOTSLEN1 = (8 * this.PARAMSN / this.WOTSLOGW);
        this.WOTSLEN = this.WOTSLEN1 + this.WOTSLEN2;
        this.WOTSSIGBYTES = this.WOTSLEN * this.PARAMSN;
    }
    generateKeyPairFrom(seed, tag) {
        // seed = sha256_ascii( seed + "seed")
        const encoder = new TextEncoder();
        const seedBytes = this.sha256(encoder.encode(seed + 'seed'));
        const pubSeed = this.sha256(encoder.encode(seed + 'publ'));
        const addr = this.sha256(encoder.encode(seed + 'addr'));
        const wots = this.wotsPublicKeyGen(seedBytes, pubSeed, addr);
        // append tag to the public key
        const tagBytes = this.hexToBytes(tag);
        const publicKey = new Uint8Array(wots.length + pubSeed.length + 20 + tagBytes.length);
        publicKey.set(wots);
        publicKey.set(pubSeed, wots.length);
        // first 20 bytes of pubseed
        publicKey.set(pubSeed.slice(0, 20), wots.length + pubSeed.length);
        publicKey.set(tagBytes, wots.length + pubSeed.length + 20);
        console.log("publicKey length", publicKey.length);
        return publicKey;
    }
    generateSignatureFrom(seed, payload) {
        const encoder = new TextEncoder();
        const seedBytes = this.sha256(encoder.encode(seed + 'seed'));
        const pubSeed = this.sha256(encoder.encode(seed + 'publ'));
        const addr = this.sha256(encoder.encode(seed + 'addr'));
        const message = this.sha256(payload);
        console.log("message to sign in hex", this.bytesToHex(message));
        return this.wotsSign(message, seedBytes, pubSeed, addr);
    }
    wotsPublicKeyGen(seed, pubSeed, addrBytes) {
        const addr = this.bytesToAddr(addrBytes);
        const privateKey = this.expandSeed(seed);
        const cachePk = new Uint8Array(this.WOTSSIGBYTES);
        let offset = 0;
        for (let i = 0; i < this.WOTSLEN; i++) {
            this.setChainAddr(i, addr);
            const privKeyPortion = privateKey.slice(i * this.PARAMSN, (i + 1) * this.PARAMSN);
            const chain = this.genChain(privKeyPortion, 0, this.WOTSW - 1, pubSeed, addr);
            cachePk.set(chain, offset);
            offset += this.PARAMSN;
        }
        return cachePk;
    }
    wotsSign(msg, seed, pubSeed, addrBytes) {
        const addr = this.bytesToAddr(addrBytes);
        const lengths = this.chainLengths(msg);
        const signature = new Uint8Array(this.WOTSSIGBYTES);
        const privateKey = this.expandSeed(seed);
        let offset = 0;
        for (let i = 0; i < this.WOTSLEN; i++) {
            this.setChainAddr(i, addr);
            const privKeyPortion = privateKey.slice(i * this.PARAMSN, (i + 1) * this.PARAMSN);
            const chain = this.genChain(privKeyPortion, 0, lengths[i], pubSeed, addr);
            signature.set(chain, offset);
            offset += this.PARAMSN;
        }
        return signature;
    }
    wotsVerify(sig, msg, pubSeed, addrBytes) {
        const addr = this.bytesToAddr(addrBytes);
        const lengths = this.chainLengths(msg);
        const publicKey = new Uint8Array(this.WOTSSIGBYTES);
        let offset = 0;
        for (let i = 0; i < this.WOTSLEN; i++) {
            this.setChainAddr(i, addr);
            const sigPortion = sig.slice(i * this.PARAMSN, (i + 1) * this.PARAMSN);
            const chain = this.genChain(sigPortion, lengths[i], this.WOTSW - 1 - lengths[i], pubSeed, addr);
            publicKey.set(chain, offset);
            offset += this.PARAMSN;
        }
        return publicKey;
    }
    sha256(input) {
        const hash = (0, crypto_1.createHash)('sha256');
        hash.update(Buffer.from(input));
        return new Uint8Array(hash.digest());
    }
    expandSeed(seed) {
        const outSeeds = new Uint8Array(this.WOTSLEN * this.PARAMSN);
        for (let i = 0; i < this.WOTSLEN; i++) {
            const ctr = this.ullToBytes(this.PARAMSN, new Uint8Array([i]));
            const expanded = this.prf(ctr, seed);
            outSeeds.set(expanded, i * this.PARAMSN);
        }
        return outSeeds;
    }
    prf(input, key) {
        const buf = new Uint8Array(this.PARAMSN * 3);
        let offset = 0;
        buf.set(this.ullToBytes(this.PARAMSN, new Uint8Array([this.XMSS_HASH_PADDING_PRF])), offset);
        offset += this.PARAMSN;
        buf.set(this.byteCopy(key, this.PARAMSN), offset);
        offset += this.PARAMSN;
        buf.set(this.byteCopy(input, this.PARAMSN), offset);
        return this.sha256(buf);
    }
    genChain(input, start, steps, pubSeed, addr) {
        let out = this.byteCopy(input, this.PARAMSN);
        for (let i = start; i < (start + steps) && i < this.WOTSW; i++) {
            this.setHashAddr(i, addr);
            out = this.tHash(out, pubSeed, addr);
        }
        return out;
    }
    tHash(input, pubSeed, addr) {
        const buf = new Uint8Array(this.PARAMSN * 3);
        let offset = 0;
        buf.set(this.ullToBytes(this.PARAMSN, new Uint8Array([this.XMSS_HASH_PADDING_F])), offset);
        offset += this.PARAMSN;
        this.setKeyAndMask(0, addr);
        const addrAsBytes = this.addrToBytes(addr);
        const key = this.prf(addrAsBytes, pubSeed);
        buf.set(key, offset);
        offset += this.PARAMSN;
        this.setKeyAndMask(1, addr);
        const bitmask = this.prf(this.addrToBytes(addr), pubSeed);
        const xorInput = new Uint8Array(input.length);
        for (let i = 0; i < input.length; i++) {
            xorInput[i] = input[i] ^ bitmask[i];
        }
        buf.set(xorInput, offset);
        return this.sha256(buf);
    }
    chainLengths(msg) {
        const lengths = this.baseW(this.WOTSLEN1, msg);
        const checksum = this.wotsChecksum(lengths);
        const result = new Uint8Array(this.WOTSLEN);
        result.set(lengths);
        result.set(checksum, this.WOTSLEN1);
        return result;
    }
    baseW(outlen, input) {
        const output = new Uint8Array(outlen);
        let inIdx = 0;
        let outIdx = 0;
        let bits = 0;
        let total = 0;
        for (let consumed = 0; consumed < outlen; consumed++) {
            if (bits === 0) {
                total = input[inIdx++] || 0;
                bits = 8;
            }
            bits -= this.WOTSLOGW;
            output[outIdx++] = (total >> bits) & (this.WOTSW - 1);
        }
        return output;
    }
    wotsChecksum(msgBaseW) {
        let csum = 0;
        for (let i = 0; i < this.WOTSLEN1; i++) {
            csum += this.WOTSW - 1 - msgBaseW[i];
        }
        csum = csum << (8 - ((this.WOTSLEN2 * this.WOTSLOGW) % 8));
        const csumBytes = this.ullToBytes(Math.ceil((this.WOTSLEN2 * this.WOTSLOGW + 7) / 8), this.fromIntToByteArray(csum));
        return this.baseW(this.WOTSLEN2, csumBytes);
    }
    byteCopy(source, numBytes) {
        const result = new Uint8Array(numBytes);
        result.set(source.slice(0, numBytes));
        return result;
    }
    fromIntToByteArray(num) {
        if (num === 0)
            return new Uint8Array([0]);
        const bytes = [];
        while (num > 0) {
            bytes.push(num & 0xff);
            num = num >> 8;
        }
        return new Uint8Array(bytes);
    }
    setChainAddr(chainAddress, addr) {
        addr['5'] = new Uint8Array([0, 0, 0, chainAddress]);
    }
    setHashAddr(hash, addr) {
        addr['6'] = new Uint8Array([0, 0, 0, hash]);
    }
    setKeyAndMask(keyAndMask, addr) {
        addr['7'] = new Uint8Array([0, 0, 0, keyAndMask]);
    }
    addrToBytes(addr) {
        const outBytes = new Uint8Array(32);
        for (let i = 0; i < 8; i++) {
            const key = i.toString();
            const value = addr[key] || new Uint8Array(4);
            outBytes.set(value, i * 4);
        }
        return outBytes;
    }
    bytesToAddr(addrBytes) {
        const addr = {};
        for (let i = 0; i < 8; i++) {
            addr[i.toString()] = this.ullToBytes(4, addrBytes.slice(i * 4, (i + 1) * 4));
        }
        return addr;
    }
    ullToBytes(numBytes, num) {
        const result = new Uint8Array(numBytes);
        result.set(num.slice(0, numBytes));
        return result;
    }
    bytesToHex(bytes) {
        return Buffer.from(bytes).toString('hex');
    }
    hexToBytes(hex) {
        return new Uint8Array(Buffer.from(hex, 'hex'));
    }
}
exports.WOTS = WOTS;
class MochimoRosettaClient {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
        this.networkIdentifier = {
            blockchain: 'mochimo',
            network: 'mainnet'
        };
    }
    async post(endpoint, data) {
        //console.log(`Sending request to ${this.baseUrl}${endpoint}`);
        console.log('Request data to:', endpoint);
        console.log(JSON.stringify(data, null, 2));
        const response = await (0, node_fetch_1.default)(`${this.baseUrl}${endpoint}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });
        const responseData = await response.json();
        //console.log('Response from', endpoint);
        //console.log(JSON.stringify(responseData, null, 2));
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
    // get mempool
    async getMempool() {
        return this.post('/mempool', {
            network_identifier: this.networkIdentifier
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
exports.MochimoRosettaClient = MochimoRosettaClient;
class TransactionManager {
    constructor(client, wots_seed, next_wots_seed, sender_tag, receiver_tag) {
        this.wots = new WOTS();
        this.status_string = 'Initializing...';
        this.client = client;
        this.status_string = 'Generating public key from seed...';
        this.public_key = this.wots.generateKeyPairFrom(wots_seed, sender_tag);
        this.wots_seed = wots_seed;
        this.status_string = 'Generating change public key from seed...';
        this.change_public_key = this.wots.generateKeyPairFrom(next_wots_seed, sender_tag);
        this.receiver_tag = receiver_tag;
        this.status_string = 'Initialized';
    }
    async sendTransaction(amount, miner_fee) {
        // Derive sender address
        this.status_string = 'Deriving the address from API...';
        const senderResponse = await this.client.constructionDerive('0x' + this.wots.bytesToHex(this.public_key));
        const senderAddress = senderResponse.account_identifier;
        // Derive change address
        this.status_string = 'Deriving the change address from API...';
        const changeResponse = await this.client.constructionDerive('0x' + this.wots.bytesToHex(this.change_public_key));
        const changeAddress = changeResponse.account_identifier;
        const operations = [
            {
                operation_identifier: { index: 0 },
                type: 'TRANSFER',
                status: 'SUCCESS',
                account: senderAddress,
                amount: {
                    value: '0',
                    currency: {
                        symbol: 'MCM',
                        decimals: 0
                    }
                }
            },
            {
                operation_identifier: { index: 1 },
                type: 'TRANSFER',
                status: 'SUCCESS',
                account: {
                    address: "0x" + this.receiver_tag,
                },
                amount: {
                    value: '0',
                    currency: {
                        symbol: 'MCM',
                        decimals: 0
                    }
                }
            },
            {
                operation_identifier: { index: 2 },
                type: 'TRANSFER',
                status: 'SUCCESS',
                account: changeAddress,
                amount: {
                    value: '0',
                    currency: {
                        symbol: 'MCM',
                        decimals: 0
                    }
                }
            }
        ];
        // Preprocess
        this.status_string = 'Preprocessing transaction...';
        console.log("status_string", this.status_string);
        const preprocessResponse = await this.client.constructionPreprocess(operations);
        // Get resolved tags and source balance
        this.status_string = 'Getting transaction metadata...';
        console.log("status_string", this.status_string);
        const metadataResponse = await this.client.constructionMetadata(preprocessResponse.options);
        const senderBalance = Number(metadataResponse.metadata.source_balance || '0');
        operations[0].amount.value = (-senderBalance).toString(); // Fix negative conversion
        operations[1].amount.value = amount.toString();
        operations[2].amount.value = (senderBalance - amount - miner_fee).toString();
        // Append operation 3 mining fee
        operations.push({
            operation_identifier: { index: 3 },
            type: 'TRANSFER',
            status: 'SUCCESS',
            account: {
                address: ''
            },
            amount: {
                value: String(miner_fee),
                currency: {
                    symbol: 'MCM',
                    decimals: 0
                }
            }
        });
        // Prepare payloads
        this.status_string = 'Preparing transaction payloads...';
        console.log("status_string", this.status_string);
        const payloadsResponse = await this.client.constructionPayloads(operations, metadataResponse.metadata);
        // Parse unsigned transaction to verify correctness
        this.status_string = 'Parsing unsigned transaction...';
        console.log("status_string", this.status_string);
        const parseResponse = await this.client.constructionParse(payloadsResponse.unsigned_transaction, false);
        // Sign the transaction
        this.status_string = 'Signing transaction...';
        console.log("status_string", this.status_string);
        //const payload = Buffer.from(payloadsResponse.unsigned_transaction, 'hex');
        const payload = this.wots.hexToBytes(payloadsResponse.unsigned_transaction);
        const payloadbytes = new Uint8Array(payload);
        console.log(" payload length", payload.length);
        // hash the transaction
        const signatureBytes = this.wots.generateSignatureFrom(this.wots_seed, payloadbytes);
        // print payload bytes lenght
        console.log("payloadbytes", payloadbytes.length);
        // convert payloadbytes to hex and
        console.log("payloadbytes", this.wots.bytesToHex(payloadbytes));
        // Try to verify the signature
        /*
      const computedPublicKey = this.wots.verifySignature(
          signatureBytes,
          payloadbytes,
          this.wots.sha256(this.wots_seed + 'publ'),
          this.wots.sha256(this.wots_seed + 'addr')
      );
      

      console.log("computedPublicKey", this.wots.bytesToHex(computedPublicKey));
      console.log("public_key", this.wots.bytesToHex(this.public_key));

      // say if they match
      const expectedPublicKeyPart = this.public_key.slice(0, 2144);
      if (this.wots.bytesToHex(computedPublicKey) !== this.wots.bytesToHex(expectedPublicKeyPart)) {
          console.error("Public key mismatch:");
          console.error("Computed:", this.wots.bytesToHex(computedPublicKey));
          console.error("Expected:", this.wots.bytesToHex(expectedPublicKeyPart));
          throw new Error("Signature verification failed");
      }*/
        // Combine transaction
        this.status_string = 'Combining transaction parts...';
        console.log("status_string", this.status_string);
        // Create signature with matching hex bytes
        const signature = {
            signing_payload: {
                hex_bytes: payloadsResponse.unsigned_transaction,
                signature_type: "wotsp"
            },
            public_key: {
                hex_bytes: this.wots.bytesToHex(this.public_key),
                curve_type: "wotsp"
            },
            signature_type: "wotsp",
            hex_bytes: this.wots.bytesToHex(signatureBytes)
        };
        // Verify the hex bytes match before sending
        if (signature.signing_payload.hex_bytes !== payloadsResponse.unsigned_transaction) {
            throw new Error("Signing payload hex bytes must match unsigned transaction");
        }
        const combineResponse = await this.client.constructionCombine(payloadsResponse.unsigned_transaction, [signature]);
        // Parse signed transaction to verify
        this.status_string = 'Verifying signed transaction...';
        const parseSignedResponse = await this.client.constructionParse(combineResponse.signed_transaction, true);
        // Submit transaction
        this.status_string = 'Submitting transaction...';
        console.log("status_string", this.status_string);
        const submitResponse = await this.client.constructionSubmit(combineResponse.signed_transaction);
        this.status_string = 'Transaction submitted successfully';
        console.log("status_string", this.status_string);
        // print the various parts of the hex signed transaction (three public keys 2208 bytes, 3 numbers 8 bytes, a signature 2144 bytes)
        const source_address = combineResponse.signed_transaction.slice(0, 2208 * 2);
        const destination_address = combineResponse.signed_transaction.slice(2208 * 2, 2208 * 2 * 2);
        const change_address = combineResponse.signed_transaction.slice(2208 * 2 * 2, 2208 * 2 * 3);
        const amount_hex = combineResponse.signed_transaction.slice(2208 * 2 * 3, 2208 * 2 * 3 + 8 * 2);
        const change_hex = combineResponse.signed_transaction.slice(2208 * 2 * 3 + 8 * 2, 2208 * 2 * 3 + 8 * 2 * 2);
        const fee_hex = combineResponse.signed_transaction.slice(2208 * 2 * 3 + 8 * 2 * 2, 2208 * 2 * 3 + 8 * 2 * 3);
        const signature_hex = combineResponse.signed_transaction.slice(2208 * 2 * 3 + 8 * 2 * 3, 2208 * 2 * 3 + 8 * 2 * 3 + 2144 * 2);
        console.log("source_address", source_address);
        console.log("destination_address", destination_address);
        console.log("change_address", change_address);
        console.log("amount_hex", amount_hex);
        console.log("change_hex", change_hex);
        console.log("fee_hex", fee_hex);
        console.log("signature_hex", signature_hex);
        console.log("signature original", this.wots.bytesToHex(signatureBytes));
        // print transaction unsigned payload
        console.log("unsigned_transaction", payloadsResponse.unsigned_transaction);
        return submitResponse.transaction_identifier;
    }
}
exports.TransactionManager = TransactionManager;
