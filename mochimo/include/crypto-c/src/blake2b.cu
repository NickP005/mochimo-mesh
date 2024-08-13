/**
 * @private
 * @headerfile blake2b.cuh <blake2b.cuh>
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_BLAKE2B_CU
#define CRYPTO_BLAKE2B_CU


#include "blake2b.cuh"
#include <string.h>  /* for memory handling */

/**
 * Blake2b compression rounds.
 * @param ctx Pointer to Blake2b context
 * @param last Flag indicating the final compression
*/
__device__ void cu_blake2b_compress(BLAKE2B_CTX *ctx, int last)
{
   /* Blake2b compression Sigma */
   __constant__ static uint8_t Sigma[12][16] = {
      { 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15 },
      { 14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3 },
      { 11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4 },
      { 7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8 },
      { 9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13 },
      { 2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9 },
      { 12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11 },
      { 13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10 },
      { 6, 15, 14, 9, 11, 3, 0, 8, 12, 2, 13, 7, 1, 4, 10, 5 },
      { 10, 2, 8, 4, 7, 6, 1, 5, 15, 11, 9, 14, 3, 12, 13, 0 },
      { 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15 },
      { 14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3 }
   };
   uint64_t v[16];

   blake2b_compress_init(v, ctx->h, ctx->t, last);
   blake2b_compress_rounds(v, ctx->in.q, Sigma);
   blake2b_compress_set(v, ctx->h);
}  /* end cu_blake2b_compress() */

/**
 * Add @a inlen bytes from @a in to a Blake2b context for hashing.
 * @param ctx Pointer to Blake2b context
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
*/
__device__ void cu_blake2b_update(BLAKE2B_CTX *ctx, const void *in,
   size_t inlen)
{
   size_t i, n;

   for(i = n = 0; inlen; i += n, inlen -= n) {
      if (ctx->c == 128) {
         ctx->t[0] += ctx->c;
         if (ctx->t[0] < ctx->c) ctx->t[1]++;
         cu_blake2b_compress(ctx, 0);
         ctx->c = 0;
      }
      /* copy memory in chunks */
      n = (ctx->c + inlen) > 128 ? 128 - ctx->c : inlen;
      memcpy(ctx->in.b + ctx->c, (const uint8_t *) in + i, n);
      ctx->c += n;
   }
}  /* end cu_blake2b_update() */

/**
 * Initialize a Blake2b context with optional @a key.
 * To hash without a key, specify `NULL` and `0` for @a key and
 * @a keylen, respectively.
 * @param ctx Pointer to Blake2b context
 * @param key Pointer to optional "key" input
 * @param keylen Length of optional @a key input, in bytes
 * @param outlen Byte length of desired digest
 * @returns 0 on success, else if initialization fails (-1).
 * @note Blake2b initialization can fail if @a keylen is greater
 * than 64 or outlen is not a supported digest length. Supported
 * lengths include: 32 (256-bits), 48 (384-bits) or 64 (512-bits).
*/
__device__ int cu_blake2b_init(BLAKE2B_CTX *ctx, const void *key,
   int keylen, int outlen)
{
   if (keylen > 64) return -1;
   if (outlen != 32 && outlen != 48 && outlen != 64) return -1;

   ctx->c = 0;
   ctx->t[0] = 0;
   ctx->t[1] = 0;
   ctx->outlen = (uint64_t) outlen;
   ctx->h[0] = 0x6A09E667F3BCC908ull/* Blake2b_iv[0] */ ^
      0x01010000 ^ (keylen << 8) ^ outlen;
   ctx->h[1] = 0xBB67AE8584CAA73Bull/* Blake2b_iv[1] */;
   ctx->h[2] = 0x3C6EF372FE94F82Bull/* Blake2b_iv[2] */;
   ctx->h[3] = 0xA54FF53A5F1D36F1ull/* Blake2b_iv[3] */;
   ctx->h[4] = 0x510E527FADE682D1ull/* Blake2b_iv[4] */;
   ctx->h[5] = 0x9B05688C2B3E6C1Full/* Blake2b_iv[5] */;
   ctx->h[6] = 0x1F83D9ABFB41BD6Bull/* Blake2b_iv[6] */;
   ctx->h[7] = 0x5BE0CD19137E2179ull/* Blake2b_iv[7] */;

   /* zero remaining input buffer */
   memset(&ctx->in.b[keylen], 0, 128 - keylen);

   if (keylen > 0) {
      cu_blake2b_update(ctx, key, keylen);
      ctx->c = 128;
   }

   return 0;
}  /* end cu_blake2b_init() */

/**
 * Finalize a Blake2b message digest.
 * Generate the Blake2b message digest and place in @a out.
 * @param ctx Pointer to Blake2b context
 * @param out Pointer to location to place the message digest
*/
__device__ void cu_blake2b_final(BLAKE2B_CTX *ctx, void *out)
{
   ctx->t[0] += ctx->c;
   if (ctx->t[0] < ctx->c) ctx->t[1]++;

   /* zero remainder of input buffer */
   if (ctx->c < 128) memset(&ctx->in.b[ctx->c], 0, 128 - ctx->c);

   /* final compression */
   cu_blake2b_compress(ctx, 1);

   /* copy digest to out */
   memcpy(out, ctx->h, ctx->outlen);
}  /* end cu_blake2b_final() */

/**
 * Convenient all-in-one Blake2b computation.
 * Performs cu_blake2b_init(), cu_blake2b_update() and cu_blake2b_final(),
 * and places the resulting hash in @a out.
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
 * @param key Pointer to optional "key" input
 * @param keylen Length of optional @a key input, in bytes
 * @param out Pointer to location to place the message digest
 * @param outlen Length* of desired message digest, in bytes<br/>
 * <sup>_*compatible message digest lengths are 32, 48 and 64_</sup>
*/
__device__ int cu_blake2b(const void *in, size_t inlen,
   const void *key, int keylen, void *out, int outlen)
{
   BLAKE2B_CTX ctx;

   if (cu_blake2b_init(&ctx, key, keylen, outlen)) {
      return -1;
   }
   cu_blake2b_update(&ctx, in, inlen);
   cu_blake2b_final(&ctx, out);

   return 0;
}  /* end cu_blake2b() */

/* CUDA kernel function */
__global__ static void kcu_blake2b(
   const void *d_in, size_t *d_inlen, size_t max_inlen,
   const void *d_key, int *d_keylen, int max_keylen,
   void *d_out, int outlen, int *d_ret, int num)
{
   int tid = blockIdx.x * blockDim.x + threadIdx.x;
   if (tid >= num) return;

   uint8_t *in = ((uint8_t *) d_in) + (tid * max_inlen);
   uint8_t *key = ((uint8_t *) d_key) + (tid * max_keylen);
   uint8_t *out = ((uint8_t *) d_out) + (tid * outlen);

   d_ret[tid] = cu_blake2b(in, d_inlen[tid], key, d_keylen[tid], out, outlen);
}  /* end kcu_blake2b() */

/* CUDA kernel testing function */
void test_kcu_blake2b(
   const void *in, size_t *inlen, size_t max_inlen,
   const void *key, int *keylen, int max_keylen,
   void *out, int outlen, int *ret, int num)
{
   uint8_t *d_in, *d_key, *d_out;
   size_t *d_inlen;
   int *d_keylen, *d_ret;

   cudaMalloc(&d_ret, num * sizeof(int));
   cudaMalloc(&d_in, num * max_inlen);
   cudaMalloc(&d_inlen, num * sizeof(size_t));
   cudaMalloc(&d_key, num * max_keylen);
   cudaMalloc(&d_keylen, num * sizeof(int));
   cudaMalloc(&d_out, num * outlen);

   cudaMemset(d_ret, 0, num * sizeof(int));
   cudaMemcpy(d_in, in, num * max_inlen, cudaMemcpyHostToDevice);
   cudaMemcpy(d_inlen, inlen, num * sizeof(size_t), cudaMemcpyHostToDevice);
   cudaMemcpy(d_key, key, num * max_keylen, cudaMemcpyHostToDevice);
   cudaMemcpy(d_keylen, keylen, num * sizeof(int), cudaMemcpyHostToDevice);
   cudaMemset(d_out, 0, num * outlen);

   kcu_blake2b<<<1, num>>>(
      d_in, d_inlen, max_inlen,
      d_key, d_keylen, max_keylen,
      d_out, outlen, d_ret, num);

   cudaMemcpy(ret, d_ret, num * sizeof(int), cudaMemcpyDeviceToHost);
   cudaMemcpy(out, d_out, num * outlen, cudaMemcpyDeviceToHost);
}  /* end test_kcu_blake2b() */

/* end include guard */
#endif
