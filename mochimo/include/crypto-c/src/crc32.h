/**
 * @file crc32.h
 * @brief CRC32 hash function support.
 * @details
 * Param  | Value
 * ------ | -----
 * Alias  | "CRC-32"
 * Check  | 0xCBF43926 (crc32 of "123456789")
 * Poly   | 0x04C11DB7
 * Init   | 0xffffffff
 * RefIn  | true
 * RefOut | true
 * XorOut | 0xffffffff
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_CRC32_H
#define CRYPTO_CRC32_H


#include "utildev.h"

#define CRC32LEN  4  /**< CRC32 message digest length, in bytes */

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

uint32_t crc32(void *in, size_t inlen);

/* CUDA testing functions */
#ifdef CUDA
   void test_kcu_crc32(const void *in, size_t *inlen, size_t max_inlen,
      uint32_t *ret, int num);
#endif

/* end extern "C" {} for C++ */
#ifdef __cplusplus
}
#endif

/* end include guard */
#endif
