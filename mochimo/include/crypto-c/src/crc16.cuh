/**
 * @file crc16.cuh
 * @brief CRC16 CUDA hash function support.
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
#ifndef CRYPTO_CRC16_CUH
#define CRYPTO_CRC16_CUH


#include "crc16.h"

__device__ uint16_t cu_crc16(void *in, size_t inlen);

/* end include guard */
#endif
