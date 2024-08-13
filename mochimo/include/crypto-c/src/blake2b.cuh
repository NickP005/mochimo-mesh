/**
 * @file blake2b.cuh
 * @brief Blake2b CUDA hash function support.
 * @details This file is based on Dr. Markku-Juhani O. Saarinen's
 * "somewhat smaller" BLAKE2 implemetation...
 * > <https://github.com/mjosaarinen/blake2_mjosref><br/>
 * ... which was released into the Public Domain under the
 * Creative Commons Zero v1.0 Universal license.
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_BLAKE2B_CUH
#define CRYPTO_BLAKE2B_CUH


#include "blake2b.h"

__device__ void cu_blake2b_compress(BLAKE2B_CTX *ctx, int last);
__device__ int cu_blake2b_init(BLAKE2B_CTX *ctx, const void *key,
   int keylen, int outlen);
__device__ void cu_blake2b_update(BLAKE2B_CTX *ctx, const void *in,
   size_t inlen);
__device__ void cu_blake2b_final(BLAKE2B_CTX *ctx, void *out);
__device__ int cu_blake2b(const void *in, size_t inlen,
   const void *key, int keylen, void *out, int outlen);

/* end include guard */
#endif
