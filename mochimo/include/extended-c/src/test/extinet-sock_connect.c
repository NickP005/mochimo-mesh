
#include "_assert.h"
#include "../extinet.h"

int main()
{  /* check; connectip() connects to listening sockets */
   sock_startup();  /* enables socket support */
   ASSERT_EQ_MSG(sock_connect_addr("localhost", 80, 3), INVALID_SOCKET,
      "connection should timeout for connection to localhost:443");
   ASSERT_NE_MSG(sock_connect_addr("example.com", 80, 3), INVALID_SOCKET,
      "connection should succeed after listener is created");
   sock_cleanup();
}
