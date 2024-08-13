/**
 * @private
 * @headerfile md5.h <md5.h>
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_MD5_C
#define CRYPTO_MD5_C


#include "md5.h"
#include <string.h>  /* for memory handling */

/**
 * @private
 * MD5 transformation rounds.
 * @param ctx Pointer to MD5 context
 * @param data Pointer to input to be transformed
*/
void md5_transform(MD5_CTX *ctx, const uint8_t data[])
{
   md5_tranform_unrolled(ctx->state, ((uint32_t *) data));
}  /* end md5_transform() */

/**
 * Initialize a MD5 context.
 * @param ctx Pointer to MD5 context
*/
void md5_init(MD5_CTX *ctx)
{
   ctx->datalen = 0;
   ctx->bitlen[0] = ctx->bitlen[1] = 0;
   ctx->state[0] = 0x67452301;
   ctx->state[1] = 0xEFCDAB89;
   ctx->state[2] = 0x98BADCFE;
   ctx->state[3] = 0x10325476;
}  /* end md5_init() */

/**
 * Add @a inlen bytes from @a in to a MD5 context for hashing.
 * @param ctx Pointer to MD5 context
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
*/
void md5_update(MD5_CTX *ctx, const void *in, size_t inlen)
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
         md5_transform(ctx, ctx->data);
         ctx->datalen = 0;
         old = ctx->bitlen[0];
         ctx->bitlen[0] += 512;
         if (ctx->bitlen[0] < old) ctx->bitlen[1]++;  /* add in carry */
      }
   }
}  /* end md5_update() */

/**
 * Finalize a MD5 message digest.
 * Generate the MD5 message digest and place in @a out.
 * @param ctx Pointer to MD5 context
 * @param out Pointer to location to place the message digest
*/
void md5_final(MD5_CTX *ctx, void *out)
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
      md5_transform(ctx, ctx->data);
      memset(ctx->data, 0, 56);
   }

   /* Append to the padding the total message's length in bits */
   old = ctx->bitlen[0];
   ctx->bitlen[0] += ctx->datalen << 3;
   if (ctx->bitlen[0] < old) ctx->bitlen[1]++;  /* add in carry */
   ((uint32_t *) ctx->data)[14] = ctx->bitlen[0];
   ((uint32_t *) ctx->data)[15] = ctx->bitlen[1];

   /* perform final transform */
   md5_transform(ctx, ctx->data);

   /* copy digest to out */
   memcpy(out, ctx->state, MD5LEN);
}  /* end md5_final() */

/**
 * Convenient all-in-one MD5 computation.
 * Performs md5_init(), md5_update() and md5_final(),
 * and places the resulting hash in @a out.
 * @param in Pointer to data to hash
 * @param inlen Length of @a in data, in bytes
 * @param out Pointer to location to place the message digest
*/
void md5(const void *in, size_t inlen, void *out)
{
   MD5_CTX ctx;

   md5_init(&ctx);
   md5_update(&ctx, in, inlen);
   md5_final(&ctx, out);
}  /* end md5() */

/* end include guard */
#endif
