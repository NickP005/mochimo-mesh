/**
 * @file sha1.cuh
 * @brief SHA1 CUDA hash function support.
 * @details This file is based on Brad Conte's basic
 * implementations of cryptography algorithms...
 * > <https://github.com/B-Con/crypto-algorithms>
 * ... and alterations to the transform function were based on
 * Steve Reid's SHA1 implementation...
 * > <https://github.com/clibs/sha1>
 * ... both of which were released into the Public Domain.
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_SHA1_CUH
#define CRYPTO_SHA1_CUH


#include "sha1.h"

__device__ void cu_sha1_transform(SHA1_CTX *ctx, const uint32_t data[]);
__device__ void cu_sha1_init(SHA1_CTX *ctx);
__device__ void cu_sha1_update(SHA1_CTX *ctx, const void *in, size_t inlen);
__device__ void cu_sha1_final(SHA1_CTX *ctx, void *out);
__device__ void cu_sha1(const void *in, size_t inlen, void *out);

/* end include guard */
#endif
