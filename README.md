Da ricordarsi

- deve essere specfificato allo startup the il middleware funziona in online o offline mode come scritto
https://docs.cdp.coinbase.com/mesh/docs/docker-deployment

transazione normale
1
src: tag1
dst: tag2
change: tag1
-> tag1-, tag2+

2
src: tag1
dst: wots1
change: tag1
-> tag1-, wots1+

3
src: wots1
dst: tag1
change: wots2
-> wots1-, tag1+, wots2+

4
src: wots1
dst: wots2
change: wots3
-> wots1-, wots2+, wots3+

5
src: wots1
dst: wots2
change: wots3
-> wots1-, wots2+, wots3+

Endpoints

/network/list
/network/status
/network/options

/block
/block/transaction

/mempool
/mempool/transaction

/account/balance
/account/coins

/construction/derive
/construction/preprocess
/construction/metadata
/construction/payloads
/construction/combine
/construction/parse
/construction/hash
/construction/submit

/call

/events/blocks

/search/transactions