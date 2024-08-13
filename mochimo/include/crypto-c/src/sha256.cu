/**
 * @private
 * @headerfile sha256.cuh <sha256.cuh>
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_SHA256_CU
#define CRYPTO_SHA256_CU


#include "sha256.cuh"
#include <string.h>

/**
 * SHA256 transformation rounds.
 * @param ctx Pointer to SHA256 context
 * @param data Pointer to input to be transformed
*/
__device__ void cu_sha256_transform(SHA256_CTX *ctx, const uint8_t data[])
{
   /* SHA256 transformation constant */
   __constant__ static uint32_t k[64] = {
      0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5,
      0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
      0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
      0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
      0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc,
      0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
      0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7,
      0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
      0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
      0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
      0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3,
      0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
      0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5,
      0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
      0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
      0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
   };

   sha256_tranform_unrolled(ctx->state, ((uint32_t *) data), k);
}  /* end cu_sha256_transform() */

/**
 * Initialize a SHA256 context.
 * @param ctx Pointer to SHA256 context
*/
__device__ void cu_sha256_init(SHA256_CTX *ctx)
{
   ctx->datalen = 0;
   ctx->bitlen[0] = ctx->bitlen[1] = 0;
   ctx->state[0] = 0x6a09e667;
   ctx->state[1] = 0xbb67ae85;
   ctx->state[2] = 0x3c6ef372;
   ctx->state[3] = 0xa54ff53a;
   ctx->state[4] = 0x510e527f;
   ctx->state[5] = 0x9b05688c;
   ctx->state[6] = 0x1f83d9ab;
   ctx->state[7] = 0x5be0cd19;
}  /* end cu_sha256_init() */

/**
 * Add @a inlen bytes from @a in to a SHA256 context for hashing.
 * @param ctx Pointer to SHA256 context
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
*/
__device__ void cu_sha256_update(SHA256_CTX *ctx, const void *in,
   size_t inlen)
{
   size_t i, n;
   uint32_t old;

   for(i = n = 0; inlen; i += n, inlen -= n) {
      /* copy memory to input buffer in chunks */
      n = (ctx->datalen + inlen) > 64 ? 64 - ctx->datalen : inlen;
      memcpy(ctx->data + ctx->datalen, (const uint8_t *) in + i, n);
      ctx->datalen += n;
      /* process input buffer */
      if (ctx->datalen == 64) {
         cu_sha256_transform(ctx, ctx->data);
         ctx->datalen = 0;
         old = ctx->bitlen[0];
         ctx->bitlen[0] += 512;
         if (ctx->bitlen[0] < old) ctx->bitlen[1]++;  /* add in carry */
      }
   }
}  /* end cu_sha256_update() */

/**
 * Finalize a SHA256 message digest.
 * Generate the SHA256 message digest and place in @a out.
 * @param ctx Pointer to SHA256 context
 * @param out Pointer to location to place the message digest
*/
__device__ void cu_sha256_final(SHA256_CTX *ctx, void *out)
{
   uint32_t i, old;

   i = ctx->datalen;

   /* Pad whatever data is left in the buffer. */
   if (ctx->datalen < 56) {
      ctx->data[i++] = 0x80;
      memset(ctx->data + i, 0, 64 - i);
   } else if (ctx->datalen >= 56) {
      ctx->data[i++] = 0x80;
      if (i < 64) memset(ctx->data + i, 0, 64 - i);
      cu_sha256_transform(ctx, ctx->data);
      memset(ctx->data, 0, 56);
   }

   /* Append to the padding the total message's length in bits and
    * transform (big endian). */
   old = ctx->bitlen[0];
   ctx->bitlen[0] += ctx->datalen << 3;
   if(ctx->bitlen[0] < old) ctx->bitlen[1]++;  /* add in carry */
   /* immitate bswap64() for bitlen */
   ((uint32_t *) ctx->data)[15] = bswap32(ctx->bitlen[0]);
   ((uint32_t *) ctx->data)[14] = bswap32(ctx->bitlen[1]);
   cu_sha256_transform(ctx, ctx->data);

   /* Since this implementation uses little endian byte ordering and
    * SHA uses big endian, reverse all the bytes when copying the
    * final state to the output hash. */
   ((uint32_t *) out)[0] = bswap32(ctx->state[0]);
   ((uint32_t *) out)[1] = bswap32(ctx->state[1]);
   ((uint32_t *) out)[2] = bswap32(ctx->state[2]);
   ((uint32_t *) out)[3] = bswap32(ctx->state[3]);
   ((uint32_t *) out)[4] = bswap32(ctx->state[4]);
   ((uint32_t *) out)[5] = bswap32(ctx->state[5]);
   ((uint32_t *) out)[6] = bswap32(ctx->state[6]);
   ((uint32_t *) out)[7] = bswap32(ctx->state[7]);
}  /* end cu_sha256_final() */

/**
 * Convenient all-in-one SHA256 computation.
 * Performs cu_sha256_init(), cu_sha256_update() and cu_sha256_final(),
 * and places the resulting hash in @a out.
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
 * @param out Pointer to location to place the message digest
*/
__device__ void cu_sha256(const void *in, size_t inlen, void *out)
{
   SHA256_CTX ctx;

   cu_sha256_init(&ctx);
   cu_sha256_update(&ctx, in, inlen);
   cu_sha256_final(&ctx, out);
}  /* end cu_sha256() */

/* CUDA kernel function */
__global__ static void kcu_sha256(const void *d_in, size_t *d_inlen,
   size_t max_inlen, void *d_out, int num)
{
   int tid = blockIdx.x * blockDim.x + threadIdx.x;
   if (tid >= num) return;

   uint8_t *in = ((uint8_t *) d_in) + (tid * max_inlen);
   uint8_t *out = ((uint8_t *) d_out) + (tid * SHA256LEN);

   cu_sha256(in, d_inlen[tid], out);
}  /* end kcu_sha256() */

/* CUDA kernel testing function */
void test_kcu_sha256(const void *in, size_t *inlen, size_t max_inlen,
   void *out, int num)
{
   uint8_t *d_in, *d_out;
   size_t *d_inlen;

   cudaMalloc(&d_in, num * max_inlen);
   cudaMalloc(&d_inlen, num * sizeof(size_t));
   cudaMalloc(&d_out, num * SHA256LEN);

   cudaMemcpy(d_in, in, num * max_inlen, cudaMemcpyHostToDevice);
   cudaMemcpy(d_inlen, inlen, num * sizeof(size_t), cudaMemcpyHostToDevice);
   cudaMemset(d_out, 0, num * SHA256LEN);

   kcu_sha256<<<1, num>>>(d_in, d_inlen, max_inlen, d_out, num);

   cudaMemcpy(out, d_out, num * SHA256LEN, cudaMemcpyDeviceToHost);
}  /* end test_kcu_sha256() */

/* end include guard */
#endif
