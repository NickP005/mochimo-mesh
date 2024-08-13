/**
 * @private
 * @headerfile extstring.h <extstring.h>
 * @copyright Adequate Systems LLC, 2018-2022. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
*/

/* include guard */
#ifndef EXTENDED_STRING_C
#define EXTENDED_STRING_C


#include "extstring.h"

#define MEMSWAPCHUNK 1024

/**
 * Swap @a count bytes between pointers @a ax and @a bx.
 * Swap occurs in, at most, 1 kibibyte chunks (defined internally
 * by the MEMSWAPCHUNK definition). https://godbolt.org/z/Wo4K4zWcP
 * @param ax Pointer to primary array of bytes to swap
 * @param bx Pointer to secondary array of bytes to swap
*/
void memswap(void *ax, void *bx, size_t count)
{
   unsigned char *pa = (unsigned char *) ax;
   unsigned char *pb = (unsigned char *) bx;
   unsigned char temp[MEMSWAPCHUNK];

   while (count >= MEMSWAPCHUNK) {
      memcpy(temp, pa, MEMSWAPCHUNK);
      memcpy(pa, pb, MEMSWAPCHUNK);
      memcpy(pb, temp, MEMSWAPCHUNK);
      count -= MEMSWAPCHUNK;
      pa += MEMSWAPCHUNK;
      pb += MEMSWAPCHUNK;
   }
   if (count) {
      memcpy(temp, pa, count);
      memcpy(pa, pb, count);
      memcpy(pb, temp, count);
   }
}  /* end memswap() */

/* end include guard */
#endif
