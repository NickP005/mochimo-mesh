/**
 * @file extint.h
 * @brief Extended integer support.
 * @details Emulates the C Standard library header file `<stdint.h>`
 * and extends on it's functionality.
 * @copyright Adequate Systems LLC, 2021-2022. All Rights Reserved.
 * <br />For license information, please refer to ../LICENSE.md
 * @note 
 * Number Constants for `INT#_MAX` have been removed due to an
 * inability to determine an adequate naming scheme that does not
 * conflict with the C Standard Library headers, "stdint.h" and
 * "inttypes.h". If Number Constants for for `INT#_MAX` values are
 * required, use exact number literals, or define the following:
 * @code
 * // Where '#' represents bit width; 8/16/32/64
 * #define INT#_MAX = ( WORD#_MAX / 2 )
 * #define INT#_MIN = ( ( WORD#_MAX / (-2) ) - 1)
 * @endcode<br/>
 * For 32-bit machine compatibility, either avoid the use of 64-bit
 * elements or provide alternative routines separated by `#ifdef`'s
 * testing for the `HAS_64BIT` definition...
 * @code
 * void someFunction(void)
 * {
 * #ifdef HAS_64BIT
 *    // ... 64-bit routine
 * #else
 *    // ... 32-bit routine
 * #endif
 * }
 * @endcode
*/

/* include guard */
#ifndef EXTENDED_INTEGER_H
#define EXTENDED_INTEGER_H


#include <limits.h>

/* ensure compatible char bit width */
#if CHAR_BIT != 8
   #error Incompatible char bit width ( != 8 ).
#endif

/* determine best 16-bit datatype and default 8-bit constant */
#if UINT_MAX == 0xFFFFU
   #define CONSTANT_8BIT(x)   x
   #define CONSTANT_16BIT(x)  x ## U
   #define INT_OR_SHORT_16BIT int

/* end UINT_MAX == 0xFFFFU */
#elif USHRT_MAX == 0xFFFF
   #define CONSTANT_8BIT(x)   x
   #define CONSTANT_16BIT(x)  x
   #define INT_OR_SHORT_16BIT short int

/* end USHRT_MAX == 0xFFFF */
#else
   #error Cannot determine datatype for 16-bit word.

#endif

/* determine best 32-bit datatype */
#if ULONG_MAX == 0xFFFFFFFFUL
   #define CONSTANT_32BIT(x)     x ## UL
   #define LONG_OR_INT_32BIT     long int
   #define PRINT_PREFIX_32BIT    "l"

/* end ULONG_MAX == 0xFFFFFFFFUL */
#elif UINT_MAX == 0xFFFFFFFFU
   #define CONSTANT_32BIT(x)     x ## U
   #define LONG_OR_INT_32BIT     int
   #define PRINT_PREFIX_32BIT    ""

/* end UINT_MAX == 0xFFFFFFFFU */
#else
   #error Cannot determine datatype for 32-bit word.

#endif

/* ensure 64-bit is available (ULLONG_MAX) and not disabled (DISABLE_64BIT) */
#if defined(ULLONG_MAX) && ! defined(DISABLE_64BIT)

   #define HAS_64BIT

   /* determine best 64-bit datatype */
   #if ULONG_MAX == 0xFFFFFFFFFFFFFFFFUL
      #define CONSTANT_64BIT(x)     x ## UL
      #define LLONG_OR_LONG_64BIT   long int
      #define PRINT_PREFIX_64BIT    "l"

   /* end ULONG_MAX == 0xFFFFFFFFFFFFFFFFUL */
   #elif ULLONG_MAX == 0xFFFFFFFFFFFFFFFFULL
      #define CONSTANT_64BIT(x)     x ## ULL
      #define LLONG_OR_LONG_64BIT   long long int
      #define PRINT_PREFIX_64BIT    "ll"

   /* end ULLONG_MAX == 0xFFFFFFFFFFFFFFFFULL */
   #else
      #error Cannot determine datatype for 64-bit word.

   #endif

/* end 64-bit available and enabled */
#endif

/**
 * Decimal `printf` specifier for `Int8`.
*/
#define P8d "d"

/**
 * Decimal `printf` specifier for `Int8`.
*/
#define P8i "i"

/**
 * Octal `printf` specifier for `Int8` & `Word8`.
*/
#define P8o "o"

/**
 * Decimal `printf` specifier for `Word8`.
*/
#define P8u "u"

/**
 * Hexadecimal `printf` specifier for `Int8` & `Word8`.
*/
#define P8x "x"

/**
 * Hexadecimal `printf` specifier for `Int8` & `Word8`.
*/
#define P8X "X"

/**
 * Decimal `printf` specifier for `Int16`.
*/
#define P16d "d"

/**
 * Decimal `printf` specifier for `Int16`.
*/
#define P16i "i"

/**
 * Octal `printf` specifier for `Int16` & `Word16`.
*/
#define P16o "o"

/**
 * Decimal `printf` specifier for `Word16`.
*/
#define P16u "u"

/**
 * Hexadecimal `printf` specifier for `Int16` & `Word16`.
*/
#define P16x "x"

/**
 * Hexadecimal `printf` specifier for `Int16` & `Word16`.
*/
#define P16X "X"


/**
 * Decimal `printf` specifier for `Int32`.
 * Depending on system/compiler configuration,
 * `PRINT_PREFIX_32BIT` may expand to: `"l"` or `""`.
*/
#define P32d   PRINT_PREFIX_32BIT "d"

/**
 * Decimal `printf` specifier for `Int32`.
 * Depending on system/compiler configuration,
 * `PRINT_PREFIX_32BIT` may expand to: `"l"` or `""`.
*/
#define P32i   PRINT_PREFIX_32BIT "i"

/**
 * Octal `printf` specifier for `Int32` or `Word32`.
 * Depending on system/compiler configuration,
 * `PRINT_PREFIX_32BIT` may expand to: `"l"` or `""`.
*/
#define P32o   PRINT_PREFIX_32BIT "o"

/**
 * Decimal `printf` specifier for `Word32`.
 * Depending on system/compiler configuration,
 * `PRINT_PREFIX_32BIT` may expand to: `"l"` or `""`.
*/
#define P32u   PRINT_PREFIX_32BIT "u"

/**
 * Hexadecimal `printf` specifier for `Int32` or `Word32`.
 * Depending on system/compiler configuration,
 * `PRINT_PREFIX_32BIT` may expand to: `"l"` or `""`.
*/
#define P32x   PRINT_PREFIX_32BIT "x"

/**
 * Hexadecimal `printf` specifier for `Int32` or `Word32`.
 * Depending on system/compiler configuration,
 * `PRINT_PREFIX_32BIT` may expand to: `"l"` or `""`.
*/
#define P32X   PRINT_PREFIX_32BIT "X"

#ifdef HAS_64BIT

   /**
    * Decimal `printf` specifier for `Int64`.
    * Depending on system/compiler/definition configuration,
    * `PRINT_PREFIX_64BIT` may expand to: `"ll"` or `"l"`.
   */
   #define P64d   PRINT_PREFIX_64BIT "d"

   /**
    * Decimal `printf` specifier for `Int64`.
    * Depending on system/compiler/definition configuration,
    * `PRINT_PREFIX_64BIT` may expand to: `"ll"` or `"l"`.
   */
   #define P64i   PRINT_PREFIX_64BIT "i"

   /**
    * Octal `printf` specifier for `Int64` or `Word64`.
    * Depending on system/compiler/definition configuration,
    * `PRINT_PREFIX_64BIT` may expand to: `"ll"` or `"l"`.
   */
   #define P64o   PRINT_PREFIX_64BIT "o"

   /**
    * Decimal `printf` specifier for `Word64`.
    * Depending on system/compiler/definition configuration,
    * `PRINT_PREFIX_64BIT` may expand to: `"ll"` or `"l"`.
   */
   #define P64u   PRINT_PREFIX_64BIT "u"

   /**
    * Hecadecimal `printf` specifier for `Int64` or `Word64`.
    * Depending on system/compiler/definition configuration,
    * `PRINT_PREFIX_64BIT` may expand to: `"ll"` or `"l"`.
   */
   #define P64x   PRINT_PREFIX_64BIT "x"

   /**
    * Hecadecimal `printf` specifier for `Int64` or `Word64`.
    * Depending on system/compiler/definition configuration,
    * `PRINT_PREFIX_64BIT` may expand to: `"ll"` or `"l"`.
   */
   #define P64X   PRINT_PREFIX_64BIT "X"

/* end HAS_64BIT */
#endif

/**
 * (re)Defines an 8-bit number constant.
 * `CONSTANT_8BIT(x)` always expands to `x`, and is only
 * present for reasons of uniformity.
*/
#define WORD8_C(x)   CONSTANT_8BIT(x)

/**
 * (re)Defines a 16-bit number constant.
 * Depending on system/compiler configuration,
 * `CONSTANT_16BIT(x)` may expand to: `xU` or `x`.
*/
#define WORD16_C(x)  CONSTANT_16BIT(x)

/**
 * (re)Defines a 32-bit number constant.
 * Depending on system/compiler configuration,
 * `CONSTANT_32BIT(x)` may expand to: `xUL` or `xU`.
*/
#define WORD32_C(x)  CONSTANT_32BIT(x)

#ifdef HAS_64BIT

   /**
    * (re)Defines a 64-bit number constant.
    * Depending on system/compiler/definition configuration,
    * `CONSTANT_64BIT(x)` may expand to: `xULL` or `xUL`.
   */
   #define WORD64_C(x)  CONSTANT_64BIT(x)

/* end HAS_64BIT */
#endif

/**
 * Largest possible value for `word8` as a Number constant.
*/
#define WORD8_MAX    WORD8_C(0xFF)

/**
 * Largest possible value for `word16` as a Number constant.
*/
#define WORD16_MAX   WORD16_C(0xFFFF)

/**
 * Largest possible value for `word32` as a Number constant.
*/
#define WORD32_MAX   WORD32_C(0xFFFFFFFF)

#ifdef HAS_64BIT

   /**
    * Largest possible value for `word64` as a Number constant.
   */
   #define WORD64_MAX   WORD64_C(0xFFFFFFFFFFFFFFFF)

/* end HAS_64BIT */
#endif

/**
 * 8-bit signed integer. Values range `[-128, 127]`.
*/
typedef char int8;

/**
 * 8-bit unsigned integer. Values range `[0, 255]`.
 * @note Historically, `byte` was used to represent an 8-bit word of
 * unsigned value. However, it's possible to encounter a redefinition
 * conflict with `std::byte`, which was introduced to the Standard in
 * C++17. Thus, it has been superseded by `word8`.
*/
typedef unsigned char word8;

/**
 * 16-bit signed integer. Values range `[-32768, 32767]`.
 * <br/>Depending on system/compiler configuration,
 * `INT_OR_SHORT_16BIT` may expand to: `int` or `short int`.
*/
typedef INT_OR_SHORT_16BIT int16;

/**
 * 16-bit unsigned integer. Values range `[0, 65535]`.
 * <br/>Depending on system/compiler configuration,
 * `INT_OR_SHORT_16BIT` may expand to: `int` or `short int`.
*/
typedef unsigned INT_OR_SHORT_16BIT word16;

/**
 * 32-bit signed integer. Values range `[-2147483648, 2147483647]`.
 * <br/>Depending on system/compiler configuration,
 * `LONG_OR_INT_32BIT` may expand to: `long int` or `int`.
*/
typedef LONG_OR_INT_32BIT int32;

/**
 * 32-bit unsigned integer. Values range `[0, 4294967295]`.
 * <br/>Depending on system/compiler configuration,
 * `LONG_OR_INT_32BIT` may expand to: `long int` or `int`.
*/
typedef unsigned LONG_OR_INT_32BIT word32;

#ifdef HAS_64BIT

   /**
    * 64-bit signed integer.
    * Values range `[-9223372036854775808, 9223372036854775807]`.
    * <br/>Depending on system/compiler/definition configuration,
    * `LLONG_OR_LONG_64BIT` may expand to: `long long int` or `long int`.
   */
   typedef LLONG_OR_LONG_64BIT int64;

   /**
    * 64-bit unsigned integer. Values range `[0, 18446744073709551616]`.
    * <br/>Depending on system/compiler/definition configuration,
    * `LLONG_OR_LONG_64BIT` may expand to: `long long int` or `long int`.
   */
   typedef unsigned LLONG_OR_LONG_64BIT word64;

/* end HAS_64BIT */
#endif

/* end include guard */
#endif
