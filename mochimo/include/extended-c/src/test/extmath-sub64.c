
#include "_assert.h"
#include "../extint.h"
#include "../extmath.h"

int main()
{  /* check; returned carry, and final result */
   word32 numA[2] = { WORD32_C(0xc0c0aace) };
   word32 numB[2] = { WORD32_C(0xdeadbeef), WORD32_C(0xcafef00d) };
   word32 expect[2] = { WORD32_C(0xe212ebdf), WORD32_C(0x35010ff2) };
   word32 result[2] = { 0 };

   /* test ext function */
   ASSERT_EQ(sub64(&result, &numB, &result), 1);
   ASSERT_EQ(sub64(&numB, &numB, &result), 0);
   ASSERT_EQ(sub64(&numB, &numA, &result), 0);
   ASSERT_EQ(sub64(&numA, &numA, &result), 0);
   ASSERT_EQ(sub64(&numA, &numB, &result), 1);
   ASSERT_CMP(result, expect, 8);
   result[0] = result[1] = 0;
   /* test x64 function */
   ASSERT_EQ(sub64_x64(&result, &numB, &result), 1);
   ASSERT_EQ(sub64_x64(&numB, &numB, &result), 0);
   ASSERT_EQ(sub64_x64(&numB, &numA, &result), 0);
   ASSERT_EQ(sub64_x64(&numA, &numA, &result), 0);
   ASSERT_EQ(sub64_x64(&numA, &numB, &result), 1);
   ASSERT_CMP(result, expect, 8);
   result[0] = result[1] = 0;
   /* test x86 function */
   ASSERT_EQ(sub64_x86(&result, &numB, &result), 1);
   ASSERT_EQ(sub64_x86(&numB, &numB, &result), 0);
   ASSERT_EQ(sub64_x86(&numB, &numA, &result), 0);
   ASSERT_EQ(sub64_x86(&numA, &numA, &result), 0);
   ASSERT_EQ(sub64_x86(&numA, &numB, &result), 1);
   ASSERT_CMP(result, expect, 8);
}
