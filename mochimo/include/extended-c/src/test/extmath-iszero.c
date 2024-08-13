
#include "_assert.h"
#include "../extint.h"
#include "../extmath.h"

int main()
{  /* check iszero() appropriately returns falsey and truthy values */
   unsigned char nonzero[9] = { 0, 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f, 0xff };
   unsigned char zero[9] = { 0 };

   /* test ext function */
   ASSERT_EQ(iszero(nonzero, 9), 0);
   ASSERT_EQ(iszero(zero, 9), 1);
   /* test x64 function */
   ASSERT_EQ(iszero_x64(nonzero, 9), 0);
   ASSERT_EQ(iszero_x64(zero, 9), 1);
   /* test x86 function */
   ASSERT_EQ(iszero_x86(nonzero, 9), 0);
   ASSERT_EQ(iszero_x86(zero, 9), 1);
}
