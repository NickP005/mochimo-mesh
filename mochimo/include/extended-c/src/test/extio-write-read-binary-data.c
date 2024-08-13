
#include "_assert.h"
#include "../extio.h"

#define FNAME  "binary.data"

int main()
{  /* check; writes binary data, reads binary data, and return code */
   unsigned char write[256] = { 0 };
   unsigned char read[256] = { 0 };
   for (int i = 0; i < 256; i++) write[i] = i;
   ASSERT_EQ(write_data(write, 256, FNAME), 256);
   ASSERT_EQ(read_data(read, 256, FNAME), 256);
   ASSERT_CMP(read, write, 256);
   remove(FNAME);
}
