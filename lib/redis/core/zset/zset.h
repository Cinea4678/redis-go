#include <stdint.h>
typedef uint32_t ZSetType;
extern const double ZSetNotFound;
extern const double ZSetSuccess;

void* NewZSet();

int ReleaseZSet(void* zs);

int ZSetLen(void* zs);

double ZSetGetScore(void* zs, ZSetType value);

double ZSetAdd(void* zs, double score, ZSetType value);

void* ZSetRemoveScore(void* zs, double score);

double ZSetRemoveValue(void* zs, ZSetType value);

double ZSetSearch(void* zs, double lscore, double rscore, int* count);

double ZSetSearchRange(void* zs, double lscore, double rscore, int* count);

double ZSetNotFoundSign();
double ZSetSuccessSign();