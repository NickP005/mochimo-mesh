
#include <time.h>
#include <string.h>

#include "../extint.h"
#include "../extlib.h"
#include "../extprint.h"

static char Metric[9][3] = { "", "K", "M", "G", "T", "P", "E", "Z", "Y" };

static inline word32 rotl(const word32 x, int k) {
	return (x << k) | (x >> (32 - k));
}


static word32 s[2] = { 0xd15ea5e5u, 0xcafef00du };

word32 next(void) {
	const word32 s0 = s[0];
	word32 s1 = s[1];
	const word32 result = s0 * 0x9E3779BB;

	s1 ^= s0;
	s[0] = rotl(s0, 26) ^ s1 ^ (s1 << 9); // a, b
	s[1] = rotl(s1, 13); // c

	return result;
}

static word64 mcg_state = 0xcafef00dd15ea5e5u;	// Must be odd
static word64 const multiplier = 6364136223846793005u;
word32 pcg32_fast(void)
{
	word64 x = mcg_state;
	unsigned count = (unsigned)(x >> 61);	// 61 = 64 - 3

	mcg_state = x * multiplier;
	x ^= x >> 22;
	return (word32)(x >> (22 + count));	// 22 = 32 - 3 - 7
}

void pcg32_fast_init(word64 seed)
{
	mcg_state = 2*seed + 1;
	(void)pcg32_fast();
}
int main()
{
   size_t Metriclen = sizeof(Metric) / sizeof(Metric[0]);
   clock_t startclock;
   double seconds, perf;
   word64 m, n, nmax;
   word32 r;

   nmax = 1;
   seconds = 0.0;

   plog("");
   plog("Analyzing raw performance of next...");
   do {
      nmax <<= 1;
      startclock = clock();
      for (n = 0; n < nmax; n++) {
         r = next();
      }
      seconds = (double) (clock() - startclock) / CLOCKS_PER_SEC;
      perf = (double) n / seconds;
   } while (seconds < 1.0);
   plog("Completed next(%u) (x%" WORDu64 ") in %lf seconds", r, nmax, seconds);
   for (m = 0; perf > 999 && m < Metriclen; m++) perf /= 1000;
   plog("   Relative performance= %lf %s/s", perf, Metric[m]);
}
