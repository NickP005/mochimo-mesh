
#include "_assert.h"
#include "../extint.h"
#include "../extmath.h"

int main()
{  /* check; returned carry, and final result */
   word32 numA[8] = {
      0xc0c0aace, 0xc0ffee00, 0xdecafbee, 0xaddedbed,
      0xdecaface, 0xbead50ff, 0xcafef00d, 0xdeadbeef
   };
   word32 numB[8] = {
      0xdeadbeef, 0xcafef00d, 0xbead50ff, 0xdecaface,
      0xaddedbed, 0xdecafbee, 0xc0ffee00, 0xc0c0aace
   };
   word32 expect[8] = {
      0x1ded1421, 0x09ff020d, 0xdfe25511, 0x30ec1ee0,
      0xcf13e11f, 0x201daaee, 0xf600fdf3, 0xe212ebde
   };
   word32 result[8];

   ASSERT_EQ(multi_sub(&numA, &numB, &result, 32), 0);
   ASSERT_EQ(multi_sub(&numB, &numA, &result, 32), 1);
   ASSERT_CMP(result, expect, 32);
}
