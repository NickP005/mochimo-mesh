
/* must be declared before includes */
#ifndef CUDA
   #define CUDA
#endif

#include <stdint.h>
#include "_assert.h"
#include "../crc16.h"

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
static uint16_t expect[NUMVECTORS] = {
   0, 0x7c87, 0x9dd6, 0x9ba6, 0x63ac, 0x7db0, 0xe73a
};

int main()
{  /* check crc16() digest results match expected */
   size_t inlen[NUMVECTORS];
   uint16_t ret[NUMVECTORS];
   int j;

   /* init memory (synchronous) */
   memset(ret, 0, sizeof(ret));
   for (j = 0; j < NUMVECTORS; j++) {
      inlen[j] = strlen(rfc_1321_vectors[j]);
   }

   /* perform bulk hash */
   test_kcu_crc16(rfc_1321_vectors, inlen, MAXVECTORLEN, ret, NUMVECTORS);

   /* analyze results */
   for (j = 0; j < NUMVECTORS; j++) {
      ASSERT_EQ(ret[j], expect[j]);
   }
}
