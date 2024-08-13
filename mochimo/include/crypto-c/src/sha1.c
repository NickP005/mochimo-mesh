/**
 * @private
 * @headerfile sha1.h <sha1.h>
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_SHA1_C
#define CRYPTO_SHA1_C


#include "sha1.h"
#include <string.h>

/**
 * @private
 * SHA1 transformation rounds.
 * @param ctx Pointer to SHA1 context
 * @param data Pointer to input to be transformed
*/
void sha1_transform(SHA1_CTX *ctx, const uint8_t data[])
{
   static const uint32_t k[4] = /* SHA1 transformation constant */
      { 0x5a827999, 0x6ed9eba1, 0x8f1bbcdc, 0xca62c1d6 };
   sha1_transform_unrolled(ctx->state, ((uint32_t *) data), k);
}  /* end sha1_transform() */

/**
 * Initialize a SHA1 context.
 * @param ctx Pointer to SHA1 context
*/
void sha1_init(SHA1_CTX *ctx)
{
   ctx->datalen = 0;
   ctx->bitlen[0] = ctx->bitlen[1] = 0;
   ctx->state[0] = 0x67452301;
   ctx->state[1] = 0xEFCDAB89;
   ctx->state[2] = 0x98BADCFE;
   ctx->state[3] = 0x10325476;
   ctx->state[4] = 0xc3d2e1f0;
}  /* end sha1_init() */

/**
 * Add @a inlen bytes from @a in to a SHA1 context for hashing.
 * @param ctx Pointer to SHA1 context
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
*/
void sha1_update(SHA1_CTX *ctx, const void *in, size_t inlen)
{
   size_t i, n;
   uint32_t old;

   for(i = n = 0; inlen; i += n, inlen -= n) {
      /* copy memory to input buffer in chunks */
      n = (ctx->datalen + inlen) > 64 ? 64 - ctx->datalen : inlen;
      memcpy(ctx->data + ctx->datalen, (const uint8_t *) in + i, n);
      ctx->datalen += n;
      /* process input buffer */
      if (ctx->datalen == 64) {
         sha1_transform(ctx, ctx->data);
         ctx->datalen = 0;
         old = ctx->bitlen[0];
         ctx->bitlen[0] += 512;
         if (ctx->bitlen[0] < old) ctx->bitlen[1]++;  /* add in carry */
      }
   }
}  /* end sha1_update() */

/**
 * Finalize a SHA1 message digest.
 * Generate the SHA1 message digest and place in @a out.
 * @param ctx Pointer to SHA1 context
 * @param out Pointer to location to place the message digest
*/
void sha1_final(SHA1_CTX *ctx, void *out)
{
   uint32_t i, old;

   i = ctx->datalen;

   /* Pad whatever data is left in the buffer. */
   if (ctx->datalen < 56) {
      ctx->data[i++] = 0x80;
      memset(ctx->data + i, 0, 64 - i);
   } else if (ctx->datalen >= 56) {
      ctx->data[i++] = 0x80;
      if (i < 64) memset(ctx->data + i, 0, 64 - i);
      sha1_transform(ctx, ctx->data);
      memset(ctx->data, 0, 56);
   }

   /* Append to the padding the total message's length in bits and
    * transform (big endian). */
   old = ctx->bitlen[0];
   ctx->bitlen[0] += (uint32_t) ctx->datalen << 3;
   if(ctx->bitlen[0] < old) ctx->bitlen[1]++;  /* add in carry */
   /* immitate bswap64() for bitlen */
   ((uint32_t *) ctx->data)[15] = bswap32(ctx->bitlen[0]);
   ((uint32_t *) ctx->data)[14] = bswap32(ctx->bitlen[1]);
   sha1_transform(ctx, ctx->data);

   /* Since this implementation uses little endian byte ordering and
    * SHA uses big endian, reverse all the bytes when copying the
    * final state to the output hash. */
   ((uint32_t *) out)[0] = bswap32(ctx->state[0]);
   ((uint32_t *) out)[1] = bswap32(ctx->state[1]);
   ((uint32_t *) out)[2] = bswap32(ctx->state[2]);
   ((uint32_t *) out)[3] = bswap32(ctx->state[3]);
   ((uint32_t *) out)[4] = bswap32(ctx->state[4]);
}  /* end sha1_final() */

/**
 * Convenient all-in-one SHA1 computation.
 * Performs sha1_init(), sha1_update() and sha1_final(),
 * and places the resulting hash in @a out.
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
 * @param out Pointer to location to place the message digest
*/
void sha1(const void *in, size_t inlen, void *out)
{
   SHA1_CTX ctx;

   sha1_init(&ctx);
   sha1_update(&ctx, in, inlen);
   sha1_final(&ctx, out);
}  /* end sha1() */

/* end include guard */
#endif
