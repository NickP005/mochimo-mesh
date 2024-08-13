
#include <stdint.h>
#include "_assert.h"
#include "../sha3.h"

#define NUMVECTORS    7
#define MAXVECTORLEN  81
#define DIGESTLEN     KECCAKLEN512

/* Test vectors used in RFC 1321 */
static char rfc_1321_vectors[NUMVECTORS][MAXVECTORLEN] = {
   "",
   "a",
   "abc",
   "message digest",
   "abcdefghijklmnopqrstuvwxyz",
   "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789",
   "1234567890123456789012345678901234567890123456789012345678901234"
   "5678901234567890"
};

/* expected results to test vectors */
static uint8_t expect[NUMVECTORS][DIGESTLEN] = {
   {
      0x0e, 0xab, 0x42, 0xde, 0x4c, 0x3c, 0xeb, 0x92, 0x35, 0xfc, 0x91,
      0xac, 0xff, 0xe7, 0x46, 0xb2, 0x9c, 0x29, 0xa8, 0xc3, 0x66, 0xb7,
      0xc6, 0x0e, 0x4e, 0x67, 0xc4, 0x66, 0xf3, 0x6a, 0x43, 0x04, 0xc0,
      0x0f, 0xa9, 0xca, 0xf9, 0xd8, 0x79, 0x76, 0xba, 0x46, 0x9b, 0xcb,
      0xe0, 0x67, 0x13, 0xb4, 0x35, 0xf0, 0x91, 0xef, 0x27, 0x69, 0xfb,
      0x16, 0x0c, 0xda, 0xb3, 0x3d, 0x36, 0x70, 0x68, 0x0e
   }, {
      0x9c, 0x46, 0xdb, 0xec, 0x5d, 0x03, 0xf7, 0x43, 0x52, 0xcc, 0x4a,
      0x4d, 0xa3, 0x54, 0xb4, 0xe9, 0x79, 0x68, 0x87, 0xee, 0xb6, 0x6a,
      0xc2, 0x92, 0x61, 0x76, 0x92, 0xe7, 0x65, 0xdb, 0xe4, 0x00, 0x35,
      0x25, 0x59, 0xb1, 0x62, 0x29, 0xf9, 0x7b, 0x27, 0x61, 0x4b, 0x51,
      0xdb, 0xfb, 0xbb, 0x14, 0x61, 0x3f, 0x2c, 0x10, 0x35, 0x04, 0x35,
      0xa8, 0xfe, 0xaf, 0x53, 0xf7, 0x3b, 0xa0, 0x1c, 0x7c
   }, {
      0x18, 0x58, 0x7d, 0xc2, 0xea, 0x10, 0x6b, 0x9a, 0x15, 0x63, 0xe3,
      0x2b, 0x33, 0x12, 0x42, 0x1c, 0xa1, 0x64, 0xc7, 0xf1, 0xf0, 0x7b,
      0xc9, 0x22, 0xa9, 0xc8, 0x3d, 0x77, 0xce, 0xa3, 0xa1, 0xe5, 0xd0,
      0xc6, 0x99, 0x10, 0x73, 0x90, 0x25, 0x37, 0x2d, 0xc1, 0x4a, 0xc9,
      0x64, 0x26, 0x29, 0x37, 0x95, 0x40, 0xc1, 0x7e, 0x2a, 0x65, 0xb1,
      0x9d, 0x77, 0xaa, 0x51, 0x1a, 0x9d, 0x00, 0xbb, 0x96
   }, {
      0xcc, 0xcc, 0x49, 0xfa, 0x63, 0x82, 0x2b, 0x00, 0x00, 0x4c, 0xf6,
      0xc8, 0x89, 0xb2, 0x8a, 0x03, 0x54, 0x40, 0xff, 0xb3, 0xef, 0x50,
      0xe7, 0x90, 0x59, 0x99, 0x35, 0x51, 0x8e, 0x2a, 0xef, 0xb0, 0xe2,
      0xf1, 0x83, 0x91, 0x70, 0x79, 0x7f, 0x77, 0x63, 0xa5, 0xc4, 0x3b,
      0x2d, 0xcf, 0x02, 0xab, 0xf5, 0x79, 0x95, 0x0e, 0x36, 0x35, 0x8d,
      0x6d, 0x04, 0xdf, 0xdd, 0xc2, 0xab, 0xac, 0x75, 0x45
   }, {
      0xe5, 0x5b, 0xdc, 0xa6, 0x4d, 0xfe, 0x33, 0xf3, 0x6a, 0xe3, 0x15,
      0x3c, 0x72, 0x78, 0x33, 0xf9, 0x94, 0x7d, 0x92, 0x95, 0x80, 0x73,
      0xf4, 0xdd, 0x02, 0xe3, 0x8a, 0x82, 0xd8, 0xac, 0xb2, 0x82, 0xb1,
      0xee, 0x13, 0x30, 0xa6, 0x82, 0x52, 0xa5, 0x4c, 0x6d, 0x3d, 0x27,
      0x30, 0x65, 0x08, 0xca, 0x76, 0x5a, 0xcd, 0x45, 0x60, 0x6c, 0xae,
      0xaf, 0x51, 0xd6, 0xbd, 0xc4, 0x59, 0xf5, 0x51, 0xf1
   }, {
      0xd5, 0xfa, 0x6b, 0x93, 0xd5, 0x4a, 0x87, 0xbb, 0xde, 0x52, 0xdb,
      0xb4, 0x4d, 0xaf, 0x96, 0xa3, 0x45, 0x5d, 0xae, 0xf9, 0xd6, 0x0c,
      0xdb, 0x92, 0x2b, 0xc4, 0xb7, 0x2a, 0x5b, 0xbb, 0xa9, 0x7c, 0x5b,
      0xf8, 0xc5, 0x98, 0x16, 0xfe, 0xde, 0x30, 0x2f, 0xc6, 0x4e, 0x98,
      0xce, 0x1b, 0x86, 0x4d, 0xf7, 0xbe, 0x67, 0x1c, 0x96, 0x8e, 0x43,
      0xd1, 0xba, 0xe2, 0x3a, 0xd7, 0x6a, 0x3e, 0x70, 0x2d
   }, {
      0xbc, 0x08, 0xa9, 0xa2, 0x45, 0xe9, 0x9f, 0x62, 0x75, 0x31, 0x66,
      0xa3, 0x22, 0x6e, 0x87, 0x48, 0x96, 0xde, 0x09, 0x14, 0x56, 0x5b,
      0xee, 0x0f, 0x8b, 0xe2, 0x9d, 0x67, 0x8e, 0x0d, 0xa6, 0x6c, 0x50,
      0x8c, 0xc9, 0x94, 0x8e, 0x8a, 0xd7, 0xbe, 0x78, 0xea, 0xa4, 0xed,
      0xce, 0xd4, 0x82, 0x25, 0x3f, 0x8a, 0xb2, 0xe6, 0x76, 0x8c, 0x9c,
      0x8f, 0x2a, 0x2f, 0x0a, 0xff, 0xf0, 0x83, 0xd5, 0x1c
    }
};

int main()
{  /* check 512-bit keccak() digest results match expected */
   size_t inlen;
   uint8_t digest[DIGESTLEN];
   int j;

   for (j = 0; j < NUMVECTORS; j++) {
      inlen = strlen(rfc_1321_vectors[j]);
      keccak(rfc_1321_vectors[j], inlen, digest, DIGESTLEN);
      ASSERT_CMP(digest, expect[j], DIGESTLEN);
   }
}
