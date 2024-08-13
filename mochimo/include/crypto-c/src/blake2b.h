/**
 * @file blake2b.h
 * @brief Blake2b hash function support.
 * @details This file is based on Dr. Markku-Juhani O. Saarinen's
 * "somewhat smaller" BLAKE2 implemetation...
 * > <https://github.com/mjosaarinen/blake2_mjosref><br/>
 * ... which was released into the Public Domain under the
 * Creative Commons Zero v1.0 Universal license.
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_BLAKE2B_H
#define CRYPTO_BLAKE2B_H


#include "utildev.h"

#define BLAKE2BLEN256   32  /**< 256-bit Blake2b digest length in bytes */
#define BLAKE2BLEN384   48  /**< 384-bit Blake2b digest length in bytes */
#define BLAKE2BLEN512   64  /**< 512-bit Blake2b digest length in bytes */

/* Number of Blake2b rounds */
#define BLAKE2BROUNDS  12

/* Initialivation vectors for Blake2b */
#define BLAKE2B_IV0   0x6A09E667F3BCC908ull
#define BLAKE2B_IV1   0xBB67AE8584CAA73Bull
#define BLAKE2B_IV2   0x3C6EF372FE94F82Bull
#define BLAKE2B_IV3   0xA54FF53A5F1D36F1ull
#define BLAKE2B_IV4   0x510E527FADE682D1ull
#define BLAKE2B_IV5   0x9B05688C2B3E6C1Full
#define BLAKE2B_IV6   0x1F83D9ABFB41BD6Bull
#define BLAKE2B_IV7   0x5BE0CD19137E2179ull

/* G Mixing function */
#define B2B_G(a, b, c, d, x, y) \
{  \
   v[a] = v[a] + v[b] + x; v[d] = ror64(v[d] ^ v[a], 32);   \
   v[c] = v[c] + v[d]; v[b] = ror64(v[b] ^ v[c], 24);       \
   v[a] = v[a] + v[b] + y; v[d] = ror64(v[d] ^ v[a], 16);   \
   v[c] = v[c] + v[d]; v[b] = ror64(v[b] ^ v[c], 63);       \
}

/* Unrolled Blake2b compression initialization */
#define blake2b_compress_init(v, st, t, last) \
{  \
   v[0] = st[0]; v[1] = st[1]; v[2] = st[2]; v[3] = st[3];  \
   v[4] = st[4]; v[5] = st[5]; v[6] = st[6]; v[7] = st[7];  \
   v[8] = BLAKE2B_IV0; v[9] = BLAKE2B_IV1;                  \
   v[10] = BLAKE2B_IV2; v[11] = BLAKE2B_IV3;                \
   v[12] = BLAKE2B_IV4 ^ t[0]; v[13] = BLAKE2B_IV5 ^ t[1];  \
   v[14] = last ? ~(BLAKE2B_IV6) : BLAKE2B_IV6;             \
   v[15] = BLAKE2B_IV7;                                     \
}

/* Unrolled Blake2b compression rounds */
#define blake2b_compress_rounds(v, in, sigma) \
{  \
   int _i;  \
   for (_i = 0; _i < BLAKE2BROUNDS; _i++) {   \
      B2B_G( 0, 4,  8, 12, in[sigma[_i][ 0]], in[sigma[_i][ 1]]);   \
      B2B_G( 1, 5,  9, 13, in[sigma[_i][ 2]], in[sigma[_i][ 3]]);   \
      B2B_G( 2, 6, 10, 14, in[sigma[_i][ 4]], in[sigma[_i][ 5]]);   \
      B2B_G( 3, 7, 11, 15, in[sigma[_i][ 6]], in[sigma[_i][ 7]]);   \
      B2B_G( 0, 5, 10, 15, in[sigma[_i][ 8]], in[sigma[_i][ 9]]);   \
      B2B_G( 1, 6, 11, 12, in[sigma[_i][10]], in[sigma[_i][11]]);   \
      B2B_G( 2, 7,  8, 13, in[sigma[_i][12]], in[sigma[_i][13]]);   \
      B2B_G( 3, 4,  9, 14, in[sigma[_i][14]], in[sigma[_i][15]]);   \
   }  \
}

/* Unrolled Blake2b compression finalization */
#define blake2b_compress_set(v, st) \
{  \
   st[0] ^= v[0] ^ v[8]; st[1] ^= v[1] ^ v[9];     \
   st[2] ^= v[2] ^ v[10]; st[3] ^= v[3] ^ v[11];   \
   st[4] ^= v[4] ^ v[12]; st[5] ^= v[5] ^ v[13];   \
   st[6] ^= v[6] ^ v[14]; st[7] ^= v[7] ^ v[15];   \
}

typedef struct {
   union {
      uint8_t b[128];   /**< 8-bit input buffer */
      uint64_t q[16];   /**< 64-bit input buffer */
   } in;                /**< Input buffer union */
   uint64_t h[8];       /**< Internal hashing state */
   uint64_t t[2];       /**< Total number of bytes */
   uint64_t c;          /**< Length of buffered input */
   uint64_t outlen;     /**< Digest length */
} BLAKE2B_CTX;    /**< Blake2b context */

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

int blake2b_init(BLAKE2B_CTX *ctx, const void *key, int keylen,
   int outlen);
void blake2b_update(BLAKE2B_CTX *ctx, const void *in, size_t inlen);
void blake2b_final(BLAKE2B_CTX *ctx, void *out);
int blake2b(const void *in, size_t inlen, const void *key,
   int keylen, void *out, int outlen);

/* CUDA testing functions */
#ifdef CUDA
   void test_kcu_blake2b(
      const void *in, size_t *inlen, size_t max_inlen,
      const void *key, int *keylen, int max_keylen,
      void *out, int outlen, int *ret, int num);
#endif

/* end extern "C" {} for C++ */
#ifdef __cplusplus
}
#endif

/* end include guard */
#endif
