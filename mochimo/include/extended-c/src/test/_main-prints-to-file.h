
#include <errno.h>
#include "_assert.h"
#include "../extio.h"

int main()
{
/* definition checks */
#ifndef FP
   FILE *FP = NULL;
   assert(FP && "definition was not defined...");
#endif
#ifndef FNAME
   char *FNAME = NULL;
   assert(FNAME && "definition was not defined...");
#endif
#ifndef CONSOLE
   int CONSOLE = 0;
   assert(CONSOLE && "definition was not defined...");
#endif
#ifndef PREFIX
   char *PREFIX = NULL;
   assert(PREFIX && "definition was not defined...");
#endif
#ifndef PFUNC
#define PFUNC(...)   do { /*nothing */ } while(0)
   assert(NULL && "PFUNC() definition was not defined...");
#endif

   /* init */
   char *inputp, input[1024];
   char *printable = " !\"#$%&'()*+,-./0123456789:;<=>?@"
      "ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~";
   size_t printablelen = strlen(printable);  /* count printable lewngth */
   size_t prefixlen = strlen(PREFIX);  /* count prefix length */
   int ii;

   /* check log file opens for writing */
   if (FP != NULL) ASSERT_NE(freopen(FNAME, "w+", FP), NULL);
   else ASSERT_NE((FP = fopen(FNAME, "w+")), NULL);
   Pconsole = CONSOLE;  /* set console log level */
   Ptimestamp = 1;  /* enable timestsamps in logs */
   PFUNC("%s", printable);  /* print to log file */
   rewind(FP);  /* rewind log file */
   /* read log file into input, byte-by-byte */
   for (ii = 0; !feof(FP) && !ferror(FP); ii++) {
      fread(&input[ii], sizeof(char), 1, FP);
   } input[ii] = '\0';  /* ensure input is NUL terminated */

   /* check for fread() failures */
   ASSERT_EQ_MSG(errno, 0, "fread() error");
   /* check prefix was included in log file, if a prefix was specified */
   if (prefixlen) {
      inputp = strstr(input, PREFIX);
      ASSERT_STR(inputp, PREFIX, prefixlen);
      inputp += prefixlen; /* fast-forward input pointer */
   }
   /* check printable ascii was replicated correctly */
   inputp = strstr(input, printable);
   ASSERT_STR(inputp, printable, printablelen);

   /* cleanup */
   if (FP) {
      fclose(FP);
      remove(FNAME);
   }
}
