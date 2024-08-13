
#include "_assert.h"
#include "../extint.h"
#include "../extmath.h"

int main()
{  /* check; final result */
   word32 numA[2] = { WORD32_C(0xc0c0aace), WORD32_C(0xc0ffee00) };
   word32 expect[2] = { WORD32_C(0x3f3f5532), WORD32_C(0x3f0011ff) };

   /* test ext function */
   negate64(&numA);
   ASSERT_CMP(numA, expect, 8);
   negate64(&numA);
   /* test x64 function */
   negate64_x64(&numA);
   ASSERT_CMP(numA, expect, 8);
   negate64_x64(&numA);
   /* test x86 function */
   negate64_x86(&numA);
   ASSERT_CMP(numA, expect, 8);
}
