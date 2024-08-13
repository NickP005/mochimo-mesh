
#include "_assert.h"
#include "../extinet.h"

int main()
{  /* check aton() converts from domain or numbers-and-dots notation */
   ASSERT_EQ(aton("localhost"), 0x100007fUL);
   ASSERT_EQ(aton("255.207.159.111"), 0x6f9fcfffUL);
}
