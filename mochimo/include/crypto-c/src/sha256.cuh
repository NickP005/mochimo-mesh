/**
 * @file sha256.cuh
 * @brief SHA256 CUDA hash function support.
 * @details This file is based on Brad Conte's basic
 * implementations of cryptography algorithms...
 * > <https://github.com/B-Con/crypto-algorithms>
 * ... and alterations to the transform function were based on
 * Igor Pavlov's SHA256 implementation...
 * > <https://github.com/jb55/sha256.c>
 * ... both of which were released into the Public Domain.
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_SHA256_CUH
#define CRYPTO_SHA256_CUH


#include "sha256.h"

__device__ void cu_sha256_transform(SHA256_CTX *ctx, const uint8_t data[]);
__device__ void cu_sha256_init(SHA256_CTX *ctx);
__device__ void cu_sha256_update(SHA256_CTX *ctx, const void *in,
   size_t inlen);
__device__ void cu_sha256_final(SHA256_CTX *ctx, void *out);
__device__ void cu_sha256(const void *in, size_t inlen, void *out);

/* end include guard */
#endif
