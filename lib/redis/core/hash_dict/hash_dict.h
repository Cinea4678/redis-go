void* NewHashDict();

int ReleaseHashDict(void* hd);

int DictAdd(void* hd, const char* key, int val);

int DictRemove(void* hd, const char* key);

int DictFind(void* hd, const char* key);

int DictLen(void* hd);
