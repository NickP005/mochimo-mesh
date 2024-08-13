
#include "_assert.h"
#include "../extlib.h"

int main()
{  /* check get16() returns 16-bit unsigned value */
   unsigned char data[9] = { 0x1, 0xff, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f, 0xff };
   ASSERT_EQ(get16(data), WORD16_C(0xff01));
}
