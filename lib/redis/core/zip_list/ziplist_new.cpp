#include <vector>
#include <cstdint>
#include <cstring>
#include <limits>
#include <cassert>
#include <iostream>
#include <cstddef>

using namespace std;

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
    int ba_length; // bit array length ⚠️ 这和书上的定义是不一样的，这里是为了方便写样例
    uint64_t value; // 当压缩列表节点存的是数字时，存在这里面
    // uint8_t content[];
    // c++中不建议使用uint8_t，建议使用vector<uint8_t>
    vector<uint8_t> content; // ⚠️ 在实现时建议换成上面那个uint8_t[]

    ziplist_node(char *bytes, int len);
    ziplist_node(int64_t integer);
    ziplist_node() {};

    // 测试 用于输出zlnode的内容
    void output_zlnode() {
        cout<<endl;
        cout<<"previous_entry_length: "<< previous_entry_length << endl;
        cout<<"encoding: "<< encoding << endl;
        cout<<"ba_length: "<< ba_length << endl;
        if (content.size()) {
            cout<<"content: ";
            for(auto& it : content) {
                cout<< it <<' ';
            }
        }
        else {
            cout<<"value: "<< value << endl;
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

    //get set zlbytes, zltail, zllen
    uint32_t getZlbytes();  //压缩列表占用的内存字节数
    uint32_t getZltail();   //压缩列表表尾节点距离压缩列表的起始地址有多少字节
    uint16_t getZllen();    //压缩列表包含的节点数量

    void setZlbytes(uint32_t zlbytes);
    void setZltail(uint32_t zltail);
    void setZllen(uint16_t zllen);

    /**
     * 用于在push操作中获取最后一个节点的length，
     * 填到新加入的节点的previous_entry_length中
    */
    size_t get_prev_len_for_push();

    /**
     * 获取在vector中起始位置为pos的节点的previous_entry_length
    */
    size_t get_prev_length(size_t pos);

public:
    /**
     * 获取以索引pos作为起始地址的长度
    */
    size_t get_node_len(size_t pos);
    /**
     * 用于给定正向索引index，返回该节点在store中起始节点的位置
    */
    size_t locate_pos(int index);

    /**
     * 底层存储到zlnode结构体
    */
    ZipListResult mem2zlnode(size_t pos, ziplist_node* & zp); 
    /**
     * zlnode结构体到底层存储
    */
    ZipListResult zlnode2mem(ziplist_node zn);

    void output_store();
    ziplist();
    // 将元素插入到表尾
    ZipListResult push(char *bytes, int len);
    ZipListResult push(int64_t integer);

    ZipListResult insert(int pos, char *bytes, int len);
    ZipListResult insert(int pos, int64_t integer);

    // 返回压缩列表给定索引上的节点。
    ziplist_node *index(int n);

    // 查找具有指定值的节点
    ziplist_node *find(char *bytes, int len);
    ziplist_node *find(int64_t integer);

    // 返回指定节点的下一个节点
    ziplist_node *next(ziplist_node *cur);

    // 返回指定节点的上一个节点
    ziplist_node *prev(ziplist_node *cur);

    static int64_t get_integer(ziplist_node *cur);
    static char *get_byte_array(ziplist_node *cur);

    ZipListResult delete_(ziplist_node *cur);
    ZipListResult delete_range(ziplist_node *start, int len);

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
ZipListResult ziplist::mem2zlnode(size_t pos, ziplist_node* & zp) {
    size_t p = pos;
    zp = new ziplist_node();
    zp->previous_entry_length = this->get_node_len(p);

    //定位encoding，并记录prev_length的长度到res中
    if(this->store[p] == (uint8_t)0xFE) {
        p += 5;
    }
    else {
        p += 1;
    }

    //store[pos] & 11000000 = 11000000说明是整数
    if(((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0xC0) {
        zp->encoding = (uint8_t)store[p];
        uint8_t encoding = store[p];
        p += 1;
        if (encoding & ZIP_INT_4b == ZIP_INT_4b) {
            //整数为0-12，content长度为1，只有低4位是有效的
            zp->value = encoding & 0x0F;
            zp->ba_length = 0;
        }
        else if(encoding == ZIP_INT_8B) {
            //8位整数
            zp->value = (uint64_t)store[p];
            zp->ba_length = 1;
        }
        else if(encoding == ZIP_INT_16B) {
            //16位整数
            zp->value = static_cast<uint16_t>(store[p]) |
           static_cast<uint16_t>(store[p + 1]) << 8;
           zp->ba_length = 2;

        }
        else if(encoding == ZIP_INT_24B) {
            //24位整数
            zp->value = static_cast<uint32_t>(store[p]) |
           static_cast<uint32_t>(store[p + 1]) << 8 |
           static_cast<uint32_t>(store[p + 2]) << 16;
           zp->ba_length = 3;
        }
        else if(encoding == ZIP_INT_32B) {
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
    else if(((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x00) {
        //encoding长度为1
        //获取后6位，即为字节的长度，并更新ba->length
        size_t len = store[p] & 0x3F;
        zp->ba_length = len;
        p += 1;
        for(size_t i = 0; i < len; i++) {
            zp->content.push_back(store[p+i]);
        }
    }
    //encoding长度2字节，字节数组长度小于16838字节
    else if(((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x40) {
        //encoding长度为2
        //获取后14位，即为字节的长度，并更新ba->length
        size_t len = ((store[p] & 0x3F) << 8) | store[p+1];
        zp->ba_length = len;
        p += 2;
        for(size_t i = 0; i < len; i++) {
            zp->content.push_back(store[p+i]);
        }
    }
    //encoding长度5字节，字节数组长度大于16838字节
    else if(((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x80) {
        //encoding长度为5
        //获取后32位，即为字节的长度，并更新res
        size_t len = ((uint64_t)store[p+1] << 24) |
                    ((uint64_t)store[p+2] << 16) |
                    ((uint64_t)store[p+3] << 8)  |
                    ((uint64_t)store[p+4]);
        p += 5;
        for(size_t i = 0; i < len; i++) {
            zp->content.push_back(store[p+i]);
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
    if(this->getZllen() == 0) {
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

ZipListResult ziplist::push(char *bytes, int len)
{
    //构造新插入节点的previous_entry_length
    size_t prev_length = this->get_prev_len_for_push();
    prev_length = (uint32_t) prev_length;
    uint8_t prev_length_buf[5];
    int prev_length_len = 0;    //新插入节点的prev_length的长度
    //前序节点长度位于0-254之间，previous_entry_length长为1字节
    if(prev_length > 0 && prev_length < 254) {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)prev_length;
    }
    //前序节点长度大于254，previous_entry_length长为5字节
    else if(prev_length>=254) {
        prev_length_len = 5;
        prev_length_buf[0] = (uint8_t)0xFE;
        for (size_t i = 0; i < sizeof(uint32_t); ++i) {
            prev_length_buf[i+1] = reinterpret_cast<uint8_t*>(&prev_length)[i];
        }
    }
    //前序节点长度为0，该节点是ziplist的第一个节点
    else {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)0;
    }
    //写入previous_entry_length
    for(int i = 0; i<prev_length_len; i++) {
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
    for(int i = 0; i<encoding_len; i++) {
        this->store.push_back(buf[i]);
    }
    /*将字符串本身写入store中*/
    for(int i = 0; i<len; i++) {
        this->store.push_back(*(bytes+i));
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
    if(prev_length > 0 && prev_length < 254) {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)prev_length;
    }
    //前序节点长度大于254，previous_entry_length长为5字节
    else if(prev_length>=254) {
        prev_length_len = 5;
        prev_length_buf[0] = (uint8_t)0xFE;
        for (size_t i = 0; i < sizeof(uint32_t); ++i) {
            prev_length_buf[i+1] = reinterpret_cast<uint8_t*>(&prev_length)[i];
        }
    }
    //前序节点长度为0，该节点是ziplist的第一个节点
    else {
        prev_length_len = 1;
        prev_length_buf[0] = (uint8_t)0;
    }
    //写入previous_entry_length
    for(int i = 0; i<prev_length_len; i++) {
        store.push_back(prev_length_buf[i]);
    }

    uint8_t encoding = 0;
    if (integer >= 0 && integer <= 12) {
        encoding = ZIP_INT_4b + integer;
    } else if (integer >= INT8_MIN && integer <= INT8_MAX) {
        encoding = ZIP_INT_8B;
    } else if (integer >= INT16_MIN && integer <= INT16_MAX) {
        encoding = ZIP_INT_16B;
    } else if (integer >= INT24_MIN && integer <= INT24_MAX) {
        encoding = ZIP_INT_24B;
    } else if (integer >= INT32_MIN && integer <= INT32_MAX) {
        encoding = ZIP_INT_32B;
    } else {
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
 * 此处pos指的是能直接放在store[pos]的索引，不是从1开始的位置
*/
size_t ziplist:: get_node_len(size_t pos) {
    size_t res = 0; //返回值
    size_t p = pos;
    //定位encoding，并记录prev_length的长度到res中
    if(this->store[p] == (uint8_t)0xFE) {
        p += 5;
        res += 5;
    }
    else {
        p += 1;
        res += 1;
    }

    //store[pos] & 11000000 = 11000000说明是整数
    //注意==的优先级高于&，要加括号
    if(((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0xC0) {
        uint8_t encoding = store[p];
        res++;  //记录encoding的长度到res中
        if (encoding & ZIP_INT_4b == ZIP_INT_4b) {
            //整数为0-12，什么也不做，没有content
        }
        else if(encoding == ZIP_INT_8B) {
            //8位整数，节点大小+1
            res += 1;
        }
        else if(encoding == ZIP_INT_16B) {
            //16位整数，节点大小+2
            res += 2;
        }
        else if(encoding == ZIP_INT_24B) {
            //24位整数，节点大小+3
            res += 3;
        }
        else if(encoding == ZIP_INT_32B) {
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
    else if(((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x00) {
        //encoding长度为1，更新res
        res += 1;
        //获取后6位，即为字节的长度，并更新res
        int len = store[p] & 0x3F;
        res += len;
    }
    //encoding长度2字节，字节数组长度小于16838字节
    else if(((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x40) {
        //encoding长度为2，更新res
        res += 2;
        //获取后14位，即为字节的长度，并更新res
        int len = ((store[p] & 0x3F) << 8) | store[p+1];
        res += len;
    }
    //encoding长度5字节，字节数组长度大于16838字节
    else if(((uint8_t)store[p] & (uint8_t)0xC0) == (uint8_t)0x80) {
        //encoding长度为5，更新res
        res += 5;
        //获取后32位，即为字节的长度，并更新res
        size_t len = ((uint64_t)store[p+1] << 24) |
                    ((uint64_t)store[p+2] << 16) |
                    ((uint64_t)store[p+3] << 8)  |
                    ((uint64_t)store[p+4]);
        res += len;
    }
    
    return res;
}

size_t ziplist::get_prev_length(size_t pos) {
    size_t res = 0; //返回值
    size_t p = pos;
    if(this->store[p] == (uint8_t)0xFE) {
        // 对于小端字节序：
        res = ((uint32_t)store[p+4] << 24) |
                    ((uint32_t)store[p+3] << 16) |
                    ((uint32_t)store[p+2] << 8)  |
                    ((uint32_t)store[p+1]);
    }   
    else {
        res = (size_t)this->store[p];
    }
    return res;
}

/**
 * 用于给定正向索引index，返回该节点在store中起始节点的位置
*/
size_t ziplist::locate_pos(int index) {
    if (index <= 0) {
        return 0;
    }
    uint16_t len = this->getZllen();
    uint32_t pos = this->getZltail();
    //若列表为空，直接返回0
    if(len == 0) {
        return 0;
    }
    int rev_index = len - index;
    for(int i = 0; i<rev_index; i++) {
        size_t prev_length = this->get_prev_length(pos);
        pos -= prev_length;
    }
    return pos;
}


//用于测试，输出底层存储全部内容
void ziplist::output_store() {
    for(auto& it : this->store) {
        cout << static_cast<unsigned int>(it) << " ";
    }
    cout<<endl;
    // cout<<endl<<endl;
    // for(auto& it : this->store) {
    //     cout << it << " ";
    // }
}

int main() {
    ziplist* zp = new ziplist();

    // for(auto& it : zp->store) {
    //     cout << static_cast<unsigned int>(it) << " ";
    // }
    // cout<<endl;
    // cout<<zp->getZlbytes()<<endl;
    // cout<<zp->getZllen()<<endl;
    // cout<<zp->getZltail()<<endl;
    // cout<<endl;
    // zp->setZlbytes(66);
    // zp->setZllen(55);
    // zp->setZltail(44);
    // cout<<zp->getZlbytes()<<endl;
    // cout<<zp->getZllen()<<endl;
    // cout<<zp->getZltail()<<endl;
    
    //测试push操作
    char testPushChar1[] = "hello"; 
    zp->push(testPushChar1, sizeof(testPushChar1));
    // char testPushChar2[] = "216549asdfkpaweigjpoiaajsoighpoawwjoeifhgqeorijgaspofidhaoiwpejsghppwaehijgpafidkn"; 
    // zp->push(testPushChar2, sizeof(testPushChar2));
    int s = 9, bi = 88;
    zp->push(s);
    zp->push(bi);
    zp->output_store();
    // 测试字符串节点长度获取
    // cout<<zp->get_node_len(10)<<endl;
    // 测试整数节点长度获取
    // cout<<zp->get_node_len(20)<<endl;
    // 测试给定一个正向索引，返回其在vector中的位置
    cout<<zp->locate_pos(1)<<endl;
    cout<<zp->locate_pos(2)<<endl;
    cout<<zp->locate_pos(3)<<endl;
    // 测试mem2zlnode
    ziplist_node* zlnode = new ziplist_node();
    // if (zp->mem2zlnode(zp->locate_pos(1), zlnode) == Ok) {
    //     zlnode->output_zlnode();
    // }
    if (zp->mem2zlnode(zp->locate_pos(1), zlnode) == Ok) {
        zlnode->output_zlnode();
    }
    else {
        cout<<"Err!"<<endl;
    }
    delete zlnode;
    delete zp;
    return 0;
}