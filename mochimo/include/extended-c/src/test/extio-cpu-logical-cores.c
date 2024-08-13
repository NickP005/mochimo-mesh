
#include "_assert.h"
#include "../extio.h"

int main()
{  /* check; CPU cores is greater than zero */
   ASSERT_GT(cpu_cores(), 0);
}
