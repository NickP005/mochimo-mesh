import fetch from 'node-fetch';
import * as CryptoJS from 'crypto-js'

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



export class WOTS {
  private readonly PARAMSN = 32
  private readonly WOTSW = 16
  private readonly WOTSLOGW = 4
  private readonly WOTSLEN1: number
  private readonly WOTSLEN2 = 3
  private readonly WOTSLEN: number
  private readonly WOTSSIGBYTES: number
  private readonly TXSIGLEN = 2144
  private readonly TXADDRLEN = 2208
  private readonly XMSS_HASH_PADDING_F = 0
  private readonly XMSS_HASH_PADDING_PRF = 3

  constructor() {
    this.WOTSLEN1 = (8 * this.PARAMSN / this.WOTSLOGW)
    this.WOTSLEN = this.WOTSLEN1 + this.WOTSLEN2
    this.WOTSSIGBYTES = this.WOTSLEN * this.PARAMSN
    this.validate_params()
  }

  public generateKeyPairFrom(wots_seed: string, tag?: string): Uint8Array {
    // if (!wots_seed) {
    //   throw new Error('Seed is required')
    // }

    // Add tag validation
    if (tag !== undefined) {
      if (tag.length !== 24) {
        // Use default tag for invalid length
        tag = undefined
      } else {
        // Check if tag contains only valid hex characters (0-9, A-F)
        const validHex = /^[0-9A-F]{24}$/i
        if (!validHex.test(tag)) {
          throw new Error('Invalid tag format')
        }
      }
    }

    const private_seed = this.sha256(wots_seed + "seed")
    const public_seed = this.sha256(wots_seed + "publ")
    const addr_seed = this.sha256(wots_seed + "addr")
    
    let wots_public = this.public_key_gen(private_seed, public_seed, addr_seed)
    
    // Create a single array with all components
    const totalLength = wots_public.length + public_seed.length + 20 + 12
    const result = new Uint8Array(totalLength)
    
    let offset = 0
    result.set(wots_public, offset)
    offset += wots_public.length
    
    result.set(public_seed, offset)
    offset += public_seed.length
    
    result.set(addr_seed.slice(0, 20), offset)
    offset += 20
    
    // Add tag
    const tagBytes = !tag || tag.length !== 24 
      ? new Uint8Array([66, 0, 0, 0, 14, 0, 0, 0, 1, 0, 0, 0])
      : this.hexToBytes(tag)
    result.set(tagBytes, offset)
    
    return result
  }

  public generateSignatureFrom(wots_seed: string, payload: Uint8Array): Uint8Array {
    const private_seed = this.sha256(wots_seed + "seed")
    const public_seed = this.sha256(wots_seed + "publ")
    const addr_seed = this.sha256(wots_seed + "addr")
    const to_sign = this.sha256(payload)
    
    return this.wots_sign(to_sign, private_seed, public_seed, addr_seed)
  }

  public sha256(input: string | Uint8Array): Uint8Array {
    if (typeof input === 'string') {
      const hash = CryptoJS.SHA256(input)
      return new Uint8Array(this.hexToBytes(hash.toString()))
    } else {
      const hash = CryptoJS.SHA256(this.bytesToHex(input))
      return new Uint8Array(this.hexToBytes(hash.toString()))
    }
  }

  public hexToBytes(hex: string): number[] {
    const bytes: number[] = []
    for (let i = 0; i < hex.length; i += 2) {
      bytes.push(parseInt(hex.substr(i, 2), 16)) 
    }
    return bytes
  }

  public bytesToHex(bytes: Uint8Array): string {
    return Array.from(bytes)
      .map(b => b.toString(16).padStart(2, '0'))
      .join('')
  }

  /**
   * Generates WOTS public key from private key
   */
  private public_key_gen(seed: Uint8Array, pub_seed: Uint8Array, addr_bytes: Uint8Array): Uint8Array {
    const private_key = this.expand_seed(seed)
    const public_key = new Uint8Array(this.WOTSSIGBYTES)
    let addr = this.bytes_to_addr(addr_bytes)

    for (let i = 0; i < this.WOTSLEN; i++) {
      this.set_chain_addr(i, addr)
      const private_key_portion = private_key.slice(i * this.PARAMSN, (i + 1) * this.PARAMSN)
      const chain = this.gen_chain(
        private_key_portion,
        0,
        this.WOTSW - 1,
        pub_seed,
        addr
      )
      public_key.set(chain, i * this.PARAMSN)
    }

    return public_key
  }

  /**
   * Signs a message using WOTS
   */
  private wots_sign(msg: Uint8Array, seed: Uint8Array, pub_seed: Uint8Array, addr_bytes: Uint8Array): Uint8Array {
    const private_key = this.expand_seed(seed)
    const signature = new Uint8Array(this.WOTSSIGBYTES)
    const lengths = this.chain_lengths(msg)
    let addr = this.bytes_to_addr(addr_bytes)

    for (let i = 0; i < this.WOTSLEN; i++) {
      this.set_chain_addr(i, addr)
      const private_key_portion = private_key.slice(i * this.PARAMSN, (i + 1) * this.PARAMSN)
      const chain = this.gen_chain(
        private_key_portion,
        0,
        lengths[i],
        pub_seed,
        addr
      )
      signature.set(chain, i * this.PARAMSN)
    }

    return signature
  }

  /**
   * Verifies a WOTS signature
   */
  private wots_publickey_from_sig(
    sig: Uint8Array,
    msg: Uint8Array,
    pub_seed: Uint8Array,
    addr_bytes: Uint8Array
  ): Uint8Array {
    let addr = this.bytes_to_addr(addr_bytes)
    const lengths = this.chain_lengths(msg)
    const public_key = new Uint8Array(this.WOTSSIGBYTES)

    for (let i = 0; i < this.WOTSLEN; i++) {
      this.set_chain_addr(i, addr)
      const sig_portion = sig.slice(i * this.PARAMSN, (i + 1) * this.PARAMSN)
      const chain = this.gen_chain(
        sig_portion,
        lengths[i],
        this.WOTSW - 1 - lengths[i],
        pub_seed,
        addr
      )
      public_key.set(chain, i * this.PARAMSN)
    }

    return public_key
  }

  /**
   * Expands seed into private key
   */
  private expand_seed(seed: Uint8Array): Uint8Array {
    const private_key = new Uint8Array(this.WOTSSIGBYTES)

    for (let i = 0; i < this.WOTSLEN; i++) {
      const ctr = this.ull_to_bytes(this.PARAMSN, [i])
      const portion = this.prf(ctr, seed)
      private_key.set(portion, i * this.PARAMSN)
    }

    return private_key
  }

  /**
   * Generates hash chain
   */
  private gen_chain(
    input: Uint8Array,
    start: number,
    steps: number,
    pub_seed: Uint8Array,
    addr: Record<string, Uint8Array>
  ): Uint8Array {
    let out = new Uint8Array(input)
    
    for (let i = start; i < start + steps && i < this.WOTSW; i++) {
      this.set_hash_addr(i, addr)
      out = this.t_hash(out, pub_seed, addr)
    }

    return out
  }

  /**
   * Computes PRF using SHA-256
   */
  private prf(input: Uint8Array, key: Uint8Array): Uint8Array {
    const buf = new Uint8Array(32 * 3)
    
    // Add padding
    buf.set(this.ull_to_bytes(this.PARAMSN, [this.XMSS_HASH_PADDING_PRF]))
    
    // Add key and input
    const byte_copied_key = this.byte_copy(key, this.PARAMSN)
    buf.set(byte_copied_key, this.PARAMSN)
    
    const byte_copied_input = this.byte_copy(input, 32)
    buf.set(byte_copied_input, this.PARAMSN * 2)
    
    return this.sha256(buf)
  }

  /**
   * Computes t_hash for WOTS chain
   */
  private t_hash(input: Uint8Array, pub_seed: Uint8Array, addr: Record<string, Uint8Array>): Uint8Array {
    const buf = new Uint8Array(32 * 3)
    let addr_bytes: Uint8Array
    
    // Add padding
    buf.set(this.ull_to_bytes(this.PARAMSN, [this.XMSS_HASH_PADDING_F]))
    
    // Get key mask
    this.set_key_and_mask(0, addr)
    addr_bytes = this.addr_to_bytes(addr)
    buf.set(this.prf(addr_bytes, pub_seed), this.PARAMSN)
    
    // Get bitmask
    this.set_key_and_mask(1, addr)
    addr_bytes = this.addr_to_bytes(addr)
    const bitmask = this.prf(addr_bytes, pub_seed)
    
    // XOR input with bitmask
    const XOR_bitmask_input = new Uint8Array(input.length)
    for (let i = 0; i < this.PARAMSN; i++) {
      XOR_bitmask_input[i] = input[i] ^ bitmask[i]
    }
    buf.set(XOR_bitmask_input, this.PARAMSN * 2)
    
    return this.sha256(buf)
  }

  /**
   * Converts number array to bytes with specified length
   */
  private ull_to_bytes(outlen: number, input: number[]): Uint8Array {
    const out = new Uint8Array(outlen)
    for (let i = outlen - 1; i >= 0; i--) {
      out[i] = input[i] || 0
    }
    return out
  }

  /**
   * Copies bytes with specified length
   */
  private byte_copy(source: Uint8Array, num_bytes: number): Uint8Array {
    const output = new Uint8Array(num_bytes)
    for (let i = 0; i < num_bytes; i++) {
      output[i] = source[i] || 0
    }
    return output
  }

  /**
   * Converts address to bytes
   */
  private addr_to_bytes(addr: Record<string, Uint8Array>): Uint8Array {
    const out_bytes = new Uint8Array(32)
    for (let i = 0; i < 8; i++) {
      const chunk = addr[i.toString()] || new Uint8Array(4)
      out_bytes.set(chunk, i * 4)
    }
    return out_bytes
  }

  /**
   * Converts bytes to address
   */
  private bytes_to_addr(addr_bytes: Uint8Array): Record<string, Uint8Array> {
    const out_addr: Record<string, Uint8Array> = {}
    for (let i = 0; i < 8; i++) {
      out_addr[i.toString()] = this.ull_to_bytes(4, Array.from(addr_bytes.slice(i * 4, (i + 1) * 4)))
    }
    return out_addr
  }

  /**
   * Sets chain address
   */
  private set_chain_addr(chain_address: number, addr: Record<string, Uint8Array>): void {
    addr['5'] = new Uint8Array([0, 0, 0, chain_address])
  }

  /**
   * Sets hash address
   */
  private set_hash_addr(hash: number, addr: Record<string, Uint8Array>): void {
    addr['6'] = new Uint8Array([0, 0, 0, hash])
  }

  /**
   * Sets key and mask
   */
  private set_key_and_mask(key_and_mask: number, addr: Record<string, Uint8Array>): void {
    addr['7'] = new Uint8Array([0, 0, 0, key_and_mask])
  }

  /**
   * Calculates chain lengths from message
   */
  private chain_lengths(msg: Uint8Array): Uint8Array {
    const msg_base_w = this.base_w(this.WOTSLEN1, msg)
    const csum_base_w = this.wots_checksum(msg_base_w)
    
    // Combine message and checksum base-w values
    const lengths = new Uint8Array(this.WOTSLEN)
    lengths.set(msg_base_w)
    lengths.set(csum_base_w, this.WOTSLEN1)
    
    return lengths
  }

  /**
   * Converts bytes to base-w representation
   */
  private base_w(outlen: number, input: Uint8Array): Uint8Array {
    const output = new Uint8Array(outlen)
    let in_ = 0
    let total = 0
    let bits = 0

    for (let i = 0; i < outlen; i++) {
      if (bits === 0) {
        total = input[in_]
        in_++
        bits += 8
      }
      bits -= this.WOTSLOGW
      output[i] = (total >> bits) & (this.WOTSW - 1)
    }

    return output
  }

  /**
   * Computes WOTS checksum
   */
  private wots_checksum(msg_base_w: Uint8Array): Uint8Array {
    let csum = 0
    
    // Calculate checksum
    for (let i = 0; i < this.WOTSLEN1; i++) {
      csum += this.WOTSW - 1 - msg_base_w[i]
    }

    // Convert checksum to base_w
    csum = csum << (8 - ((this.WOTSLEN2 * this.WOTSLOGW) % 8))
    
    const csum_bytes = this.int_to_bytes(csum)
    const csum_base_w = this.base_w(
      this.WOTSLEN2, 
      this.byte_copy(csum_bytes, Math.floor((this.WOTSLEN2 * this.WOTSLOGW + 7) / 8))
    )
    
    return csum_base_w
  }

  /**
   * Converts integer to bytes
   */
  private int_to_bytes(value: number): Uint8Array {
    const bytes = new Uint8Array(8)
    for (let i = 7; i >= 0; i--) {
      bytes[i] = value & 0xff
      value = value >> 8
    }
    return bytes
  }


  /**
   * Validates input parameters
   */
  private validate_params(): void {
    if (this.PARAMSN !== 32) {
      throw new Error('PARAMSN must be 32')
    }
    if (this.WOTSW !== 16) {
      throw new Error('WOTSW must be 16')
    }
    if (this.WOTSLOGW !== 4) {
      throw new Error('WOTSLOGW must be 4')
    }
  }

  /**
   * Initializes WOTS instance
   */
  public init(): void {
    this.validate_params()
  }

  // Add array extension functionality
  private concatUint8Arrays(...arrays: Uint8Array[]): Uint8Array {
    const totalLength = arrays.reduce((acc, arr) => acc + arr.length, 0)
    const result = new Uint8Array(totalLength)
    let offset = 0
    
    for (const arr of arrays) {
      result.set(arr, offset)
      offset += arr.length
    }
    
    return result
  }

  /**
   * Verifies a signature
   */
  public verifySignature(
    signature: Uint8Array,
    message: Uint8Array,
    pubSeed: Uint8Array,
    addrSeed: Uint8Array
  ): Uint8Array {
    const messageHash = this.sha256(message)
    return this.wots_publickey_from_sig(
      signature,
      messageHash,
      pubSeed,
      addrSeed
    )
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

    constructor(baseUrl: string = 'http://0.0.0.0:8080') {
        this.baseUrl = baseUrl;
        this.networkIdentifier = {
            blockchain: 'mochimo',
            network: 'mainnet'
        };
    }

    private async post<T>(endpoint: string, data: any): Promise<T> {
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
        operations[0].amount.value = String(-(senderBalance));  // Convert to string
        operations[1].amount.value = String(amount);  // Convert to string
        operations[2].amount.value = String(senderBalance - amount - miner_fee);  // Convert to string

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
        const to_sign = new Uint8Array(this.wots.hexToBytes(payloadsResponse.unsigned_transaction));
        // hash the transaction
        const hash = this.wots.sha256(to_sign);
        const signatureBytes = this.wots.generateSignatureFrom(
            this.wots_seed,
            hash
            );

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

        console.log("Debug - Validating signature payload match:");
        console.log("Unsigned tx:", payloadsResponse.unsigned_transaction);
        console.log("Signing payload:", signature.signing_payload.hex_bytes);

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
        const submitResponse = await this.client.constructionSubmit(
            combineResponse.signed_transaction
        );

        this.status_string = 'Transaction submitted successfully';
        return submitResponse.transaction_identifier;
    }
}