
#include "_assert.h"
#include "../extinet.h"

int main()
{  /* check setting socket to blocking and non-blocking I/O */
   SOCKET sd;

   sock_startup();  /* enables socket support */
   sd = sock_connect_addr("example.com", 80, 3);
   ASSERT_NE_MSG(sd, INVALID_SOCKET, "connection is required for testing");
   ASSERT_EQ_MSG(sock_set_blocking(sd), 0, "should set socket to blocking");
   ASSERT_EQ_MSG(sock_set_nonblock(sd), 0, "should set socket to nonblocking");
   sock_cleanup();
}
