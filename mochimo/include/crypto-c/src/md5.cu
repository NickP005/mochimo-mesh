/**
 * @private
 * @headerfile md5.cuh <md5.cuh>
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_MD5_CU
#define CRYPTO_MD5_CU


#include "md5.cuh"
#include <string.h>  /* for memory handling */

/**
 * MD5 transformation rounds.
 * @param ctx Pointer to MD5 context
 * @param data Pointer to input to be transformed
*/
__device__ void cu_md5_transform(MD5_CTX *ctx, const uint8_t data[])
{
   md5_tranform_unrolled(ctx->state, ((uint32_t *) data));
}  /* end cu_md5_transform() */

/**
 * Initialize a MD5 context.
 * @param ctx Pointer to MD5 context
*/
__device__ void cu_md5_init(MD5_CTX *ctx)
{
   ctx->datalen = 0;
   ctx->bitlen[0] = ctx->bitlen[1] = 0;
   ctx->state[0] = 0x67452301;
   ctx->state[1] = 0xEFCDAB89;
   ctx->state[2] = 0x98BADCFE;
   ctx->state[3] = 0x10325476;
}  /* end cu_md5_init() */

/**
 * Add @a inlen bytes from @a in to a MD5 context for hashing.
 * @param ctx Pointer to MD5 context
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
*/
__device__ void cu_md5_update(MD5_CTX *ctx, const void *in, size_t inlen)
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
         cu_md5_transform(ctx, ctx->data);
         ctx->datalen = 0;
         old = ctx->bitlen[0];
         ctx->bitlen[0] += 512;
         if (ctx->bitlen[0] < old) ctx->bitlen[1]++;  /* add in carry */
      }
   }
}  /* end cu_md5_update() */

/**
 * Finalize a MD5 message digest.
 * Generate the MD5 message digest and place in @a out.
 * @param ctx Pointer to MD5 context
 * @param out Pointer to location to place the message digest
*/
__device__ void cu_md5_final(MD5_CTX *ctx, void *out)
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
      cu_md5_transform(ctx, ctx->data);
      memset(ctx->data, 0, 56);
   }

   /* Append to the padding the total message's length in bits */
   old = ctx->bitlen[0];
   ctx->bitlen[0] += ctx->datalen << 3;
   if (ctx->bitlen[0] < old) ctx->bitlen[1]++;  /* add in carry */
   ((uint32_t *) ctx->data)[14] = ctx->bitlen[0];
   ((uint32_t *) ctx->data)[15] = ctx->bitlen[1];

   /* perform final transform */
   cu_md5_transform(ctx, ctx->data);

   /* copy digest to out */
   memcpy(out, ctx->state, MD5LEN);
}  /* end cu_md5_final() */

/**
 * Convenient all-in-one MD5 computation.
 * Performs cu_md5_init(), cu_md5_update() and cu_md5_final(),
 * and places the resulting hash in @a out.
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
 * @param out Pointer to location to place the message digest
*/
__device__ void cu_md5(const void *in, size_t inlen, void *out)
{
   MD5_CTX ctx;

   cu_md5_init(&ctx);
   cu_md5_update(&ctx, in, inlen);
   cu_md5_final(&ctx, out);
}  /* end cu_md5() */

/* CUDA kernel function */
__global__ static void kcu_md5(const void *d_in, size_t *d_inlen,
   size_t max_inlen, void *d_out, int num)
{
   int tid = blockIdx.x * blockDim.x + threadIdx.x;
   if (tid >= num) return;

   uint8_t *in = ((uint8_t *) d_in) + (tid * max_inlen);
   uint8_t *out = ((uint8_t *) d_out) + (tid * MD5LEN);

   cu_md5(in, d_inlen[tid], out);
}  /* end kcu_md5() */

/* CUDA kernel testing function */
void test_kcu_md5(const void *in, size_t *inlen, size_t max_inlen,
   void *out, int num)
{
   uint8_t *d_in, *d_out;
   size_t *d_inlen;

   cudaMalloc(&d_in, num * max_inlen);
   cudaMalloc(&d_inlen, num * sizeof(size_t));
   cudaMalloc(&d_out, num * MD5LEN);

   cudaMemcpy(d_in, in, num * max_inlen, cudaMemcpyHostToDevice);
   cudaMemcpy(d_inlen, inlen, num * sizeof(size_t), cudaMemcpyHostToDevice);
   cudaMemset(d_out, 0, num * MD5LEN);

   kcu_md5<<<1, num>>>(d_in, d_inlen, max_inlen, d_out, num);

   cudaMemcpy(out, d_out, num * MD5LEN, cudaMemcpyDeviceToHost);
}  /* end test_kcu_md5() */

/* end include guard */
#endif
