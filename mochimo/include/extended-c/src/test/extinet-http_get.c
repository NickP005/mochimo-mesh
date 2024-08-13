
#include "_assert.h"
#include "../extinet.h"
#include <stdio.h>

int main()
{  /* check; GET request with NULL/bogus url, with/without fname specified */
   ASSERT_GT(http_get(NULL, NULL, 10), 0);
   ASSERT_GT(http_get("", NULL, 10), 0);
   ASSERT_GT(http_get("https://invalid@domain.name", NULL, 10), 0);
   ASSERT_EQ(http_get("https://www.mochimap.net/startnodes.lst", NULL, 10), 0);
   ASSERT_EQ(remove("startnodes.lst"), 0);
   ASSERT_EQ(http_get("example.com", "example.html", 10), 0);
   ASSERT_EQ(remove("example.html"), 0);
}
