#include <vector>
#include <cstdint>
#include <cstring>
#include <limits>
#include <cassert>
#include <iostream>
#include <cstddef>

using namespace std;

extern "C" {
#include "ziplist.h"
}

#define Ok 0
#define Err 1

#ifndef LLONG_MAX
#define LLONG_MAX 9223372036854775807LL
#endif

#ifndef ULLONG_MAX
#define ULLONG_MAX 18446744073709551615ULL
#endif

#ifndef LLONG_MIN
#define LLONG_MIN (-9223372036854775807LL - 1)
#endif

/*
 * ziplist 末端标识符，以及 5 字节长长度标识符
 */
#define ZIP_END 255
#define ZIP_BIGLEN 254

 /* Different encoding/length possibilities */
 /*
  * 字符串编码和整数编码的掩码
  */
#define ZIP_STR_MASK 0xc0   /*11000000*/
#define ZIP_INT_MASK 0x30   /*00110000*/

  /*
   * 字符串编码类型
   */
#define ZIP_STR_06B (0 << 6)    /* 字符串长度被直接编码在接下来的 6 位中，适用于长度小于等于 63 的字符串*/
#define ZIP_STR_14B (1 << 6)    /*01 << 6 字符串长度被编码在接下来的 14 位中，适用于长度在 64 到 16383 之间的字符串*/
#define ZIP_STR_32B (2 << 6)    /*10 << 6 字符串长度被编码在接下来的 32 位中，适用于更长的字符串。*/

   /*
    * 整数编码类型
    * 0xc0 (11000000)用于标识整数编码
    */
#define ZIP_INT_16B (0xc0 | 0<<4)   /*表示 16 位整数的编码。0xc0 | 0 的结果仍然是 0xc0。表示 16 位整数的编码以 11000000 开头。*/
#define ZIP_INT_32B (0xc0 | 1<<4)   /*表示 32 位整数的编码。1 << 4 得到 00010000 与 0xc0 进行或，得到  11010000。表示 32 位整数的编码以 11010000 开头。*/
#define ZIP_INT_64B (0xc0 | 2<<4)   /*表示 64 位整数的编码。10 << 4得到 00100000 与 0xc0 进行或操作得到11100000。表示 64 位整数的编码以 11100000 开头*/
#define ZIP_INT_24B (0xc0 | 3<<4)   /*表示 24 位整数的编码。11 << 4得到 00110000。与 0xc0 进行或操作为 11110000 表示 24 位整数的编码以 11110000 开头。*/
#define ZIP_INT_8B 0xfe             /*表示 8 位整数的编码，直接使用 0xfe（11111110）作为其标识。*/

    /* 4 bit integer immediate encoding
     *
     * 4 位整数编码的掩码和类型
     */
#define ZIP_INT_IMM_MASK 0x0f   /*00001111用作掩码，目的是从一个编码过的字节中提取出实际的整数值。结果是字节的低4位，也就是实际存储的整数值。*/
     /*
     * 下面两个宏定义标识了使用即时编码可以表示的整数值的范围。
     * 0xf1（11110001）到0xfd（11111101）之间的每个值都可以被用来表示一个小的整数值。
     * 注意这里没有使用到完整的字节范围，这是因为一些特殊的标识符（如0xfe和0xff）已被预留用于其他目的。
     */
#define ZIP_INT_4b 0xf0         /* 11110000 */  //用于4位大小的整数0-12，写到encode的后4位
#define ZIP_INT_IMM_MIN 0xf1    /* 11110001 */
#define ZIP_INT_IMM_MAX 0xfd    /* 11111101 */
     /*
     * 用于从编码过的字节v中提取出实际的整数值。
     * 通过将输入字节与ZIP_INT_IMM_MASK进行AND操作，可以去除字节高位的编码信息，仅留下低4位的整数值。
     */
#define ZIP_INT_IMM_VAL(v) (v & ZIP_INT_IMM_MASK)

     /*
      * 24 位整数的最大值和最小值
      */
#define INT24_MAX 0x7fffff
#define INT24_MIN (-INT24_MAX - 1)

typedef int ZipListResult;

// ⚠️ 压缩列表的节点不能照抄这段struct，应该照书上来（例如，书上说previous_entry_length的长度是可变的，它可能只占1个字节，也可能占5个字节
// ⚠️ 实现建议：在底层不存储这个struct。可以实现两个把这个struct和uint8_t[]互相转换的函数
struct ziplist_node
{
    int previous_entry_length;
    uint8_t encoding;
    int ba_length;  // 压缩列表节点存储的内容的长度
    uint64_t value; // 当压缩列表节点存的是数字时，存在这里面
    vector<uint8_t> content; // 当压缩列表节点存的是字符串时，存在这里面

    ziplist_node(char* bytes, int len);
    ziplist_node(int64_t integer);
    ziplist_node() {};

    // 测试 用于输出zlnode的内容
    void output_zlnode() {
        cout << endl;
        cout << "previous_entry_length: " << previous_entry_length << endl;
        cout << "encoding: " << encoding << endl;
        cout << "ba_length: " << ba_length << endl;
        if (content.size()) {
            cout << "content: ";
            for (auto& it : content) {
                cout << it << ' ';
            }
        }
        else {
            cout << "value: " << value << endl;
        }

    }

    ~ziplist_node() {};
};

class ziplist
{
private:
    // vector<ziplist_node> store; // ⚠️ 建议使用 vector<uint8_t> 而不是 uint8_t* 这种C风格的东西，可参考intset的用法
    vector<uint8_t> store;
    //整数写入store中
    void int2uint8(uint16_t num);
    void int2uint8(uint32_t num);
    void int2uint8(uint64_t num);

    void setZlbytes(uint32_t zlbytes);
    void setZltail(uint32_t zltail);
    void setZllen(uint16_t zllen);

    /**
     * 在store<uint8_t>中位置为pos的地方插入一段新的节点，返回值为插入成功或失败
     * 插入失败返回Err,原因为指定插入的位置当前超出store.size()
    */
    ZipListResult insertEntry(vector<uint8_t>& new_node, size_t position);
    /**
     * 传入当前节点的索引（是第几个节点，从1开始）
     * 获得其前序节点的节点长度并返回
    */
    size_t get_prev_len(int pos);

    /**
     * 用于在push操作中获取最后一个节点的length，
     * 填到新加入的节点的previous_entry_length中
    */
    size_t get_prev_len_for_push();

    /**
     * 获取在vector中起始位置为pos的节点的previous_entry_length
    */
    size_t get_prev_length(size_t pos);
    /**
     * 获取以索引pos作为起始地址的长度
    */
    size_t get_node_len(size_t pos);
    /**
     * 用于给定正向索引index，返回该节点在store中起始节点的位置
    */
    size_t locate_pos(int index);
    /**
      * 用于给定节点指针*cur，以引用的形式返回该节点在store中起始节点的位置
      * 会返回Ok 或 Err
     */
    ZipListResult locate_node(ziplist_node* cur, size_t& pos);

    /**
     * 底层存储到zlnode结构体
    */
    ZipListResult mem2zlnode(size_t pos, ziplist_node*& zp);

    /**
     * 连锁更新，为删除操作特化
     * 传入的pos是待删除节点在store中的起始的位置
    */
    void chain_renew_for_delete(size_t pos, size_t former_node_len);

    //get set zlbytes, zltail, zllen
    uint32_t getZlbytes();  //压缩列表占用的内存字节数
    uint32_t getZltail();   //压缩列表表尾节点距离压缩列表的起始地址有多少字节
    uint16_t getZllen();    //压缩列表包含的节点数量

public:
    /**
     * zlnode结构体到底层存储
    */
    // ZipListResult zlnode2mem(ziplist_node zn);
    void output_store();

    /**
     * 按顺序输出当前ziplist中的内容
    */
    void output_node_content();

    ziplist();
    // 将元素插入到表尾
    ZipListResult push(char* bytes, int len);
    ZipListResult push(int64_t integer);

    /**
     * 在pos位置插入新的节点，pos是从1开始的，新插入节点的位置为第pos个
    */
    ZipListResult insert(int pos, char* bytes, int len);
    ZipListResult insert(int pos, int64_t integer);

    // 返回压缩列表给定索引上的节点，此处索引是从1开始的
    ziplist_node* index(int n);

    // 查找具有指定值的节点
    ziplist_node* find(char* bytes, int len);
    ziplist_node* find(int64_t integer);

    // 返回指定节点的下一个节点
    ziplist_node* next(ziplist_node* cur);

    // 返回指定节点的上一个节点
    ziplist_node* prev(ziplist_node* cur);

    static int64_t get_integer(ziplist_node* cur);
    static vector<uint8_t> get_byte_array(ziplist_node* cur);

    //TODO
    ZipListResult delete_(ziplist_node* cur);
    ZipListResult delete_range(ziplist_node* start, int len);
    /**
     * 通过传入的参数pos来
    */
    ZipListResult delete_by_pos(size_t pos);

    /**
     * 连锁更新，从ziplist的第pos个（pos从1开始）之后（不包括pos节点）开始连锁更新
     * 第pos个节点前一个节点的长度为former_node_len
    */
    void chain_renew(size_t pos, size_t former_node_len);

    int blob_len();
    int len();
};

void ziplist::int2uint8(uint16_t num) {
    for (size_t i = 0; i < sizeof(uint16_t); ++i) {
        // 按字节添加到store中，考虑小端字节序
        this->store.push_back(reinterpret_cast<uint8_t*>(&num)[i]);
    }
}

void ziplist::int2uint8(uint32_t num) {
    for (size_t i = 0; i < sizeof(uint32_t); ++i) {
        this->store.push_back(reinterpret_cast<uint8_t*>(&num)[i]);
    }
}

void ziplist::int2uint8(uint64_t num) {
    for (size_t i = 0; i < sizeof(uint64_t); ++i) {
        this->store.push_back(reinterpret_cast<uint8_t*>(&num)[i]);
    }
}

uint32_t ziplist::getZlbytes() {
    uint32_t zlbytes;
    memcpy(&zlbytes, this->store.data(), sizeof(zlbytes));
    return zlbytes;
}

uint32_t ziplist::getZltail() {
    uint32_t zltail;
    memcpy(&zltail, this->store.data() + sizeof(uint32_t), sizeof(zltail));
    return zltail;
}

uint16_t ziplist::getZllen() {
    uint16_t zllen;
    memcpy(&zllen, this->store.data() + 2 * sizeof(uint32_t), sizeof(zllen));
    return zllen;
}

void ziplist::setZlbytes(uint32_t zlbytes) {
    for (size_t i = 0; i < sizeof(zlbytes); ++i) {
        // 更新 vector 中相应位置的字节
        this->store[i] = reinterpret_cast<uint8_t*>(&zlbytes)[i];
    }
}

void ziplist::setZltail(uint32_t zltail) {
    for (size_t i = 0; i < sizeof(zltail); ++i) {
        // 假设 zltail 紧跟在 zlbytes 后面
        this->store[sizeof(uint32_t) + i] = reinterpret_cast<uint8_t*>(&zltail)[i];
    }
}

void ziplist::setZllen(uint16_t zllen) {
    for (size_t i = 0; i < sizeof(zllen); ++i) {
        // 假设 zllen 紧跟在 zltail 后面
        this->store[2 * sizeof(uint32_t) + i] = reinterpret_cast<uint8_t*>(&zllen)[i];
    }
}

/**
 * 底层存储到zlnode结构体
 * 此处pos指的是能直接放在store[pos]的索引，不是从1开始的位置
 * 以引用的形式返回ziplist_node
 * 返回值是操作成功或失败，操作失败原因为编码encoding找不到对应
*/
ZipListResult ziplist::mem2zlnode(size_t pos, ziplist_node*& zp) {
    size_t p = pos;
    zp = new ziplist_node();
    zp->previous_entry_length = this->get_prev_length(p);

    //定位encoding，并记录prev_length的长度到res中
    if (this->store[p] == (uint8_t)0xFE) {
        p += 5;
    }
    else {
        p += 1;
    }

    //store[pos] & 11000000 = 11000000说明是整数
    if (((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0xC0) {
        zp->encoding = (uint8_t)store[p];
        uint8_t encoding = store[p];
        p += 1;
        if ((encoding & ZIP_INT_4b) == ZIP_INT_4b) {
            //整数为0-12，content长度为1，只有低4位是有效的
            zp->value = encoding & 0x0F;
            zp->ba_length = 0;
        }
        else if (encoding == ZIP_INT_8B) {
            //8位整数
            zp->value = (uint64_t)store[p];
            zp->ba_length = 1;
        }
        else if (encoding == ZIP_INT_16B) {
            //16位整数
            zp->value = static_cast<uint16_t>(store[p]) |
                static_cast<uint16_t>(store[p + 1]) << 8;
            zp->ba_length = 2;

        }
        else if (encoding == ZIP_INT_24B) {
            //24位整数
            zp->value = static_cast<uint32_t>(store[p]) |
                static_cast<uint32_t>(store[p + 1]) << 8 |
                static_cast<uint32_t>(store[p + 2]) << 16;
            zp->ba_length = 3;
        }
        else if (encoding == ZIP_INT_32B) {
            //32位整数
            zp->value = static_cast<uint32_t>(store[p]) |
                static_cast<uint32_t>(store[p + 1]) << 8 |
                static_cast<uint32_t>(store[p + 2]) << 16 |
                static_cast<uint32_t>(store[p + 3]) << 24;
            zp->ba_length = 4;
        }
        else {
            //64位整数，节点大小+8
            zp->value = static_cast<uint64_t>(store[p]) |
                static_cast<uint64_t>(store[p + 1]) << 8 |
                static_cast<uint64_t>(store[p + 2]) << 16 |
                static_cast<uint64_t>(store[p + 3]) << 24 |
                static_cast<uint64_t>(store[p + 4]) << 32 |
                static_cast<uint64_t>(store[p + 5]) << 40 |
                static_cast<uint64_t>(store[p + 6]) << 48 |
                static_cast<uint64_t>(store[p + 7]) << 56;
            zp->ba_length = 8;
        }
    }
    //encoding长度1字节，字节数组长度小于63字节
    else if (((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x00) {
        //encoding长度为1
        //获取后6位，即为字节的长度，并更新ba->length
        size_t len = store[p] & 0x3F;
        zp->ba_length = len;
        p += 1;
        for (size_t i = 0; i < len; i++) {
            zp->content.push_back(store[p + i]);
        }
    }
    //encoding长度2字节，字节数组长度小于16838字节
    else if (((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x40) {
        //encoding长度为2
        //获取后14位，即为字节的长度，并更新ba->length
        size_t len = ((store[p] & 0x3F) << 8) | store[p + 1];
        zp->ba_length = len;
        p += 2;
        for (size_t i = 0; i < len; i++) {
            zp->content.push_back(store[p + i]);
        }
    }
    //encoding长度5字节，字节数组长度大于16838字节
    else if (((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x80) {
        //encoding长度为5
        //获取后32位，即为字节的长度，并更新res
        size_t len = ((uint64_t)store[p + 1] << 24) |
            ((uint64_t)store[p + 2] << 16) |
            ((uint64_t)store[p + 3] << 8) |
            ((uint64_t)store[p + 4]);
        p += 5;
        for (size_t i = 0; i < len; i++) {
            zp->content.push_back(store[p + i]);
        }
    }
    else {
        // 解码错误
        return Err;
    }
    return Ok;
}

size_t ziplist::get_prev_len_for_push() {
    size_t zltail = this->getZltail();
    //若当前压缩列表内没有节点，则返回0
    if (this->getZllen() == 0) {
        return 0;
    }
    return store.size() - zltail;
}

ziplist::ziplist()
{
    uint32_t zlbytes = 0x0A;
    uint32_t zltail = 0x0A;
    uint16_t zllen = 0;
    this->int2uint8(zlbytes);
    this->int2uint8(zltail);
    this->int2uint8(zllen);
}

void ziplist::output_node_content() {
    cout << endl;
    int zl_len = this->getZllen();
    ziplist_node* zlnode = nullptr;
    for (int i = 1; i <= zl_len; i++) {
        zlnode = this->index(i);
        int64_t val = ziplist::get_integer(zlnode);
        cout << i << ": ";
        if (val == LLONG_MAX) {
            vector<uint8_t> cur_content = ziplist::get_byte_array(zlnode);
            for (uint8_t byte : cur_content) {
                cout << byte;
            }
            cout << endl;
        }
        else {
            int64_t num = ziplist::get_integer(zlnode);
            cout << num << endl;
        }
    }
    cout << endl;
}

ZipListResult ziplist::push(char* bytes, int len)
{
    //构造新插入节点的previous_entry_length
    size_t prev_length = this->get_prev_len_for_push();
    prev_length = (uint32_t)prev_length;
    uint8_t prev_length_buf[5];
    int prev_length_len = 0;    //新插入节点的prev_length的长度
    //前序节点长度位于0-254之间，previous_entry_length长为1字节
    if (prev_length > 0 && prev_length < 254) {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)prev_length;
    }
    //前序节点长度大于254，previous_entry_length长为5字节
    else if (prev_length >= 254) {
        prev_length_len = 5;
        prev_length_buf[0] = (uint8_t)0xFE;
        for (size_t i = 0; i < sizeof(uint32_t); ++i) {
            prev_length_buf[i + 1] = reinterpret_cast<uint8_t*>(&prev_length)[i];
        }
    }
    //前序节点长度为0，该节点是ziplist的第一个节点
    else {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)0;
    }
    //写入previous_entry_length
    for (int i = 0; i < prev_length_len; i++) {
        store.push_back(prev_length_buf[i]);
    }

    uint8_t buf[5];   //encoding数组编码
    int encoding_len = 0;    //encoding数组长度
    size_t node_len = 0; // 保存节点长度
    /*确定字符串的encoding*/
    if (len <= 0x3f) {
        buf[0] = ZIP_STR_06B | len;
        encoding_len = 1;
    }
    else if (len <= 0x3fff) {
        buf[0] = ZIP_STR_14B | ((len >> 8) & 0x3f);
        buf[1] = len & 0xff;
        encoding_len = 2;
    }
    else {
        len += 4;
        buf[0] = ZIP_STR_32B;
        buf[1] = (len >> 24) & 0xff;
        buf[2] = (len >> 16) & 0xff;
        buf[3] = (len >> 8) & 0xff;
        buf[4] = len & 0xff;
        encoding_len = 5;
    }
    /*将字符串的encoding写入store中*/
    for (int i = 0; i < encoding_len; i++) {
        this->store.push_back(buf[i]);
    }
    /*将字符串本身写入store中*/
    for (int i = 0; i < len; i++) {
        this->store.push_back(*(bytes + i));
    }
    this->setZlbytes(this->store.size());
    this->setZllen(this->getZllen() + 1);
    this->setZltail(this->getZltail() + prev_length);
    return Ok;
}

ZipListResult ziplist::push(int64_t integer)
{
    size_t prev_length = this->get_prev_len_for_push();
    uint8_t prev_length_buf[5];
    int prev_length_len = 0;
    //前序节点长度位于0-254之间，previous_entry_length长为1字节
    if (prev_length > 0 && prev_length < 254) {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)prev_length;
    }
    //前序节点长度大于254，previous_entry_length长为5字节
    else if (prev_length >= 254) {
        prev_length_len = 5;
        prev_length_buf[0] = (uint8_t)0xFE;
        for (size_t i = 0; i < sizeof(uint32_t); ++i) {
            prev_length_buf[i + 1] = reinterpret_cast<uint8_t*>(&prev_length)[i];
        }
    }
    //前序节点长度为0，该节点是ziplist的第一个节点
    else {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)0;
    }
    //写入previous_entry_length
    for (int i = 0; i < prev_length_len; i++) {
        store.push_back(prev_length_buf[i]);
    }

    uint8_t encoding = 0;
    if (integer >= 0 && integer <= 12) {
        encoding = ZIP_INT_4b + integer;
    }
    else if (integer >= INT8_MIN && integer <= INT8_MAX) {
        encoding = ZIP_INT_8B;
    }
    else if (integer >= INT16_MIN && integer <= INT16_MAX) {
        encoding = ZIP_INT_16B;
    }
    else if (integer >= INT24_MIN && integer <= INT24_MAX) {
        encoding = ZIP_INT_24B;
    }
    else if (integer >= INT32_MIN && integer <= INT32_MAX) {
        encoding = ZIP_INT_32B;
    }
    else {
        encoding = ZIP_INT_64B;
    }
    this->store.push_back(encoding);

    // 对于非立即数编码，将integer的字节添加到store
    if (!(integer >= 0 && integer <= 12)) {
        size_t size = 0; // 根据encoding确定需要存储的字节数
        switch (encoding) {
        case ZIP_INT_8B: size = 1; break;
        case ZIP_INT_16B: size = 2; break;
        case ZIP_INT_24B: size = 3; break;
        case ZIP_INT_32B: size = 4; break;
        case ZIP_INT_64B: size = 8; break;
        }

        for (size_t i = 0; i < size; ++i) {
            // 按字节添加到store中，考虑系统可能的小端字节序
            this->store.push_back(reinterpret_cast<uint8_t*>(&integer)[i]);
        }
    }
    this->setZlbytes(this->store.size());
    this->setZllen(this->getZllen() + 1);
    this->setZltail(this->getZltail() + prev_length);
    return Ok;
}

/**
 * 获取以索引pos作为起始地址的节点长度
 * 此处pos指的是能直接放在store[pos]的索引，不是从1开始的位置
 * 错误处理：如果pos < 0, 则直接返回0；如果pos > store.size()，则返回LLONG_MAX
*/
size_t ziplist::get_node_len(size_t pos) {
    if (pos < 0) {
        return 0;
    }
    else if (pos >= this->store.size()) {
        return LLONG_MAX;
    }
    size_t res = 0; //返回值
    size_t p = pos;
    //定位encoding，并记录prev_length的长度到res中
    if (this->store[p] == (uint8_t)0xFE) {
        p += 5;
        res += 5;
    }
    else {
        p += 1;
        res += 1;
    }

    //store[pos] & 11000000 = 11000000说明是整数
    //注意==的优先级高于&，要加括号
    if (((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0xC0) {
        uint8_t encoding = store[p];
        res++;  //记录encoding的长度到res中
        if (encoding & ZIP_INT_4b == ZIP_INT_4b) {
            //整数为0-12，什么也不做，没有content
        }
        else if (encoding == ZIP_INT_8B) {
            //8位整数，节点大小+1
            res += 1;
        }
        else if (encoding == ZIP_INT_16B) {
            //16位整数，节点大小+2
            res += 2;
        }
        else if (encoding == ZIP_INT_24B) {
            //24位整数，节点大小+3
            res += 3;
        }
        else if (encoding == ZIP_INT_32B) {
            //32位整数，节点大小+4
            res += 4;
        }
        // else if(encoding == ZIP_INT_64B) {
        //     //64位整数，节点大小+8
        //     res += 8;
        // }
        else {
            //64位整数，节点大小+8
            res += 8;
        }
    }
    //encoding长度1字节，字节数组长度小于63字节
    else if (((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x00) {
        //encoding长度为1，更新res
        res += 1;
        //获取后6位，即为字节的长度，并更新res
        int len = store[p] & 0x3F;
        res += len;
    }
    //encoding长度2字节，字节数组长度小于16838字节
    else if (((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x40) {
        //encoding长度为2，更新res
        res += 2;
        //获取后14位，即为字节的长度，并更新res
        int len = ((store[p] & 0x3F) << 8) | store[p + 1];
        res += len;
    }
    //encoding长度5字节，字节数组长度大于16838字节
    else if (((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x80) {
        //encoding长度为5，更新res
        res += 5;
        //获取后32位，即为字节的长度，并更新res
        size_t len = ((uint64_t)store[p + 1] << 24) |
            ((uint64_t)store[p + 2] << 16) |
            ((uint64_t)store[p + 3] << 8) |
            ((uint64_t)store[p + 4]);
        res += len;
    }

    return res;
}

size_t ziplist::get_prev_length(size_t pos) {
    size_t res = 0; //返回值
    size_t p = pos;
    if (this->store[p] == (uint8_t)0xFE) {
        // 对于小端字节序：
        res = ((uint32_t)store[p + 4] << 24) |
            ((uint32_t)store[p + 3] << 16) |
            ((uint32_t)store[p + 2] << 8) |
            ((uint32_t)store[p + 1]);
        //cout << res << endl;  
    }
    else {
        res = (size_t)this->store[p];
    }
    return res;
}

// 返回压缩列表给定索引上的节点，此处的索引是从1开始的
ziplist_node* ziplist::index(int n)
{
    ziplist_node* zlnode = new ziplist_node();
    if (this->mem2zlnode(this->locate_pos(n), zlnode) == Ok) {
        return zlnode;
    }
    else {
        return nullptr;
    }
}

/**
 * 用于给定正向索引index，返回该节点在store中起始节点的位置
 * 错误处理：若index <= 0，则返回0；若index > 当前长度，则返回LLONG_MAX
*/
size_t ziplist::locate_pos(int index) {
    if (index <= 0) {
        return 0;
    }
    else if (index > this->getZllen()) {
        return LLONG_MAX;
    }
    uint16_t len = this->getZllen();
    uint32_t pos = this->getZltail();
    //若列表为空，直接返回0
    if (len == 0) {
        return 0;
    }
    int rev_index = len - index;
    for (int i = 0; i < rev_index; i++) {
        size_t prev_length = this->get_prev_length(pos);
        pos -= prev_length;
    }
    return pos;
}

/**
 * 如果没找到，返回nullptr
*/
ziplist_node* ziplist::find(char* bytes, int len) {
    size_t zl_len = this->getZllen();
    ziplist_node* zlnode = nullptr;
    for (int i = 1; i <= zl_len; i++) {
        zlnode = this->index(i);
        vector<uint8_t> cur_content = ziplist::get_byte_array(zlnode);
        if (cur_content.empty()) {
            continue;
        }
        if (equal(cur_content.begin(), cur_content.end(), bytes, bytes + len)) {
            return zlnode;
        }
    }
    return nullptr;
}

/**
 * pos为store中的指定位置，会插在这个位置的后面
*/
ZipListResult ziplist::insertEntry(vector<uint8_t>& new_node, size_t pos) {
    if (pos > store.size()) {
        return Err;
    }
    // 扩展store的大小以容纳新节点
    store.resize(store.size() + new_node.size());
    // 将从position开始的旧元素向后移动new_node.size()个位置
    move_backward(store.begin() + pos, store.end() - new_node.size(), store.end());
    /*if (pos == 31) {
        cout << 1111111 << endl;
        this->output_store();
    }*/
    // 复制new_node到store的指定位置
    copy(new_node.begin(), new_node.end(), store.begin() + pos);
    /*if (pos == 31) {
        cout << 1111111 << endl;
        this->output_store();
    }*/
    return Ok;
}

/**
 * 传入当前节点的索引（是第几个节点，从1开始）
 * 获得其节点长度并返回
*/
size_t ziplist::get_prev_len(int pos) {
    size_t zl_len = this->getZllen();
    if (pos > zl_len) {
        return LLONG_MAX;
    }
    else if (pos <= 0) {
        return 0;
    }
    return this->get_node_len(this->locate_pos(pos));
}

/**
 * 在压缩列表第pos个节点之后插入新的节点
 * 返回错误的原因：pos > this.getZllen() || pos < 0
 * 若pos = 0，则插入在开头
*/
ZipListResult ziplist::insert(int pos, char* bytes, int len) {
    if (this->getZllen() == 0) {
        return this->push(bytes, len);
    }
    vector<uint8_t> new_node;   //构造新的待写入节点
    if (pos > this->getZllen() || pos < 0) {
        return Err;
    }
    //构造新插入节点的previous_entry_length
    size_t prev_length = this->get_prev_len(pos);
    prev_length = (uint32_t)prev_length;
    uint8_t prev_length_buf[5];
    int prev_length_len = 0;    //新插入节点的prev_length的长度
    //前序节点长度位于0-254之间，previous_entry_length长为1字节
    if (prev_length > 0 && prev_length < 254) {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)prev_length;
    }
    //前序节点长度大于254，previous_entry_length长为5字节
    else if (prev_length >= 254) {
        prev_length_len = 5;
        prev_length_buf[0] = (uint8_t)0xFE;
        for (size_t i = 0; i < sizeof(uint32_t); ++i) {
            prev_length_buf[i + 1] = reinterpret_cast<uint8_t*>(&prev_length)[i];
        }
    }
    //前序节点长度为0，该节点是ziplist的第一个节点
    else {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)0;
    }
    //写入previous_entry_length
    for (int i = 0; i < prev_length_len; i++) {
        new_node.push_back(prev_length_buf[i]);
    }

    uint8_t buf[5];   //encoding数组编码
    int encoding_len = 0;    //encoding数组长度
    size_t node_len = 0; // 保存节点长度
    /*确定字符串的encoding*/
    if (len <= 0x3f) {
        buf[0] = ZIP_STR_06B | len;
        encoding_len = 1;
    }
    else if (len <= 0x3fff) {
        buf[0] = ZIP_STR_14B | ((len >> 8) & 0x3f);
        buf[1] = len & 0xff;
        encoding_len = 2;
    }
    else {
        len += 4;
        buf[0] = ZIP_STR_32B;
        buf[1] = (len >> 24) & 0xff;
        buf[2] = (len >> 16) & 0xff;
        buf[3] = (len >> 8) & 0xff;
        buf[4] = len & 0xff;
        encoding_len = 5;
    }
    /*将字符串的encoding写入*/
    for (int i = 0; i < encoding_len; i++) {
        new_node.push_back(buf[i]);
    }
    /*将字符串本身写入*/
    for (int i = 0; i < len; i++) {
        new_node.push_back(*(bytes + i));
    }
    if (pos == 0) {
        // 10是排除掉ziplist头的位置
        this->insertEntry(new_node, 10);
    }
    else {
        this->insertEntry(new_node, this->locate_pos(pos) + get_node_len(this->locate_pos(pos)));
    }
    //如果该节点被添加在最后，则zltail+prev_len
    if (pos == this->getZllen()) {
        this->setZltail(this->getZltail() + prev_length);
    }
    //否则，添加的是他本身的长度
    else {
        this->setZltail(this->getZltail() + new_node.size());
    }
    this->setZllen(this->getZllen() + 1);
    //插入位置在最后，没必要更新后续节点的prev_len
    if (pos == this->store.size()) {
        return Ok;
    }

    // 从新插入的节点（pos+1）之后开始更新
    this->chain_renew(pos + 2, new_node.size());
    this->setZlbytes(this->store.size());
    return Ok;
}

ZipListResult ziplist::insert(int pos, int64_t integer) {
    if (this->getZllen() == 0) {
        return this->push(integer);
    }
    vector<uint8_t> new_node;   //构造新的待写入节点
    if (pos > this->getZllen() || pos < 0) {
        return Err;
    }
    //构造新插入节点的previous_entry_length
    size_t prev_length = this->get_prev_len(pos);
    prev_length = (uint32_t)prev_length;
    uint8_t prev_length_buf[5];
    int prev_length_len = 0;    //新插入节点的prev_length的长度
    //前序节点长度位于0-254之间，previous_entry_length长为1字节
    if (prev_length > 0 && prev_length < 254) {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)prev_length;
    }
    //前序节点长度大于254，previous_entry_length长为5字节
    else if (prev_length >= 254) {
        prev_length_len = 5;
        prev_length_buf[0] = (uint8_t)0xFE;
        for (size_t i = 0; i < sizeof(uint32_t); ++i) {
            prev_length_buf[i + 1] = reinterpret_cast<uint8_t*>(&prev_length)[i];
        }
    }
    //前序节点长度为0，该节点是ziplist的第一个节点
    else {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)0;
    }
    //写入previous_entry_length
    for (int i = 0; i < prev_length_len; i++) {
        new_node.push_back(prev_length_buf[i]);
    }
    uint8_t encoding = 0;
    if (integer >= 0 && integer <= 12) {
        encoding = ZIP_INT_4b + integer;
    }
    else if (integer >= INT8_MIN && integer <= INT8_MAX) {
        encoding = ZIP_INT_8B;
    }
    else if (integer >= INT16_MIN && integer <= INT16_MAX) {
        encoding = ZIP_INT_16B;
    }
    else if (integer >= INT24_MIN && integer <= INT24_MAX) {
        encoding = ZIP_INT_24B;
    }
    else if (integer >= INT32_MIN && integer <= INT32_MAX) {
        encoding = ZIP_INT_32B;
    }
    else {
        encoding = ZIP_INT_64B;
    }
    new_node.push_back(encoding);
    // 对于非立即数编码，将integer的字节添加到new_node
    if (!(integer >= 0 && integer <= 12)) {
        size_t size = 0; // 根据encoding确定需要存储的字节数
        switch (encoding) {
        case ZIP_INT_8B: size = 1; break;
        case ZIP_INT_16B: size = 2; break;
        case ZIP_INT_24B: size = 3; break;
        case ZIP_INT_32B: size = 4; break;
        case ZIP_INT_64B: size = 8; break;
        }
        for (size_t i = 0; i < size; ++i) {
            new_node.push_back(reinterpret_cast<uint8_t*>(&integer)[i]);
        }
    }
    if (pos == 0) {
        this->insertEntry(new_node, 10);
    }
    else {
        this->insertEntry(new_node, this->locate_pos(pos) + get_node_len(this->locate_pos(pos)));
    }
    //如果该节点被添加在最后，则zltail+prev_len
    if (pos == this->getZllen()) {
        this->setZltail(this->getZltail() + prev_length);
    }
    //否则，添加的是他本身的长度
    else {
        this->setZltail(this->getZltail() + new_node.size());
    }
    this->setZllen(this->getZllen() + 1);
    //插入位置在最后，没必要更新后续节点的prev_len
    if (pos == this->store.size()) {
        return Ok;
    }

    // 从新插入的节点（pos+1）之后开始更新
    this->chain_renew(pos + 2, new_node.size());
    this->setZlbytes(this->store.size());
    return Ok;
}

ZipListResult ziplist::locate_node(ziplist_node* cur, size_t& pos) {
    vector<uint8_t> cur_content = ziplist::get_byte_array(cur);
    int64_t integer = 0;
    bool str_or_int = false;    //false:存的字符串, int:存的整数值
    if (cur_content.empty()) {
        str_or_int = true;
        integer = ziplist::get_integer(cur);
    }
    if (integer == LLONG_MAX) {
        return Err;
    }
    // 存的整数值
    if (str_or_int) {
        size_t zl_len = this->getZllen();
        uint32_t cur_pos = this->getZltail();
        int i = 0;
        ziplist_node* zl_node = nullptr;
        for (i = 0; i < zl_len; i++) {
            if (this->mem2zlnode(cur_pos, zl_node) == Ok) {
                if (zl_node->content.size() == 0) {
                    if (ziplist::get_integer(zl_node) == integer) {
                        pos = cur_pos;
                        return Ok;
                    }
                }
            }
            else {
                return Err;
            }
            size_t prev_length = this->get_prev_length(cur_pos);
            cur_pos -= prev_length;
        }
        if (i == zl_len) {
            return Err;
        }
    }
    else {
        size_t zl_len = this->getZllen();
        uint32_t cur_pos = this->getZltail();
        int i = 0;
        ziplist_node* zl_node = nullptr;
        for (i = 0; i < zl_len; i++) {
            if (this->mem2zlnode(cur_pos, zl_node) == Ok) {
                if (zl_node->content.size() != 0) {
                    if (ziplist::get_byte_array(zl_node) == cur_content) {
                        pos = cur_pos;
                        return Ok;
                    }
                }
            }
            else {
                return Err;
            }
            size_t prev_length = this->get_prev_length(cur_pos);
            cur_pos -= prev_length;
        }
        if (i == zl_len) {
            return Err;
        }
    }
    return Ok;
}

/**
 * 指定一个节点cur，删除该节点
 * 正确删除返回Ok，
 * 错误情况返回Err：传入节点结构体cur值出错(content没有或value没有)
 *          节点的值未找到
*/
ZipListResult ziplist::delete_(ziplist_node* cur) {
    size_t pos; //此处的pos是store的索引，从0开始
    if (this->locate_node(cur, pos) == Ok) {
        size_t node_len = this->get_node_len(pos);
        size_t prev_len = this->get_prev_length(pos);
        bool flag = false;
        if (pos == this->getZltail()) { //说明删除的是最后一个节点
            flag = true;
        }
        if (this->delete_by_pos(pos) == Ok) {
            this->setZllen(this->getZllen() - 1);
            if (flag) {
                //若前序节点长度也为0，说明删去后ziplist长度为0
                if (prev_len == 0) {
                    this->setZltail(0x0A);
                }
                else {
                    this->setZltail(this->getZltail() - prev_len);
                }
            }
            else {
                this->setZltail(this->getZltail() - node_len);
            }
            this->chain_renew_for_delete(pos, prev_len);
            this->setZlbytes(this->store.size());
            return Ok;
        }
        else {
            return Err;
        }
    }
    else {
        return Err;
    }
    return Ok;
}

/**
 * 当传入的pos >= store.size() || pos < 0时，返回Err
*/
ZipListResult ziplist::delete_by_pos(size_t pos) {
    if (pos >= store.size() || pos < 0) {
        return Err;
    }
    auto node_len = this->get_node_len(pos);
    this->store.erase(store.begin() + pos, store.begin() + pos + node_len);
    return Ok;
}

/**
 * 从start节点开始，删除包括start节点在内的len个节点，
 * 当要删除的部分超出最后一个节点时，返回Err
*/
ZipListResult ziplist::delete_range(ziplist_node* start, int len) {
    size_t pos; //此处的pos是store的索引，从0开始
    if (this->locate_node(start, pos) == Ok) {
        size_t del_len = 0;
        size_t cur_pos = pos;
        size_t prev_len = this->get_prev_length(pos);
        bool flag = false;  //判断最后一个节点是否被删除，若被删除则需要对zltail特殊处理
        //先确定要删除的长度
        for (int i = 0; i < len; i++) {
            size_t node_len = this->get_node_len(cur_pos);
            del_len += node_len;
            cur_pos += node_len;
            if (cur_pos >= store.size() || cur_pos < 0) {
                return Err;
            }
        }
        if (cur_pos == this->store.size() - 1) {
            flag = true;
        }
        this->store.erase(store.begin() + pos, store.begin() + pos + del_len);
        this->setZllen(this->getZllen() - len);
        if (flag) {
            if (prev_len == 0) {
                this->setZltail(0x0A);
            }
            else {
                this->setZltail(this->getZltail() - del_len - prev_len);
            }
        }
        else {
            this->setZltail(this->getZltail() - del_len);
        }
        this->chain_renew_for_delete(pos, prev_len);
        this->setZlbytes(this->store.size());
    }
    else {
        return Err;
    }
    return Ok;
}

/*
* 注意，在更新时，若前序节点长度为4位，则低位在前 高位在后，
* previou_length字段若为0xFE 80 1 0 0 ，则应该倒序组装为
* 00000000 00000000 00000001 01010000 即为336
*/

void ziplist::chain_renew(size_t pos, size_t former_node_len) {
    if (pos > this->getZllen()) {
        return;
    }
    //此时，zltail已更新，如果是中间插入的场景，是可以定位到新插入的节点的后一个节点的
    ziplist_node* zlnode = this->index(pos);
    //暂存当前待修改节点的长度和起始位置，方便递归调用
    size_t temp_length = this->get_node_len(this->locate_pos(pos));
    size_t cur_pos = this->locate_pos(pos);
    //若前序节点和待插入节点的长度均小于254字节，则直接修改内存中的prev_length字段
    if (former_node_len < 254 && zlnode->previous_entry_length < 254) {
        this->store[this->locate_pos(pos)] = (uint8_t)former_node_len;
        return;
    }
    //若前序节点和待修改节点均长于254字节
    else if (former_node_len >= 254 && zlnode->previous_entry_length >= 254) {
        for (size_t i = 1; i <= sizeof(uint32_t); i++) {
            this->store[this->locate_pos(pos) + i] = reinterpret_cast<uint8_t*>(&former_node_len)[i];
        }
        return;
    }
    //若前序节点长于254字节，而待修改节点短于254字节，则需要resize store
    else if (former_node_len >= 254 && zlnode->previous_entry_length < 254) {
        temp_length += 4;   //节点变长
        this->store[cur_pos] = (uint8_t)0xFE;
        vector<uint8_t> prev_length_buf;
        for (size_t i = 0; i < sizeof(uint32_t); ++i) {
            prev_length_buf.push_back(reinterpret_cast<uint8_t*>(&former_node_len)[i]);
        }
        this->store.insert(store.begin() + cur_pos + 1, prev_length_buf.begin(), prev_length_buf.end());
        if (pos != this->store.size()) {
            this->setZltail(this->getZltail() + 4);
        }
        this->chain_renew(pos + 1, temp_length);
    }
    else if (former_node_len < 254 && zlnode->previous_entry_length >= 254) {
        temp_length -= 4;   //节点变短  
        this->store[cur_pos] = (uint8_t)former_node_len;
        //清除多余的4个字节
        this->store.erase(store.begin() + cur_pos + 1, store.begin() + cur_pos + 1 + 4);
        if (pos != this->store.size()) {
            this->setZltail(this->getZltail() - 4);
        }
        this->chain_renew(pos + 1, temp_length);
    }
    else {
        //其实此时应该抛错
        return;
    }
}

void ziplist::chain_renew_for_delete(size_t pos, size_t former_node_len) {
    if (pos >= this->store.size()) {
        return;
    }
    ziplist_node* zlnode;
    this->mem2zlnode(pos, zlnode);
    //暂存当前待修改节点的长度和起始位置，方便递归调用
    size_t temp_length = this->get_node_len(pos);
    //若前序节点和待插入节点的长度均小于254字节，则直接修改内存中的prev_length字段
    if (former_node_len < 254 && zlnode->previous_entry_length < 254) {
        this->store[pos] = (uint8_t)former_node_len;
        return;
    }
    //若前序节点和待修改节点均长于254字节
    else if (former_node_len >= 254 && zlnode->previous_entry_length >= 254) {
        for (size_t i = 1; i <= sizeof(uint32_t); i++) {
            this->store[pos + i] = reinterpret_cast<uint8_t*>(&former_node_len)[i];
        }
        return;
    }
    //若前序节点长于254字节，而待修改节点短于254字节，则需要resize store
    else if (former_node_len >= 254 && zlnode->previous_entry_length < 254) {
        temp_length += 4;   //节点变长
        this->store[pos] = (uint8_t)0xFE;
        vector<uint8_t> prev_length_buf;
        for (size_t i = 0; i < sizeof(uint32_t); ++i) {
            prev_length_buf.push_back(reinterpret_cast<uint8_t*>(&former_node_len)[i]);
        }
        store.insert(store.begin() + pos + 1, prev_length_buf.begin(), prev_length_buf.end());
        if (pos != this->store.size()) {
            this->setZltail(this->getZltail() + 4);
        }
        this->chain_renew_for_delete(pos + temp_length, temp_length);
    }
    else if (former_node_len < 254 && zlnode->previous_entry_length >= 254) {
        temp_length -= 4;   //节点变短
        this->store[pos] = (uint8_t)former_node_len;
        //清除多余的4个字节
        this->store.erase(store.begin() + pos + 1, store.begin() + pos + 1 + 4);
        if (pos != this->store.size()) {
            this->setZltail(this->getZltail() - 4);
        }
        this->chain_renew_for_delete(pos + temp_length, temp_length);
    }
    else {
        //其实此时应该抛错
        return;
    }
}

/**
 * 如果没找到，返回nullptr
*/
ziplist_node* ziplist::find(int64_t integer) {
    size_t zl_len = this->getZllen();
    ziplist_node* zlnode = nullptr;
    for (int i = 1; i <= zl_len; i++) {
        zlnode = this->index(i);
        int64_t val = ziplist::get_integer(zlnode);
        if (val == LLONG_MAX) {
            continue;
        }
        if (val == integer) {
            return zlnode;
        }
    }
    return nullptr;
}

//用于测试，输出底层存储全部内容
void ziplist::output_store() {
    for (auto& it : this->store) {
        cout << static_cast<unsigned int>(it) << " ";
    }
    /*cout << endl;
    for (auto& it : this->store) {
        cout << it << " ";
    }*/
    cout << endl << endl;
}

/**
 * 若该节点存储的是字符数组，则返回LLONG_MAX
*/
int64_t ziplist::get_integer(ziplist_node* cur) {
    if (cur->content.size()) {
        return LLONG_MAX;
    }
    return cur->value;
}

/**
 * 若该节点存储的是数字，则返回空vector
*/
vector<uint8_t> ziplist::get_byte_array(ziplist_node* cur) {
    if (!cur->content.size()) {
        return {};
    }
    return cur->content;
}

int ziplist::blob_len() {
    return store.size();
}

int ziplist::len() {
    return this->getZllen();
}

/**
 * 如果cur是最后一个或没找到，则返回nullptr
*/
ziplist_node* ziplist::next(ziplist_node* cur) {
    size_t zl_len = this->getZllen();
    ziplist_node* zlnode = nullptr;
    // content字段为空，说明存储的是value值
    if (cur->content.empty()) {
        int64_t integer = cur->value;
        for (int i = 1; i <= zl_len; i++) {
            zlnode = this->index(i);
            int64_t val = ziplist::get_integer(zlnode);
            if (val == LLONG_MAX) {
                continue;
            }
            if (val == integer) {
                if (i == zl_len) {
                    zlnode = nullptr;
                    break;
                }
                else {
                    zlnode = this->index(i + 1);
                    break;
                }
            }
        }
    }
    // content字段不为空，说明存储的是字符串
    else {
        vector<uint8_t> bytes = cur->content;
        for (int i = 1; i <= zl_len; i++) {
            zlnode = this->index(i);
            vector<uint8_t> cur_content = ziplist::get_byte_array(zlnode);
            if (cur_content.empty()) {
                continue;
            }
            if (equal(cur_content.begin(), cur_content.end(), bytes.begin(), bytes.end())) {
                if (i == zl_len) {
                    zlnode = nullptr;
                    break;
                }
                else {
                    zlnode = this->index(i + 1);
                    break;
                }
            }
        }
    }
    return zlnode;
}

/**
 * 如果cur是第一个或没找到，则返回nullptr
*/
ziplist_node* ziplist::prev(ziplist_node* cur) {
    size_t zl_len = this->getZllen();
    ziplist_node* zlnode = nullptr;
    // content字段为空，说明存储的是value值
    if (cur->content.empty()) {
        int64_t integer = cur->value;
        for (int i = 1; i <= zl_len; i++) {
            zlnode = this->index(i);
            int64_t val = ziplist::get_integer(zlnode);
            if (val == LLONG_MAX) {
                continue;
            }
            if (val == integer) {
                if (i == 1) {
                    zlnode = nullptr;
                    break;
                }
                else {
                    zlnode = this->index(i - 1);
                    break;
                }
            }
        }
    }
    // content字段不为空，说明存储的是字符串
    else {
        vector<uint8_t> bytes = cur->content;
        for (int i = 1; i <= zl_len; i++) {
            zlnode = this->index(i);
            vector<uint8_t> cur_content = ziplist::get_byte_array(zlnode);
            if (cur_content.empty()) {
                continue;
            }
            if (equal(cur_content.begin(), cur_content.end(), bytes.begin(), bytes.end())) {
                if (i == 1) {
                    zlnode = nullptr;
                    break;
                }
                else {
                    zlnode = this->index(i - 1);
                    break;
                }
            }
        }
    }
    return zlnode;
}

// int main() {
//     ziplist* zp = new ziplist();

//     //测试push操作
//     char testPushChar1[] = "hello";
//     zp->push(testPushChar1, sizeof(testPushChar1));
//     int s = 9, bi = 88;
//     zp->push(s);
//     zp->push(bi);
//     // zp->output_store();
//     // 测试字符串节点长度获取
//     // cout<<zp->get_node_len(10)<<endl;
//     // 测试整数节点长度获取
//     // cout<<zp->get_node_len(20)<<endl;
//     // 测试给定一个正向索引，返回其在vector中的位置
//     // cout<<zp->locate_pos(1)<<endl;
//     // cout<<zp->locate_pos(2)<<endl;
//     // cout<<zp->locate_pos(3)<<endl;
//     // 测试mem2zlnode
//     ziplist_node* zlnode = new ziplist_node();

//     //测试查找具有指定值的节点
//     char testPushChar2[] = "testPushChar2";
//     char testNullStr[] = "1235465";
//     zp->push(testPushChar2, sizeof(testPushChar2));
//     zp->push(21);
//     // zlnode = zp->find(testPushChar2, sizeof(testPushChar2));
//     // zlnode = zp->find(testNullStr, sizeof(testNullStr));
//     // if (zlnode) {
//     //     cout<< zlnode->content.data() <<endl;
//     // }
//     // else {
//     //     cout<< "no str"<<endl;
//     // }

//     // zlnode = zp->find(9);
//     // zlnode = zp->find(64);
//     // if(zlnode) {
//     //     cout<< zlnode->value << endl;
//     // }
//     // else {
//     //     cout<<"no number"<<endl;
//     // }

//     // 测试返回指定节点的上一个下一个节点
//     // zlnode = zp->find(9);
//     // zlnode = zp->next(zlnode);
//     // if(zlnode) {
//     //     //output 88
//     //     cout<<zlnode->value<<endl;
//     // }
//     // zlnode = zp->next(zlnode);
//     // if(zlnode) {
//     //     //output testPushChar2
//     //     cout<<zlnode->content.data()<<endl;
//     // }
//     // zlnode = zp->find(21);
//     // zlnode = zp->next(zlnode);
//     // if(zlnode) {
//     //     cout<< zlnode->value <<endl;
//     // }
//     // else {
//     //     //output err
//     //     cout<< "Err!" <<endl;
//     // }
//     // zlnode = zp->find(21);
//     // zlnode = zp->prev(zlnode);
//     // if(zlnode) {
//     //     // output testPushChar2
//     //     cout<< zlnode->content.data() <<endl;
//     // }
//     // zlnode = zp->find(testPushChar1, sizeof(testPushChar1));
//     // zlnode = zp->prev(zlnode);
//     // if(zlnode) {
//     //     cout<< zlnode->content.data() <<endl;
//     // }
//     // else {
//     //     //output err
//     //     cout<< "Err!" <<endl;
//     // }

//     //测试insert
//     // zp->output_store();
//     char testInsert[] = "after second";
//     zp->insert(2, testInsert, sizeof(testInsert));
//     // zp->output_store();

//     char testStartInsert[] = "be first";
//     zp->insert(0, testStartInsert, sizeof(testStartInsert));
//     // zp->output_store();

//     int testInsertInt = 777;
//     zp->insert(4, testInsertInt);
//     // cout<<zp->index(5)->value<<endl;
//     // zp->output_store();
//     // zp->output_node_content();

//     //测试delete_
//     // zlnode = zp->find(777);
//     // if(zlnode == nullptr) {
//     //     cout<<"Err"<<endl;
//     // }
//     // else {
//     //     if (zp->delete_(zlnode) == Ok) {
//     //         zp->output_store();
//     //     }
//     //     else {
//     //         cout<< "Err!" <<endl;
//     //     }
//     // }

//     // zlnode = zp->find(testStartInsert, sizeof(testStartInsert));
//     // if(zlnode == nullptr) {
//     //     cout<<"Err"<<endl;
//     // }
//     // else {
//     //     if (zp->delete_(zlnode) == Ok) {
//     //         zp->output_store();
//     //     }
//     //     else {
//     //         cout<< "Err!" <<endl;
//     //     }
//     // }

//     //测试delete_range
//     zlnode = zp->find(testInsert, sizeof(testInsert));
//     if (zlnode == nullptr) {
//         cout << "Err" << endl;
//     }
//     else {
//         if (zp->delete_range(zlnode, 3) == Ok) {
//             zp->output_node_content();
//             zp->output_store();
//         }
//         else {
//             cout << "Err!" << endl;
//         }
//     }

//     //测试较长字符串的插入
//     char testLongStr[] = "longlonglonglonglonglonglonglong\
//     longlonglonglonglonglonglonglonglonglonglonglonglonglong\
//     longlonglonglonglonglonglonglonglonglonglonglonglonglong\
//     longlonglonglonglonglonglonglonglonglonglonglonglonglong\
//     longlonglonglonglonglonglonglonglonglonglonglonglonglong\
//     longlonglonglonglonglonglonglonglonglonglonglonglonglong";
//     zp->insert(3, testLongStr, sizeof(testLongStr));
//     zp->output_node_content();
//     //zp->output_store();

//     zp->insert(5, 333);
//     zp->output_node_content();
//     //zp->output_store();

//     //测试长字符串删除
//     if (zp->delete_(zp->find(testLongStr, sizeof(testLongStr))) == Ok) {
//         zp->output_node_content();
//     }
//     else {
//         cout<< "Err!" <<endl;
//     }

//     //较短字符串的插入
//     //char testShortStr[] = "shortshort";
//     //zp->insert(4, testShortStr, sizeof(testShortStr));
//     //zp->output_node_content();

//     // zlnode = zp->find(testShortStr, sizeof(testShortStr));
//     // if(zlnode == nullptr) {
//     //     cout<<"Err"<<endl;
//     // }
//     // else {
//     //     if (zp->delete_(zlnode) == Ok) {
//     //         zp->output_node_content();
//     //     }
//     //     else {
//     //         cout<< "Err!" <<endl;
//     //     }
//     // }


//     //测试blob_len()和len()
//     // cout<<zp->blob_len()<<endl;
//     // cout<<zp->len()<<endl;

//     delete zlnode;
//     delete zp;
//     return 0;
// }


ZiplistHandle NewZiplist() {
    return new ziplist();
}

void ReleaseZiplist(ZiplistHandle handle) {
    delete static_cast<ziplist*>(handle);
}

int ZiplistPushBytes(ZiplistHandle handle, char *bytes, int len) {
    return static_cast<ziplist*>(handle)->push(bytes, len);
}

int ZiplistPushInteger(ZiplistHandle handle, int64_t integer) {
    return static_cast<ziplist*>(handle)->push(integer);
}

int ZiplistInsertBytes(ZiplistHandle handle, int pos, char *bytes, int len) {
    return static_cast<ziplist*>(handle)->insert(pos, bytes, len);
}

int ZiplistInsertInteger(ZiplistHandle handle, int pos, int64_t integer) {
    return static_cast<ziplist*>(handle)->insert(pos, integer);
}

ZiplistNodeHandle ZiplistIndex(ZiplistHandle handle, int index) {
    return static_cast<ziplist*>(handle)->index(index);
}

ZiplistNodeHandle ZiplistFindBytes(ZiplistHandle handle, char* bytes, int len) {
    return static_cast<ziplist*>(handle)->find(bytes, len);
}

ZiplistNodeHandle ZiplistFindInteger(ZiplistHandle handle, int64_t integer) {
    return static_cast<ziplist*>(handle)->find(integer);
}

ZiplistNodeHandle ZiplistNext(ZiplistHandle handle, ZiplistNodeHandle currentNode) {
    // 假设 ziplist_node 是指向相应节点的指针，并且 ziplist 类有一个返回下一个节点指针的方法
    return static_cast<ziplist*>(handle)->next(static_cast<ziplist_node*>(currentNode));
}

ZiplistNodeHandle ZiplistPrev(ZiplistHandle handle, ZiplistNodeHandle currentNode) {
    // 假设 ziplist_node 是指向相应节点的指针，并且 ziplist 类有一个返回上一个节点指针的方法
    return static_cast<ziplist*>(handle)->prev(static_cast<ziplist_node*>(currentNode));
}

int64_t ZiplistGetInteger(ZiplistNodeHandle nodeHandle) {
    return ziplist::get_integer(static_cast<ziplist_node*>(nodeHandle));
}

void ZiplistGetByteArray(ZiplistNodeHandle nodeHandle, uint8_t **array, int *len) {
    auto byte_vector = ziplist::get_byte_array(static_cast<ziplist_node*>(nodeHandle));
    *len = static_cast<int>(byte_vector.size());
    *array = new uint8_t[*len];
    std::copy(byte_vector.begin(), byte_vector.end(), *array);
}

int ZiplistDelete(ZiplistNodeHandle nodeHandle) {
    return static_cast<ziplist*>(nodeHandle)->delete_(static_cast<ziplist_node*>(nodeHandle));
}

int ZiplistDeleteRange(ZiplistHandle handle, ZiplistNodeHandle startNodeHandle, int len) {
    return static_cast<ziplist*>(handle)->delete_range(static_cast<ziplist_node*>(startNodeHandle), len);
}

int ZiplistDeleteByPos(ZiplistHandle handle, size_t pos) {
    return static_cast<ziplist*>(handle)->delete_by_pos(pos);
}

int ZiplistBlobLen(ZiplistHandle handle) {
    return static_cast<ziplist*>(handle)->blob_len();
}

int ZiplistLen(ZiplistHandle handle) {
    return static_cast<ziplist*>(handle)->len();
}
