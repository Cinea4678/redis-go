#include <stdint.h>
#include <stdlib.h>
typedef uint32_t ZSetType;
extern const double ZSetNotFound;
extern const double ZSetSuccess;

struct ZNode {
    ZSetType value;
    double score;
};

void* NewZSet();

int ReleaseZSet(void* zs);

int ZSetLen(void* zs);

double ZSetGetScore(void* zs, ZSetType value);

double ZSetAdd(void* zs, double score, ZSetType value);

void* ZSetRemoveScore(void* zs, double score, int* length);

double ZSetRemoveValue(void* zs, ZSetType value);

void* ZSetSearch(void* zs, double score, int* length);

void* ZSetSearchRange(void* zs, double lscore, double rscore, int* length);

ZSetType ZSetSearchRank(void* zs, int rank);

void* ZSetSearchRankRange(void* zs, int lrank, int rrank, int* length);

double ZSetNotFoundSign();
double ZSetSuccessSign();