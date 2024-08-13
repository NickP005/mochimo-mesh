
#include "_assert.h"
#include "../extprint.h"
#include <time.h>

int main()
{  /* this test is a visual test for sticky messages */
   char sticky[128] = "Loading";
   time_t start;
   double i, last;

   set_print_level(PLEVEL_DEBUG);

   i = 0;
   last = (-1);
   time(&start);
   /* print progress over 5 seconds, lazy spin loops */
   while(i < 5.0) {
      i = difftime(time(NULL), start);
      if (i == last) continue;
      plog("Update %d.0...", (int) i);
      psticky(strncat(sticky, ".", sizeof(sticky) - strlen(sticky) - 1));
      if ((int) i % 2) plog("Update %d.5...", (int) i);
      last = i;
   }
   psticky("");
   plog("Update final");
}
