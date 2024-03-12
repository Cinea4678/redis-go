#define IntsetHandle void*

void* NewIntset();

void ReleaseIntset(IntsetHandle handle);

int IntsetAdd(IntsetHandle handle, long long val);

int IntsetRemove(IntsetHandle handle, long long val);

int IntsetFind(IntsetHandle handle, long long val);

long long IntsetRandom(IntsetHandle handle);

long long IntsetGet(IntsetHandle handle, int index);

int IntsetLen(IntsetHandle handle);

int IntsetBlobLen(IntsetHandle handle);
