/**
 * @file crc32.cuh
 * @brief CRC32 CUDA hash function support.
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
#ifndef CRYPTO_CRC32_CUH
#define CRYPTO_CRC32_CUH


#include "crc32.h"

__device__ uint32_t cu_crc32(void *in, size_t inlen);

/* end include guard */
#endif
