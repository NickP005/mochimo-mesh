
#include "_assert.h"
#include "../extinet.h"

int main()
{  /* check sock state, and sock_startup/cleanup() */
   ASSERT_EQ_MSG(sock_state(), 0, "should initialize Zero (0)");
   sock_startup();
   ASSERT_EQ_MSG(sock_state(), 1, "should increment (1)");
   sock_startup();
   ASSERT_EQ_MSG(sock_state(), 2, "should increment again (2)");
   sock_cleanup();
   ASSERT_EQ_MSG(sock_state(), 0, "should decrement to Zero (0)");
}
