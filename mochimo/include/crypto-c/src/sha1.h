/**
 * @file sha1.h
 * @brief SHA1 hash function support.
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
#ifndef CRYPTO_SHA1_H
#define CRYPTO_SHA1_H


#include "utildev.h"

#define SHA1LEN   20 /**< SHA1 message digest length, in bytes */

/* SHA1 transform routines */
#define sha1_blk0(i, in)  ( W[i] = bswap32(in[i]) )
#define sha1_blk_xor4(i) \
   xor4(W[(i + 13) & 15], W[(i + 8) & 15], W[(i + 2) & 15], W[i & 15])
#define sha1_blk(i) ( W[i & 15] = rol32(sha1_blk_xor4(i), 1) )
/* SHA1 round1 input */
#define sha1_r0(a, b, c, d, e, k, in, i) \
   e += xandx(b, c, d) + sha1_blk0(i, in) + k[0] + rol32(a, 5); \
   b = rol32(b, 30)
/* SHA1 round1 extended */
#define sha1_r1(a, b, c, d, e, k, i) \
   e += xandx(b, c, d) + sha1_blk(i) + k[0] + rol32(a, 5); \
   b = rol32(b, 30)
/* SHA1 rounds 2/3/4 */
#define sha1_r2(a, b, c, d, e, k, i) \
   e += xor3(b, c, d) + sha1_blk(i) + k[1] + rol32(a, 5); \
   b = rol32(b, 30)
#define sha1_r3(a, b, c, d, e, k, i) \
   e += (((b | c) & d) | (b & c)) + sha1_blk(i) + k[2] + rol32(a, 5); \
   b = rol32(b, 30)
#define sha1_r4(a, b, c, d, e, k, i) \
   e += xor3(b, c, d) + sha1_blk(i) + k[3] + rol32(a, 5); \
   b = rol32(b, 30)

/* Unrolled SHA1 transformation */
#define sha1_transform_unrolled(st, in, k)   \
{                                            \
   uint32_t W[16];                           \
	uint32_t a = st[0];                       \
   uint32_t b = st[1];                       \
   uint32_t c = st[2];                       \
   uint32_t d = st[3];                       \
   uint32_t e = st[4];                       \
	/* SHA1 round 1 */                        \
   sha1_r0(a, b, c, d, e, k, in, 0);         \
   sha1_r0(e, a, b, c, d, k, in, 1);         \
   sha1_r0(d, e, a, b, c, k, in, 2);         \
   sha1_r0(c, d, e, a, b, k, in, 3);         \
   sha1_r0(b, c, d, e, a, k, in, 4);         \
   sha1_r0(a, b, c, d, e, k, in, 5);         \
   sha1_r0(e, a, b, c, d, k, in, 6);         \
   sha1_r0(d, e, a, b, c, k, in, 7);         \
   sha1_r0(c, d, e, a, b, k, in, 8);         \
   sha1_r0(b, c, d, e, a, k, in, 9);         \
   sha1_r0(a, b, c, d, e, k, in, 10);        \
   sha1_r0(e, a, b, c, d, k, in, 11);        \
   sha1_r0(d, e, a, b, c, k, in, 12);        \
   sha1_r0(c, d, e, a, b, k, in, 13);        \
   sha1_r0(b, c, d, e, a, k, in, 14);        \
   sha1_r0(a, b, c, d, e, k, in, 15);        \
   /* alternate round computation */         \
   sha1_r1(e, a, b, c, d, k, 16);            \
   sha1_r1(d, e, a, b, c, k, 17);            \
   sha1_r1(c, d, e, a, b, k, 18);            \
   sha1_r1(b, c, d, e, a, k, 19);            \
   sha1_r2(a, b, c, d, e, k, 20);            \
	/* SHA1 round 2 */                        \
   sha1_r2(e, a, b, c, d, k, 21);            \
   sha1_r2(d, e, a, b, c, k, 22);            \
   sha1_r2(c, d, e, a, b, k, 23);            \
   sha1_r2(b, c, d, e, a, k, 24);            \
   sha1_r2(a, b, c, d, e, k, 25);            \
   sha1_r2(e, a, b, c, d, k, 26);            \
   sha1_r2(d, e, a, b, c, k, 27);            \
   sha1_r2(c, d, e, a, b, k, 28);            \
   sha1_r2(b, c, d, e, a, k, 29);            \
   sha1_r2(a, b, c, d, e, k, 30);            \
   sha1_r2(e, a, b, c, d, k, 31);            \
   sha1_r2(d, e, a, b, c, k, 32);            \
   sha1_r2(c, d, e, a, b, k, 33);            \
   sha1_r2(b, c, d, e, a, k, 34);            \
   sha1_r2(a, b, c, d, e, k, 35);            \
   sha1_r2(e, a, b, c, d, k, 36);            \
   sha1_r2(d, e, a, b, c, k, 37);            \
   sha1_r2(c, d, e, a, b, k, 38);            \
   sha1_r2(b, c, d, e, a, k, 39);            \
	/* SHA1 round 3 */                        \
   sha1_r3(a, b, c, d, e, k, 40);            \
   sha1_r3(e, a, b, c, d, k, 41);            \
   sha1_r3(d, e, a, b, c, k, 42);            \
   sha1_r3(c, d, e, a, b, k, 43);            \
   sha1_r3(b, c, d, e, a, k, 44);            \
   sha1_r3(a, b, c, d, e, k, 45);            \
   sha1_r3(e, a, b, c, d, k, 46);            \
   sha1_r3(d, e, a, b, c, k, 47);            \
   sha1_r3(c, d, e, a, b, k, 48);            \
   sha1_r3(b, c, d, e, a, k, 49);            \
   sha1_r3(a, b, c, d, e, k, 50);            \
   sha1_r3(e, a, b, c, d, k, 51);            \
   sha1_r3(d, e, a, b, c, k, 52);            \
   sha1_r3(c, d, e, a, b, k, 53);            \
   sha1_r3(b, c, d, e, a, k, 54);            \
   sha1_r3(a, b, c, d, e, k, 55);            \
   sha1_r3(e, a, b, c, d, k, 56);            \
   sha1_r3(d, e, a, b, c, k, 57);            \
   sha1_r3(c, d, e, a, b, k, 58);            \
   sha1_r3(b, c, d, e, a, k, 59);            \
	/* SHA1 round 4 */                        \
   sha1_r4(a, b, c, d, e, k, 60);            \
   sha1_r4(e, a, b, c, d, k, 61);            \
   sha1_r4(d, e, a, b, c, k, 62);            \
   sha1_r4(c, d, e, a, b, k, 63);            \
   sha1_r4(b, c, d, e, a, k, 64);            \
   sha1_r4(a, b, c, d, e, k, 65);            \
   sha1_r4(e, a, b, c, d, k, 66);            \
   sha1_r4(d, e, a, b, c, k, 67);            \
   sha1_r4(c, d, e, a, b, k, 68);            \
   sha1_r4(b, c, d, e, a, k, 69);            \
   sha1_r4(a, b, c, d, e, k, 70);            \
   sha1_r4(e, a, b, c, d, k, 71);            \
   sha1_r4(d, e, a, b, c, k, 72);            \
   sha1_r4(c, d, e, a, b, k, 73);            \
   sha1_r4(b, c, d, e, a, k, 74);            \
   sha1_r4(a, b, c, d, e, k, 75);            \
   sha1_r4(e, a, b, c, d, k, 76);            \
   sha1_r4(d, e, a, b, c, k, 77);            \
   sha1_r4(c, d, e, a, b, k, 78);            \
   sha1_r4(b, c, d, e, a, k, 79);            \
   /* add transformed state */               \
	st[0] += a;                               \
   st[1] += b;                               \
   st[2] += c;                               \
   st[3] += d;                               \
   st[4] += e;                               \
}

typedef struct {
   uint8_t data[64];    /**< Input buffer */
   uint32_t bitlen[2];  /**< Total bit length of updated input */
   uint32_t state[5];   /**< Internal hashing state */
   uint32_t datalen;    /**< Length of buffered input */
} SHA1_CTX;  /**< SHA1 Context */

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

void sha1_init(SHA1_CTX *ctx);
void sha1_update(SHA1_CTX *ctx, const void *in, size_t inlen);
void sha1_final(SHA1_CTX *ctx, void *out);
void sha1(const void *in, size_t inlen, void *out);

/* CUDA testing functions */
#ifdef CUDA
   void test_kcu_sha1(const void *in, size_t *inlen, size_t max_inlen,
      void *out, int num);
#endif

/* end extern "C" {} for C++ */
#ifdef __cplusplus
}
#endif

/* end include guard */
#endif
