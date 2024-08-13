/**
 * @private
 * @headerfile extio.h <extio.h>
 * @copyright Adequate Systems LLC, 2018-2021. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
*/

/* include guard */
#ifndef EXTENDED_IO_C
#define EXTENDED_IO_C


#include "extio.h"
#include "extos.h"

#include <errno.h>   /* for access to errno and error codes */
#include <stdarg.h>  /* for va_list functionality */
#include <string.h>  /* for string manipulation and comparison */

#if OS_WINDOWS
   #include <direct.h>  /* for _mkdir() */

#elif OS_UNIX
   #include <sys/stat.h>  /* for mkdir() */

#endif

/**
 * Get the number of logical cores available for use.
 * Includes hyper threaded cores.
 * @returns Number of logical cores available for use.
*/
int cpu_cores(void)
{
   static int cores = 0;

   if (cores == 0) {
   #if OS_WINDOWS
      SYSTEM_INFO sysinfo;
      GetSystemInfo(&sysinfo);
      cores = (int) sysinfo.dwNumberOfProcessors;
   #elif OS_UNIX
      cores = (int) sysconf(_SC_NPROCESSORS_ONLN);
   #else
      #warning Unexpected OS configuration; cpu_cores() affected.
      /* assume single core */
      cores = 1;
   #endif
   }

   return cores;
}  /* end cpu_cores() */

/**
 * Copy a file from one location to another.
 * @param srcpath Path of the source file.
 * @param dstpath Path of the destination file.
 * @return 0 on success, or 1 on error (check errno for details).
*/
int fcopy(char *srcpath, char *dstpath)
{
   char buf[BUFSIZ];
   FILE *sfp, *dfp;
   size_t nBytes;
   int ecode = 1;

   /* open source file */
   sfp = fopen(srcpath, "rb");
   if (sfp != NULL) {
      /* open destination file */
      dfp = fopen(dstpath, "wb");
      if (dfp != NULL) {
         /* transfer bytes in BUFSIZ chunks (set by stdio) */
         do {  /* nBytes represents number of bytes read */
            nBytes = fread(buf, 1, BUFSIZ, sfp);
         } while(nBytes && nBytes == fwrite(buf, 1, nBytes, dfp));
         /* if no file errors, set operation success (0) */
         if (ferror(sfp) == 0 && ferror(dfp) == 0) ecode = 0;
         fclose(dfp);
      }
      fclose(sfp);
   }

   return ecode;
}  /* end fcopy() */

/**
 * Check if a file exists.
 * @param fpath Path to file to check
 * @return 1 if file exists, else 0.
*/
int fexists(char *fpath)
{
   FILE *fp;

   fp = fopen(fpath, "rb");
   if(!fp) return 0;
   fclose(fp);

   return 1;
}  /* end fexists() */

/**
 * Check if a file exists and contains data.
 * Attribution: Thanks David!
 * @param fpath Path o file to check
 * @return 1 if file exists and contains data, else 0.
*/
int fexistsnz(char *fpath)
{
   FILE *fp;
   long len;

   fp = fopen(fpath, "rb");
   if(!fp) return 0;
   fseek(fp, 0, SEEK_END);
   len = ftell(fp);
   fclose(fp);

   return len ? 1 : 0;
}  /* end fexistsnz() */

/**
 * Touch a file. Opens @a fpath in "ab" mode, and closes it.
 * Performs no other operations on the file.
 * @param fpath Path o file to touch
 * @return 0 on success, else 1 on error (check errno for details).
*/
int ftouch(char *fpath)
{
   FILE *fp;

   fp = fopen(fpath, "ab");
   if (fp == NULL) return 1;
   fclose(fp);

   return 0;
}  /* end ftouch() */

/**
 * Create a directory at dirpath (including any parent directories).
 * Immitates the shell command @code mkdir -p <dirpath> @endcode
 * @param dirpath Path of directory to be created
 * @return 0 on success, or 1 on error (check errno for details).
 * @note Where `dirpath` already exists, `mkdir_p()` always succeeds.
*/
int mkdir_p(char *dirpath)
{
   char path[BUFSIZ] = { 0 };
   char *dirpathp = dirpath;
   size_t len = 0;
   int ecode = 1;

   if (dirpath == NULL) errno = EINVAL;
   else if (strlen(dirpath) + 1 >= BUFSIZ) errno = ENAMETOOLONG;
   else do {
      /* split (excl. preceding path separator)... */
      dirpathp = strpbrk(++dirpathp, "\\/");
      if (dirpathp == NULL) len = strlen(dirpath);
      else len = dirpathp - dirpath;
      /* ... and rebuild directory path... */
      strncpy(path, dirpath, len);
      path[len] = '\0';  /* ensure nul-termination */
      /* ... creating parent directories, and ignoring EEXIST */
   #if OS_WINDOWS
      ecode = (_mkdir(path) && errno != EEXIST) ? 1 : 0;
   #elif OS_UNIX
      ecode = (mkdir(path, 0777) && errno != EEXIST) ? 1 : 0;
   #else
      #warning Unexpected OS configuration; mkdir_p() affected.
   #endif
   } while(ecode == 0 && len < strlen(dirpath));

   return ecode;
}

/**
 * Read data directly into a buffer.
 * Reads, at most, @a len bytes from @a fpath into @a buff
 * @param buff Buffer pointer to place read bytes
 * @param len Maximum number of bytes to read
 * @param fpath Path to file to read from
 * @returns Number of bytes read into buff, or (-1) on error
*/
int read_data(void *buff, int len, char *fpath)
{
   FILE *fp;
   size_t count;

   if(len == 0) return 0;
   fp = fopen(fpath, "rb");
   if(fp == NULL) return -1;
   count = fread(buff, 1, len, fp);
   fclose(fp);

   return (int) count;
}  /* end read_data() */

/**
 * Append the result of snprintf to `buf[buflen]`. In this function,
 * @a buflen represents the entire length of @a buf and appends only
 * the remaining amount of characters available in @a buf defined by
 * `( buflen - strlen(buf) )`.
 * @param buf Pointer to character string to write to
 * @param buflen Length of characters to be written (includes null
 * terminator)
 * @param fmt Pointer to null terminated character string specifying
 * how to interpet the variable argument list
 * @param ... arguments specifying data to print
 * @returns The number of characters written if successful or
 * negative value if an error occurred.
*/
int snprintf_append(char *buf, size_t buflen, const char *fmt, ...)
{
   size_t current, remaining;
   va_list args;
   int res = 0;

   current = strlen(buf);
   if (current <= buflen - 1) {
      remaining = buflen - current;
      va_start(args, fmt);
      res = vsnprintf(&buf[current], remaining, fmt, args);
      va_end(args);
   }

   return res;
}

/**
 * Write data from a buffer, directly to file.
 * Writes, at most, @a len bytes from @a buff to @a fpath
 * @param buff Buffer pointer of bytes to write
 * @param len Maximum number of bytes to write
 * @param fpath Path to file to write to
 * @returns Number of bytes written, or (-1) on error.
*/
int write_data(void *buff, int len, char *fpath)
{
   FILE *fp;
   size_t count;

   fp = fopen(fpath, "wb");
   if(fp == NULL) return -1;
   count = fwrite(buff, 1, len, fp);
   fclose(fp);

   return (int) count;
}  /* end write_data() */

/* end include guard */
#endif
