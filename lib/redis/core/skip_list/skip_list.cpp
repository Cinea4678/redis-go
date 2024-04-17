#include <chrono>
#include <cmath>
#include <cstdint>
#include <cstdlib>
#include <ctime>
#include <iostream>
#include <vector>

using namespace std;

extern "C" {
#include "skip_list.h"
}

#define SKIP_LIST_MAX_LEVEL 32 // 最大层数

/*
跳表节点设计方式：同一score的不同值使用newnode方式实现，而非将value换成vector形式
适用场景：适用于同score的值较少的场景
原因：
vector对[每个节点]都会有额外24B占用(动态数组指针、大小、容量)，但可以减少一个forward指针开销，即额外开销len*16B；
而newnode方式，除了与vector同样的val占的8B以外，额外对[重复score的节点]需要(16)(+scr+*bwd)+(24+8*2)(vector(*fwd)的大小期望)=60B；
extra*60=len*16->除非重复score元素总数占总元素26.7%以上，否则newnode内存开销总比vector方式少；
此处仅从内存方面讨论，未考虑性能开销，如vector动态分配内存/newnode维护开销。

可能可以考虑的方式；
实现一个siblingNode，对于重复score的元素，不维护大部分指针和分数，仅维护值+forward[0]
*/

// 跳表节点
class SkipListNode {
    friend class SkipList;

public:
    void print() const;

private:
    uint32_t value; // 值(代表go对象的实际元素索引)
    double score;   // 分数

    // 与redis3.0不同，有level而非level-1个forward指针
    // 为了实现同一score的不同节点
    vector<SkipListNode*>
        forward; // 前向指针:
                 // forward[0]指向同一score的后继节点，其它的才指向其它score

    // 后向指针: 指向直接前驱节点，但目前只是维护了其更新，还没用到
    // 可能的用途: 查询与某个val相同score的所有节点;用于优化范围查询等
    SkipListNode* backward;

    SkipListNode(double scr, uint32_t val, uint32_t level)
        : score(scr), value(val), forward(level + 1, nullptr),
          backward(nullptr) {} // 注意forward的size是level+1（痛苦的debug）

    ~SkipListNode() {}
};

// 跳表
class SkipList {
public:
    // 头结点，不会实际使用所以value设置为0也没关系
    SkipList()
        : level(0),
          header(new SkipListNode(-INFINITY, 0, SKIP_LIST_MAX_LEVEL + 1)),
          tail(nullptr) {
        if (!header) {
            throw runtime_error("Memory allocation for header failed.");
        }
    }

    // 循环释放
    ~SkipList() {
        SkipListNode* cur = header;
        while (cur) {
            SkipListNode* next = cur->forward[1];

            // 类似的，删除相同score的所有元素
            SkipListNode* cur2 = cur->forward[0];
            while (cur2) {
                SkipListNode* next2 = cur2->forward[0];
                cout << "del " << cur2->value << endl;
                delete cur2;
                cur2 = next2;
            }
            cout << "del " << cur->value << endl;
            delete cur;
            cur = next;
        }
    }

    SkipListNode* Header() const { return this->header; }
    SkipListNode* Tail() const { return this->tail; }
    uint32_t Level() const { return this->level; }
    uint32_t Len() const { return this->length; }

    // 插入randomLevel计算的随机层数
    void insert(double score, uint32_t value);

    // 这里不做按值查找，而是在zset的哈希表中存其score，将value查找置换为score
    // 按值寻找(是否重复在哈希表中判定，此处不做约束)
    // SkipListNode* find(uint32_t value);

    // 按score寻找(由于score可以重复，所以可能找到多个)
    vector<SkipListNode*> search(double score);

    // 在上层调用中检查l与r的大小与范围关系，此处不检查
    vector<SkipListNode*> searchRange(double lscore, double rscore);

    // 删除同一score的所有节点并返回删除的点
    vector<SkipListNode*> remove(double score);

    // 删除指定val的所有节点并返回删除的点
    vector<SkipListNode*> del(double score);

    // 打印跳表，测试用
    void print() const;
    // 打印指定层数跳表
    void printLevel(uint32_t lvl) const;

private:
    // 随机生成层数(算法有待改善)
    uint32_t randomLevel() const;

    SkipListNode* header; // 头节点
    SkipListNode* tail;   // 尾节点
    uint32_t level;       // 当前最高层数
    uint32_t length;      // 跳表长度
};

void SkipListNode::print() const {
    cout << "[SLNode] s=" << score << ", v=" << value << endl;
    if (forward[0]) {
        cout << "forward[0]: " << forward[0]->score << ", ";
    }
    // for (int i = 1; i < SKIP_LIST_MAX_LEVEL; i++) {
    for (int i = 1; i < forward.size() - 1; i++) {
        // 如果越界了，检查可能不为空，但根本不是forward节点
        if (!forward[i]) {
            break;
        }
        cout << "forward[" << i << "]: " << forward[i]->value << " "
             << forward[i]->score << ", ";
    }
    if (backward) {
        cout << "backward: " << backward->score;
    }
    cout << endl;
}

// 随机生成节点层数
// 50%概率加层
// 少量高层节点快速跳过大部分低层节点
uint32_t SkipList::randomLevel() const {
    int lvl = 1;
    while ((rand() & 1) && lvl < SKIP_LIST_MAX_LEVEL) {
        lvl++;
    }
    return lvl;
}

void SkipList::insert(double score, uint32_t value) {
    // 新节点层数
    const uint32_t newNodeLevel = randomLevel();

    // cout << "Inserting [" << score << "] " << value << ", l=" << newNodeLevel
    //      << endl;

    // 存储新节点对应的前驱节点
    vector<SkipListNode*> update(max(level, newNodeLevel) + 1, nullptr);
    SkipListNode* cur = header;

    // 从最高层开始，查找插入位置
    // 渐进式的查找，下一层查找起点在上一层终点
    // 注意遍历查找要从level开始才能效率较大，但在小于等于newNodeLevel部分才会更新
    for (uint32_t i = level; i > 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            cur = cur->forward[i];
        }
        if (i <= newNodeLevel) {
            update[i] = cur;
        }
    }

    // 经过上面的循环后，cur抵达小于score的最后一个节点(其第一层后继节点大于等于score)
    // forward[0]要另外判断，而不能直接在上面score判断取等
    if (cur->forward[1] && cur->forward[1]->score == score) {
        // 如果判定有相等的，则抵达score对应的节点
        cur = cur->forward[1];
        // 找到当前score对应顶节点
        while (cur->forward[0]) {
            cur = cur->forward[0];
        }
        update[0] = cur;
    }

    // 如果新节点层数大于目前最大节点层数，则多出来的部分指向header
    // 对首次插入也适用
    if (newNodeLevel > level) {
        for (uint32_t i = level + 1; i <= newNodeLevel; i++) {
            update[i] = header;
        }
        level = newNodeLevel;
    }

    SkipListNode* newNode = new SkipListNode(score, value, newNodeLevel);
    // 将新节点插入到前驱与后继之间，其实就是链表插入操作

    // 如果score已存在，则不更新其他点的指针，而只是复制一份前面节点的指针
    if (update[0]) {
        // update[0]是该score的顶节点
        update[0]->forward[0] = newNode;
        // for (uint32_t i = 1; i <= newNodeLevel; i++) { //
        // 注意，考虑newNode比update[0]层数高的情况
        for (uint32_t i = 1; i < update[0]->forward.size(); i++) {
            newNode->forward[i] = update[0]->forward[i];
        }
        // newnode高出update[0]的部分
        for (uint32_t i = update[0]->forward.size(); i <= newNodeLevel; i++) {
            newNode->forward[i] = update[i]->forward[i];
        }

        newNode->backward = update[0]->backward;

        length++;

        return;
    }

    // 如果score不存在
    for (uint32_t i = 1; i <= newNodeLevel; i++) {
        newNode->forward[i] = update[i]->forward[i];
        update[i]->forward[i] = newNode;
    }

    // 设置后向指针

    newNode->backward = update[1];

    // 设置后继节点的后向指针
    if (newNode->forward[1]) {
        newNode->forward[1]->backward = newNode;
    } else {
        // 如果后继为null，则设置尾节点
        tail = newNode;
    }

    length++;
}

vector<SkipListNode*> SkipList::search(double score) {
    vector<SkipListNode*> result;
    SkipListNode* cur = header;
    for (uint32_t i = level; i > 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            cur = cur->forward[i];
        }
    }
    cur = cur->forward[1];

    // Score not found
    if (!cur || cur->score != score) {
        return result;
    }

    // Score found
    while (cur) {
        result.push_back(cur);
        cur = cur->forward[0];
    }

    return result;
}

vector<SkipListNode*> SkipList::searchRange(double lscore, double rscore) {
    vector<SkipListNode*> result;
    SkipListNode* cur = header;
    for (uint32_t i = level; i > 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < lscore) {
            cur = cur->forward[i];
        }
    }
    // cur抵达小于lscore的最后一个节点

    // 从>=lscore的第一个节点开始，按顺序找直系后继
    while (cur->forward[1] && cur->forward[1]->score < rscore) {
        cur = cur->forward[1];
        // 遍历当前score所有并列节点
        while (cur) {
            result.push_back(cur);
            cur = cur->forward[0];
        }
    }

    return result;
}

vector<SkipListNode*> SkipList::remove(double score) {
    vector<SkipListNode*> removedNodes;
    vector<SkipListNode*> update(level + 1, nullptr);
    SkipListNode* cur = header;

    // 寻找所有层的前驱节点，应该从1开始，而不是0
    for (uint32_t i = level; i > 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            cur = cur->forward[i];
        }
        update[i] = cur;
    }

    // 寻找同score的链表的起始节点
    cur = cur->forward[1];

    // 检查是否有这个score
    if (!cur || cur->score != score) {
        return removedNodes;
    }

    // 逐个删除
    while (cur) {

        // 从第一层往上找（因为第一层一定有前驱节点，而往上则不一定
        // 此处是用判定update[i]->forward[i]是否为cur的方式来确认cur是否为某个节点的后继节点的
        for (uint32_t i = 1; i <= level; i++) {

            if (update[i]->forward[i] != cur)
                break;
            update[i]->forward[i] = cur->forward[i];
        }

        removedNodes.push_back(cur);
        SkipListNode* next = cur->forward[0];
        delete cur;
        cur = next;
    }

    // 如果有节点被删除，更新backward和tail指针
    if (!removedNodes.empty()) {
        // 如果对第一层而言，删除节点为最后一个节点，则更新tail
        if (update[1]->forward[1] == nullptr) {
            tail = update[1];
        }

        // for (uint32_t i = 1; i <= level; i++) {
        //     if (update[i]->forward[i]) {
        //         update[i]->forward[i]->backward = update[i];
        //     }
        // }

        // 更新删除前直接后继的backward指针
        if (update[1]->forward[1]) {
            update[1]->forward[1]->backward = update[1];
        }

        // 更新跳表层数
        while (level > 1 && header->forward[level] == nullptr) {
            level--;
        }
    }

    return removedNodes;
}

void SkipList::print() const {
    cout << "{Skip List}-------------------------------" << endl;
    for (uint32_t i = 1; i <= level; i++) {
        printLevel(i);
    }
    cout << "{Skip List}-------------------------------" << endl;
}

void SkipList::printLevel(uint32_t lvl) const {
    cout << "[Skip List] level: " << lvl << endl;
    SkipListNode* cur = header->forward[lvl];

    while (cur) {
        cout << "(" << cur->value << " " << cur->score;
        while (cur->forward[0]) {
            cur = cur->forward[0];
            cout << ", " << cur->value << " " << cur->score << endl;
        }
        cout << ")->";
        cur = cur->forward[lvl];
    }
    cout << "null" << endl;
}

// int main() {
//     srand(514); // 初始化随机种子

//     SkipList list;
//     cout << "Finish new list" << endl;
//     // 记录开始时间
//     auto start = chrono::high_resolution_clock::now();

//     // 插入元素
//     for (int i = 0; i < 100; i++) {
//         list.insert(rand() % 500 * 0.1, i + 1);
//         // list.print();
//     }
//     list.print();

//     // for (int i = 0; i < 100; i++) {
//     //     auto r = list.remove(rand() % 500 * 0.1);
//     //     if (r.size() != 0) {
//     //         for (int j = 0; j < r.size(); j++) {
//     //             cout << "remove node: " << endl;
//     //             r[j]->print();
//     //         }
//     //     }
//     //     // list.print();
//     // }

//     // 搜索元素
//     auto result = list.search(8);
//     cout << "Search results for score 8.0: " << result.size()
//          << " entries found." << endl;

//     for (int i = 0; i < 100; i++) {
//         auto r = list.search(rand() % 500 * 0.1);
//         if (r.size() != 0) {
//             for (int j = 0; j < r.size(); j++) {
//                 cout << "search node: " << endl;
//                 r[j]->print();
//             }
//         }
//         // list.print();
//     }

//     // TODO：补充测试删除
//     //  删除元素
//     auto removed = list.remove(8);
//     auto r = list.remove(rand() % 500 * 0.1);
//     if (r.size() != 0) {
//         for (int j = 0; j < r.size(); j++) {
//             cout << "remove node: " << endl;
//             r[j]->print();
//         }
//     }

//     // 打印跳表
//     // list.print();

//     // 记录结束时间并计算运行时间
//     auto end = chrono::high_resolution_clock::now();
//     chrono::duration<double, std::milli> elapsed = end - start;
//     cout << "Elapsed time: " << elapsed.count() << " ms" << endl;

//     return 0;
// }