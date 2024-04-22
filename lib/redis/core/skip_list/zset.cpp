#include "zset.h"
#include <algorithm>
#include <vector>

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

bool zset::remove(double score, ZSetType value) {
    // value不存在则返回false
    auto it = map.find(value);
    if (it == map.end()) {
        return false;
    }

    map.erase(it);
    return true;
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

pair<double, ZSetType> zset::searchRank(int rank) {
    auto result = list.searchRank(rank);
    if (result.first == SkipListNotFound) {
        return {ZSetNotFound, 0};
    }
    return result;
}

vector<pair<double, ZSetType>> zset::searchRankRange(int lrank, int rrank) {
    return searchRankRange(lrank, rrank);
}