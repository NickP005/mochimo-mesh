
#include "_assert.h"
#include "../extlib.h"

int main()
{  /* check; random numbers do not exceed WORD16_MAX, and final seed */
   word32 iterations = WORD16_MAX;
   word32 expect = WORD32_C(0x19eb1f6);

   srand16fast(iterations);
   while (iterations--) {
      if (rand16fast() > WORD16_MAX) {
         ASSERT_LE_MSG(get_rand16fast(), WORD16_MAX,
            "the resulting seed resulted in rand16fast() > WORD16_MAX");
      }
   }
   ASSERT_EQ(get_rand16fast(), expect);
}
