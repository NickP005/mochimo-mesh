/**
 * @file md5.h
 * @brief MD5 hash function support.
 * @details This file is based on Brad Conte's basic
 * implementations of cryptography algorithms...
 * > <https://github.com/B-Con/crypto-algorithms>
 * ... which was released into the Public Domain.
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_MD5_H
#define CRYPTO_MD5_H


#include "utildev.h"

#define MD5LEN  16   /**< MD5 message digest length, in bytes */

/* MD5 specific routines */
#define F(x,y,z)  ( (x & y) | (~x & z) )
#define G(x,y,z)  ( (x & z) | (y & ~z) )
#define H(x,y,z)  ( x ^ y ^ z )
#define I(x,y,z)  ( y ^ (x | ~z) )
#define FF(a,b,c,d,m,s,t)  { a += F(b,c,d) + m + t; a = b + rol32(a,s); }
#define GG(a,b,c,d,m,s,t)  { a += G(b,c,d) + m + t; a = b + rol32(a,s); }
#define HH(a,b,c,d,m,s,t)  { a += H(b,c,d) + m + t; a = b + rol32(a,s); }
#define II(a,b,c,d,m,s,t)  { a += I(b,c,d) + m + t; a = b + rol32(a,s); }

/* Unrolled MD5 transformation */
#define md5_tranform_unrolled(st, in)        \
{                                            \
   uint32_t a = st[0];                       \
   uint32_t b = st[1];                       \
   uint32_t c = st[2];                       \
   uint32_t d = st[3];                       \
   FF(a, b, c, d, in[0],   7, 0xd76aa478);   \
   FF(d, a, b, c, in[1],  12, 0xe8c7b756);   \
   FF(c, d, a, b, in[2],  17, 0x242070db);   \
   FF(b, c, d, a, in[3],  22, 0xc1bdceee);   \
   FF(a, b, c, d, in[4],   7, 0xf57c0faf);   \
   FF(d, a, b, c, in[5],  12, 0x4787c62a);   \
   FF(c, d, a, b, in[6],  17, 0xa8304613);   \
   FF(b, c, d, a, in[7],  22, 0xfd469501);   \
   FF(a, b, c, d, in[8],   7, 0x698098d8);   \
   FF(d, a, b, c, in[9],  12, 0x8b44f7af);   \
   FF(c, d, a, b, in[10], 17, 0xffff5bb1);   \
   FF(b, c, d, a, in[11], 22, 0x895cd7be);   \
   FF(a, b, c, d, in[12],  7, 0x6b901122);   \
   FF(d, a, b, c, in[13], 12, 0xfd987193);   \
   FF(c, d, a, b, in[14], 17, 0xa679438e);   \
   FF(b, c, d, a, in[15], 22, 0x49b40821);   \
   GG(a, b, c, d, in[1],   5, 0xf61e2562);   \
   GG(d, a, b, c, in[6],   9, 0xc040b340);   \
   GG(c, d, a, b, in[11], 14, 0x265e5a51);   \
   GG(b, c, d, a, in[0],  20, 0xe9b6c7aa);   \
   GG(a, b, c, d, in[5],   5, 0xd62f105d);   \
   GG(d, a, b, c, in[10],  9, 0x02441453);   \
   GG(c, d, a, b, in[15], 14, 0xd8a1e681);   \
   GG(b, c, d, a, in[4],  20, 0xe7d3fbc8);   \
   GG(a, b, c, d, in[9],   5, 0x21e1cde6);   \
   GG(d, a, b, c, in[14],  9, 0xc33707d6);   \
   GG(c, d, a, b, in[3],  14, 0xf4d50d87);   \
   GG(b, c, d, a, in[8],  20, 0x455a14ed);   \
   GG(a, b, c, d, in[13],  5, 0xa9e3e905);   \
   GG(d, a, b, c, in[2],   9, 0xfcefa3f8);   \
   GG(c, d, a, b, in[7],  14, 0x676f02d9);   \
   GG(b, c, d, a, in[12], 20, 0x8d2a4c8a);   \
   HH(a, b, c, d, in[5],   4, 0xfffa3942);   \
   HH(d, a, b, c, in[8],  11, 0x8771f681);   \
   HH(c, d, a, b, in[11], 16, 0x6d9d6122);   \
   HH(b, c, d, a, in[14], 23, 0xfde5380c);   \
   HH(a, b, c, d, in[1],   4, 0xa4beea44);   \
   HH(d, a, b, c, in[4],  11, 0x4bdecfa9);   \
   HH(c, d, a, b, in[7],  16, 0xf6bb4b60);   \
   HH(b, c, d, a, in[10], 23, 0xbebfbc70);   \
   HH(a, b, c, d, in[13],  4, 0x289b7ec6);   \
   HH(d, a, b, c, in[0],  11, 0xeaa127fa);   \
   HH(c, d, a, b, in[3],  16, 0xd4ef3085);   \
   HH(b, c, d, a, in[6],  23, 0x04881d05);   \
   HH(a, b, c, d, in[9],   4, 0xd9d4d039);   \
   HH(d, a, b, c, in[12], 11, 0xe6db99e5);   \
   HH(c, d, a, b, in[15], 16, 0x1fa27cf8);   \
   HH(b, c, d, a, in[2],  23, 0xc4ac5665);   \
   II(a, b, c, d, in[0],   6, 0xf4292244);   \
   II(d, a, b, c, in[7],  10, 0x432aff97);   \
   II(c, d, a, b, in[14], 15, 0xab9423a7);   \
   II(b, c, d, a, in[5],  21, 0xfc93a039);   \
   II(a, b, c, d, in[12],  6, 0x655b59c3);   \
   II(d, a, b, c, in[3],  10, 0x8f0ccc92);   \
   II(c, d, a, b, in[10], 15, 0xffeff47d);   \
   II(b, c, d, a, in[1],  21, 0x85845dd1);   \
   II(a, b, c, d, in[8],   6, 0x6fa87e4f);   \
   II(d, a, b, c, in[15], 10, 0xfe2ce6e0);   \
   II(c, d, a, b, in[6],  15, 0xa3014314);   \
   II(b, c, d, a, in[13], 21, 0x4e0811a1);   \
   II(a, b, c, d, in[4],   6, 0xf7537e82);   \
   II(d, a, b, c, in[11], 10, 0xbd3af235);   \
   II(c, d, a, b, in[2],  15, 0x2ad7d2bb);   \
   II(b, c, d, a, in[9],  21, 0xeb86d391);   \
   st[0] += a;                               \
   st[1] += b;                               \
   st[2] += c;                               \
   st[3] += d;                               \
}

typedef struct {
   uint8_t data[64];    /**< Input buffer */
   uint32_t state[4];   /**< Internal hashing state */
   uint32_t bitlen[2];  /**< Total bit length */
   uint32_t datalen;    /**< Length of buffered input */
} MD5_CTX;  /**< MD5 Context */

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

void md5_init(MD5_CTX *ctx);
void md5_update(MD5_CTX *ctx, const void *in, size_t inlen);
void md5_final(MD5_CTX *ctx, void *out);
void md5(const void *in, size_t inlen, void *out);

/* CUDA testing functions */
#ifdef CUDA
   void test_kcu_md5(const void *in, size_t *inlen, size_t max_inlen,
      void *out, int num);
#endif

/* end extern "C" {} for C++ */
#ifdef __cplusplus
}
#endif

/* end include guard */
#endif
