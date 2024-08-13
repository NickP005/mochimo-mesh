/**
 * @file extio.h
 * @brief Extended input/output support.
 * @details Provides extended support for system input and output.
 * @copyright Adequate Systems LLC, 2018-2021. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
*/

/* include guard */
#ifndef EXTENDED_IO_H
#define EXTENDED_IO_H


#include <stdio.h>

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

int cpu_cores(void);
int fcopy(char *srcpath, char *dstpath);
int fexists(char *fpath);
int fexistsnz(char *fpath);
int ftouch(char *fpath);
int mkdir_p(char *dirpath);
int read_data(void *buff, int len, char *fpath);
int snprintf_append(char *buf, size_t buflen, const char *fmt, ...);
int write_data(void *buff, int len, char *fpath);

#ifdef __cplusplus
}  /* end extern "C" */
#endif

/* end include guard */
#endif
