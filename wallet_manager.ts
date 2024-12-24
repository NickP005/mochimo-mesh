import fetch from 'node-fetch';
import * as CryptoJS from 'crypto-js'
import { createHash } from 'crypto';

export interface WOTSKeyPair {
    privateKey: string
    publicKey: string
}

export interface WalletAccount {
    index: number
    baseSeed: string
    currentWOTS: WOTSKeyPair
    nextWOTS: WOTSKeyPair
    usedAddresses: string[]  // Track used addresses
    tag: string  // Add tag field
    isActivated?: boolean  // Add activation status
}

export interface MasterWallet {
    mnemonic: string
    masterSeed: Uint8Array
    accounts: { [index: number]: WalletAccount }
    password?: string  // Add password to the interface
}

/////

export interface WOTSAddress {
    [key: string]: Uint8Array;
}

export interface WOTSParams {
    n?: number;      // 32 bytes default
    w?: number;      // 16 default
    logW?: number;   // 4 default
}

export class WOTS {
    private readonly PARAMSN: number;
    private readonly WOTSW: number;
    private readonly WOTSLOGW: number;
    private readonly WOTSLEN1: number;
    private readonly WOTSLEN2: number = 3;
    private readonly WOTSLEN: number;
    private readonly WOTSSIGBYTES: number;
    private readonly XMSS_HASH_PADDING_F: number = 0;
    private readonly XMSS_HASH_PADDING_PRF: number = 3;

    constructor(params?: WOTSParams) {
        this.PARAMSN = params?.n ?? 32;
        this.WOTSW = params?.w ?? 16;
        this.WOTSLOGW = params?.logW ?? 4;
        this.WOTSLEN1 = (8 * this.PARAMSN / this.WOTSLOGW);
        this.WOTSLEN = this.WOTSLEN1 + this.WOTSLEN2;
        this.WOTSSIGBYTES = this.WOTSLEN * this.PARAMSN;
    }

    public generateKeyPairFrom(seed: string, tag: string): Uint8Array {
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

    public generateSignatureFrom(seed: string, payload: Uint8Array): Uint8Array {
        const encoder = new TextEncoder();
        const seedBytes = this.sha256(encoder.encode(seed + 'seed'));
        const pubSeed = this.sha256(encoder.encode(seed + 'publ'));
        const addr = this.sha256(encoder.encode(seed + 'addr'));

        const message = this.sha256(payload);

        console.log("message to sign in hex", this.bytesToHex(message));

        return this.wotsSign(message, seedBytes, pubSeed, addr);
    }


    public wotsPublicKeyGen(
        seed: Uint8Array, 
        pubSeed: Uint8Array, 
        addrBytes: Uint8Array
    ): Uint8Array {
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

    public wotsSign(
        msg: Uint8Array,
        seed: Uint8Array,
        pubSeed: Uint8Array,
        addrBytes: Uint8Array
    ): Uint8Array {
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

    public wotsVerify(
        sig: Uint8Array,
        msg: Uint8Array,
        pubSeed: Uint8Array,
        addrBytes: Uint8Array
    ): Uint8Array {
        const addr = this.bytesToAddr(addrBytes);
        const lengths = this.chainLengths(msg);
        const publicKey = new Uint8Array(this.WOTSSIGBYTES);
        let offset = 0;

        for (let i = 0; i < this.WOTSLEN; i++) {
            this.setChainAddr(i, addr);
            const sigPortion = sig.slice(i * this.PARAMSN, (i + 1) * this.PARAMSN);
            const chain = this.genChain(
                sigPortion,
                lengths[i],
                this.WOTSW - 1 - lengths[i],
                pubSeed,
                addr
            );
            publicKey.set(chain, offset);
            offset += this.PARAMSN;
        }

        return publicKey;
    }

    private sha256(input: Uint8Array): Uint8Array {
        const hash = createHash('sha256');
        hash.update(Buffer.from(input));
        return new Uint8Array(hash.digest());
    }

    private expandSeed(seed: Uint8Array): Uint8Array {
        const outSeeds = new Uint8Array(this.WOTSLEN * this.PARAMSN);
        for (let i = 0; i < this.WOTSLEN; i++) {
            const ctr = this.ullToBytes(this.PARAMSN, new Uint8Array([i]));
            const expanded = this.prf(ctr, seed);
            outSeeds.set(expanded, i * this.PARAMSN);
        }
        return outSeeds;
    }

    private prf(input: Uint8Array, key: Uint8Array): Uint8Array {
        const buf = new Uint8Array(this.PARAMSN * 3);
        let offset = 0;

        buf.set(this.ullToBytes(this.PARAMSN, new Uint8Array([this.XMSS_HASH_PADDING_PRF])), offset);
        offset += this.PARAMSN;
        buf.set(this.byteCopy(key, this.PARAMSN), offset);
        offset += this.PARAMSN;
        buf.set(this.byteCopy(input, this.PARAMSN), offset);

        return this.sha256(buf);
    }

    private genChain(
        input: Uint8Array,
        start: number,
        steps: number,
        pubSeed: Uint8Array,
        addr: WOTSAddress
    ): Uint8Array {
        let out = this.byteCopy(input, this.PARAMSN);
        
        for (let i = start; i < (start + steps) && i < this.WOTSW; i++) {
            this.setHashAddr(i, addr);
            out = this.tHash(out, pubSeed, addr);
        }
        return out;
    }

    private tHash(
        input: Uint8Array,
        pubSeed: Uint8Array,
        addr: WOTSAddress
    ): Uint8Array {
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

    private chainLengths(msg: Uint8Array): Uint8Array {
        const lengths = this.baseW(this.WOTSLEN1, msg);
        const checksum = this.wotsChecksum(lengths);
        const result = new Uint8Array(this.WOTSLEN);
        result.set(lengths);
        result.set(checksum, this.WOTSLEN1);
        return result;
    }

    private baseW(outlen: number, input: Uint8Array): Uint8Array {
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

    private wotsChecksum(msgBaseW: Uint8Array): Uint8Array {
        let csum = 0;
        for (let i = 0; i < this.WOTSLEN1; i++) {
            csum += this.WOTSW - 1 - msgBaseW[i];
        }

        csum = csum << (8 - ((this.WOTSLEN2 * this.WOTSLOGW) % 8));
        const csumBytes = this.ullToBytes(
            Math.ceil((this.WOTSLEN2 * this.WOTSLOGW + 7) / 8),
            this.fromIntToByteArray(csum)
        );

        return this.baseW(this.WOTSLEN2, csumBytes);
    }

    private byteCopy(source: Uint8Array, numBytes: number): Uint8Array {
        const result = new Uint8Array(numBytes);
        result.set(source.slice(0, numBytes));
        return result;
    }

    private fromIntToByteArray(num: number): Uint8Array {
        if (num === 0) return new Uint8Array([0]);
        
        const bytes = [];
        while (num > 0) {
            bytes.push(num & 0xff);
            num = num >> 8;
        }
        return new Uint8Array(bytes);
    }

    private setChainAddr(chainAddress: number, addr: WOTSAddress): void {
        addr['5'] = new Uint8Array([0, 0, 0, chainAddress]);
    }

    private setHashAddr(hash: number, addr: WOTSAddress): void {
        addr['6'] = new Uint8Array([0, 0, 0, hash]);
    }

    private setKeyAndMask(keyAndMask: number, addr: WOTSAddress): void {
        addr['7'] = new Uint8Array([0, 0, 0, keyAndMask]);
    }

    private addrToBytes(addr: WOTSAddress): Uint8Array {
        const outBytes = new Uint8Array(32);
        for (let i = 0; i < 8; i++) {
            const key = i.toString();
            const value = addr[key] || new Uint8Array(4);
            outBytes.set(value, i * 4);
        }
        return outBytes;
    }

    private bytesToAddr(addrBytes: Uint8Array): WOTSAddress {
        const addr: WOTSAddress = {};
        for (let i = 0; i < 8; i++) {
            addr[i.toString()] = this.ullToBytes(4, addrBytes.slice(i * 4, (i + 1) * 4));
        }
        return addr;
    }

    private ullToBytes(numBytes: number, num: Uint8Array): Uint8Array {
        const result = new Uint8Array(numBytes);
        result.set(num.slice(0, numBytes));
        return result;
    }

    public bytesToHex(bytes: Uint8Array): string {
        return Buffer.from(bytes).toString('hex');
    }

    public hexToBytes(hex: string): Uint8Array {
        return new Uint8Array(Buffer.from(hex, 'hex'));
    }
}



//////

export interface Amount {
    value: string;
    currency: {
        symbol: string;
        decimals: number;
    };
}

export interface NetworkIdentifier {
    blockchain: string;
    network: string;
}

export interface BlockIdentifier {
    index?: number;
    hash?: string;
}

export interface TransactionIdentifier {
    hash: string;
}

export interface Operation {
    operation_identifier: {
        index: number;
    };
    type: string;
    status: string;
    account: {
        address: string;
        metadata?: Record<string, any>;
    };
    amount: {
        value: string;  // Changed from number to string
        currency: {
            symbol: string;
            decimals: number;
        };
    };
}

export interface Transaction {
    transaction_identifier: TransactionIdentifier;
    operations: Operation[];
}

export interface Block {
    block_identifier: BlockIdentifier;
    parent_block_identifier: BlockIdentifier;
    timestamp: number;
    transactions: Transaction[];
}

export interface NetworkStatus {
    current_block_identifier: BlockIdentifier;
    genesis_block_identifier: BlockIdentifier;
    current_block_timestamp: number;
}

export interface NetworkOptions {
    version: {
        rosetta_version: string;
        node_version: string;
        middleware_version: string;
    };
    allow: {
        operation_statuses: Array<{
            status: string;
            successful: boolean;
        }>;
        operation_types: string[];
        errors: Array<{
            code: number;
            message: string;
            retriable: boolean;
        }>;
        mempool_coins: boolean;
        transaction_hash_case: string;
    };
}

export interface PublicKey {
    hex_bytes: string;
    curve_type: string;
}

export interface ConstructionDeriveRequest {
    network_identifier: NetworkIdentifier;
    public_key: PublicKey;
    metadata?: Record<string, any>;
}

export interface ConstructionDeriveResponse {
    account_identifier: {
        address: string;
        metadata?: {
            tag?: string;
        };
    };
    metadata?: Record<string, any>;
}

export interface ConstructionPreprocessRequest {
    network_identifier: NetworkIdentifier;
    operations: Operation[];
    metadata?: Record<string, any>;
}

export interface ConstructionPreprocessResponse {
    required_public_keys?: Array<{
        address: string;
        metadata?: {
            tag?: string;
        };
    }>;
    options?: {
        source_address?: string;
        source_tag?: string;
        change_address?: string;
        change_tag?: string;
        destination_tag?: string;
        amount?: string;
        fee?: string;
    };
}

export interface ConstructionMetadataRequest {
    network_identifier: NetworkIdentifier;
    options?: Record<string, any>;
    public_keys?: PublicKey[];
}

export interface ConstructionMetadataResponse {
    metadata: {
        source_balance?: string;
        source_nonce?: number;
        source_tag?: string;
        destination_tag?: string;
        change_tag?: string;
        suggested_fee?: string;
    };
    suggested_fee?: Amount[];
}

export interface ConstructionPayloadsRequest {
    network_identifier: NetworkIdentifier;
    operations: Operation[];
    metadata?: Record<string, any>;
    public_keys?: PublicKey[];
}

export interface ConstructionPayloadsResponse {
    unsigned_transaction: string;
    payloads: Array<{
        address: string;
        hex_bytes: string;
        signature_type: string;
        metadata?: {
            tag?: string;
        };
    }>;
}

export interface ConstructionParseRequest {
    network_identifier: NetworkIdentifier;
    signed: boolean;
    transaction: string;
}

export interface ConstructionParseResponse {
    operations: Operation[];
    account_identifier_signers?: { address: string }[];
    metadata?: Record<string, any>;
}

export interface ConstructionCombineRequest {
    network_identifier: NetworkIdentifier;
    unsigned_transaction: string;
    signatures: Signature[];
}

export interface ConstructionCombineResponse {
    signed_transaction: string;
}

export interface ConstructionHashRequest {
    network_identifier: NetworkIdentifier;
    signed_transaction: string;
}

export interface ConstructionHashResponse {
    transaction_identifier: TransactionIdentifier;
    metadata?: Record<string, any>;
}

export interface ConstructionSubmitRequest {
    network_identifier: NetworkIdentifier;
    signed_transaction: string;
}

export interface ConstructionSubmitResponse {
    transaction_identifier: TransactionIdentifier;
    metadata?: Record<string, any>;
}

export interface SigningPayload {
    hex_bytes: string;
    signature_type: string;
    address?: string;
}

export interface Signature {
    signing_payload: SigningPayload;
    public_key: PublicKey;
    signature_type: string;
    hex_bytes: string;
}

export class MochimoRosettaClient {
    private baseUrl: string;
    public networkIdentifier: NetworkIdentifier;

    constructor(baseUrl: string = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
        this.networkIdentifier = {
            blockchain: 'mochimo',
            network: 'mainnet'
        };
    }

    private async post<T>(endpoint: string, data: any): Promise<T> {
        //console.log(`Sending request to ${this.baseUrl}${endpoint}`);
        console.log('Request data to:', endpoint);
        console.log(JSON.stringify(data, null, 2));
        
        const response = await fetch(`${this.baseUrl}${endpoint}`, {
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

    async initialize(): Promise<{status: NetworkStatus, options: NetworkOptions}> {
        const [status, options] = await Promise.all([
            this.getNetworkStatus(),
            this.getNetworkOptions()
        ]);
        return { status, options };
    }

    async getNetworkStatus(): Promise<NetworkStatus> {
        return this.post<NetworkStatus>('/network/status', {
            network_identifier: this.networkIdentifier
        });
    }

    async getNetworkOptions(): Promise<NetworkOptions> {
        return this.post<NetworkOptions>('/network/options', {
            network_identifier: this.networkIdentifier
        });
    }

    async getBlock(identifier: BlockIdentifier): Promise<{ block: Block }> {
        return this.post<{ block: Block }>('/block', {
            network_identifier: this.networkIdentifier,
            block_identifier: identifier
        });
    }

    async getAccountBalance(address: string): Promise<any> {
        return this.post('/account/balance', {
            network_identifier: this.networkIdentifier,
            account_identifier: { address }
        });
    }

    // get mempool
    async getMempool(): Promise<any> {
        return this.post('/mempool', {
            network_identifier: this.networkIdentifier
        });
    }

    async constructionDerive(publicKey: string, curveType: string = 'wotsp'): Promise<ConstructionDeriveResponse> {
        const request: ConstructionDeriveRequest = {
            network_identifier: this.networkIdentifier,
            public_key: {
                hex_bytes: publicKey,
                curve_type: curveType
            }
        };

        return this.post<ConstructionDeriveResponse>('/construction/derive', request);
    }

    async constructionPreprocess(operations: Operation[], metadata?: Record<string, any>): Promise<ConstructionPreprocessResponse> {
        const request: ConstructionPreprocessRequest = {
            network_identifier: this.networkIdentifier,
            operations,
            metadata
        };
        return this.post<ConstructionPreprocessResponse>('/construction/preprocess', request);
    }

    async constructionMetadata(options?: Record<string, any>, publicKeys?: PublicKey[]): Promise<ConstructionMetadataResponse> {
        const request: ConstructionMetadataRequest = {
            network_identifier: this.networkIdentifier,
            options,
            public_keys: publicKeys
        };
        return this.post<ConstructionMetadataResponse>('/construction/metadata', request);
    }

    async constructionPayloads(
        operations: Operation[], 
        metadata?: Record<string, any>,
        publicKeys?: PublicKey[]
    ): Promise<ConstructionPayloadsResponse> {
        const request: ConstructionPayloadsRequest = {
            network_identifier: this.networkIdentifier,
            operations,
            metadata,
            public_keys: publicKeys
        };
        return this.post<ConstructionPayloadsResponse>('/construction/payloads', request);
    }

    async constructionParse(
        transaction: string,
        signed: boolean
    ): Promise<ConstructionParseResponse> {
        const request: ConstructionParseRequest = {
            network_identifier: this.networkIdentifier,
            signed,
            transaction
        };
        return this.post<ConstructionParseResponse>('/construction/parse', request);
    }

    async constructionCombine(
        unsignedTransaction: string,
        signatures: Signature[]
    ): Promise<ConstructionCombineResponse> {
        const request: ConstructionCombineRequest = {
            network_identifier: this.networkIdentifier,
            unsigned_transaction: unsignedTransaction,
            signatures
        };
        return this.post<ConstructionCombineResponse>('/construction/combine', request);
    }

    async constructionHash(signedTransaction: string): Promise<ConstructionHashResponse> {
        const request: ConstructionHashRequest = {
            network_identifier: this.networkIdentifier,
            signed_transaction: signedTransaction
        };
        return this.post<ConstructionHashResponse>('/construction/hash', request);
    }

    async constructionSubmit(signedTransaction: string): Promise<ConstructionSubmitResponse> {
        const request: ConstructionSubmitRequest = {
            network_identifier: this.networkIdentifier,
            signed_transaction: signedTransaction
        };
        return this.post<ConstructionSubmitResponse>('/construction/submit', request);
    }
}

export class TransactionManager {
    private wots = new WOTS();
    private client: MochimoRosettaClient;
    private public_key: Uint8Array;
    private change_public_key: Uint8Array;
    private receiver_tag: string;
    private wots_seed: string;

    public status_string: string = 'Initializing...';

    constructor(client: MochimoRosettaClient, wots_seed: string, next_wots_seed: string, sender_tag: string, receiver_tag: string) {
        this.client = client;
        this.status_string = 'Generating public key from seed...';
        this.public_key = this.wots.generateKeyPairFrom(wots_seed, sender_tag);
        this.wots_seed = wots_seed;
        this.status_string = 'Generating change public key from seed...';
        this.change_public_key = this.wots.generateKeyPairFrom(next_wots_seed, sender_tag);
        this.receiver_tag = receiver_tag;
        this.status_string = 'Initialized';
    }

    async sendTransaction(amount: number, miner_fee: number): Promise<TransactionIdentifier> {
        // Derive sender address
        this.status_string = 'Deriving the address from API...';
        const senderResponse = await this.client.constructionDerive('0x' + this.wots.bytesToHex(this.public_key));
        const senderAddress = senderResponse.account_identifier;

        // Derive change address
        this.status_string = 'Deriving the change address from API...';
        const changeResponse = await this.client.constructionDerive('0x' + this.wots.bytesToHex(this.change_public_key));
        const changeAddress = changeResponse.account_identifier;

        const operations: Operation[] = [
            {
                operation_identifier: { index: 0 },
                type: 'TRANSFER',
                status: 'SUCCESS',
                account: senderAddress,
                amount: {
                    value: '0',  // Changed to string
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
                    value: '0',  // Changed to string
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
                    value: '0',  // Changed to string
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

        const senderBalance: number = Number(metadataResponse.metadata.source_balance || '0');
        operations[0].amount.value = (-senderBalance).toString();  // Fix negative conversion
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
                value: String(miner_fee),  // Convert to string
                currency: {
                    symbol: 'MCM',
                    decimals: 0
                }
            }
        });

        // Prepare payloads
        this.status_string = 'Preparing transaction payloads...';
        console.log("status_string", this.status_string);
        const payloadsResponse = await this.client.constructionPayloads(
            operations,
            metadataResponse.metadata,
        );

        // Parse unsigned transaction to verify correctness
        this.status_string = 'Parsing unsigned transaction...';
        console.log("status_string", this.status_string);
        const parseResponse = await this.client.constructionParse(
            payloadsResponse.unsigned_transaction,
            false
        );

        // Sign the transaction
        this.status_string = 'Signing transaction...';
        console.log("status_string", this.status_string);
        //const payload = Buffer.from(payloadsResponse.unsigned_transaction, 'hex');
        const payload = this.wots.hexToBytes(payloadsResponse.unsigned_transaction);
        const payloadbytes = new Uint8Array(payload);
        console.log(" payload length", payload.length);
        // hash the transaction

        const signatureBytes = this.wots.generateSignatureFrom(
            this.wots_seed,
            payloadbytes);

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
        const signature: Signature = {
            signing_payload: {
                hex_bytes: payloadsResponse.unsigned_transaction, // Must match unsigned_transaction exactly
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

        const combineResponse = await this.client.constructionCombine(
            payloadsResponse.unsigned_transaction,
            [signature]
        );

        // Parse signed transaction to verify
        this.status_string = 'Verifying signed transaction...';
        const parseSignedResponse = await this.client.constructionParse(
            combineResponse.signed_transaction,
            true
        );

        // Submit transaction
        this.status_string = 'Submitting transaction...';
        console.log("status_string", this.status_string);
        const submitResponse = await this.client.constructionSubmit(
            combineResponse.signed_transaction
        );

        this.status_string = 'Transaction submitted successfully';
        console.log("status_string", this.status_string);

        // print the various parts of the hex signed transaction (three public keys 2208 bytes, 3 numbers 8 bytes, a signature 2144 bytes)
        const source_address = combineResponse.signed_transaction.slice(0, 2208*2);
        const destination_address = combineResponse.signed_transaction.slice(2208*2, 2208*2*2);
        const change_address = combineResponse.signed_transaction.slice(2208*2*2, 2208*2*3);
        const amount_hex = combineResponse.signed_transaction.slice(2208*2*3, 2208*2*3 + 8*2);
        const change_hex = combineResponse.signed_transaction.slice(2208*2*3 + 8*2, 2208*2*3 + 8*2*2);
        const fee_hex = combineResponse.signed_transaction.slice(2208*2*3 + 8*2*2, 2208*2*3 + 8*2*3);
        const signature_hex = combineResponse.signed_transaction.slice(2208*2*3 + 8*2*3, 2208*2*3 + 8*2*3 + 2144*2);

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