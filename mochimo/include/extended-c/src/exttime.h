/**
 * @file exttime.h
 * @brief Extended time support.
 * @copyright Adequate Systems LLC, 2021-2022. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
*/

/* include guard */
#ifndef EXTENDED_TIME_H
#define EXTENDED_TIME_H


#include "extos.h"
#include <time.h>

/**
 * Suspend the current thread for a specified number of milliseconds.
 * @param ms Number of milliseconds to suspend the current thread
*/
static inline void millisleep(unsigned long ms)
{
#if OS_WINDOWS
   Sleep(ms);

#elif OS_UNIX
   /* nanosleep() is specified by POSIX.1b-1993... */
   #if _POSIX_VERSION >= 199309L
      struct timespec ts;

      /* split seconds and nanoseconds from ms parameter */
      ts.tv_sec = (time_t) (ms / 1000UL);
      ts.tv_nsec = (long) ((ms - (ts.tv_sec * 1000UL)) * 1000000L);
      /* perform POSIX compliant sleep */
      /* while(nanosleep(&ts, &ts) != 0); // uninterruptible */
      nanosleep(&ts, &ts);

   /* end #if _POSIX_VERSION... */
   #else
      usleep(ms * 1000UL);

   #endif

/* end OS_UNIX */
#endif
}

/* end include guard */
#endif
