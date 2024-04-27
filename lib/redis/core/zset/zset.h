extern const double ZSetNotFound;
extern const double ZSetSuccess;

void* NewZSet();

int ReleaseZSet(void* zs);

int ZSetLen(void* zs);

double ZSetGetScore(void* zs, int value);

double ZSetAdd(void* zs, double score, int value);

void* ZSetRemove(void* zs, double score);

int ZSetRemoveByValue(void* hd, double score);

double ZSetSearch(void* zs, double lscore, double rscore, int* count);

double ZSetSearchRange(void* zs, double lscore, double rscore, int* count);

double ZSetNotFoundSign();
double ZSetSuccessSign();