
#include "_assert.h"
#include "../extint.h"
#include "../extmath.h"

int main()
{  /* check; final result */
   word32 numA[2] = { WORD32_C(0xc0c0aace), WORD32_C(0xc0ffee00) };
   word32 expect[2] = { WORD32_C(0x60605567), WORD32_C(0x607ff700) };
   word32 expect2[2] = { WORD32_C(0x30302ab3), WORD32_C(0x303ffb80) };
   word32 expect3[2] = { WORD32_C(0x18181559), WORD32_C(0x181ffdc0) };

   /* test ext function */
   shiftr64(&numA);
   ASSERT_CMP(numA, expect, 8);
   /* test x64 function */
   shiftr64_x64(&numA);
   ASSERT_CMP(numA, expect2, 8);
   /* test x86 function */
   shiftr64_x86(&numA);
   ASSERT_CMP(numA, expect3, 8);
}
