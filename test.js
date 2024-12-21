import { MochimoRosettaClient } from './wallet_manager';
async function test() {
    const client = new MochimoRosettaClient();
    try {
        console.log('Initializing Mochimo Rosetta client...');
        const { status, options } = await client.initialize();
        console.log('\nNetwork Status:');
        console.log('Current Block:', status.current_block_identifier);
        console.log('Current Timestamp:', new Date(status.current_block_timestamp).toISOString());
        console.log('\nNetwork Options:');
        console.log('Rosetta Version:', options.version.rosetta_version);
        console.log('Node Version:', options.version.node_version);
        console.log('Middleware Version:', options.version.middleware_version);
        // Get account balance
        console.log('\nFetching Account Balance:');
        const balance = await client.getAccountBalance('0x0135306fa8b44d0bf87aedfc');
        console.log('Account Balance:', JSON.stringify(balance, null, 2));
        // Or more specifically for the balance:
        if (balance.balances && balance.balances.length > 0) {
            console.log('Balance:', {
                value: balance.balances[0].value,
                currency: {
                    symbol: balance.balances[0].currency.symbol,
                    decimals: balance.balances[0].currency.decimals
                }
            });
            // Show human-readable balance
            const decimalValue = parseInt(balance.balances[0].value) / Math.pow(10, balance.balances[0].currency.decimals);
            console.log(`Human readable balance: ${decimalValue} ${balance.balances[0].currency.symbol}`);
        }
        // Get current block data
        /*
        console.log('\nFetching Current Block Data:');
        const blockData = await client.getBlock(status.current_block_identifier);
        console.log('Block Data:', JSON.stringify(blockData, null, 2));*/
    }
    catch (error) {
        console.error('Error:', error);
    }
}
test();
