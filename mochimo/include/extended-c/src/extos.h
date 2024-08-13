/**
 * @file extos.h
 * @brief Extended operating system.
 * @details Defines an operating system type definition of the format
 * `OS_<system>` and includes a base operating system header file.
 * <br/>On Windows, defines `OS_WINDOWS` and includes `<Windows.h>`
 * (ensuring `WIN32_LEAN_AND_MEAN` is defined and returned to its
 * original state, before and after the include, respectively).
 * <br/>On UNIX, defines `OS_UNIX` and includes `<unistd.h>`.
 * @copyright Adequate Systems LLC, 2022. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
*/

/* include guard */
#ifndef EXTENDED_OS_H
#define EXTENDED_OS_H


#if defined(_WIN32)
   /* Windows operating system */
   #define OS_WINDOWS 1

   #ifdef WIN32_LEAN_AND_MEAN
      /* WIN32_LEAN_AND_MEAN is already defined, no change */
      #include <Windows.h>

   #else
      /* define WIN32_LEAN_AND_MEAN */
      #define WIN32_LEAN_AND_MEAN
      #include <Windows.h>
      /* return to previous state (undefined) */
      #undef WIN32_LEAN_AND_MEAN

   #endif

/* end #if defined(_WIN32) */
#elif defined(__unix__) || defined(__unix) || \
      (defined(__APPLE__) && defined(__MACH__))
   /* UNIX-style operating system */
   #define OS_UNIX 1
   #include <unistd.h>

/* end #elif defined(__unix__) || ... */
#endif

/* end include guard */
#endif
