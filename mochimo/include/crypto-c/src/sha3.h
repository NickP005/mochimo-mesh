/**
 * @file sha3.h
 * @brief SHA3 hash function support.
 * @details This file is based on Dr. Markku-Juhani O. Saarinen's
 * "cooked up" compact and readable keccak implemetation...
 * > <https://github.com/mjosaarinen/tiny_sha3>
 * ... which was released under the MIT license (MIT).
 * <br/><br/>
 * Alterations to the transform function were based on
 * Marko Kreen's optimized C implementation...
 * > <https://github.com/markokr/spongeshaker>
 * ... which was released under the ISC License (ISC).
 *
 * @copyright 2014 Marko Kreen <markokr@gmail.com>
 *
 * Permission to use, copy, modify, and/or distribute this software
 * for any purpose with or without fee is hereby granted, provided
 * that the above copyright notice and this permission notice appear
 * in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL
 * WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL
 * THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR
 * CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
 * LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT,
 * NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN
 * CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 * @copyright 2015 Markku-Juhani O. Saarinen <mjos@iki.fi>
 * @copyright 2020-2022 Adequate Systems, LLC.
 *
 * Permission is hereby granted, free of charge, to any person
 * obtaining a copy of this software and associated documentation
 * files (the "Software"), to deal in the Software without
 * restriction, including without limitation the rights to use, copy,
 * modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * * The above copyright notice and this permission notice shall be
 *   included in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
 * OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
 * HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
*/

/* include guard */
#ifndef CRYPTO_SHA3_H
#define CRYPTO_SHA3_H


#include "utildev.h"

#define SHA3LEN224      28 /**< SHA3 224-bit digest length, in bytes */
#define SHA3LEN256      32 /**< SHA3 256-bit digest length, in bytes */
#define SHA3LEN384      48 /**< SHA3 384-bit digest length, in bytes */
#define SHA3LEN512      64 /**< SHA3 512-bit digest length, in bytes */
#define KECCAKLEN224    28 /**< KECCAK 224-bit digest length, in bytes */
#define KECCAKLEN256    32 /**< KECCAK 256-bit digest length, in bytes */
#define KECCAKLEN384    48 /**< KECCAK 384-bit digest length, in bytes */
#define KECCAKLEN512    64 /**< KECCAK 512-bit digest length, in bytes */

#define KECCAKFROUNDS   24

#define sha3_keccakf_unrolled(st, k)   \
{                                                     \
   uint64_t Ba, Be, Bi, Bo, Bu;                       \
	uint64_t Ca, Ce, Ci, Co, Cu;                       \
	uint64_t Da, De, Di, Do, Du;                       \
	int r;                                             \
	for (r = 0; r < KECCAKFROUNDS; r += 4) {           \
		/* Unrolled 4 rounds at a time */               \
		Ca = st[0] ^ st[5] ^ st[10] ^ st[15] ^ st[20];  \
		Ce = st[1] ^ st[6] ^ st[11] ^ st[16] ^ st[21];  \
		Ci = st[2] ^ st[7] ^ st[12] ^ st[17] ^ st[22];  \
		Co = st[3] ^ st[8] ^ st[13] ^ st[18] ^ st[23];  \
		Cu = st[4] ^ st[9] ^ st[14] ^ st[19] ^ st[24];  \
		Da = Cu ^ rol64(Ce, 1);                         \
		De = Ca ^ rol64(Ci, 1);                         \
		Di = Ce ^ rol64(Co, 1);                         \
		Do = Ci ^ rol64(Cu, 1);                         \
		Du = Co ^ rol64(Ca, 1);                         \
		Ba = (st[0] ^ Da);                              \
		Be = rol64((st[6] ^ De), 44);                   \
		Bi = rol64((st[12] ^ Di), 43);                  \
		Bo = rol64((st[18] ^ Do), 21);                  \
		Bu = rol64((st[24] ^ Du), 14);                  \
		st[0]  = Ba ^ ((~Be) & Bi) ^ k[r];              \
		st[6]  = Be ^ ((~Bi) & Bo);                     \
		st[12] = Bi ^ ((~Bo) & Bu);                     \
		st[18] = Bo ^ ((~Bu) & Ba);                     \
		st[24] = Bu ^ ((~Ba) & Be);                     \
		Bi = rol64((st[10] ^ Da), 3);                   \
		Bo = rol64((st[16] ^ De), 45);                  \
		Bu = rol64((st[22] ^ Di), 61);                  \
		Ba = rol64((st[3] ^ Do), 28);                   \
		Be = rol64((st[9] ^ Du), 20);                   \
		st[10] = Ba ^ ((~Be) & Bi);                     \
		st[16] = Be ^ ((~Bi) & Bo);                     \
		st[22] = Bi ^ ((~Bo) & Bu);                     \
		st[3] = Bo ^ ((~Bu) & Ba);                      \
		st[9] = Bu ^ ((~Ba) & Be);                      \
		Bu = rol64((st[20] ^ Da), 18);                  \
		Ba = rol64((st[1] ^ De), 1);                    \
		Be = rol64((st[7] ^ Di), 6);                    \
		Bi = rol64((st[13] ^ Do), 25);                  \
		Bo = rol64((st[19] ^ Du), 8);                   \
		st[20] = Ba ^ ((~Be) & Bi);                     \
		st[1] = Be ^ ((~Bi) & Bo);                      \
		st[7] = Bi ^ ((~Bo) & Bu);                      \
		st[13] = Bo ^ ((~Bu) & Ba);                     \
		st[19] = Bu ^ ((~Ba) & Be);                     \
		Be = rol64((st[5] ^ Da), 36);                   \
		Bi = rol64((st[11] ^ De), 10);                  \
		Bo = rol64((st[17] ^ Di), 15);                  \
		Bu = rol64((st[23] ^ Do), 56);                  \
		Ba = rol64((st[4] ^ Du), 27);                   \
		st[5] = Ba ^ ((~Be) & Bi);                      \
		st[11] = Be ^ ((~Bi) & Bo);                     \
		st[17] = Bi ^ ((~Bo) & Bu);                     \
		st[23] = Bo ^ ((~Bu) & Ba);                     \
		st[4] = Bu ^ ((~Ba) & Be);                      \
		Bo = rol64((st[15] ^ Da), 41);                  \
		Bu = rol64((st[21] ^ De), 2);                   \
		Ba = rol64((st[2] ^ Di), 62);                   \
		Be = rol64((st[8] ^ Do), 55);                   \
		Bi = rol64((st[14] ^ Du), 39);                  \
		st[15] = Ba ^ ((~Be) & Bi);                     \
		st[21] = Be ^ ((~Bi) & Bo);                     \
		st[2] = Bi ^ ((~Bo) & Bu);                      \
		st[8] = Bo ^ ((~Bu) & Ba);                      \
		st[14] = Bu ^ ((~Ba) & Be);                     \
		Ca = st[0] ^ st[10] ^ st[20] ^ st[5] ^ st[15];  \
		Ce = st[6] ^ st[16] ^ st[1] ^ st[11] ^ st[21];  \
		Ci = st[12] ^ st[22] ^ st[7] ^ st[17] ^ st[2];  \
		Co = st[18] ^ st[3] ^ st[13] ^ st[23] ^ st[8];  \
		Cu = st[24] ^ st[9] ^ st[19] ^ st[4] ^ st[14];  \
		Da = Cu ^ rol64(Ce, 1);                         \
		De = Ca ^ rol64(Ci, 1);                         \
		Di = Ce ^ rol64(Co, 1);                         \
		Do = Ci ^ rol64(Cu, 1);                         \
		Du = Co ^ rol64(Ca, 1);                         \
		Ba = (st[0] ^ Da);                              \
		Be = rol64((st[16] ^ De), 44);                  \
		Bi = rol64((st[7] ^ Di), 43);                   \
		Bo = rol64((st[23] ^ Do), 21);                  \
		Bu = rol64((st[14] ^ Du), 14);                  \
		st[0] = Ba ^ ((~Be) & Bi) ^ k[r + 1];           \
		st[16] = Be ^ ((~Bi) & Bo);                     \
		st[7] = Bi ^ ((~Bo) & Bu);                      \
		st[23] = Bo ^ ((~Bu) & Ba);                     \
		st[14] = Bu ^ ((~Ba) & Be);                     \
		Bi = rol64((st[20] ^ Da), 3);                   \
		Bo = rol64((st[11] ^ De), 45);                  \
		Bu = rol64((st[2] ^ Di), 61);                   \
		Ba = rol64((st[18] ^ Do), 28);                  \
		Be = rol64((st[9] ^ Du), 20);                   \
		st[20] = Ba ^ ((~Be) & Bi);                     \
		st[11] = Be ^ ((~Bi) & Bo);                     \
		st[2] = Bi ^ ((~Bo) & Bu);                      \
		st[18] = Bo ^ ((~Bu) & Ba);                     \
		st[9] = Bu ^ ((~Ba) & Be);                      \
		Bu = rol64((st[15] ^ Da), 18);                  \
		Ba = rol64((st[6] ^ De), 1);                    \
		Be = rol64((st[22] ^ Di), 6);                   \
		Bi = rol64((st[13] ^ Do), 25);                  \
		Bo = rol64((st[4] ^ Du), 8);                    \
		st[15] = Ba ^ ((~Be) & Bi);                     \
		st[6] = Be ^ ((~Bi) & Bo);                      \
		st[22] = Bi ^ ((~Bo) & Bu);                     \
		st[13] = Bo ^ ((~Bu) & Ba);                     \
		st[4] = Bu ^ ((~Ba) & Be);                      \
		Be = rol64((st[10] ^ Da), 36);                  \
		Bi = rol64((st[1] ^ De), 10);                   \
		Bo = rol64((st[17] ^ Di), 15);                  \
		Bu = rol64((st[8] ^ Do), 56);                   \
		Ba = rol64((st[24] ^ Du), 27);                  \
		st[10] = Ba ^ ((~Be) & Bi);                     \
		st[1] = Be ^ ((~Bi) & Bo);                      \
		st[17] = Bi ^ ((~Bo) & Bu);                     \
		st[8] = Bo ^ ((~Bu) & Ba);                      \
		st[24] = Bu ^ ((~Ba) & Be);                     \
		Bo = rol64((st[5] ^ Da), 41);                   \
		Bu = rol64((st[21] ^ De), 2);                   \
		Ba = rol64((st[12] ^ Di), 62);                  \
		Be = rol64((st[3] ^ Do), 55);                   \
		Bi = rol64((st[19] ^ Du), 39);                  \
		st[5] = Ba ^ ((~Be) & Bi);                      \
		st[21] = Be ^ ((~Bi) & Bo);                     \
		st[12] = Bi ^ ((~Bo) & Bu);                     \
		st[3] = Bo ^ ((~Bu) & Ba);                      \
		st[19] = Bu ^ ((~Ba) & Be);                     \
		Ca = st[0] ^ st[20] ^ st[15] ^ st[10] ^ st[5];  \
		Ce = st[16] ^ st[11] ^ st[6] ^ st[1] ^ st[21];  \
		Ci = st[7] ^ st[2] ^ st[22] ^ st[17] ^ st[12];  \
		Co = st[23] ^ st[18] ^ st[13] ^ st[8] ^ st[3];  \
		Cu = st[14] ^ st[9] ^ st[4] ^ st[24] ^ st[19];  \
		Da = Cu ^ rol64(Ce, 1);                         \
		De = Ca ^ rol64(Ci, 1);                         \
		Di = Ce ^ rol64(Co, 1);                         \
		Do = Ci ^ rol64(Cu, 1);                         \
		Du = Co ^ rol64(Ca, 1);                         \
		Ba = (st[0] ^ Da);                              \
		Be = rol64((st[11] ^ De), 44);                  \
		Bi = rol64((st[22] ^ Di), 43);                  \
		Bo = rol64((st[8] ^ Do), 21);                   \
		Bu = rol64((st[19] ^ Du), 14);                  \
		st[0] = Ba ^ ((~Be) & Bi) ^ k[r + 2];           \
		st[11] = Be ^ ((~Bi) & Bo);                     \
		st[22] = Bi ^ ((~Bo) & Bu);                     \
		st[8] = Bo ^ ((~Bu) & Ba);                      \
		st[19] = Bu ^ ((~Ba) & Be);                     \
		Bi = rol64((st[15] ^ Da), 3);                   \
		Bo = rol64((st[1] ^ De), 45);                   \
		Bu = rol64((st[12] ^ Di), 61);                  \
		Ba = rol64((st[23] ^ Do), 28);                  \
		Be = rol64((st[9] ^ Du), 20);                   \
		st[15] = Ba ^ ((~Be) & Bi);                     \
		st[1] = Be ^ ((~Bi) & Bo);                      \
		st[12] = Bi ^ ((~Bo) & Bu);                     \
		st[23] = Bo ^ ((~Bu) & Ba);                     \
		st[9] = Bu ^ ((~Ba) & Be);                      \
		Bu = rol64((st[5] ^ Da), 18);                   \
		Ba = rol64((st[16] ^ De), 1);                   \
		Be = rol64((st[2] ^ Di), 6);                    \
		Bi = rol64((st[13] ^ Do), 25);                  \
		Bo = rol64((st[24] ^ Du), 8);                   \
		st[5] = Ba ^ ((~Be) & Bi);                      \
		st[16] = Be ^ ((~Bi) & Bo);                     \
		st[2] = Bi ^ ((~Bo) & Bu);                      \
		st[13] = Bo ^ ((~Bu) & Ba);                     \
		st[24] = Bu ^ ((~Ba) & Be);                     \
		Be = rol64((st[20] ^ Da), 36);                  \
		Bi = rol64((st[6] ^ De), 10);                   \
		Bo = rol64((st[17] ^ Di), 15);                  \
		Bu = rol64((st[3] ^ Do), 56);                   \
		Ba = rol64((st[14] ^ Du), 27);                  \
		st[20] = Ba ^ ((~Be) & Bi);                     \
		st[6] = Be ^ ((~Bi) & Bo);                      \
		st[17] = Bi ^ ((~Bo) & Bu);                     \
		st[3] = Bo ^ ((~Bu) & Ba);                      \
		st[14] = Bu ^ ((~Ba) & Be);                     \
		Bo = rol64((st[10] ^ Da), 41);                  \
		Bu = rol64((st[21] ^ De), 2);                   \
		Ba = rol64((st[7] ^ Di), 62);                   \
		Be = rol64((st[18] ^ Do), 55);                  \
		Bi = rol64((st[4] ^ Du), 39);                   \
		st[10] = Ba ^ ((~Be) & Bi);                     \
		st[21] = Be ^ ((~Bi) & Bo);                     \
		st[7] = Bi ^ ((~Bo) & Bu);                      \
		st[18] = Bo ^ ((~Bu) & Ba);                     \
		st[4] = Bu ^ ((~Ba) & Be);                      \
		Ca = st[0] ^ st[15] ^ st[5] ^ st[20] ^ st[10];  \
		Ce = st[11] ^ st[1] ^ st[16] ^ st[6] ^ st[21];  \
		Ci = st[22] ^ st[12] ^ st[2] ^ st[17] ^ st[7];  \
		Co = st[8] ^ st[23] ^ st[13] ^ st[3] ^ st[18];  \
		Cu = st[19] ^ st[9] ^ st[24] ^ st[14] ^ st[4];  \
		Da = Cu ^ rol64(Ce, 1);                         \
		De = Ca ^ rol64(Ci, 1);                         \
		Di = Ce ^ rol64(Co, 1);                         \
		Do = Ci ^ rol64(Cu, 1);                         \
		Du = Co ^ rol64(Ca, 1);                         \
		Ba = (st[0] ^ Da);                              \
		Be = rol64((st[1] ^ De), 44);                   \
		Bi = rol64((st[2] ^ Di), 43);                   \
		Bo = rol64((st[3] ^ Do), 21);                   \
		Bu = rol64((st[4] ^ Du), 14);                   \
		st[0] = Ba ^ ((~Be) & Bi) ^ k[r + 3];           \
		st[1] = Be ^ ((~Bi) & Bo);                      \
		st[2] = Bi ^ ((~Bo) & Bu);                      \
		st[3] = Bo ^ ((~Bu) & Ba);                      \
		st[4] = Bu ^ ((~Ba) & Be);                      \
		Bi = rol64((st[5] ^ Da), 3);                    \
		Bo = rol64((st[6] ^ De), 45);                   \
		Bu = rol64((st[7] ^ Di), 61);                   \
		Ba = rol64((st[8] ^ Do), 28);                   \
		Be = rol64((st[9] ^ Du), 20);                   \
		st[5] = Ba ^ ((~Be) & Bi);                      \
		st[6] = Be ^ ((~Bi) & Bo);                      \
		st[7] = Bi ^ ((~Bo) & Bu);                      \
		st[8] = Bo ^ ((~Bu) & Ba);                      \
		st[9] = Bu ^ ((~Ba) & Be);                      \
		Bu = rol64((st[10] ^ Da), 18);                  \
		Ba = rol64((st[11] ^ De), 1);                   \
		Be = rol64((st[12] ^ Di), 6);                   \
		Bi = rol64((st[13] ^ Do), 25);                  \
		Bo = rol64((st[14] ^ Du), 8);                   \
		st[10] = Ba ^ ((~Be) & Bi);                     \
		st[11] = Be ^ ((~Bi) & Bo);                     \
		st[12] = Bi ^ ((~Bo) & Bu);                     \
		st[13] = Bo ^ ((~Bu) & Ba);                     \
		st[14] = Bu ^ ((~Ba) & Be);                     \
		Be = rol64((st[15] ^ Da), 36);                  \
		Bi = rol64((st[16] ^ De), 10);                  \
		Bo = rol64((st[17] ^ Di), 15);                  \
		Bu = rol64((st[18] ^ Do), 56);                  \
		Ba = rol64((st[19] ^ Du), 27);                  \
		st[15] = Ba ^ ((~Be) & Bi);                     \
		st[16] = Be ^ ((~Bi) & Bo);                     \
		st[17] = Bi ^ ((~Bo) & Bu);                     \
		st[18] = Bo ^ ((~Bu) & Ba);                     \
		st[19] = Bu ^ ((~Ba) & Be);                     \
		Bo = rol64((st[20] ^ Da), 41);                  \
		Bu = rol64((st[21] ^ De), 2);                   \
		Ba = rol64((st[22] ^ Di), 62);                  \
		Be = rol64((st[23] ^ Do), 55);                  \
		Bi = rol64((st[24] ^ Du), 39);                  \
		st[20] = Ba ^ ((~Be) & Bi);                     \
		st[21] = Be ^ ((~Bi) & Bo);                     \
		st[22] = Bi ^ ((~Bo) & Bu);                     \
		st[23] = Bo ^ ((~Bu) & Ba);                     \
		st[24] = Bu ^ ((~Ba) & Be);                     \
	}                                                  \
}

/**
 * SHA3/Keccak hashing context
*/
typedef struct {
   union {
      uint8_t b[200];   /**< 8-bit input buffer */
      uint64_t q[25];   /**< 64-bit input buffer */
   } st;                /**< Input buffer union */
   uint32_t outlen;     /**< Digest length, in bytes */
   uint32_t rsiz;       /**< Rate size, in bytes */
   uint32_t pt;         /**< Length of buffered input */
} SHA3_KECCAK_CTX;

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

void sha3_init(SHA3_KECCAK_CTX *ctx, int outlen);
void sha3_update(SHA3_KECCAK_CTX *ctx, const void *in, size_t inlen);
void sha3_final(SHA3_KECCAK_CTX *ctx, void *out);
void sha3(const void *in, size_t inlen, void *out, int outlen);
void keccak_init(SHA3_KECCAK_CTX *ctx, int outlen);
void keccak_update(SHA3_KECCAK_CTX *ctx, const void *in, size_t inlen);
void keccak_final(SHA3_KECCAK_CTX *ctx, void *out);
void keccak(const void *in, size_t inlen, void *out, int outlen);

/* CUDA testing functions */
#ifdef CUDA
   void test_kcu_sha3(const void *in, size_t *inlen, size_t max_inlen,
      void *out, int outlen, int num);
   void test_kcu_keccak(const void *in, size_t *inlen, size_t max_inlen,
      void *out, int outlen, int num);
#endif

/* end extern "C" {} for C++ */
#ifdef __cplusplus
}
#endif

/* end include guard */
#endif
