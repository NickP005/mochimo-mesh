
#include "_assert.h"
#include "../extlib.h"

int main()
{  /* check put64() places 64-bit unsigned value in array */
   unsigned char array[8] = { 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f, 0xff };
   unsigned char expect[8] = { 0xff, 0x7f, 0x3f, 0x1f, 0xf, 0x7, 0x3, 0x1 };
   word32 array64[2] = { WORD32_C(0x0f070301), WORD32_C(0xff7f3f1f) };
   word32 expect64[2] = { WORD32_C(0x1f3f7fff), WORD32_C(0x0103070f) };

   /* text ext function */
   put64(array, &expect64);
   ASSERT_CMP(array, expect, 8);
   put64(array, &array64);
   /* text x64 function */
   put64(array, &expect64);
   ASSERT_CMP(array, expect, 8);
   put64_x64(array, &array64);
   /* text x86 function */
   put64_x86(array, &expect64);
   ASSERT_CMP(array, expect, 8);
}
