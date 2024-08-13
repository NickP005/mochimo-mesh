/**
 * @private
 * @headerfile extlib.h <extlib.h>
 * @copyright Adequate Systems LLC, 2018-2022. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
*/

/* include guard */
#ifndef EXTENDED_UTILITIES_C
#define EXTENDED_UTILITIES_C


#include "extlib.h"
#include "extmath.h"  /* for iszero() in *nz() functions */
#include "extstring.h"  /* for memory manipulation support */

/* Internal state seeds for PRNG's rand16fast() and rand16() */
volatile word32 Lseed = 1;
volatile word32 Lseed2 = 1;
volatile word32 Lseed3 = 362436069;
volatile word32 Lseed4 = 123456789;

/* 64-bit guard */
#ifdef HAS_64BIT

   /* Place a 64-bit unsigned @a value in @a buff. */
   void put64_x64(void *buff, void *value)
   {
      *((word64 *) buff) = *((word64 *) value);
   }

#endif  /* end WORD64_MAX */

/* Place a 64-bit unsigned @a value in @a buff. */
void put64_x86(void *buff, void *value)
{
   ((word32 *) buff)[0] = ((word32 *) value)[0];
   ((word32 *) buff)[1] = ((word32 *) value)[1];
}

/**
 * Get a 16-bit unsigned value from @a buff.
 * @param buff Pointer to buffer to get value from
 * @returns 16-bit unsigned value from @a buff.
*/
word16 get16(void *buff)
{
   return *((word16 *) buff);
}

/**
 * Place a 16-bit unsigned @a value in @a buff.
 * @param buff Pointer to buffer to place value
 * @param value 16-bit unsigned value
*/
void put16(void *buff, word16 value)
{
   *((word16 *) buff) = value;
}

/**
 * Get a 32-bit unsigned value from @a buff.
 * @param buff Pointer to buffer to get value from
 * @returns 32-bit unsigned value from @a buff.
*/
word32 get32(void *buff)
{
   return *((word32 *) buff);
}

/**
 * Place a 32-bit @a value in @a buff.
 * @param buff Pointer to buffer to place value
*/
void put32(void *buff, word32 value)
{
   *((word32 *) buff) = value;
}

/**
 * Place a 64-bit unsigned @a value in @a buff.
 * @param buff Pointer to buffer to place value
 * @param value Pointer to 64-bit value
*/
void put64(void *buff, void *value)
{
#ifdef WORD64_MAX
   put64_x64(buff, value);
#else
   put64_x86(buff, value);
#endif
}

/**
 * Set the internal state seed for rand16fast().
 * @param x Value to set the internal state seed @a Lseed to.
*/
void srand16fast(word32 x)
{
   Lseed = x;
}

/**
 * Get the current internal state seed used by rand16fast().
 * @returns Value of internal state seed @a Lseed. */
word32 get_rand16fast(void)
{
   return Lseed;
}

/**
 * Set the internal state seed for rand16().
 * @param x Value to set the internal state seed @a Lseed to
 * @param y Value to set the internal state seed @a Lseed to
 * @param z Value to set the internal state seed @a Lseed to
*/
void srand16(word32 x, word32 y, word32 z)
{
   Lseed2 = x;
   Lseed3 = y;
   Lseed4 = z;
}

/**
 * Get the current internal state seeds used by rand16().
 * @param x Pointer to location to place internal state seed @a Lseed2
 * @param y Pointer to location to place internal state seed @a Lseed3
 * @param z Pointer to location to place internal state seed @a Lseed4
*/
void get_rand16(word32 *x, word32 *y, word32 *z)
{
   *x = Lseed2;
   *y = Lseed3;
   *z = Lseed4;
}

/**
 * Fast 16-bit PRNG using internal state seed @a Lseed.
 * Based on Dr. Marsaglia's Usenet post of a linear
 * congruential generator.
 * @returns Random number range [0, 65535].
 * @warning Exhibits an increase in PRNG generation at the
 * cost of statistical randomness, when compared to rand16().
*/
word32 rand16fast(void)
{
   Lseed = Lseed * WORD32_C(69069) + WORD32_C(262145);

   return Lseed >> 16;
}

/**
 * 16-bit PRNG using internal state seeds @a Lseed2/3/4.
 * Based on Dr. Marsaglia's KISS method. Produces
 * reasonable 16-bit statistical randomness.
 * @returns Random number range [0, 65535].
*/
word32 rand16(void)
{
   /* linear congruential generator */
   Lseed2 = Lseed2 * WORD32_C(69069) + WORD32_C(262145);
   /* multiply with carry */
   if(Lseed3 == 0) Lseed3 = WORD32_C(362436069);
   Lseed3 = WORD32_C(36969) * (Lseed3 & WORD32_C(65535)) + (Lseed3 >> 16);
   /* linear-feedback shift register */
   if(Lseed4 == 0) Lseed4 = WORD32_C(123456789);
   Lseed4 ^= (Lseed4 << 17);
   Lseed4 ^= (Lseed4 >> 13);
   Lseed4 ^= (Lseed4 << 5);
   /* the KISS method (combination of methods) */
   return (Lseed2 ^ (Lseed3 << 16) ^ Lseed4) >> 16;
}  /* end rand16() */

/**
 * Shuffle a `list[count]` of @a size byte elements.
 * Uses Durstenfeld's implementation of the Fisher-Yates
 * shuffling algorithm.
 * @note Shuffle @a count is bound to 16 bits due to rand16().
 * @note Set random seed with srand16() before use.
*/
void shuffle(void *list, size_t size, size_t count)
{
   unsigned char *listp = (unsigned char *) list;
   unsigned char *elemp, *swapp;

   if (count < 2) return;  /* list is not worth shuffling */
   elemp = &listp[(count - 1) * size];
   for( ; count > 1; count--, elemp -= size) {
      /* for every element (in reverse order), swap with random
       * element whose index is less than the current */
      swapp = &listp[(rand16() % count) * size];
      memswap(elemp, swapp, size);
   }
}  /* end shuffle() */

/**
 * Shuffle a `list[count]` of non-zero, @a size byte elements.
 * A zero value marks the end of list, and shuffling does
 * NOT occur for said value onwards.<br/>Uses Durstenfeld's
 * implementation of the Fisher-Yates shuffling algorithm.
 * @note Shuffle @a count is bound to 16 bits due to rand16().
 * @note Set random seed with srand16() before use.
*/
void shufflenz(void *list, size_t size, size_t count)
{
   char *listp = ((char *) list) + (size * count) - size;

   /* reduce count for every zero value at the end of the list */
   for( ; count > 0 && iszero(listp, size); listp -= size, count--);
   shuffle(list, size, count);
}  /* end shufflenz() */

/**
 * Search a `list[count]` of non-zero, @a size byte elements
 * for a @a value.
 * @returns Pointer to found value, else NULL if not found.
*
void *search(void *value, void *list, size_t size, size_t count)
{
   unsigned char listp = (unsigned char *) list;

   for( ; count--; listp += size) {
      if(memcmp(value, listp, size)) continue;
      return ((void *) listp);
   }

   return NULL;
}  // end search() */

/**
 * Search a `list[count]` of non-zero, @a size byte elements
 * for a @a value. A zero value marks the end of list.
 * @returns Pointer to found value, else NULL if not found.
*
void *searchnz(void *value, void *list, size_t size, size_t count)
{
   unsigned char listp = (unsigned char *) list;

   for( ; count-- && iszero(listp, size); listp += size) {
      if(memcmp(value, listp, size)) continue;
      return ((void *) listp);
   }

   return NULL;
}  // end searchnz() */

/**
 * Extract a @a value from a `list[count]` of non-zero, @a size
 * byte elements.
 * @returns Pointer to extracted value, else NULL if not found.
*
void *extractnz(void *value, void *list, size_t size, size_t count)
{
   char *end = &((char *) list)[size * count];
   char *last = end - size;
   void *found;

   found = searchnz(value, list, size, count);
   if (found == NULL) return NULL;
   if (found < ((size_t) end - size)) {
      memmove(found, ((char *) found)[size], (size_t) (end - found));
   }  // set final element zero
   memset(last, 0, size);

   return value;
}  // end extractnz() */

/**
 * Append a non-zero, @a size byte value to a `list[count]`.
 * @returns Pointer to appended value, else NULL if not appended.
*
void *appendnz(void *value, void *list, size_t size, size_t count)
{
   char *listp = (char *) list;

   if (value == NULL || iszero(value, size)) return NULL;
   for( ; listp < &listp[size * count]; listp += size) {
      if (iszero(listp, size)) {
         memset(listp, value, size);
         return ((void *) listp);
      }
   }

   return NULL;
}  // end appendnz() */

/* end include guard */
#endif
