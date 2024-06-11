#include "hash_table.h"
#include <string>

using namespace std;

extern "C" {
#include "hash_dict.h"
}

#define OK 0
#define Err 1

class hash_dict {
    int dict_add(string key);

    hash_table map;

public:
    int dict_add(string key, int val);
    int dict_remove(string key);
    int dict_find(string key);
    int dict_len();
    void dict_foreach(uintptr_t callback_h);
    int dict_randomval(const size_t n = 1);
};

int hash_dict::dict_add(string key, int val) {
    auto res = map.insert(key, val);
    return res == hashOk ? OK : Err;
}

void hash_dict::dict_foreach(uintptr_t callback_h) {
    for (hash_table_iterator it = map.begin(); it != map.end(); ++it) {
        goCallbackCharInt(callback_h, (char*)it.key().c_str(), it.val());
    }
}

int hash_dict::dict_remove(string key) {
    return map.remove(key);
}

int hash_dict::dict_find(string key) {
    hash_table_iterator it = map.find(key);
    if (it == map.end()) {
        return -1;
    } else {
        return it.val();
    }
}

int hash_dict::dict_len() {
    return map.getSize();
}

// 不知道怎么返回int或iter列表
// TODO: 支持查询n个random元素
int hash_dict::dict_randomval(const size_t n) {
    vector<hash_table_iterator> its = map.random(n);
    // 目前只返回一个int
    return its[0].val();
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

int DictLen(void* hd) {
    return static_cast<hash_dict*>(hd)->dict_len();
}

void DictForEach(void* hd, uintptr_t callback_h) {
    return static_cast<hash_dict*>(hd)->dict_foreach(callback_h);
}

int DictRandom(void* hd, const size_t n) {
    return static_cast<hash_dict*>(hd)->dict_randomval(n);
}