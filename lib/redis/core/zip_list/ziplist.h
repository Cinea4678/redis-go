#include <stdint.h>

#define ZiplistHandle void*
#define ZiplistNodeHandle void*

ZiplistHandle NewZiplist();

void ZiplistPush();

void ReleaseZiplist(ZiplistHandle handle);

int ZiplistPushBytes(ZiplistHandle handle, char *bytes, int len);

int ZiplistPushInteger(ZiplistHandle handle, int64_t integer);

int ZiplistInsertBytes(ZiplistHandle handle, int pos, char *bytes, int len);

int ZiplistInsertInteger(ZiplistHandle handle, int pos, int64_t integer);

ZiplistNodeHandle ZiplistIndex(ZiplistHandle handle, int index);

ZiplistNodeHandle ZiplistFindBytes(ZiplistHandle handle, char* bytes, int len);

ZiplistNodeHandle ZiplistFindInteger(ZiplistHandle handle, int64_t integer);

ZiplistNodeHandle ZiplistNext(ZiplistHandle handle, ZiplistNodeHandle currentNode);

ZiplistNodeHandle ZiplistPrev(ZiplistHandle handle, ZiplistNodeHandle currentNode);

int64_t ZiplistGetInteger(ZiplistNodeHandle nodeHandle);

void ZiplistGetByteArray(ZiplistNodeHandle nodeHandle, uint8_t **array, int *len);

int ZiplistDelete(ZiplistNodeHandle nodeHandle);

int ZiplistDeleteRange(ZiplistHandle handle, ZiplistNodeHandle startNodeHandle, int len);

int ZiplistDeleteByPos(ZiplistHandle handle, size_t pos);

int ZiplistBlobLen(ZiplistHandle handle);

int ZiplistLen(ZiplistHandle handle);