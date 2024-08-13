/**
 * @file sha256.h
 * @brief SHA256 hash function support.
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
#ifndef CRYPTO_SHA256_H
#define CRYPTO_SHA256_H


#include "utildev.h"

#define SHA256LEN 32 /**< SHA256 message digest length, in bytes */

/* SHA256 transform routines */
#define EP0(x) xor3(ror32(x, 2), ror32(x, 13), ror32(x, 22))
#define EP1(x) xor3(ror32(x, 6), ror32(x, 11), ror32(x, 25))
#define S0(x)  xor3(ror32(x, 7), ror32(x, 18), ((x) >> 3))
#define S1(x)  xor3(ror32(x, 17), ror32(x, 19), ((x) >> 10))

#define sha256_blk0(i, in) ( W[i] = bswap32(in[i]) )
#define sha256_blk(i)      ( W[i & 15] += \
   S1(W[(i - 2) & 15]) + W[(i - 7) & 15] + S0(W[(i - 15) & 15]) )

#define Ch(x, y, z)  xandx(x, y, z)
#define Maj(x, y, z) ( (x & y) | (z & (x | y)) )

/* initializer type round */
#define R0(a, b, c, d, e, f, g, h, i, k, in)             \
   h += EP1(e) + Ch(e,f,g) + k[i] + sha256_blk0(i, in);  \
   d += h; h += EP0(a) + Maj(a, b, c)
/* normal type round */
#define R(a, b, c, d, e, f, g, h, i, j, k)            \
   h += EP1(e) + Ch(e,f,g) + k[i+j] + sha256_blk(i);  \
   d += h; h += EP0(a) + Maj(a, b, c)
/* round expansion for initializer type rounds */
#define RX0_8(i, k, in)                         \
  R0(a, b, c, d, e, f, g, h, i, k, in);         \
  R0(h, a, b, c, d, e, f, g, (i + 1), k, in);   \
  R0(g, h, a, b, c, d, e, f, (i + 2), k, in);   \
  R0(f, g, h, a, b, c, d, e, (i + 3), k, in);   \
  R0(e, f, g, h, a, b, c, d, (i + 4), k, in);   \
  R0(d, e, f, g, h, a, b, c, (i + 5), k, in);   \
  R0(c, d, e, f, g, h, a, b, (i + 6), k, in);   \
  R0(b, c, d, e, f, g, h, a, (i + 7), k, in)
/* round expansion for normal type rounds */
#define RX_8(i, j, k)                        \
  R(a, b, c, d, e, f, g, h, i, j, k);        \
  R(h, a, b, c, d, e, f, g, (i + 1), j, k);  \
  R(g, h, a, b, c, d, e, f, (i + 2), j, k);  \
  R(f, g, h, a, b, c, d, e, (i + 3), j, k);  \
  R(e, f, g, h, a, b, c, d, (i + 4), j, k);  \
  R(d, e, f, g, h, a, b, c, (i + 5), j, k);  \
  R(c, d, e, f, g, h, a, b, (i + 6), j, k);  \
  R(b, c, d, e, f, g, h, a, (i + 7), j, k)

/* Unrolled SHA256 transformation */
#define sha256_tranform_unrolled(st, in, k)        \
{                                                  \
   uint32_t W[16], a, b, c, d, e, f, g, h;         \
   a = st[0]; b = st[1]; c = st[2]; d = st[3];     \
   e = st[4]; f = st[5]; g = st[6]; h = st[7];     \
   /* initial 16 rounds */                         \
   RX0_8(0, k, in); RX0_8(8, k, in);               \
   /* rounds 16 - 32 */                            \
   RX_8(0, 16, k); RX_8(8, 16, k);                 \
   /* rounds 32 - 48 */                            \
   RX_8(0, 32, k); RX_8(8, 32, k);                 \
   /* rounds 48 - 64 */                            \
   RX_8(0, 48, k); RX_8(8, 48, k);                 \
   st[0] += a; st[1] += b; st[2] += c; st[3] += d; \
   st[4] += e; st[5] += f; st[6] += g; st[7] += h; \
}

typedef struct {
   uint8_t data[64];    /**< Input buffer */
   uint32_t bitlen[2];  /**< Total bit length of updated input */
   uint32_t state[8];   /**< Internal hashing state */
   uint32_t datalen;    /**< Length of buffered input */
} SHA256_CTX;  /**< SHA256 Context */

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

void sha256_init(SHA256_CTX *ctx);
void sha256_update(SHA256_CTX *ctx, const void *in, size_t inlen);
void sha256_final(SHA256_CTX *ctx, void *out);
void sha256(const void *in, size_t inlen, void *out);

/* CUDA testing functions */
#ifdef CUDA
   void test_kcu_sha256(const void *in, size_t *inlen, size_t max_inlen,
      void *out, int num);
#endif

/* end extern "C" {} for C++ */
#ifdef __cplusplus
}
#endif

/* end include guard */
#endif
