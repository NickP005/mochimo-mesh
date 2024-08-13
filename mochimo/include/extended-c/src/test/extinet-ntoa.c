
#include "_assert.h"
#include "../extinet.h"

int main()
{  /* check ntoa() converts 32-bit unsigned value to hexadecimal string */
   unsigned long decimalIPv4 = 0x6f9fcfff;
   ASSERT_STR(ntoa(&decimalIPv4, NULL), "255.207.159.111", 15);
}
