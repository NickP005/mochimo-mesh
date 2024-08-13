/**
 * @file extprint.h
 * @brief Extended printing and logging support.
 * @copyright Adequate Systems LLC, 2018-2022. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
*/

/* include guard */
#ifndef EXTENDED_PRINT_H
#define EXTENDED_PRINT_H


#include <stdio.h>

/**
 * No print level, none, zip. Indicates no printing
*/
#define PLEVEL_NONE  0

/**
 * Fatal error logging print level. Indicates fatal error level printing
*/
#define PLEVEL_FATAL 1

/**
 * Error logging print level. Indicates standard error level printing
*/
#define PLEVEL_ERROR 2

/**
 * Warning logging print level. Indicates standard warning level printing
*/
#define PLEVEL_WARN  3

/**
 * Standard logging print level. Indicates standard log level printing
*/
#define PLEVEL_LOG   4

/**
 * Debug logging print level. Indicates debug level printing
*/
#define PLEVEL_DEBUG 5

/**
 * Print/log a FATAL error message.
 * @param ... arguments you would normally pass to printf()
 * @returns 2, as integer per print_ext()
*/
#define pfatal(...)  \
   print_ext((-1), PLEVEL_FATAL, __LINE__, __FILE__, __VA_ARGS__)

/**
 * Print/log an error message, with description of @a errnum.
 * @param E @a errno associated with error log message
 * @param ... arguments you would normally pass to printf()
 * @returns 1, as integer per print_ext()
*/
#define perrno(E, ...)  \
   print_ext(E, PLEVEL_ERROR, __LINE__, __FILE__, __VA_ARGS__)

/**
 * Print/log an error message.
 * @param ... arguments you would normally pass to printf()
 * @returns 1, as integer per print_ext()
*/
#define perr(...)    \
   print_ext((-1), PLEVEL_ERROR, __LINE__, __FILE__, __VA_ARGS__)

/**
 * Print/log a warning message.
 * @param ... arguments you would normally pass to printf()
 * @returns 0, as integer per print_ext()
*/
#define pwarn(...)   \
   print_ext((-1), PLEVEL_WARN, __LINE__, __FILE__, __VA_ARGS__)

/**
 * Print/log a message.
 * @param ... arguments you would normally pass to printf()
 * @returns 0, as integer per print_ext()
*/
#define plog(...)    \
   print_ext((-1), PLEVEL_LOG, __LINE__, __FILE__, __VA_ARGS__)

/**
 * Print/log a debug message.
 * @param ... arguments you would normally pass to printf()
 * @returns 0, as integer per print_ext()
*/
#define pdebug(...)  \
   print_ext((-1), PLEVEL_DEBUG, __LINE__, __FILE__, __VA_ARGS__)

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

unsigned get_num_errs(void);
unsigned get_num_logs(void);
int set_output_file(char *fname, char *mode);
void set_output_level(int level);
void set_print_level(int level);
void print(const char *fmt, ...);
void psticky(const char *fmt, ...);
int print_ext(int e, int ll, int line, const char *file, const char *fmt, ...);

#ifdef __cplusplus
}  /* end extern "C" */
#endif

/* end include guard */
#endif
