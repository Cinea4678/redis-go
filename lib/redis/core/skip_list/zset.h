#include <cstdint>
#include <limits>
#include <unordered_map>
#include <vector>

#include "skip_list.h"

using namespace std;

// 使用两个特殊的double值表示状态，避免额外状态位的设定
const double ZSetNotFound = numeric_limits<double>::min();
const double ZSetSuccess = numeric_limits<double>::max();

// 有序集合
class zset {

public:
    // 获取对应元素的score
    double getScore(ZSetType value) const {
        auto it = map.find(value);
        if (it != map.end()) {
            return it->second;
        }
        return ZSetNotFound;
    }

    // 添加元素，若已存在则返回value对应score
    double add(double score, ZSetType value);
    // 移除元素，若不存在则返回false
    bool remove(double score, ZSetType value);
    // 查找元素
    vector<ZSetType> search(double score);
    // 如果l大于r，则反向输出
    vector<pair<double, ZSetType>> searchRange(double lscore, double rscore);
    // 按值查找(本质还是按score查找)
    vector<ZSetType> searchValue(ZSetType value);
    // 按排名查找(负数代表倒数，同score按照插入先后排序)
    // 返回score，value对
    pair<double, ZSetType> searchRank(int rank);
    // 按排名范围查找(允许l<=0或r>length)
    // 返回score，value对的vector
    vector<pair<double, ZSetType>> searchRankRange(int lrank, int rrank);

private:
    // value->score哈希表
    unordered_map<ZSetType, double> map;

    // 跳跃表
    SkipList list;
};
