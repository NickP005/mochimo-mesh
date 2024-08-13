
#include "_assert.h"
#include "../extint.h"
#include "../extmath.h"

int main()
{  /* check; returned carry */
   word32 numA[8] = {
      0xdeadbeef, 0xcafef00d, 0xbead50ff, 0xdecaface,
      0xaddedbed, 0xdecafbee, 0xc0ffee00, 0xc0c0aace
   };
   word32 numB[8] = {
      0xc0c0aace, 0xc0ffee00, 0xdecafbee, 0xaddedbed,
      0xdecaface, 0xbead50ff, 0xcafef00d, 0xdeadbeef
   };

   /* test ext function */
   ASSERT_EQ(cmp256(&numA, &numB), -1);
   ASSERT_EQ(cmp256(&numB, &numA), 1);
   ASSERT_EQ(cmp256(&numA, &numA), 0);
   /* test x64 function */
   ASSERT_EQ(cmp256_x64(&numA, &numB), -1);
   ASSERT_EQ(cmp256_x64(&numB, &numA), 1);
   ASSERT_EQ(cmp256_x64(&numA, &numA), 0);
   /* test x86 function */
   ASSERT_EQ(cmp256_x86(&numA, &numB), -1);
   ASSERT_EQ(cmp256_x86(&numB, &numA), 1);
   ASSERT_EQ(cmp256_x86(&numA, &numA), 0);
}
