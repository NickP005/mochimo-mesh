/**
 * @file crc16.h
 * @brief CRC16 hash function support.
 * @details
 * Param  | Value
 * ------ | -----
 * Alias  | "XMODEM", "ZMODEM", "CRC-16/ACORN"
 * Check  | 0x31c3 (crc16 of "123456789")
 * Poly   | 0x1021
 * Init   | 0x0000
 * RefIn  | false
 * RefOut | false
 * XorOut | 0x0000
 * @copyright This file is released into the Public Domain under
 * the Creative Commons Zero v1.0 Universal license.
*/

/* include guard */
#ifndef CRYPTO_CRC16_H
#define CRYPTO_CRC16_H


#include "utildev.h"

#define CRC16LEN  2  /**< 16-bit CRC16 digest length in bytes */

/* C/C++ compatible function prototypes */
#ifdef __cplusplus
extern "C" {
#endif

uint16_t crc16(void *in, size_t inlen);

/* CUDA testing functions */
#ifdef CUDA
   void test_kcu_crc16(const void *in, size_t *inlen, size_t max_inlen,
      uint16_t *ret, int num);
#endif

/* end extern "C" {} for C++ */
#ifdef __cplusplus
}
#endif

/* end include guard */
#endif
