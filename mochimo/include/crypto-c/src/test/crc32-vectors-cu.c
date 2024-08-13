
/* must be declared before includes */
#ifndef CUDA
   #define CUDA
#endif

#include <stdint.h>
#include "_assert.h"
#include "../crc32.h"

#define NUMVECTORS    7
#define MAXVECTORLEN  81

/* Test vectors used in RFC 1321 */
static char rfc_1321_vectors[NUMVECTORS][MAXVECTORLEN] = {
   "",
   "a",
   "abc",
   "message digest",
   "abcdefghijklmnopqrstuvwxyz",
   "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789",
   "1234567890123456789012345678901234567890123456789012345678901234"
   "5678901234567890"
};

/* expected results to test vectors */
static uint32_t expect[NUMVECTORS] = {
   0, 0xe8b7be43, 0x352441c2, 0x20159d7f, 0x4c2750bd, 0x1fc2e6d2, 0x7ca94a72
};

int main()
{  /* check crc32() digest results match expected */
   size_t inlen[NUMVECTORS];
   uint32_t ret[NUMVECTORS];
   int j;

   /* init memory (synchronous) */
   memset(ret, 0, sizeof(ret));
   for (j = 0; j < NUMVECTORS; j++) {
      inlen[j] = strlen(rfc_1321_vectors[j]);
   }

   /* perform bulk hash */
   test_kcu_crc32(rfc_1321_vectors, inlen, MAXVECTORLEN, ret, NUMVECTORS);

   /* analyze results */
   for (j = 0; j < NUMVECTORS; j++) {
      ASSERT_EQ(ret[j], expect[j]);
   }
}
