#include <unordered_map>
#include <string>
using namespace std;

extern "C" {
#include "hash_dict.h"
}

#define OK 0
#define Err 1

class hash_dict {
    unordered_map<string, int> map;

public:
    int dict_add(string key, int val);
    int dict_remove(string key);
    int dict_find(string key);
};

int hash_dict::dict_add(string key, int val)
{
    auto res = map.insert({ key, val });
    return res.second == true ? OK : Err;
}

int hash_dict::dict_remove(string key)
{
    auto res = map.find(key);
    if (res != map.end()) {
        auto val = res->second;
        map.erase(res);
        return val;
    }
    else {
        return -1;
    }
}

int hash_dict::dict_find(string key)
{
    auto iter = map.find(key);
    if (iter == map.end()) {
        return -1;
    }
    else {
        return iter->second;
    }
}

void* NewHashDict() {
    auto hd = new hash_dict();
    return static_cast<hash_dict*>(hd);
}

int ReleaseHashDict(void* hd) {
    delete static_cast<hash_dict*>(hd);
    return OK;
}


int DictAdd(void* hd, const char* key, int val) {
    return static_cast<hash_dict*>(hd)->dict_add(key, val);
}

int DictRemove(void* hd, const char* key) {
    return static_cast<hash_dict*>(hd)->dict_remove(key);
}

int DictFind(void* hd, const char* key) {
    return static_cast<hash_dict*>(hd)->dict_find(key);
}
