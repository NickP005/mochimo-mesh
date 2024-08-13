/**
 * @private
 * @headerfile extprint.h <extprint.h>
 * @copyright Adequate Systems LLC, 2018-2022. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
*/

/* include guard */
#ifndef EXTENDED_PRINT_C
#define EXTENDED_PRINT_C


#include "extos.h"
#include "extprint.h"
#include "extthread.h"  /* for mutex support */
#include <stdarg.h>     /* for va_list support */
#include <string.h>     /* for string support */
#include <time.h>       /* for timestamp support */

/* Initialize default runtime configuration */
Mutex Printlock = MUTEX_INITIALIZER;
Mutex Outputlock = MUTEX_INITIALIZER;
FILE *Outputfp;
int Printlevel;
int Outputlevel;
unsigned Nprinterrs;
unsigned Nprintlogs;

/* localtime_r() is not specified by Windows... */
#ifdef OS_WINDOWS
   #define localtime_r(t, tm) localtime_s(tm, t)

#endif

/* strerror_r() is specified by POSIX.1-2001... */
#if ! defined(_POSIX_VERSION) || _POSIX_VERSION < 200112L

static void strerror_r(int errnum, char *buf, size_t buflen)
{  /* ... provides fallback, when strerror_r() is unavailable */
   static Mutex lock = MUTEX_INITIALIZER;

   /* a NULL buf or a bufflen of zero (0) is inappropriate */
   if (buf == NULL || buflen == 0) return;
   /* secure the integrity of strerror()'s static string */
   mutex_lock(&lock);
   strncpy(buf, strerror(errnum), buflen);
   mutex_unlock(&lock);
}  /* end strerror_r() */

/* end _POSIX_VERSION */
#endif

/**
 * @private
 * Move the cursor position around the screen.
 * @param x number of x axis steps to move, negative for left
 * @param y number of y axis steps to move, negative for up
*/
static void move_cursor(int x, int y)
{
#ifdef OS_WINDOWS
   COORD coord;
   CONSOLE_SCREEN_BUFFER_INFO csbi;
   HANDLE handle = GetStdHandle(STD_OUTPUT_HANDLE);

   if (handle == INVALID_HANDLE_VALUE) return;
   if (!GetConsoleScreenBufferInfo(handle, &csbi)) return;

   coord.X = csbi.dwCursorPosition.X + x;
   coord.Y = csbi.dwCursorPosition.Y + y;
   SetConsoleCursorPosition(handle, coord);

#else
   if (x < 0) printf("\x1b[%dD", -(x));
   else if (x) printf("\x1b[%dC", x);
   if (y < 0) printf("\x1b[%dA", -(y));
   else if (y) printf("\x1b[%dB", y);

#endif
}

/**
 * @private
 * Clear line from cursor right.
 * @param fp file pointer to location of clear
*/
static void clear_right(FILE *fp)
{
#ifdef OS_WINDOWS
   COORD coord;
   HANDLE handle;
   DWORD cells, count;
   CONSOLE_SCREEN_BUFFER_INFO csbi;

   if (fp == stdout) handle = GetStdHandle(STD_OUTPUT_HANDLE);
   else if (fp == stderr) handle = GetStdHandle(STD_ERROR_HANDLE);
   else return;

   if (handle == INVALID_HANDLE_VALUE) return;
   if (!GetConsoleScreenBufferInfo(handle, &csbi)) return;

   coord = csbi.dwCursorPosition;
   cells = csbi.dwSize.X - csbi.dwCursorPosition.X;
   FillConsoleOutputCharacter(handle, (TCHAR) ' ', cells, coord, &count);

#else
   if (fp == stdout || fp == stderr) fprintf(fp, "\x1b[K");

#endif
}

/**
 * @private
 * Prints a "clear" string, by "clearing right" every newline.
 * When printing to screen, it's possible to get garbled
 * text due to movement of the cursor (e.g. psticky() ), so this
 * splits str into lines and clears right after every line.
 * @param fp file pointer to location of output
 * @param str string to print
*/
static void print_clear(FILE *fp, char *str)
{
   char *strp = str;
   char *nextp;

   /* print str in parts separated by any "\r\n" characters */
   while( (nextp = strpbrk(strp + 1, "\r\n")) ) {
      fprintf(fp, "%.*s", (int) (nextp - strp), strp);
      if (nextp[-1] != '\r') clear_right(fp);
      strp = nextp;
   }
   /* print remaining str and clear right */
   fprintf(fp, "%s", strp);
}

/**
 * @private
 * Prints a "clear" string format, with variable arguments.
 * Writes format string and arguments to intermediate buffer
 * of size BUFSIZ, and utilizes print_clear() for final output.
 * @param fp file pointer to location of output
 * @param fmt string fmt to print
 * @param ... variable arguments supporting @a fmt
*/
static void fprintf_clear(FILE *fp, const char *fmt, ...)
{
   char str[BUFSIZ];
   va_list args;

   va_start(args, fmt);
   vsnprintf(str, BUFSIZ, fmt, args);
   va_end(args);

   print_clear(fp, str);
}

/**
 * @private
 * Prints a "clear" string format and variable arguments list,
 * passed from another function. Writes format string and arguments
 * to intermediate buffer `str[BUFSIZ]`, and utilizes print_clear()
 * for final output.
 * @param fp file pointer to location of output
 * @param fmt string fmt to print
 * @param args variable arguments list supporting @a fmt
 * @note Expects va_start() to be performed on @a args before calling
 * and calls va_end() on @a args before returning
*/
static void vfprintf_clear_end(FILE *fp, const char *fmt, va_list args)
{
   char str[BUFSIZ];

   vsnprintf(str, BUFSIZ, fmt, args);
   va_end(args);

   print_clear(fp, str);
}

/**
 * Get the number of errors counted by extprint().
 * @returns Number of logs counted by extprint()
*/
unsigned get_num_errs(void)
{
   unsigned counter;

   mutex_lock(&Outputlock);
   counter = Nprinterrs;
   mutex_unlock(&Outputlock);

   return counter;
}

/**
 * Get the number of logs counted by extprint().
 * @returns Number of logs counted by extprint()
*/
unsigned get_num_logs(void)
{
   unsigned counter;

   mutex_lock(&Outputlock);
   counter = Nprintlogs;
   mutex_unlock(&Outputlock);

   return counter;
}

/**
 * Set the output file for printing logs to file.
 * If @a fname or @a mode are `NULL`, output file is closed.
 * @param fname file name of output file
 * @param mode file mode of output file (e.g. "w" or "a")
 * @returns 0 on success, else 1 on error (see @a errno for details)
*/
int set_output_file(char *fname, char *mode)
{
   int ecode = 0;

   mutex_lock(&Outputlock);
   if (Outputfp) fclose(Outputfp);
   if (fname && mode) {
      Outputfp = fopen(fname, mode);
      if (Outputfp == NULL) ecode = 1;
   } else Outputfp = NULL;
   mutex_unlock(&Outputlock);

   return ecode;
}

/**
 * Set the print level for printing to output.
 * @param level The print level allowed to print to output file
*/
void set_output_level(int level)
{
   mutex_lock(&Outputlock);
   Outputlevel = level;
   mutex_unlock(&Outputlock);
}

/**
 * Set the print level for printing to screen.
 * @param level The print level allowed to print to screen
*/
void set_print_level(int level)
{
   mutex_lock(&Printlock);
   Printlevel = level;
   mutex_unlock(&Printlock);
}

/**
 * Print a sticky message. Prints and "sticks" a message to the
 * bottom of a terminal.
 * @param fmt A string format (or message) for printing
 * @param ... Variable arguments supporting the format string
 * @note To clear sticky message use @code psticky(""); @endcode
 * @note To reprint sticky message use @code psticky(NULL); @endcode
 * @note Message is only sticky when using the extended print
 * functions and print MACROS, ie. print(), perr(), plog(), ...
*/
void psticky(const char *fmt, ...)
{
   static Mutex lock = MUTEX_INITIALIZER;
   static char sticky[BUFSIZ] = "";
   static int nls = 0;
   va_list args;
   char *cp;
   int i;

   /* bail on print level NONE, or NULL update */
   if (Printlevel == PLEVEL_NONE) return;
   if (fmt == NULL && strlen(sticky) == 0) return;

   mutex_lock(&lock);

   if (fmt) {
      if (!fmt[0]) {
         mutex_lock(&Printlock);

         /* clear sticky */
         for(i = 0; i <= nls; i++) {
            printf("\n");
            clear_right(stdout);
         }
         /* return cursor */
         move_cursor(0, -(1 + nls));

         mutex_unlock(&Printlock);

         sticky[0] = '\0';
         nls = 0;
      } else {
         /* update sticky message */
         va_start(args, fmt);
         vsnprintf(sticky, BUFSIZ, fmt, args);
         va_end(args);

         /* count newlines in sticky */
         for(cp = sticky, nls = 0; (cp = strchr(cp, '\n')); cp++, nls++);
      }
   }

   mutex_lock(&Printlock);

   /* (re)print sticky message */
   fprintf_clear(stdout, "\n%s\r", sticky);
   /* return cursor */
   move_cursor(0, -(1 + nls));
   clear_right(stdout);
   /* flush output */
   fflush(stdout);

   mutex_unlock(&Printlock);

   mutex_unlock(&lock);
}

/**
 * Print a clean message to stdout. When printing to screen, it's
 * possible to get garbled text due to movement of the cursor by
 * psticky(), so this splits the resulting string into lines and
 * clears right after every line.
 * @param fmt A string format (or message) for printing
 * @param ... Variable arguments supporting the format string
 * @note Prints to screen, regardless of specified print level.
*/
void print(const char *fmt, ...)
{
   va_list args;

   mutex_lock(&Printlock);

   va_start(args, fmt);
   vfprintf_clear_end(stdout, fmt, args);
   fflush(stdout);

   mutex_unlock(&Printlock);

   /* reprint sticky message */
   psticky(NULL);
}

/**
 * Print to screen and log to file.
 * @param e error number ( @a errno ) associated with log
 * @param ll print level of log
 * @param line the line number to place in the trace details
 * @param file the filename to place in the trace details
 * @param fmt A string format (or message) to log
 * @param ... Variable arguments supporting @a fmt
 * @returns 2 for PLEVEL_FATAL, 1 for PLEVEL_ERROR, else 0
 * @note This function is a multipurpose function designed for
 * use through the extended print MACROS; pfatal(), perrno(),
 * perr(), pwarn(), plog(), pdebug()
*/
int print_ext
(int e, int ll, int line, const char *file, const char *fmt, ...)
{
   char *fname = "<unknown>";
   char trace[BUFSIZ] = "";
   char timestamp[32] = "";
   char prefix[128] = "";
   char suffix[64] = "";
   unsigned *counterp;
   int code = 0;
   struct tm dt;
   time_t t;
   FILE *fp;
   va_list args;

   /* set return code */
   if (ll == PLEVEL_FATAL) code = 2;
   if (ll == PLEVEL_ERROR) code = 1;

   /* ignore NULL fmt's */
   if (fmt == NULL) return code;

   /* derive base filename */
   if (file) {
      if (strrchr(file, '\\')) fname = strrchr(file, '\\') + 1;
      else if (strrchr(file, '/')) fname = strrchr(file, '/') + 1;
      else fname = (char *) file;
   }

   /* combine base file and line number in trace */
   snprintf(trace, BUFSIZ, "[%s:%d] ", fname, line);

   /* build print configuration - based on level */
   switch(ll) {
      case PLEVEL_FATAL:
         strncat(prefix, trace, sizeof(prefix) - 1);
         strncat(prefix, "!!! FATAL ", sizeof(prefix) - strlen(prefix) - 1);
         /* fall through */
      case PLEVEL_ERROR:
         strncat(prefix, "Error. ", sizeof(prefix) - strlen(prefix) - 1);
         if (e >= 0) {
            suffix[0] = ':'; suffix[1] = ' ';
            strerror_r(e, &suffix[2], sizeof(suffix) - 3);
         }
         counterp = &Nprinterrs;
         fp = stderr;
         break;
      case PLEVEL_WARN:
         strncat(prefix, "Warning... ", sizeof(prefix) - 1);
         counterp = &Nprintlogs;
         fp = stdout;
         break;
      case PLEVEL_DEBUG:
         strncat(prefix, "DEBUG: ", sizeof(prefix) - 1);
         /* fall through */
      case PLEVEL_LOG:
         /* fall through */
      default:
         counterp = &Nprintlogs;
         fp = stdout;
   }

   /* grab local time for log */
   time(&t);
   localtime_r(&t, &dt);

   /* print to screen, timestamp: "hh:mm:ss" */
   if (Printlevel >= ll) {
      mutex_lock(&Printlock);

      fprintf(fp, "%s", prefix);
      va_start(args, fmt);
      vfprintf_clear_end(fp, fmt, args);
      fprintf(fp, "%s", suffix);
      clear_right(fp);
      fprintf(fp, "\n");
      fflush(fp);

      mutex_unlock(&Printlock);

      /* update sticky messages */
      psticky(NULL);
   }

   mutex_lock(&Outputlock);

   /* print to output, timestamp: "yyyy-mm-ddThh:mm:ss+0000" */
   if (Outputfp && Outputlevel >= ll) {
      strftime(timestamp, sizeof(timestamp) - 1, "%FT%T%z - ", &dt);
      fprintf(Outputfp, "%s%s", timestamp, prefix);
      va_start(args, fmt);
      vfprintf(Outputfp, fmt, args);
      va_end(args);
      fprintf(Outputfp, "%s\n", suffix);
      fflush(Outputfp);
   }

   /* increment appropriate print counter */
   (*counterp)++;

   mutex_unlock(&Outputlock);

   return code;
}

/* end include guard */
#endif
