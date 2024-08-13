
#include "_assert.h"
#include "../extio.h"

int main()
{  /* check fexists() returns both falsey and truthy for associated files */
   ASSERT_EQ(fexists("non-existant-file.log"), 0);
   ASSERT_EQ(fexists(__FILE__), 1);
}
