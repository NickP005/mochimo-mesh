
/* must be declared before includes */
#ifndef CUDA
   #define CUDA
#endif

#include <stdint.h>
#include "_assert.h"
#include "../sha3.h"

#define NUMVECTORS    7
#define MAXVECTORLEN  81
#define DIGESTLEN     KECCAKLEN256

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
      0xc5, 0xd2, 0x46, 0x01, 0x86, 0xf7, 0x23, 0x3c, 0x92, 0x7e, 0x7d,
      0xb2, 0xdc, 0xc7, 0x03, 0xc0, 0xe5, 0x00, 0xb6, 0x53, 0xca, 0x82,
      0x27, 0x3b, 0x7b, 0xfa, 0xd8, 0x04, 0x5d, 0x85, 0xa4, 0x70
   }, {
      0x3a, 0xc2, 0x25, 0x16, 0x8d, 0xf5, 0x42, 0x12, 0xa2, 0x5c, 0x1c,
      0x01, 0xfd, 0x35, 0xbe, 0xbf, 0xea, 0x40, 0x8f, 0xda, 0xc2, 0xe3,
      0x1d, 0xdd, 0x6f, 0x80, 0xa4, 0xbb, 0xf9, 0xa5, 0xf1, 0xcb
   }, {
      0x4e, 0x03, 0x65, 0x7a, 0xea, 0x45, 0xa9, 0x4f, 0xc7, 0xd4, 0x7b,
      0xa8, 0x26, 0xc8, 0xd6, 0x67, 0xc0, 0xd1, 0xe6, 0xe3, 0x3a, 0x64,
      0xa0, 0x36, 0xec, 0x44, 0xf5, 0x8f, 0xa1, 0x2d, 0x6c, 0x45
   }, {
      0x85, 0x6a, 0xb8, 0xa3, 0xad, 0x0f, 0x61, 0x68, 0xa4, 0xd0, 0xba,
      0x8d, 0x77, 0x48, 0x72, 0x43, 0xf3, 0x65, 0x5d, 0xb6, 0xfc, 0x5b,
      0x0e, 0x16, 0x69, 0xbc, 0x05, 0xb1, 0x28, 0x7e, 0x01, 0x47
   }, {
      0x92, 0x30, 0x17, 0x5b, 0x13, 0x98, 0x1d, 0xa1, 0x4d, 0x2f, 0x33,
      0x34, 0xf3, 0x21, 0xeb, 0x78, 0xfa, 0x04, 0x73, 0x13, 0x3f, 0x6d,
      0xa3, 0xde, 0x89, 0x6f, 0xeb, 0x22, 0xfb, 0x25, 0x89, 0x36
   }, {
      0x6e, 0x61, 0xc0, 0x13, 0xae, 0xf4, 0xc6, 0x76, 0x53, 0x89, 0xff,
      0xcd, 0x40, 0x6d, 0xd7, 0x2e, 0x7e, 0x06, 0x19, 0x91, 0xf4, 0xa3,
      0xa8, 0x01, 0x81, 0x90, 0xdb, 0x86, 0xbd, 0x21, 0xeb, 0xb4
   }, {
      0x15, 0x23, 0xa0, 0xcd, 0x0e, 0x7e, 0x1f, 0xaa, 0xba, 0x17, 0xe1,
      0xc1, 0x22, 0x10, 0xfa, 0xbc, 0x49, 0xfa, 0x99, 0xa7, 0xab, 0xc0,
      0x61, 0xe3, 0xd6, 0xc9, 0x78, 0xee, 0xf4, 0xf7, 0x48, 0xc4
   }
};

int main()
{  /* check 256-bit keccak() digest results match expected */
   size_t size_digest;
   size_t inlen[NUMVECTORS];
   uint8_t digest[NUMVECTORS][DIGESTLEN];
   int j;

   /* calc sizes */
   size_digest = sizeof(digest);

   /* init memory (synchronous) */
   memset(digest, 0, size_digest);

   for (j = 0; j < NUMVECTORS; j++) {
      inlen[j] = strlen(rfc_1321_vectors[j]);
   }

   /* perform bulk hash */
   test_kcu_keccak(rfc_1321_vectors, inlen, MAXVECTORLEN,
      digest, DIGESTLEN, NUMVECTORS);

   /* analyze results */
   for (j = 0; j < NUMVECTORS; j++) {
      ASSERT_CMP(digest[j], expect[j], DIGESTLEN);
   }
}
