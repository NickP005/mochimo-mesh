
#include <stdlib.h>
#include "_assert.h"
#include "../extio.h"

#define EMPTYFNAME   "empty.log"

int main()
{  /* check fexistsnz() returns both falsey and truthy for associated files */
   system("touch " EMPTYFNAME);
   ASSERT_EQ(fexistsnz(EMPTYFNAME), 0);
   ASSERT_EQ(fexistsnz(__FILE__), 1);
   remove(EMPTYFNAME);
}
