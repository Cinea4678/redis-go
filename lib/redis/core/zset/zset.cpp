#include <algorithm>
#include <limits>
#include <unordered_map>
#include <vector>

#include "skip_list.h"

using namespace std;

extern "C" {
#include "zset.h"
}

#define OK 0
#define Err 1

// 使用两个特殊的double值表示状态，避免额外状态位的设定
const double ZSetNotFound = numeric_limits<double>::min();
const double ZSetSuccess = numeric_limits<double>::max();

// 有序集合
class zset {

public:
    zset(){

    };
    ~zset(){

    };

    // 获取对应元素的score
    double getScore(ZSetType value) const {
        auto it = map.find(value);
        if (it != map.end()) {
            return it->second;
        }
        return ZSetNotFound;
    }

    int len() const { return map.size(); }

    // 添加元素，若已存在则返回value对应score
    double add(double score, ZSetType value);
    // 移除元素，返回对应value
    vector<ZSetType> remove(double score);
    // 按值移除，若不存在则返回false
    double remove(ZSetType value);
    // 查找元素
    vector<ZSetType> search(double score);
    // 如果l大于r，则反向输出
    vector<pair<double, ZSetType>> searchRange(double lscore, double rscore);
    // 按值查找(本质还是按score查找)
    vector<ZSetType> searchValue(ZSetType value);
    // 按排名查找(负数代表倒数，同score按照插入先后排序)
    // 返回score，value对
    // pair<double, ZSetType> searchRank(int rank);
    ZSetType searchRank(int rank);
    // 按排名范围查找(允许l<=0或r>length)
    // 返回score，value对的vector
    vector<pair<double, ZSetType>> searchRankRange(int lrank, int rrank);

private:
    // value->score哈希表
    unordered_map<ZSetType, double> map;

    // 跳跃表
    SkipList list;
};

double zset::add(double score, ZSetType value) {
    // value已存在则返回其对应score
    auto it = map.find(value);
    if (it != map.end()) {
        return it->second;
    }

    map[value] = score;
    list.insert(score, value);
    return ZSetSuccess;
}

vector<ZSetType> zset::remove(double score) {
    vector<ZSetType> result;
    result = list.remove(score);
    for (auto r : result) {
        auto it = map.find(r);
        // 理论上是都会找到的
        if (it != map.end()) {
            map.erase(it);
        }
    }
    return result;
}

double zset::remove(ZSetType value) {
    // value不存在则返回ZSetNotFound
    auto it = map.find(value);
    if (it == map.end()) {
        return ZSetNotFound;
    }

    bool result = list.remove(it->second, value);
    map.erase(it);
    // 返回对应的score
    return it->second;
}

vector<ZSetType> zset::search(double score) {
    auto it = map.find(score);
    if (it != map.end()) {
        return list.search(score);
    }
    return vector<ZSetType>(0);
}

vector<pair<double, ZSetType>> zset::searchRange(double lscore, double rscore) {
    vector<pair<double, ZSetType>> result = list.searchRange(lscore, rscore);
    // 如果l大于r，则反向输出
    if (lscore > rscore) {
        reverse(result.begin(), result.end());
    }
    return result;
}

vector<ZSetType> zset::searchValue(ZSetType value) {
    double s = getScore(value);
    if (s != ZSetNotFound) {
        return list.search(s);
    }
    return vector<ZSetType>(0);
};

// pair<double, ZSetType> zset::searchRank(int rank) {
//     auto result = list.searchRank(rank);
//     if (result.first == SkipListNotFound) {
//         return {ZSetNotFound, -1};
//     }
//     return result;
// }

ZSetType zset::searchRank(int rank) {
    auto result = list.searchRank(rank);
    if (result.first == SkipListNotFound) {
        return -1;
    }
    return result.second;
}

vector<pair<double, ZSetType>> zset::searchRankRange(int lrank, int rrank) {
    return searchRankRange(lrank, rrank);
}

void* NewZSet() {
    zset* zs = new zset();
    return static_cast<zset*>(zs);
}

int ReleaseZSet(void* zs) {
    delete static_cast<zset*>(zs);
    return OK;
}

int ZSetLen(void* zs) {
    return static_cast<zset*>(zs)->len();
}
double ZSetGetScore(void* zs, ZSetType value) {
    return static_cast<zset*>(zs)->getScore(value);
}

double ZSetAdd(void* zs, double score, ZSetType value) {
    cout << "Adding " << score << " " << value << endl;
    return static_cast<zset*>(zs)->add(score, value);
}

void* ZSetRemoveScore(void* zs, double score, int* length) {
    // 使用c风格
    // 新分配了一块内存，用于存储返回给go的数组（否则返回局部变量被清理掉到那边就是空的）
    // TODO: go那边用完需要调用C.free释放
    vector<ZSetType> v = static_cast<zset*>(zs)->remove(score);
    ZSetType* res = (ZSetType*)malloc(v.size() * sizeof(ZSetType));
    res = &v[0];
    *length = v.size();
    return static_cast<void*>(res);

    // 这里返回了临时变量数组的地址，内存管理可能会有问题
    // return static_cast<void*>(static_cast<zset*>(zs)->remove(score));
}

double ZSetRemoveValue(void* zs, ZSetType value) {
    return (static_cast<zset*>(zs)->remove(value));
}

void* ZSetSearch(void* zs, double score, int* length) {
    vector<ZSetType> v = static_cast<zset*>(zs)->search(score);
    ZSetType* res = (ZSetType*)malloc(v.size() * sizeof(ZSetType));
    res = &v[0];
    *length = v.size();
    return static_cast<void*>(res);
}

void* ZSetSearchRange(void* zs, double lscore, double rscore, int* length) {
    // 这里的pair对应go端的ZNode
    vector<pair<double, ZSetType>> v =
        static_cast<zset*>(zs)->searchRange(lscore, rscore);
    *length = v.size();
    ZSetType* res = (ZSetType*)malloc(*length * sizeof(ZSetType));

    for (int i = 0; i < *length; i++) {
        res[i] = v[i].second;
    }

    return static_cast<void*>(res);
}

ZSetType ZSetSearchRank(void* zs, int rank) {
    return (static_cast<zset*>(zs)->searchRank(rank));
}

void* ZSetSearchRankRange(void* zs, int lrank, int rrank, int* length) {
    vector<pair<double, ZSetType>> v =
        static_cast<zset*>(zs)->searchRankRange(lrank, rrank);
    pair<double, ZSetType>* res = (pair<double, ZSetType>*)malloc(
        v.size() * sizeof(pair<double, ZSetType>));
    res = &v[0];
    *length = v.size();
    return static_cast<void*>(res);
}

double ZSetNotFoundSign() {
    return ZSetNotFound;
}

double ZSetSuccessSign() {
    return ZSetSuccess;
}