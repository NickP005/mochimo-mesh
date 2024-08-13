
#include "_assert.h"
#include "../extinet.h"

int main()
{  /* check; send and receive support functions operate as expected */
   char *request = "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n";
   char response[4096] = "";
   SOCKET sd;

   sock_startup();  /* enables socket support */
   sd = sock_connect_addr("example.com", 80, 3);
   ASSERT_NE_MSG(sd, INVALID_SOCKET, "connection is required for testing");
   ASSERT_EQ_MSG(sock_send(sd, request, (int) strlen(request), 0, 5), 0,
      "sock_send() should successfully send request packet");
   ASSERT_EQ_MSG(sock_recv(sd, response, 1, 0, 5), 0,
      "sock_recv() should recv a single byte as a packet");
   ASSERT_EQ_MSG(sock_recv(sd, &response[1], 4095, 0, 5), (-1),
      "sock_recv() should timeout after receiving entire response");
   ASSERT_GT_MSG(strlen(response), 0, "response should be non-zero");
   sock_close(sd);
   ASSERT_EQ_MSG(sock_send(sd, request, (int) strlen(request), 0, 5), (-1),
      "sock_send() should timeout, socket is closed");
   sock_cleanup();
}
