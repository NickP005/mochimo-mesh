/**
 * @private
 * @headerfile sha3.cuh <sha3.cuh>
 * @copyright This file is released under The MIT License (MIT).
 * For more information, see the header file...
*/

/* include guard */
#ifndef CRYPTO_SHA3_CU
#define CRYPTO_SHA3_CU


#include "sha3.cuh"
#include <string.h>

/**
 * SHA3_Keccak permutation rounds.
 * @param st Pointer to context state array
*/
__device__ void cu_sha3_keccakf(uint64_t st[])
{
	/* Keccakf round constant */
	__constant__ static uint64_t keccakf_rndc[24] = {
		0x0000000000000001, 0x0000000000008082, 0x800000000000808a,
		0x8000000080008000, 0x000000000000808b, 0x0000000080000001,
		0x8000000080008081, 0x8000000000008009, 0x000000000000008a,
		0x0000000000000088, 0x0000000080008009, 0x000000008000000a,
		0x000000008000808b, 0x800000000000008b, 0x8000000000008089,
		0x8000000000008003, 0x8000000000008002, 0x8000000000000080,
		0x000000000000800a, 0x800000008000000a, 0x8000000080008081,
		0x8000000000008080, 0x0000000080000001, 0x8000000080008008
	};

	sha3_keccakf_unrolled(st, keccakf_rndc);
}  /* end cu_sha3_keccakf() */

/**
 * Initialize a SHA3 context.
 * @param ctx Pointer to SHA3 context
 * @param outlen Length* of desired message digest, in bytes<br/>
 * <sup>_*compatible message digest lengths are 28, 32, 48 and 64_</sup>
*/
__device__ void cu_sha3_init(SHA3_KECCAK_CTX *ctx, int outlen)
{
   memset(ctx->st.b, 0, 200);
   ctx->outlen = (uint32_t) outlen;
   ctx->rsiz = 200 - (ctx->outlen << 1);
   ctx->pt = 0;
}

/**
 * Initialize a Keccak context.
 * @param ctx Pointer to Keccak context
 * @param outlen Length* of desired message digest, in bytes<br/>
 * <sup>_*compatible message digest lengths are 28, 32, 48 and 64_</sup>
*/
__device__ void cu_keccak_init(SHA3_KECCAK_CTX *ctx, int outlen)
{
   cu_sha3_init(ctx,outlen);
}

/**
 * Add @a inlen bytes from @a in to a SHA3 context for hashing.
 * @param ctx Pointer to SHA3 context
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
*/
__device__ void cu_sha3_update(SHA3_KECCAK_CTX *ctx, const void *in,
	size_t inlen)
{
   size_t i;

   for (i = 0; i < inlen; i++) {
      ctx->st.b[ctx->pt++] ^= ((const uint8_t *) in)[i];
      if (ctx->pt == ctx->rsiz) {
         cu_sha3_keccakf(ctx->st.q);
         ctx->pt = 0;
      }
   }
}

/**
 * Add @a inlen bytes from @a in to a Keccak context for hashing.
 * @param ctx Pointer to Keccak context
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
*/
__device__ void cu_keccak_update(SHA3_KECCAK_CTX *ctx, const void *in,
   size_t inlen)
{
   cu_sha3_update(ctx, in, inlen);
}

/**
 * Finalize a SHA3 message digest.
 * Generate the SHA3 message digest and place in @a out.
 * @param ctx Pointer to SHA3 context
 * @param out Pointer to location to place the message digest
*/
__device__ void cu_sha3_final(SHA3_KECCAK_CTX *ctx, void *out)
{
   ctx->st.b[ctx->pt] ^= 0x06;
   ctx->st.b[ctx->rsiz - 1] ^= 0x80;
   cu_sha3_keccakf(ctx->st.q);

   /* copy digest to out */
   memcpy(out, ctx->st.q, ctx->outlen);
}

/**
 * Finalize a Keccak message digest.
 * Generate the Keccak message digest and place in @a out.
 * @param ctx Pointer to Keccak context
 * @param out Pointer to location to place the message digest
*/
__device__ void cu_keccak_final(SHA3_KECCAK_CTX *ctx, void *out)
{
   /* This next step essentially converts the cu_sha3_final() step
    * `c->st.b[c->pt] ^= 0x06;`  (to)  `c->st.b[c->pt] ^= 0x01;`
    * as per the original Keccak implementation. */
   ctx->st.b[ctx->pt] ^= 0x07;
   cu_sha3_final(ctx, out);
}

/**
 * Convenient all-in-one SHA3 computation.
 * Performs sha3_init(), sha3_update() and sha3_final(),
 * and places the resulting hash in @a out.
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
 * @param out Pointer to location to place the message digest
 * @param outlen Length* of desired message digest, in bytes<br/>
 * <sup>_*compatible message digest lengths are 28, 32, 48 and 64_</sup>
*/
__device__ void cu_sha3(const void *in, size_t inlen, void *out,
   int outlen)
{
   SHA3_KECCAK_CTX ctx;

   cu_sha3_init(&ctx, outlen);
   cu_sha3_update(&ctx, in, inlen);
   cu_sha3_final(&ctx, out);
}

/**
 * Convenient all-in-one Keccak computation.
 * Performs keccak_init(), keccak_update() and keccak_final(),
 * and places the resulting hash in @a out.
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
 * @param out Pointer to location to place the message digest
 * @param outlen Length* of desired message digest, in bytes<br/>
 * <sup>_*compatible message digest lengths are 28, 32, 48 and 64_</sup>
*/
__device__ void cu_keccak(const void *in, size_t inlen, void *out,
   int outlen)
{
   SHA3_KECCAK_CTX ctx;

   cu_keccak_init(&ctx, outlen);
   cu_keccak_update(&ctx, in, inlen);
   cu_keccak_final(&ctx, out);
}

/* CUDA kernel function */
__global__ static void kcu_sha3(const void *d_in, size_t *d_inlen,
   size_t max_inlen, void *d_out, int outlen, int num)
{
   int tid = blockIdx.x * blockDim.x + threadIdx.x;
   if (tid >= num) return;

   uint8_t *in = ((uint8_t *) d_in) + (tid * max_inlen);
   uint8_t *out = ((uint8_t *) d_out) + (tid * outlen);

   cu_sha3(in, d_inlen[tid], out, outlen);
}  /* end kcu_sha3() */

/* CUDA kernel function */
__global__ static void kcu_keccak(const void *d_in, size_t *d_inlen,
   size_t max_inlen, void *d_out, int outlen, int num)
{
   int tid = blockIdx.x * blockDim.x + threadIdx.x;
   if (tid >= num) return;

   uint8_t *in = ((uint8_t *) d_in) + (tid * max_inlen);
   uint8_t *out = ((uint8_t *) d_out) + (tid * outlen);

   cu_keccak(in, d_inlen[tid], out, outlen);
}  /* end kcu_keccak() */

/* CUDA kernel testing function */
void test_kcu_sha3(const void *in, size_t *inlen, size_t max_inlen,
   void *out, int outlen, int num)
{
   uint8_t *d_in, *d_out;
   size_t *d_inlen;

   cudaMalloc(&d_in, num * max_inlen);
   cudaMalloc(&d_inlen, num * sizeof(size_t));
   cudaMalloc(&d_out, num * outlen);

   cudaMemcpy(d_in, in, num * max_inlen, cudaMemcpyHostToDevice);
   cudaMemcpy(d_inlen, inlen, num * sizeof(size_t), cudaMemcpyHostToDevice);
   cudaMemset(d_out, 0, num * outlen);

   kcu_sha3<<<1, num>>>(d_in, d_inlen, max_inlen, d_out, outlen, num);

   cudaMemcpy(out, d_out, num * outlen, cudaMemcpyDeviceToHost);
}  /* end test_kcu_sha3() */

/* CUDA kernel testing function */
void test_kcu_keccak(const void *in, size_t *inlen, size_t max_inlen,
   void *out, int outlen, int num)
{
   uint8_t *d_in, *d_out;
   size_t *d_inlen;

   cudaMalloc(&d_in, num * max_inlen);
   cudaMalloc(&d_inlen, num * sizeof(size_t));
   cudaMalloc(&d_out, num * outlen);

   cudaMemcpy(d_in, in, num * max_inlen, cudaMemcpyHostToDevice);
   cudaMemcpy(d_inlen, inlen, num * sizeof(size_t), cudaMemcpyHostToDevice);
   cudaMemset(d_out, 0, num * outlen);

   kcu_keccak<<<1, num>>>(d_in, d_inlen, max_inlen, d_out, outlen, num);

   cudaMemcpy(out, d_out, num * outlen, cudaMemcpyDeviceToHost);
}  /* end test_kcu_keccak() */

/* end include guard */
#endif
