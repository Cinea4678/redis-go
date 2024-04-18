#include <cmath>
#include <cstdint>
#include <cstdlib>
#include <iostream>
#include <vector>

using namespace std;

#define SKIP_LIST_MAX_LEVEL 32 // 最大层数

// 前向声明
class SkipListNode;
class SkipListSiblingNode;
class SkipList;

typedef uint32_t ZSetType;

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
// redis的设计: 将重复节点当做普通节点，同score间按照字典序排序
class SkipListNode {
    friend class SkipList;

public:
    ZSetType getValue() const { return value; }
    double getScore() const { return score; }

    inline void AddSibling(ZSetType value);

    // 调试用
    void print() const;

private:
    ZSetType value; // 值(代表go对象的实际元素索引)
    double score;   // 分数

    vector<SkipListNode*>
        forward; // 前向指针: forward[0]对应level 1的后继节点，以此类推

    // 后向指针: 指向直接前驱节点，但目前只是维护了其更新，还没用到
    // 可能的用途: 查询与某个val相同score的所有节点;用于优化范围查询等
    SkipListNode* backward;
    SkipListSiblingNode* nextSibling;

    SkipListNode(double scr, ZSetType val, ZSetType level)
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
    ZSetType value; // 值(代表go对象的实际元素索引)
    SkipListSiblingNode* nextSibling;

    SkipListSiblingNode(ZSetType val) : value(val), nextSibling(nullptr) {}

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
    ZSetType Level() const { return this->level; }
    ZSetType Len() const { return this->length; }

    void insert(double score, ZSetType value);

    // 此处不做按值查找，而是在zset的哈希表中存其score，将value查找置换为score
    // SkipListNode* find(ZSetType value);

    // 该部分操作函数返回值为数值(外部调用使用这一部分)
    //
    // 单点查询，寻找对应score的所有值
    vector<ZSetType> search(double score);
    // 范围查询，在上层调用中检查l与r的大小与范围关系，此处不检查
    vector<pair<double, ZSetType>> searchRange(double lscore, double rscore);
    //
    // 删除同一score的所有节点并返回删除的值
    vector<ZSetType> remove(double score);
    // 删除指定val的节点并返回是否成功(score由hash表传入)
    // 检查是否存在value，以及返回对应score应该在上层处理
    bool remove(double score, ZSetType value);

    // 该部分函数为测试用
    // 打印跳表
    void print() const;
    // 打印指定层数跳表
    void printLevel(ZSetType lvl) const;

private:
    // 随机生成层数(算法有待改善)
    ZSetType randomLevel() const;

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

    SkipListNode* header; // 头节点
    SkipListNode* tail;   // 尾节点
    ZSetType level;       // 当前最高层数
    ZSetType length;      // 跳表长度
};