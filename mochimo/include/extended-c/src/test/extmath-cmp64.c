
#include "_assert.h"
#include "../extint.h"
#include "../extmath.h"

int main()
{  /* check; returned carry, and final result */
   word32 numA[2] = { WORD32_C(0xc0c0aace), WORD32_C(0xc0ffee00) };
   word32 numB[2] = { WORD32_C(0xdeadbeef), WORD32_C(0xcafef00d) };

   /* test ext function */
   ASSERT_EQ(cmp64(&numA, &numB), -1);
   ASSERT_EQ(cmp64(&numB, &numA), 1);
   ASSERT_EQ(cmp64(&numA, &numA), 0);
   /* test x64 function */
   ASSERT_EQ(cmp64_x64(&numA, &numB), -1);
   ASSERT_EQ(cmp64_x64(&numB, &numA), 1);
   ASSERT_EQ(cmp64_x64(&numA, &numA), 0);
   /* test x86 function */
   ASSERT_EQ(cmp64_x86(&numA, &numB), -1);
   ASSERT_EQ(cmp64_x86(&numB, &numA), 1);
   ASSERT_EQ(cmp64_x86(&numA, &numA), 0);
}
