/**
 * @private
 * @headerfile sha3.h <sha3.h>
 * @copyright This file is released under The MIT License (MIT).
 * For more information, see the header file...
*/

/* include guard */
#ifndef CRYPTO_SHA3_C
#define CRYPTO_SHA3_C


#include "sha3.h"
#include <string.h>

/**
 * @private
 * SHA3_Keccak permutation.
 * @param st Pointer to context state array
*/
static void sha3_keccakf(uint64_t st[])
{
	/* Keccakf round constant */
	static const uint64_t keccakf_rndc[24] = {
		0x0000000000000001, 0x0000000000008082, 0x800000000000808a,
		0x8000000080008000, 0x000000000000808b, 0x0000000080000001,
		0x8000000080008081, 0x8000000000008009, 0x000000000000008a,
		0x0000000000000088, 0x0000000080008009, 0x000000008000000a,
		0x000000008000808b, 0x800000000000008b, 0x8000000000008089,
		0x8000000000008003, 0x8000000000008002, 0x8000000000000080,
		0x000000000000800a, 0x800000008000000a, 0x8000000080008081,
		0x8000000000008080, 0x0000000080000001, 0x8000000080008008
	};

	sha3_keccakf_unrolled(st, keccakf_rndc);
}  /* end sha3_keccakf() */

/**
 * Initialize a SHA3 context.
 * @param ctx Pointer to SHA3 context
 * @param outlen Length* of desired message digest, in bytes<br/>
 * <sup>_*compatible message digest lengths are 28, 32, 48 and 64_</sup>
*/
void sha3_init(SHA3_KECCAK_CTX *ctx, int outlen)
{
   memset(ctx->st.b, 0, 200);
   ctx->outlen = (uint32_t) outlen;
   ctx->rsiz = 200 - (ctx->outlen << 1);
   ctx->pt = 0;
}

/**
 * Add @a inlen bytes from @a in to a SHA3 context for hashing.
 * @param ctx Pointer to SHA3 context
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
*/
void sha3_update(SHA3_KECCAK_CTX *ctx, const void *in, size_t inlen)
{
   size_t i;

   for (i = 0; i < inlen; i++) {
      ctx->st.b[ctx->pt++] ^= ((const uint8_t *) in)[i];
      if (ctx->pt == ctx->rsiz) {
         sha3_keccakf(ctx->st.q);
         ctx->pt = 0;
      }
   }
}

/**
 * Finalize a SHA3 message digest.
 * Generate the SHA3 message digest and place in @a out.
 * @param ctx Pointer to SHA3 context
 * @param out Pointer to location to place the message digest
*/
void sha3_final(SHA3_KECCAK_CTX *ctx, void *out)
{
   ctx->st.b[ctx->pt] ^= 0x06;
   ctx->st.b[ctx->rsiz - 1] ^= 0x80;
   sha3_keccakf(ctx->st.q);

   /* copy digest to out */
   memcpy(out, ctx->st.q, ctx->outlen);
}

/**
 * Convenient all-in-one SHA3 computation.
 * Performs sha3_init(), sha3_update() and sha3_final(),
 * and places the resulting hash in @a out.
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
 * @param out Pointer to location to place the message digest
 * @param outlen Length* of desired message digest, in bytes<br/>
 * <sup>_*compatible message digest lengths are 28, 32, 48 and 64_</sup>
*/
void sha3(const void *in, size_t inlen, void *out, int outlen)
{
   SHA3_KECCAK_CTX sha3;

   sha3_init(&sha3, outlen);
   sha3_update(&sha3, in, inlen);
   sha3_final(&sha3, out);
}

/**
 * Initialize a Keccak context.
 * @param ctx Pointer to Keccak context
 * @param outlen Length* of desired message digest, in bytes<br/>
 * <sup>_*compatible message digest lengths are 28, 32, 48 and 64_</sup>
*/
void keccak_init(SHA3_KECCAK_CTX *ctx, int outlen)
{
   sha3_init(ctx, outlen);
}

/**
 * Add @a inlen bytes from @a in to a Keccak context for hashing.
 * @param ctx Pointer to Keccak context
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
*/
void keccak_update(SHA3_KECCAK_CTX *ctx, const void *in, size_t inlen)
{
   sha3_update(ctx, in, inlen);
}

/**
 * Finalize a Keccak message digest.
 * Generate the Keccak message digest and place in @a out.
 * @param ctx Pointer to Keccak context
 * @param out Pointer to location to place the message digest
*/
void keccak_final(SHA3_KECCAK_CTX *ctx, void *out)
{
   /* This next step essentially converts the sha3_final() step
    * `c->st.b[c->pt] ^= 0x06;`  (to)  `c->st.b[c->pt] ^= 0x01;`
    * as per the original Keccak implementation. */
   ctx->st.b[ctx->pt] ^= 0x07;
   sha3_final(ctx, out);
}

/**
 * Convenient all-in-one Keccak computation.
 * Performs keccak_init(), keccak_update() and keccak_final(),
 * and places the resulting hash in @a out.
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
 * @param out Pointer to location to place the message digest
 * @param outlen Length* of desired message digest, in bytes<br/>
 * <sup>_*compatible message digest lengths are 28, 32, 48 and 64_</sup>
*/
void keccak(const void *in, size_t inlen, void *out, int outlen)
{
   SHA3_KECCAK_CTX keccak;

   keccak_init(&keccak, outlen);
   keccak_update(&keccak, in, inlen);
   keccak_final(&keccak, out);
}

/* end include guard */
#endif
