#include <vector>
#include <iostream>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <cstdint>
#include <climits>
using namespace std;

extern "C" {
#include "intset.h"
}

#define OK      0
#define Err     1

#define ENC_INT8    0
#define ENC_INT16   1
#define ENC_INT32   2
#define ENC_INT64   3

typedef uint8_t Encoding;

class intset {
private:
    Encoding encoding;
    int length;
    vector<uint8_t> store;

    void upgrade(Encoding target);

    int insert(int64_t val, int pos);

    /// @brief 查找元素的位置
    /// @param val
    /// @return 若元素存在，则返回元素位置（大于等于0）；否则，返回应该插入的位置（之前）的相反数减一。
    int find_index(int64_t val);
public:
    intset();
    int add(int64_t val);
    int remove(int64_t val);
    int find(int64_t val);
    int64_t random();
    int64_t get(int index);
    int len();
    int blob_len();
    void debug();
};

void intset::upgrade(Encoding target)
{
    int current_bit_size = 1 << encoding;
    int target_bit_size = 1 << target;

    int current_blob_len = current_bit_size * length;
    int target_blob_len = target_bit_size * length;

    store.resize(target_blob_len, 0);
    auto data = store.data();

    for (int i = length - 1; i >= 0; i--) {
        // 需要根据大小端字序来条件编译
#if defined(__BYTE_ORDER__) && __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
        int ol = i * current_bit_size;
        int nl = i * target_bit_size + target_bit_size - current_bit_size;
        memcpy(data + nl, data + ol, current_bit_size);
#else
        int ol = i * current_bit_size;
        int nl = i * target_bit_size;
        memcpy(data + nl, data + ol, current_bit_size);
#endif
    }

    encoding = target;
}

Encoding encoding_level(int64_t val) {
    if (val < 0) {
        if (val >= INT8_MIN) {
            return ENC_INT8;
        }
        else if (val >= INT16_MIN) {
            return ENC_INT16;
        }
        else if (val >= INT32_MIN) {
            return ENC_INT32;
        }
        else {
            return ENC_INT64;
        }
    }
    else {
        if (val <= INT8_MAX) {
            return ENC_INT8;
        }
        else if (val <= INT16_MAX) {
            return ENC_INT16;
        }
        else if (val <= INT32_MAX) {
            return ENC_INT32;
        }
        else {
            return ENC_INT64;
        }
    }
}

int intset::insert(int64_t val, int pos)
{
    auto target_encoding = encoding_level(val);
    if (target_encoding > encoding) {
        // 需要升级
        upgrade(target_encoding);
    }

    // 扩容
    auto bit_size = 1 << encoding;
    auto new_size = bit_size * (length + 1);
    store.resize(new_size, 0);

    // 挪动数据
    auto data = store.data();
    int ol = bit_size * pos;
    int nl = bit_size * (pos + 1);
    memcpy(data + nl, data + ol, (length - pos) * bit_size);

    // 复制数据
    uint8_t val_ptr[8]{};
    *(int64_t*)val_ptr = val;
#if defined(__BYTE_ORDER__) && __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
    memcpy(data + bit_size * pos,val_ptr + 8 - bit_size, bit_size);
#else
    memcpy(data + bit_size * pos, val_ptr, bit_size);
#endif

    length++;

    return OK;
}

int intset::find_index(int64_t val)
{
    auto check = [this, &val](const int i) {
        return this->get(i) <= val;
        };

    if (length == 0) {
        return -1;
    }

    int l = 0, r = length - 1;
    int result = -1; // 使用一个变量来保存找到的满足条件的最右边元素的索引
    while (l <= r) {
        int mid = l + (r - l) / 2; // 防止溢出
        if (check(mid)) {
            result = mid; // 更新满足条件的索引
            l = mid + 1; // 尝试找到更右边的满足条件的元素
        } else {
            r = mid - 1;
        }
    }

    if (get(result) != val) {
//        cout << "find_index(" << val << ") = " << -result - 2 << " ; result = " << result << " ; get(result) = "<< get(result) << endl;
        return -result - 2;
    }

//     cout << "find_index(" << val << ") = " << result << endl;

    return result;
}

intset::intset() : encoding(ENC_INT8), length(0)
{}

int intset::add(int64_t val)
{
    auto fi = find_index(val);
    if (fi < 0) {
        fi = -(fi + 1);
        insert(val, fi);
        return OK;
    }
    return Err;
}

int intset::remove(int64_t val)
{
    auto fi = find_index(val);
    if (fi < 0) {
        return Err;
    }

    // 挪动元素
    auto bit_size = 1 << encoding;
    auto data = store.data();
    int ol = bit_size * (fi + 1);
    int nl = bit_size * fi;
    memcpy(data + nl,  data + ol, (length - fi - 1) * bit_size);

    length--;
    auto new_size = bit_size * length;
    store.resize(new_size);

    return 0;
}

int intset::find(int64_t val)
{
    return find_index(val) >= 0 ? OK : Err;
}

int64_t intset::random()
{
    if (length == 0) {
        return -1;
    }

    return get(rand() % length);
}

int64_t intset::get(int index)
{
    if (index < 0 || index >= length) {
        return -1;
    }

    switch (this->encoding)
    {
    case ENC_INT8:
    {
        auto ptr = this->store.data() + index;
        return *(int8_t*)ptr;
    }
    case ENC_INT16:
    {
        auto ptr = this->store.data() + index * 2;
        return *(int16_t*)ptr;
    }
    case ENC_INT32:
    {
        auto ptr = this->store.data() + index * 4;
        return *(int32_t*)ptr;
    }
    case ENC_INT64:
    {
        auto ptr = this->store.data() + index * 8;
        return *(int64_t*)ptr;
    }
    default:
        return 0;
    }
}

int intset::len()
{
    return length;
}

int intset::blob_len()
{
    auto bit_size = 1 << encoding;
    return bit_size * length;
}

void intset::debug()
{
    for (int i = 0;i < length;i++) {
        cout << get(i) << ", ";
    }
    cout << endl;
}

IntsetHandle NewIntset()
{
    return static_cast<IntsetHandle>(new intset());
}

void ReleaseIntset(IntsetHandle handle)
{
    delete static_cast<intset*>(handle);
}

int IntsetAdd(IntsetHandle handle, long long val)
{
    return static_cast<intset*>(handle)->add(val);
}

int IntsetRemove(IntsetHandle handle, long long val)
{
    return static_cast<intset*>(handle)->remove(val);
}

int IntsetFind(IntsetHandle handle, long long val)
{
    return static_cast<intset*>(handle)->find(val);
}

long long IntsetRandom(IntsetHandle handle)
{
    return static_cast<intset*>(handle)->random();
}

long long IntsetGet(IntsetHandle handle, int index)
{
    return static_cast<intset*>(handle)->get(index);
}

int IntsetLen(IntsetHandle handle)
{
    return static_cast<intset*>(handle)->len();
}

int IntsetBlobLen(IntsetHandle handle)
{
    return static_cast<intset*>(handle)->blob_len();
}

// int main() {
//     intset s;
//     // s.debug();
//     cout << s.add(1) << endl;
//     // s.debug();
//     cout << s.find(1) << endl;
//     cout << s.remove(1) << endl;
//     cout << s.find(1) << endl;
//     cout << s.add(2) << endl;
//     cout << s.add(3) << endl;
//     cout << s.random() << endl;
//     cout << s.len() << endl;
//     cout << s.blob_len() << endl;
//     cout << "Now add 300" << endl;
//     cout << s.add(300) << endl;
//     cout << s.len() << endl;
//     cout << s.blob_len() << endl;
// }
