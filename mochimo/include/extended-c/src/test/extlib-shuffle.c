
#include "_assert.h"
#include "../extlib.h"

#define LISTLENGTH   10

int main()
{  /* check; shuffle on list length < 2, and normal list */
   long shuffled[LISTLENGTH] = { 1, 5, 3, 6, 7, 0, 2, 4, 0, 8 };
   long nochange[LISTLENGTH] = { 1, 2, 3, 4, 5, 6, 7, 8, 0, };
   long list[LISTLENGTH] = { 1, 2, 3, 4, 5, 6, 7, 8, 0, };

   shuffle(list, sizeof(list[0]), 1);
   ASSERT_CMP(list, nochange, sizeof(nochange));
   shuffle(list, sizeof(list[0]), LISTLENGTH);
   ASSERT_CMP(list, shuffled, sizeof(shuffled));
}
