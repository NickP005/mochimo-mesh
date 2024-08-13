/**
 * @file md5.cuh
 * @brief MD5 CUDA hash function support.
 * @details This file is based on Brad Conte's basic
 * implementations of cryptography algorithms...
 * > <https://github.com/B-Con/crypto-algorithms>
 * ... which was released into the Public Domain.
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_MD5_CUH
#define CRYPTO_MD5_CUH


#include "md5.h"

__device__ void cu_md5_transform(MD5_CTX *ctx, const uint32_t m[]);
__device__ void cu_md5_init(MD5_CTX *ctx);
__device__ void cu_md5_update(MD5_CTX *ctx, const void *in, size_t inlen);
__device__ void cu_md5_final(MD5_CTX *ctx, void *out);
__device__ void cu_md5(const void *in, size_t inlen, void *out);

/* end include guard */
#endif
