extern void goCallbackCharInt(uintptr_t h, char* p1, int p2);

void* NewHashDict();

int ReleaseHashDict(void* hd);

int DictAdd(void* hd, const char* key, int val);

int DictRemove(void* hd, const char* key);

int DictFind(void* hd, const char* key);

int DictLen(void* hd);

void DictForEach(void* hd, uintptr_t callback_h);

int DictRandom(void* hd, const size_t n = 1);