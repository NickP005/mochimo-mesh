
#include "_assert.h"
#include "../extlib.h"

int main()
{  /* check put32() places 32-bit unsigned value in array */
   unsigned char array[8] = { 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f, 0xff };
   unsigned char expect[8] = { 0xff, 0x7f, 0x3f, 0x1f, 0xf, 0x7, 0x3, 0x1 };
   put32(array, WORD32_C(0x1f3f7fff)); ASSERT_CMP(array, expect, 4);
}
