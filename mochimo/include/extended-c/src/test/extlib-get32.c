
#include "_assert.h"
#include "../extlib.h"

int main()
{  /* check get32() returns 32-bit unsigned value */
   unsigned char data[9] = { 0x1, 0x3, 0x7, 0xff, 0xf, 0x1f, 0x3f, 0x7f, 0xff };
   ASSERT_EQ(get32(data), WORD32_C(0xff070301));
}
