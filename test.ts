import { MochimoRosettaClient, Operation, PublicKey, WOTS, TransactionManager} from './wallet_manager';

async function testMochimoRosettaClient() {
    const client = new MochimoRosettaClient();
    const wots = new WOTS();
    
    try {
        /*
        // Test network status and options
        console.log('1. Testing Network Initialization...');
        const { status, options } = await client.initialize();
        console.log('Network Status:', status);
        console.log('Network Options:', options);

        // Test account balance
        console.log('\n2. Testing Account Balance...');
        const testAddress = '0x0135306fa8b44d0bf87aedfc';
        const balance = await client.getAccountBalance(testAddress);
        console.log('Balance for', testAddress, ':', balance);

        // Test block retrieval by hash
        console.log('\n3. Testing Block Retrieval...');
        if (status.current_block_identifier.hash) {
            const blockData = await client.getBlock({
                hash: status.current_block_identifier.hash
            });
            console.log('Current Block Data:', blockData);
        }

        // Test construction derive
        console.log('\n4. Testing Construction Derive...');
        const testSeed = "test_seed_123";
        const testTag = "000000000000000000000000";
        const publicKey = wots.generateKeyPairFrom(testSeed, testTag);
        const deriveResponse = await client.constructionDerive(
            Buffer.from(publicKey).toString('hex')
        );
        console.log('Derive Response:', deriveResponse);

        // Test construction preprocess
        console.log('\n5. Testing Construction Preprocess...');
        const operations: Operation[] = [{
            operation_identifier: { index: 0 },
            type: 'TRANSACTION',
            status: 'SUCCESS',
            account: { address: testAddress },
            amount: {
                value: 1000000,
                currency: { symbol: 'MCM', decimals: 8 }
            }
        }];
        const preprocessResponse = await client.constructionPreprocess(operations);
        console.log('Preprocess Response:', preprocessResponse);

        // Test construction metadata
        console.log('\n6. Testing Construction Metadata...');
        const metadataResponse = await client.constructionMetadata(
            preprocessResponse.options
        );
        console.log('Metadata Response:', metadataResponse);

        // Test construction payloads
        console.log('\n7. Testing Construction Payloads...');
        const payloadsResponse = await client.constructionPayloads(
            operations,
            metadataResponse.metadata
        );
        console.log('Payloads Response:', payloadsResponse);

        // Test construction parse
        console.log('\n8. Testing Construction Parse...');
        const parseResponse = await client.constructionParse(
            payloadsResponse.unsigned_transaction,
            false
        );
        console.log('Parse Response:', parseResponse);

        // Test construction combine
        console.log('\n9. Testing Construction Combine...');
        const signature = wots.generateSignatureFrom(
            testSeed,
            Buffer.from(payloadsResponse.unsigned_transaction)
        );
        const signatures = [{
            hex_bytes: Buffer.from(signature).toString('hex'),
            public_key: {
                hex_bytes: Buffer.from(publicKey).toString('hex'),
                curve_type: 'wotsp'
            },
            signature_type: 'wotsp'
        }];
        const combineResponse = await client.constructionCombine(
            payloadsResponse.unsigned_transaction,
            signatures
        );
        console.log('Combine Response:', combineResponse);

        // Test construction hash
        console.log('\n10. Testing Construction Hash...');
        const hashResponse = await client.constructionHash(
            combineResponse.signed_transaction
        );
        console.log('Hash Response:', hashResponse);

        */

        // Optional: Test construction submit
        // Commented out to avoid actual transaction submission
        /*
        console.log('\n11. Testing Construction Submit...');
        const submitResponse = await client.constructionSubmit(
            combineResponse.signed_transaction
        );
        console.log('Submit Response:', submitResponse);
        */

        // get mempool
        console.log('\n11. Testing Mempool...');
        const mempool = await client.getMempool();
        console.log('Mempool:', mempool);

        // Test TransactionManager 
        console.log('\n12. Testing Transaction Manager...');
        // convert to hex   
        const txManager = new TransactionManager(
            client,
            "c54572cb24e810fc2285aa4f310ce07ad3158a34e7fe8fc63287130935189800",      // wots_seed
            "b3b8c474d47198ba3237a16397ca267a6b2324eee3d8541ff074b74fc0d21101",      // next_wots_seed
            "0e5989d23edfb582db3e730f",  // sender tag
            "985dfa821c48b8b1ff6802ca"   // receiver tag
        );
        const txResponse = await txManager.sendTransaction(
            1,  // amount
            500   // fee
        );
    } catch (error) {
        console.error('Test Error:', error);
        if (error instanceof Error) {
            console.error('Error Details:', error.message);
            console.error('Stack:', error.stack);
        }
    }
}

// Run all tests
testMochimoRosettaClient().catch(console.error);
