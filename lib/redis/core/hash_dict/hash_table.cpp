#include "hash_table.h"

hash_table_iterator hash_table::find(const string& key) {
    if (size == 0) // 哈希表为空
        return nullptr;

    unsigned long index = hashFunction(key) & sizemask; // 计算哈希表索引

    hash_entry* entry = table[index];
    while (entry != nullptr) {
        if (entry->key == key) {
            return hash_table_iterator(this, index, entry);
        }
        entry = entry->next; // 找到匹配的键，返回对应的条目
    }

    return end(); // 未找到，返回尾迭代器
}

int hash_table::findval(const string& key, int& val) {
    if (size == 0) // 哈希表为空
        return hashErr;

    unsigned long index = hashFunction(key) & sizemask; // 计算哈希表索引

    hash_entry* entry = table[index];
    while (entry != nullptr) {
        if (entry->key == key) {
            val = entry->val;
            return hashOk;
        }
        entry = entry->next; // 找到匹配的键，返回对应的条目
    }

    return hashErr; // 未找到，返回空指针
}

// 不安全
int hash_table::insert(const string& key, const int& val) {
    size_t hash = hashFunction(key);
    unsigned long index = hash & sizemask;

    // 检查是否已存在key
    hash_table_iterator it = find(key);
    if (it != end()) {
        // TODO: 可能需要新的返回格式?(扩充状态码类型or自定义Response结构体)
        cerr << "Hash_table insert failed: The key " << key
             << " is already in hash table,its value is " << it.val() << endl;
        return hashErr;
    }

    // XXX:
    // 因为更新在hash_dict.go内定义了根据find的insertOrUpdate，所以并不需要检查并做更新？
    hash_entry* new_entry = new hash_entry(key, val, table[index]);

    try {
        hash_entry* new_entry = new hash_entry(key, val, table[index]);
        table[index] = new_entry;
        used++;

        // 负载因子大于阈值，哈希表大小expand为2倍并rehash
        if (load_factor() > expand_threshold) {
            rehash(size * 2);
        }
        return hashOk;
    } catch (const std::bad_alloc& e) {
        std::cerr << "[HashTable] Memory allocation failed during insert: "
                  << e.what() << endl;
        delete new_entry; // 确保释放分配失败前的内存
        return hashErr;
    }
}

int hash_table::remove(const string& key) {
    unsigned long hash = hashFunction(key);
    unsigned long index = hash & sizemask;

    hash_entry* prevEntry = nullptr;
    hash_entry* entry = table[index];

    // 遍历链表
    while (entry != nullptr) {
        if (entry->key == key) {
            int val = entry->val;
            if (prevEntry == nullptr) {
                // 要删除的键位于链表头部
                table[index] = entry->next;
            } else {
                // 要删除的键位于链表中间或尾部
                prevEntry->next = entry->next; // 跳过当前条目
                entry->next = nullptr; // 断开当前条目与链表的连接
            }
            delete entry;
            used--;
            return val;
        }
        prevEntry = entry;
        entry = entry->next;
    }

    // 负载因子小于阈值，并且大小大于2*default，哈希表大小shrink为一半并rehash
    if (load_factor() > expand_threshold && size / 2 >= default_ht_size) {
        rehash(size / 2);
        return hashOk;
    }
    return hashErr; // 未找到要删除的键
}

void hash_table::clear() {
    for (unsigned long i = 0; i < size; ++i) {
        hash_entry* entry = table[i];
        while (entry != nullptr) {
            hash_entry* temp = entry;
            entry = entry->next;
            delete temp;
        }
        table[i] = nullptr;
    }
    used = 0;

    /*
    // XXX: 重新分配到初始大小，似乎不需要
    // delete []table;
    // this->table = new hash_entry *[default_ht_size];
    // this->size = default_ht_size;
    // this->sizemask = default_ht_size - 1;
    */
}

// 随机返回一个迭代器指向哈希表中的随机条目
vector<hash_table_iterator> hash_table::random(size_t n) {
    vector<hash_table_iterator> result;

    if (n <= 0) {
        cerr << "[HashTable] Invalid number in getting random elements" << endl;
        return result;
    }

    if (isEmpty()) {
        cerr << "[HashTable] Getting elements in empty hash table" << endl;
        return result;
    }

    // 随机数生成器初始化
    random_device rd;
    mt19937 gen(rd());
    vector<int> randomIndexes(n);
    uniform_int_distribution<> distrib(0, sizemask);

    // 生成n个随机数并排序
    for (int& index : randomIndexes) {
        index = distrib(gen);
    }
    sort(randomIndexes.begin(), randomIndexes.end());

    hash_table_iterator it = begin();
    int cur = 0;
    for (int target : randomIndexes) {
        // 移动迭代器到下一个目标位置
        for (int i = 0; i < target - cur; ++i) {
            it = it.next();
        }

        cur = target;
        // 添加迭代器到结果
        if (it != end()) {
            result.push_back(it);
        } else {
            // 如果迭代器到达尾后，停止添加
            break;
        }
    }
    return result;
}

// TODO: 目前暂不考虑分步式rehash
void hash_table::rehash(const unsigned long newSize) {
    hash_entry** newTable = nullptr;
    try {
        newTable = new hash_entry*[newSize]();
        unsigned long newSizemask = newSize - 1;

        // 遍历旧哈希表，重新哈希每个元素到新哈希表
        for (unsigned long i = 0; i < size; ++i) {
            hash_entry* entry = table[i];
            while (entry != nullptr) {
                hash_entry* nextEntry = entry->next;
                // 重新计算哈希值的索引
                unsigned long newIndex = hashFunction(entry->key) & newSizemask;

                // 将元素插入新哈希表
                entry->next = newTable[newIndex];
                newTable[newIndex] = entry;

                entry = nextEntry;
            }
        }

        // 释放旧哈希表内存
        delete[] table;

        // 更新哈希表属性为新哈希表
        table = newTable;
        size = newSize;
        sizemask = newSizemask;
    } catch (const std::bad_alloc& e) {
        std::cerr << "[HashTable] Memory allocation failed during rehash: "
                  << e.what() << endl;
        // 确保释放分配失败前的内存
        delete[] newTable;
    }
}

// 调试用输出
void hash_table::print() const {
    // return;
    cout << "Hash Table:" << endl;
    cout << "used:" << used << endl;
    for (unsigned long i = 0; i < size; ++i) {
        hash_entry* entry = table[i];
        if (entry == nullptr) {
            cout << "Bucket " << i << ": empty" << endl;
        } else {
            cout << "Bucket " << i << ":";
            while (entry != nullptr) {
                cout << " (" << entry->key << ", " << entry->val << ")";
                entry = entry->next;
            }
            cout << endl;
        }
    }
    cout << endl;
}

hash_table_iterator hash_table::begin() const {
    for (unsigned long i = 0; i < size; ++i) {
        if (table[i] != nullptr) {
            return hash_table_iterator(this, i, table[i]);
        }
    }
    // 如果哈希表为空，则返回尾后迭代器
    return end();
}

hash_table_iterator hash_table::end() const {
    return hash_table_iterator(this, true);
}

/*
1e7循环测试结果：
ht insert start time: 0
ht insert end time: 12061
cost time: 12061

map insert start time: 12069
map insert end time: 27349
cost time: 15280

ht find start time: 27359
ht find end time: 34559
cost time: 7200

map find start time: 34569
map find end time: 43127
cost time: 8558


*/

// #include <cassert>
// #include <iostream>
// #include <time.h>
// #include <unordered_map>

// using namespace std;

// int main() {
//   hash_table ht(8);
//   unordered_map<string, int> map;

//   // // 测试插入
//   // ht.insert("apple", 1);
//   // ht.insert("banana", 2);
//   // ht.insert("orange", 3);

//   // ht.print();
//   // 测试查找
//   //   hash_table_iterator e = ht.find("apple");
//   //   assert(e != nullptr && e->getval() == 1);

//   // // 测试删除
//   // int status = ht.remove("apple");
//   // ht.print();
//   // assert(status == 1);
//   // cout << (ht.find("apple") == ht.end()) << endl;

//   // // 测试清空
//   // ht.clear();
//   // ht.print();
//   // cout << (ht.find("banana") == ht.end()) << endl;

//   // 测试插入与rehash

//   // cout << "Load factor before rehash: " << ht.load_factor() << endl;

//   const int test_num = 1e7;

//   clock_t start = clock();
//   cout << "ht insert start time: " << start << endl;

//   for (int i = 0; i < test_num; i++) {
//     ht.insert(to_string(i), i);
//     // ht.print();
//   }

//   clock_t end = clock();
//   cout << "ht insert end time: " << end << endl;

//   cout << "cost time: " << end - start << endl << endl;

//   start = clock();
//   cout << "map insert start time: " << start << endl;

//   for (int i = 0; i < test_num; i++) {
//     map.insert({to_string(i), i});
//   }

//   end = clock();
//   cout << "map insert end time: " << end << endl;

//   cout << "cost time: " << end - start << endl << endl;

//   start = clock();
//   cout << "ht find start time: " << start << endl;

//   for (int i = 0; i < test_num; i++) {
//     ht.find(to_string(i));
//     // ht.print();
//   }

//   end = clock();
//   cout << "ht find end time: " << end << endl;

//   cout << "cost time: " << end - start << endl << endl;

//   start = clock();
//   cout << "map find start time: " << start << endl;

//   for (int i = 0; i < test_num; i++) {
//     map.find(to_string(i));
//   }

//   end = clock();
//   cout << "map find end time: " << end << endl;

//   cout << "cost time: " << end - start << endl << endl;

//   // ht.print();
//   // cout << "Load factor before rehash: " << ht.load_factor() << endl;
//   // ht.rehash(4); // 手动指定大小
//   // ht.print();
//   // cout << "Load factor after rehash : " << ht.load_factor() << endl;

//   cout << "All tests passed." << endl;
//   return 0;
// }