
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
      0x9f6e69bd, 0x8bfede0e, 0x9d784cee, 0x8ca9d6bc,
      0x8ca9d6bc, 0x9d784cee, 0x8bfede0e, 0x9f6e69be
   };
   word32 result[8] = { 0 };

   ASSERT_EQ(multi_add(&result, &numB, &result, 32), 0);
   ASSERT_EQ(multi_add(&numA, &numB, &result, 32), 1);
   ASSERT_CMP(result, expect, 32);
}
