#include <chrono>
#include <cmath>
#include <cstdint>
#include <cstdlib>
#include <ctime>
#include <iostream>
#include <list>
#include <queue>
#include <vector>

using namespace std;

extern "C" {
#include "skip_list.h"
}

#define SKIP_LIST_MAX_LEVEL 32 // 最大层数

// 前向声明
class SkipListNode;
class SkipListSiblingNode;
class SkipList;

// 其实有些检查(例如针对score、value的单点查询中)是否存在是不必要的(因为哈希表要存)

/*
旧设计方式：
设计方式：重复score的不同值使用newnode方式实现，而非将value换成vector形式
适用场景：适用于重复score的值较少的场景

不使用vector的原因：
vector对[每个节点]都会有额外24B占用(动态数组指针、大小、容量)，但可以减少一个forward指针开销，即额外开销len*16B；
而newnode方式，除了与vector同样的val占的8B以外，额外对[重复score的节点]需要(16)(+scr+*bwd)+(24+8*2)(vector(*fwd)的大小期望)=60B；
extra*60=len*16->除非重复score元素总数占总元素26.7%以上，否则newnode内存开销总比vector方式少；
此处仅从内存方面讨论，未考虑性能开销，如vector动态分配内存/newnode维护开销。

放弃原因：插入删除节点时，所需的操作过于繁琐，效率底下且不直观
*/

// 跳表节点

// 重复的节点不使用SkipListNode而是一个SkipListSiblingNode
// 重复的部分相当于就是一个单纯的双向链表
class SkipListNode {
    friend class SkipList;

public:
    uint32_t getValue() const { return value; }
    double getScore() const { return score; }

    inline void AddSibling(uint32_t value);

    // 调试用
    void print() const;

private:
    uint32_t value; // 值(代表go对象的实际元素索引)
    double score;   // 分数

    vector<SkipListNode*>
        forward; // 前向指针: forward[0]对应level 1的后继节点，以此类推

    // 后向指针: 指向直接前驱节点，但目前只是维护了其更新，还没用到
    // 可能的用途: 查询与某个val相同score的所有节点;用于优化范围查询等
    SkipListNode* backward;
    SkipListSiblingNode* nextSibling;

    SkipListNode(double scr, uint32_t val, uint32_t level)
        : score(scr), value(val), forward(level, nullptr), backward(nullptr),
          nextSibling(nullptr) {}

    ~SkipListNode() {}
};

// 重复score节点的实现
// extra*16=len*24->只要重复score不超过length的1.5倍，使用链表就是赚的
class SkipListSiblingNode {
    friend class SkipListNode;
    friend class SkipList;

private:
    uint32_t value; // 值(代表go对象的实际元素索引)
    SkipListSiblingNode* nextSibling;

    SkipListSiblingNode(uint32_t val) : value(val), nextSibling(nullptr) {}

    ~SkipListSiblingNode() {}
};

// 跳表
class SkipList {
public:
    // 头结点，不会实际使用所以value设置为0也没关系
    SkipList()
        : level(0), header(new SkipListNode(-INFINITY, 0, SKIP_LIST_MAX_LEVEL)),
          tail(nullptr) {
        if (!header) {
            throw runtime_error("Memory allocation for header failed.");
            delete header;
        }
    }

    // 循环释放
    ~SkipList() {
        SkipListNode* cur = header;
        while (cur) {
            // 直接后缀
            SkipListNode* next = cur->forward[0];

            // 删除重复score的所有元素
            SkipListSiblingNode* cur2 = cur->nextSibling;
            while (cur2) {
                SkipListSiblingNode* next2 = cur2->nextSibling;
                delete cur2;
                cur2 = next2;
            }

            // 删除当前节点
            delete cur;
            cur = next;
        }
    }

    SkipListNode* Header() const { return this->header; }
    SkipListNode* Tail() const { return this->tail; }
    uint32_t Level() const { return this->level; }
    uint32_t Len() const { return this->length; }

    void insert(double score, uint32_t value);

    // 此处不做按值查找，而是在zset的哈希表中存其score，将value查找置换为score
    // SkipListNode* find(uint32_t value);

    // 该部分操作函数返回值为节点指针
    //
    // 寻找对应score的首结点
    SkipListNode* searchNode(double score);
    //
    // 在上层调用中检查l与r的大小与范围关系，此处不检查
    // 寻找对应范围的首节点
    vector<SkipListNode*> searchRangeNode(double lscore, double rscore);
    //
    // 删除同一score的所有节点并返回删除的首节点
    SkipListNode* removeNode(double score);

    // 该部分操作函数返回值为数值(外部调用使用这一部分)
    //
    // 寻找对应score的所有值
    vector<uint32_t> search(double score);
    //
    // 在上层调用中检查l与r的大小与范围关系，此处不检查
    vector<pair<double, uint32_t>> searchRange(double lscore, double rscore);
    //
    // 删除同一score的所有节点并返回删除的值
    vector<uint32_t> remove(double score);
    // 删除指定val的节点并返回是否成功(score由hash表传入)
    // 检查是否存在value，以及返回对应score应该在上层处理
    bool remove(double score, uint32_t value);

    // 该部分函数为测试用
    // 打印跳表
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

inline void SkipListNode::AddSibling(uint32_t value) {
    SkipListSiblingNode* last = this->nextSibling;
    // 若目前暂无Sibling，则直接插入并更新首节点的nextSibling指针
    if (!last) {
        this->nextSibling = new SkipListSiblingNode(value);
        return;
    }
    // 如果有Sibling，则按照链表push_back操作
    while (last->nextSibling) {
        last = last->nextSibling;
    }
    last->nextSibling = new SkipListSiblingNode(value);
    return;
}

void SkipListNode::print() const {
    cout << "[SLNode] s=" << score << ", v=" << value << endl;

    // 打印重复score的节点
    if (nextSibling) {
        SkipListSiblingNode* s = nextSibling;
        cout << "[";
        while (s) {
            cout << ", v=" << s->value;
            s = s->nextSibling;
        }
        cout << "]" << endl;
    }

    // for (int i = 1; i < SKIP_LIST_MAX_LEVEL; i++) {
    for (int i = 0; i < forward.size(); i++) {
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
    vector<SkipListNode*> update(max(level, newNodeLevel), nullptr);
    SkipListNode* cur = header;
    // 从最高层开始，查找插入位置
    // 渐进式的查找，下一层查找起点在上一层终点
    // 注意遍历查找要从level开始才能效率较大，但在小于等于newNodeLevel部分才会更新
    for (int i = level - 1; i >= 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            cur = cur->forward[i];
        }
        if (i <= newNodeLevel) {
            update[i] = cur;
        }
    }
    // 经过上面的循环后，cur抵达小于score的最后一个Node(其第一层后继节点大于等于score)

    // 如果有重复score
    if (cur->forward[0] && cur->forward[0]->score == score) {
        // 如果判定有相等的，则先抵达score对应的Node
        cur = cur->forward[0];
        cur->AddSibling(value);
        return;
    } // 直接return，无需更新其他forward

    // 如果无重复score

    // 如果新节点层数大于目前最大节点层数，则多出来的部分指向header
    if (newNodeLevel > level) {
        for (uint32_t i = level; i < newNodeLevel; i++) {
            update[i] = header;
        }
        level = newNodeLevel;
    }

    SkipListNode* newNode = new SkipListNode(score, value, newNodeLevel);
    // 将新节点插入到前驱与后继之间，其实就是链表插入操作

    for (uint32_t i = 0; i < newNodeLevel; i++) {
        newNode->forward[i] = update[i]->forward[i];
        update[i]->forward[i] = newNode;
    }

    // 设置后向指针
    newNode->backward = update[0];

    // 设置后继节点的后向指针
    if (newNode->forward[0]) {
        newNode->forward[0]->backward = newNode;
    } else {
        // 如果后继为null，则设置尾节点
        tail = newNode;
    }

    length++;
}

SkipListNode* SkipList::searchNode(double score) {
    SkipListNode* cur = header;

    for (int i = level - 1; i >= 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            cur = cur->forward[i];
        }
    }
    cur = cur->forward[0];

    // Score not found
    if (!cur || cur->score != score) {
        return nullptr;
    }
    // Score found
    return cur;
}

vector<SkipListNode*> SkipList::searchRangeNode(double lscore, double rscore) {
    vector<SkipListNode*> result;
    SkipListNode* cur = header;
    for (int i = level - 1; i >= 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < lscore) {
            cur = cur->forward[i];
        }
    }
    // cur抵达小于lscore的最后一个节点

    // 从>=lscore的第一个节点开始，按顺序找直系后继
    while (cur->forward[0] && cur->forward[0]->score < rscore) {
        cur = cur->forward[0];
        result.push_back(cur);
    }

    return result;
}

SkipListNode* SkipList::removeNode(double score) {
    vector<SkipListNode*> update(level, nullptr);
    SkipListNode* cur = header;

    // 寻找所有层的前驱节点，应该从1开始，而不是0
    for (int i = level; i > 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            cur = cur->forward[i];
        }
        update[i] = cur;
    }

    // 寻找同score的链表的起始节点
    cur = cur->forward[0];

    // 检查是否有这个score
    if (!cur || cur->score != score) {
        return nullptr;
    }

    // 循环删除所有兄弟节点
    SkipListSiblingNode* last = cur->nextSibling;
    while (last) {
        SkipListSiblingNode* nxt = last->nextSibling;
        delete last;
        last = nxt;
    }

    // 从第一层往上找（因为一定是较低层有前驱节点而较高层无
    // 此处是用判定update[i]->forward[i]是否为cur的方式来确认cur是否为某个节点的后继节点的
    for (int i = 0; i < level; i++) {
        if (update[i]->forward[i] != cur)
            break;
        update[i]->forward[i] = cur->forward[i];
    }

    delete cur;

    // 更新backward和tail指针
    // 如果对第一层而言，删除节点为最后一个节点，则更新tail
    if (update[0]->forward[0] == nullptr) {
        tail = update[0];
    }
    // 如果删除前存在直接后继，则更新其的backward指针
    if (update[0]->forward[0]) {
        update[0]->forward[0]->backward = update[1];
    }

    // 更新跳表层数
    while (level > 1 && header->forward[level] == nullptr) {
        level--;
    }

    return cur;
}

vector<uint32_t> SkipList::search(double score) {
    vector<uint32_t> result;
    SkipListNode* cur = searchNode(score);
    if (!cur) { // not found
        return result;
    }

    // 遍历
    result.push_back(cur->value);
    SkipListSiblingNode* last = cur->nextSibling;
    while (last) {
        result.push_back(last->value);
        last = last->nextSibling;
    }
    return result;
}

vector<pair<double, uint32_t>> SkipList::searchRange(double lscore,
                                                     double rscore) {
    vector<pair<double, uint32_t>> result;
    vector<SkipListNode*> r = searchRangeNode(lscore, rscore);
    if (r.empty()) { // not found
        return result;
    }

    // 对每一个节点遍历
    for (SkipListNode* cur : r) {
        result.push_back({cur->score, cur->value});
        SkipListSiblingNode* last = cur->nextSibling;
        while (last) {
            result.push_back({cur->score, last->value});
            last = last->nextSibling;
        }
    }
    return result;
}

vector<uint32_t> SkipList::remove(double score) {
    vector<uint32_t> result;
    SkipListNode* cur = removeNode(score);
    if (!cur) { // not found
        return result;
    }

    // 遍历
    result.push_back(cur->value);
    SkipListSiblingNode* last = cur->nextSibling;
    while (last) {
        result.push_back(last->value);
        last = last->nextSibling;
    }
    return result;
}

bool SkipList::remove(double score, uint32_t value) {
    vector<SkipListNode*> update(level, nullptr);
    SkipListNode* cur = header;

    // 寻找所有层的前驱节点，应该从1开始，而不是0
    for (int i = level; i > 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            cur = cur->forward[i];
        }
        update[i] = cur;
    }

    // 寻找同score的链表的起始节点
    cur = cur->forward[0];

    // 检查是否有这个score
    if (!cur || cur->score != score) {
        return false;
    }

    // 如果首节点就是对应value，则需要判断其是否有兄弟节点
    //      如果有兄弟节点，则仅需将firstSibling晋升为Node(对cur节点的value、nextSibling进行赋值并删除firstSibling)
    //      如果无，则直接删除并连接
    // 如果首节点非对应value，则遍历Sibling后进行简单的链表删除(理论上在上层判断，此处不会有找不到的情况，不过还是处理了)

    // 如果首节点就是对应value
    if (cur->value == value) {
        SkipListSiblingNode* firstSibling = cur->nextSibling;
        if (firstSibling) { // 有兄弟节点
            SkipListNode* result = cur;
            cur->value = value;
            cur->nextSibling = firstSibling->nextSibling;

            delete firstSibling;
            return true;
        }
        // 无兄弟节点，则与只传入score的重载的操作一致
        for (int i = 0; i < level; i++) {
            if (update[i]->forward[i] != cur)
                break;
            update[i]->forward[i] = cur->forward[i];
        }
        delete cur;
        if (update[0]->forward[0] == nullptr) {
            tail = update[0];
        }
        if (update[0]->forward[0]) {
            update[0]->forward[0]->backward = update[1];
        }
        while (level > 1 && header->forward[level] == nullptr) {
            level--;
        }
        return true;
    }
    // 首节点非对应value

    // 遍历所有兄弟节点
    SkipListSiblingNode* sibling = cur->nextSibling;

    // 如果第一个兄弟节点就是对应value
    if (sibling->value == value) {
        cur->nextSibling = sibling->nextSibling;
        delete sibling;
    } else {
        SkipListSiblingNode* next = sibling->nextSibling;
        while (next) {
            if (next->value == value) {
                sibling->nextSibling = next->nextSibling;
                delete next;
            }
        }
    }
    return false;
}

void SkipList::print() const {
    cout << "{Skip List}-------------------------------" << endl;
    for (int i = 0; i < level; i++) {
        printLevel(i);
    }
    cout << "{Skip List}-------------------------------" << endl;
}

void SkipList::printLevel(uint32_t lvl) const {
    cout << "[Skip List] level: " << lvl << endl;
    SkipListNode* cur = header->forward[lvl];

    while (cur) {
        cout << "(" << cur->score << " " << cur->value;
        SkipListSiblingNode* sibling = cur->nextSibling;
        while (sibling) {
            cout << ", " << sibling->value;
            sibling = sibling->nextSibling;
        }
        cout << ")->";
        cur = cur->forward[lvl];
    }
    cout << "null" << endl;
}

// int main() {
//     const int testCount = int(1e6);
//     srand(514); // 初始化随机种子

//     chrono::time_point<chrono::system_clock> start, end;
//     chrono::duration<double, std::milli> elapsed;
//     SkipList list;
//     cout << "Finish new list" << endl;

//     // Test: insert
//     start = chrono::system_clock::now();
//     for (int i = 0; i < testCount; i++) {
//         list.insert(rand() % testCount * 0.1, i);
//     }
//     end = chrono::system_clock::now();
//     elapsed = end - start;
//     cout << "SkipList insert time: " << elapsed.count() << " ms" << endl;

//     // std::priority_queue<pair<double, uint32_t>> stdList;
//     // start = chrono::system_clock::now();
//     // for (int i = 0; i < testCount; i++) {
//     //     stdList.push({rand() % testCount * 0.1, i});
//     // }
//     // end = chrono::system_clock::now();
//     // elapsed = end - start;
//     // cout << "SkipList insert time: " << elapsed.count() << " ms" << endl;

//     // Test: search
//     start = chrono::system_clock::now();
//     for (int i = 0; i < testCount; i++) {
//         list.search(rand() % testCount * 0.1);
//     }
//     end = chrono::system_clock::now();
//     elapsed = end - start;
//     cout << "SkipList search time: " << elapsed.count() << " ms" << endl;

//     // for (auto i : list.searchRange(20, 30))
//     // cout << i.first << " " << i.second << endl;
//     // list.print();

//     // TODO: 补充查找和删除的测试

//     return 0;
// }