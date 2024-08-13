/**
 * @file md2.cuh
 * @brief MD2 CUDA hash function support.
 * @details This file is based on Brad Conte's basic
 * implementations of cryptography algorithms...
 * > <https://github.com/B-Con/crypto-algorithms>
 * ... which was released into the Public Domain.
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

#ifndef CRYPTO_MD2_CUH
#define CRYPTO_MD2_CUH  /* include guard */


#include "md2.h"

__device__ void cu_md2_transform(MD2_CTX *ctx, uint8_t data[]);
__device__ void cu_md2_init(MD2_CTX *ctx);
__device__ void cu_md2_update(MD2_CTX *ctx, const void *in, size_t inlen);
__device__ void cu_md2_final(MD2_CTX *ctx, void *out);
__device__ void cu_md2(const void *in, size_t inlen, void *out);

/* end include guard */
#endif
