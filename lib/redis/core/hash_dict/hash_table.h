// #include "MurmurHash3.h"
#include <cstdint>
#include <cstdlib>
#include <functional>
#include <iostream>
#include <optional>
#include <random>
#include <string>

using namespace std;

// 前向声明
class hash_entry;
class hash_table;
class hash_table_iterator;

// // Murmurhash随机种子
// const uint64_t dict_hash_func_seed = 42;

// 哈希表默认大小
const uint64_t default_ht_size = 4;

// rehash阈值(load_factor)
const static float expand_threshold = 0.8;
const static float shrink_threshold = 0.2;

enum {
    hashOk = 0,
    hashErr = -1,
    hashAllocateErr = -2,
};

// 哈希表节点
// 占48字节内存(32+8+8)
class hash_entry {
    friend class hash_table;
    friend class hash_table_iterator;

public:
    // 实际以迭代器调用
    inline string getkey() const { return key; };
    inline int getval() const { return val; };
    inline hash_entry* getnext() const { return next; };

private:
    // 键值对，key为string类型，val为int类型的内存索引(?)
    string key;
    int val;

    // 指向下个哈希表节点，形成链表
    hash_entry* next;

    // 不包含next指针的构造函数
    hash_entry(const string& key, const int& val)
        : key(key), val(val), next(nullptr){};

    // 包含next指针的构造函数
    hash_entry(const string& key, const int& val, hash_entry* next)
        : key(key), val(val), next(next){};

    ~hash_entry(){};
};

// 哈希表

class hash_table {
    friend class hash_table_iterator;

public:
    static inline size_t hashFunction(string key) {
        // murmurhash在测试中性能不佳，冲突较多（测试1~100整数键值），故采用标准库
        static hash<std::string> hash_fn;
        return hash_fn(key);
    }

    // 分配内存
    hash_table(const unsigned long size = default_ht_size)
        : size(size), sizemask(size - 1), used(0) {
        try {
            table = size > 0 ? new hash_entry*[size]() : nullptr;
            for (unsigned long i = 0; i < size; ++i)
                table[i] = nullptr;
        } catch (const std::bad_alloc& e) {
            std::cerr << "Memory allocation failed during hash_table init: "
                      << e.what() << endl;
            delete[] table; // 确保释放分配失败前的内存
        }
    }
    ~hash_table() { delete[] table; }

    // 负载因子
    inline float load_factor() const {
        return size > 0 ? static_cast<float>(used) / size : 0.0f;
    }

    /* 查找对应键值对应hash_table_iterator
       返回值：键值是否存在?对应迭代器:end */
    hash_table_iterator find(const string& key);

    /* 查找对应键值对应val，返回值以传输引用方式获得
       返回值：键值是否存在?hashOk:hashErr */
    int findval(const string& key, int& val);

    /* 插入键值对，并判断是否需要expand
       返回值：插入是否成功 */
    int insert(const string& key, const int& val);

    /* 插入键值对，并判断是否需要shrink
       返回值：删除key对应的val */
    int remove(const string& key);

    /* 随机返回n个指向哈希表条目的迭代器
       返回值：指向随机条目的迭代器数组 */
    vector<hash_table_iterator> random(size_t n);

    // 清空哈希表（不重置为初始大小）
    void clear();

    // rehash为指定大小
    void rehash(const unsigned long newSize);

    // 打印哈希表
    void print() const;

    // 起始位置迭代器
    hash_table_iterator begin() const;

    // 尾后迭代器
    hash_table_iterator end() const;

    int getSize() const { return this->size; };

    bool isEmpty() const { return this->used == 0; }

private:
    // 哈希表数组，存指向哈希表第一排节点的指针
    hash_entry** table;

    // 哈希表大小
    unsigned long size;

    // 哈希表大小掩码(size - 1)，用于计算索引值(使用按位与sizemask代替取余)
    unsigned long sizemask;

    // 该哈希表已有节点的数量
    unsigned long used;
};

// hash_table_iterator迭代器
// 不安全迭代器，可以直接删除当前迭代器指向的节点
class hash_table_iterator {
public:
    // 初始化构造函数，用于构造begin或end迭代器
    hash_table_iterator(const hash_table* ht, bool end = false)
        : ht(ht), bucket(0), entry(nullptr) {
        if (!end) {
            // 如果不是创建一个尾后迭代器，则初始化到第一个有效元素
            advanceToFirst();
        } else {
            // 创建一个尾后迭代器
            bucket = ht->size;
        }
    }
    // 构造函数，用于对指定hash_entry进行迭代
    hash_table_iterator(const hash_table* ht, unsigned long bucket,
                        hash_entry* entry)
        : ht(ht), bucket(bucket), entry(entry) {}

    // 复制构造函数
    hash_table_iterator(const hash_table_iterator& clone)
        : ht(clone.ht), bucket(clone.bucket), entry(clone.entry) {}

    // 前缀自增，移动到下一个元素
    hash_table_iterator& operator++() {
        advance();
        return *this;
    }

    // 对应entry中的方法
    inline const string key() { return this->entry->key; };
    inline const int val() { return this->entry->val; };
    inline hash_table_iterator next() {
        hash_table_iterator nxt(*this);
        nxt.advance();
        return nxt;
    };

    // 解引用迭代器，返回当前的哈希表节点
    const hash_entry& operator*() const { return *entry; }

    // 通过指针访问
    const hash_entry* operator->() const { return entry; }

    // 比较两个迭代器是否相等
    bool operator==(const hash_table_iterator& other) const {
        return ht == other.ht && bucket == other.bucket && entry == other.entry;
    }

    // 比较两个迭代器是否不相等
    bool operator!=(const hash_table_iterator& other) const {
        return !(*this == other);
    }

    // // 删除当前节点
    // 直接调用找到的entry的remove即可
    // bool erase() {
    //     if (!entry)
    //         return false; // 如果当前节点为空，则不执行删除

    //     hash_entry* toDelete = entry;
    //     // 预先移动到下一个节点
    //     advance();

    //     // 执行删除操作
    //     // 如果不使用prevEntry则需要重新遍历过
    //     if (prevEntry) {
    //         // 如果不是第一个节点
    //         prevEntry->next = entry->next;
    //     } else {
    //         // 如果是第一个节点
    //         ht->table[bucket] = entry->next;
    //     }

    //     delete toDelete; // 删除节点
    //     ht->used--;      // 更新已使用节点的数量

    //     return true;
    // }

private:
    const hash_table* ht;
    unsigned long bucket; // 当前桶的索引
    hash_entry* entry;    // 当前节点的指针

    // 移动到下一个有效元素
    void advance() {
        // 如果为有效节点，则直接到下一个
        if (entry) {
            entry = entry->next;
        }

        // next为空，到下一个桶的开头找
        // 注意这里是sizemask，不然会超索引
        // while (!entry && bucket < ht->size) {
        while (!entry && bucket < ht->sizemask) {
            bucket++;
            entry = ht->table[bucket];
        }

        // 检查是否到达了最后一个节点，或者当前已经是尾后迭代器
        if (!entry && bucket >= ht->sizemask) {
            // 设置为尾后迭代器
            bucket = ht->size;
            entry = nullptr;
            return;
        }
    }

    // 初始化到第一个有效元素
    void advanceToFirst() {
        while (bucket < ht->size && !ht->table[bucket]) {
            ++bucket;
        }
        if (bucket < ht->size) {
            entry = ht->table[bucket];
        }
    }
};
