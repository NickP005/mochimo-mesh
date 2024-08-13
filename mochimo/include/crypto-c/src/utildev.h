/**
 * @file utildev.h
 * @brief Device utilities and includes support.
 * @details Provides utilities and includes for device architectures.
 * | Architecture | Supported |
 * | :----------: | :-------: |
 * | CPU          | Yes       |
 * | CUDA         | Yes       |
 * | OPENCL       | No        |
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_UTILS_H
#define CRYPTO_UTILS_H


#include <stddef.h>  /* for size_t */
#include <stdint.h>  /* for standard integer types */

/* check for CUDA compiler */
#if defined(__NVCC__)
   /* ensure CUDA definition exists */
   #undef CUDA
   #define CUDA
   /* dynamic library includes for MSVC */
   #ifdef _WIN32
      #pragma comment(lib, "cudart.lib")
      #pragma comment(lib, "nvml.lib")
   #endif
   /* library header includes */
   #include <cuda_runtime.h>
   #include <nvml.h>

/* end CUDA compilation */
/* #elif defined(__OPENCL_CPP_VERSION__) */
   /* currently NOT IMPLEMENTED */
   /* #define OPENCL_ENABLED */

/* end OPENCL compilation */
#endif

/**
 * Swap the byte order of a 32-bit word.
 * @details Alternate expansion for `nvcc`: `__byte_perm(x, 0, 0x0123)`
 * @param x 32-bit word to be swapped
 * @returns 32-bit word with swapped byte ordering.
*/
#ifdef __CUDA_ARCH__
   #define bswap32(x)   __byte_perm(x, 0, 0x0123)

#else
   #define bswap32(x) \
      ( (rol32(x, 24) & 0xFF00FF00) | (rol32(x, 8) & 0x00FF00FF) )

/* end #ifdef __CUDA_ARCH__... else... */
#endif

/**
 * Rotate a 32-bit word left by @a n bits.
 * @param x 32-bit word to rotate
 * @param n Number of bits to rotate left by
 * @returns 32-bit word rotated left by @a n bits.
*/
#define rol32(x, n)  ( ((x) << (n)) | ((x) >> (32 - (n))) )

/**
 * Rotate a 64-bit word left by @a n bits.
 * @param x 64-bit word to rotate
 * @param n Number of bits to rotate left by
 * @returns 64-bit word rotated left by @a n bits.
*/
#define rol64(x, n)  ( ((x) << (n)) | ((x) >> (64 - (n))) )

/**
 * Rotate a 32-bit word right by @a n bits.
 * @param x 32-bit word to rotate
 * @param n Number of bits to rotate right by
 * @returns 32-bit word rotated right by @a n bits.
*/
#define ror32(x, n)  ( ((x) >> (n)) | ((x) << (32 - (n))) )

/**
 * Rotate a 64-bit word right by @a n bits.
 * @param x 64-bit word to rotate
 * @param n Number of bits to rotate right by
 * @returns 64-bit word rotated right by @a n bits.
*/
#define ror64(x, n)  ( ((x) >> (n)) | ((x) << (64 - (n))) )

/**
 * Perform an XANDX operation. An XANDX operation is composed
 * of a XOR, an AND and another XOR in the form:
 * @code ( (((a) & ((b) ^ (c))) ^ (c) ) @endcode
 * @param a first parameter
 * @param b second parameter
 * @param c third parameter
 * @returns result of XANDX operation.
*/
#define xandx(a, b, c)  ( ((a) & ((b) ^ (c))) ^ (c) )

/**
 * XOR 3 values together. Performs 2x XOR operations.
 * @param a first parameter
 * @param b second parameter
 * @param c third parameter
 * @returns result of 2x XOR operations.
*/
#define xor3(a, b, c)   ( (a) ^ (b) ^ (c) )

/**
 * XOR 4 values together. Performs 3x XOR operations.
 * @param a first parameter
 * @param b second parameter
 * @param c third parameter
 * @param d fourth parameter
 * @returns result of 3x XOR operations.
*/
#define xor4(a, b, c, d)   ( (a) ^ (b) ^ (c) ^ (d) )

/* end include guard */
#endif
