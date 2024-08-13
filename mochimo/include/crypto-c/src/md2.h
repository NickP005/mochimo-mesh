/**
 * @file md2.h
 * @brief MD2 hash function support.
 * @details This file is based on Brad Conte's basic
 * implementations of cryptography algorithms...
 * > <https://github.com/B-Con/crypto-algorithms>
 * ... which was released into the Public Domain.
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

#ifndef CRYPTO_MD2_H
#define CRYPTO_MD2_H  /* include guard */


#include "utildev.h"

#define MD2LEN 16   /**< MD2 message digest length, in bytes */

/* MD2 transformation initialization */
#define md2_transform_init64(st, in) \
{ st[4] = (st[2] = in[0]) ^ st[0]; st[5] = (st[3] = in[1]) ^ st[1]; }

/* MD2 state transformation */
#define md2_transform_state(st, s)                          \
{                                                           \
   int _j, _k;                                              \
   st[0] ^= s[0];                                           \
   for (_k = 1; _k < 48; _k++) st[_k] ^= s[st[_k - 1]];     \
	for (_j = 1; _j < 18; _j++) {                            \
      st[0] ^= s[(st[47] + (_j - 1)) & 0xFF];               \
      for (_k = 1; _k < 48; _k++) st[_k] ^= s[st[_k - 1]];  \
	}                                                        \
}

/* MD2 checksum transformation */
#define md2_transform_checksum(c, in, s)  \
{                                         \
   int _j;                                \
   c[0] ^= s[in[0] ^ c[15]];              \
	for (_j = 1; _j < 16; ++_j) {          \
      c[_j] ^= s[in[_j] ^ c[_j - 1]];     \
   }                                      \
}

typedef struct {
   uint8_t state[48];      /**< Internal hashing state */
   uint8_t checksum[16];   /**< Internal hashing checksum */
   uint8_t data[16];       /**< Input buffer */
   uint32_t datalen;       /**< Length of buffered input */
} MD2_CTX;  /**< MD2 context */

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

/* Function prototypes for md2.c */
void md2_init(MD2_CTX *ctx);
void md2_update(MD2_CTX *ctx, const void *in, size_t inlen);
void md2_final(MD2_CTX *ctx, void *out);
void md2(const void *in, size_t inlen, void *out);

/* CUDA testing functions */
#ifdef CUDA
   void test_kcu_md2(const void *in, size_t *inlen, size_t max_inlen,
      void *out, int num);
#endif

/* end extern "C" {} for C++ */
#ifdef __cplusplus
}
#endif

/* end include guard */
#endif
