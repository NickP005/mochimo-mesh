
#include "_assert.h"
#include "../extinet.h"

int main()
{  /* check get_sock_ip() returns appropriate value */
   sock_startup();  /* enables socket support */
   SOCKET sd = sock_connect_addr("example.com", 80, 3);

   ASSERT_NE_MSG(sd, INVALID_SOCKET, "connection is required for testing");
   ASSERT_EQ(get_sock_ip(sd), aton("example.com"));
   ASSERT_EQ(get_sock_ip(INVALID_SOCKET), (unsigned long) SOCKET_ERROR);
   sock_cleanup();
}
