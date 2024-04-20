#include "skip_list.cpp"

#include <chrono>
#include <iostream>

int main() {
    const int testCount = int(1e3);
    srand(514); // 初始化随机种子

    chrono::time_point<chrono::system_clock> start, end;
    chrono::duration<double, std::milli> elapsed;
    SkipList list;
    cout << "Finish new list" << endl;

    // Test: insert
    start = chrono::system_clock::now();
    for (int i = 0; i < testCount; i++) {
        list.insert(rand() % testCount * 0.1, i);
        // list.print();
    }
    // list.print();
    end = chrono::system_clock::now();
    elapsed = end - start;
    cout << "SkipList insert time: " << elapsed.count() << " ms" << endl;

    // std::priority_queue<pair<double, ZSetType>> stdList;
    // start = chrono::system_clock::now();
    // for (int i = 0; i < testCount; i++) {
    //     stdList.push({rand() % testCount * 0.1, i});
    // }
    // end = chrono::system_clock::now();
    // elapsed = end - start;
    // cout << "SkipList insert time: " << elapsed.count() << " ms" << endl;

    // Test: search
    start = chrono::system_clock::now();
    for (int i = 0; i < testCount; i++) {
        list.search(rand() % testCount * 0.1);
    }
    end = chrono::system_clock::now();
    elapsed = end - start;
    cout << "SkipList search time: " << elapsed.count() << " ms" << endl;

    // for (auto i : list.searchRange(20, 30))
    // cout << i.first << " " << i.second << endl;
    // list.print();

    // TODO: 补充查找和删除的测试

    return 0;
}