
#include "_assert.h"
#include "../extinet.h"

int main()
{  /* check gethostip() success */
   char ipv4addr[24] = "";
   int ipv4addrlen = (int) sizeof(ipv4addr);

   /* socket initialization */
   sock_startup();

   ASSERT_EQ_MSG(gethostip(ipv4addr, ipv4addrlen), 0, "gethostip() failed");
   ASSERT_NE_MSG(*ipv4addr, '\0', "ipv4addr string is empty");
   ASSERT_NE_MSG(gethostip(ipv4addr, 0), 0, "gethostip() didn't fail");
}
