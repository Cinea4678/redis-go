void* NewHashDict();

int ReleaseHashDict(int hd);

int DictAdd(void* hd, const char* key, int val);

int DictRemove(void* hd, const char* key);

int DictFind(void* hd, const char* key);
