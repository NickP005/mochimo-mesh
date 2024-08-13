
#include "_assert.h"
#include "../extlib.h"

int main()
{  /* check; random numbers do not exceed WORD16_MAX, and final seed */
   word32 iterations = WORD16_MAX;
   word32 expect = WORD32_C(0x18b1e5c0);
   word32 seedA, seedB, seedC;

   srand16(iterations, 0, 0);
   while (iterations--) {
      if (rand16() > WORD16_MAX) {
         get_rand16(&seedA, &seedB, &seedC);
         ASSERT_EQ2_MSG(seedA, seedB, seedC,
            "the resulting seeds resulted in rand16() > WORD16_MAX");
      }
   }
   get_rand16(&seedA, &seedB, &seedC);
   ASSERT_EQ((word32) (seedA + seedB + seedC), expect);
}
