
#include "_assert.h"
#include "../extio.h"
#include <errno.h>

int main()
{
   char input[1024], nametoolong[FILENAME_MAX + 1];
   char *printable = " !\"#$%&'()*+,-./0123456789:;<=>?@"
      "ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~";
   unsigned i;
   FILE *fp;

   /* check mkdir_p() fault conditions */
   for (i = 0; i < sizeof(nametoolong); i++) nametoolong[i] = 'A';
   ASSERT_EQ(mkdir_p(NULL), 1);
   ASSERT_EQ(errno, EINVAL);
   ASSERT_EQ(mkdir_p(nametoolong), 1);
   ASSERT_EQ(errno, ENAMETOOLONG);
   ASSERT_EQ(mkdir_p(__FILE__"/abc"), 1);
   /* check ftouch fault conditions */
   ASSERT_EQ(ftouch("tmp/dir/tree/test.tmp"), 1);
   /* check fcopy() source fault conditions */
   ASSERT_EQ(fcopy(NULL, NULL), 1);
   ASSERT_EQ(fcopy("tmp/dir/tree/test.tmp", NULL), 1);
   /* create directory structure and touch a file within directory */
   ASSERT_EQ_MSG(mkdir_p("tmp/dir/tree"), 0, "mkdir_p() should return 0");
   ASSERT_EQ_MSG(mkdir_p("tmp/dir/tree"), 0, "mkdir_p() shouldn't complain");
   ASSERT_EQ(ftouch("tmp/dir/tree/test.tmp"), 0);
   ASSERT_EQ(fexists("tmp/dir/tree/test.tmp"), 1);
   /* write data to file for copy test */
   fp = fopen("tmp/dir/tree/test.tmp", "w");
   fwrite(printable, strlen(printable), 1, fp);
   fclose(fp);
   /* check fcopy() destination fault conditions */
   ASSERT_EQ(fcopy("tmp/dir/tree/test.tmp", NULL), 1);
   ASSERT_EQ(fcopy("tmp/dir/tree/test.tmp", "tmp/dir/tree/bad/file"), 1);
   /* copy file to same location */
   ASSERT_EQ(fcopy("tmp/dir/tree/test.tmp", "tmp/dir/tree/test2.tmp"), 0);
   ASSERT_EQ(fexists("tmp/dir/tree/test2.tmp"), 1);
   /* check copied data is exact */
   fp = fopen("tmp/dir/tree/test2.tmp", "r");
   fread(input, 1, sizeof(input), fp);
   fclose(fp);
   ASSERT_STR(input, printable, strlen(printable));

   /* cleanup */
   remove("tmp/dir/tree/test2.tmp");
   remove("tmp/dir/tree/test.tmp");
   remove("tmp/dir/tree");
   remove("tmp/dir");
   remove("tmp");
}
