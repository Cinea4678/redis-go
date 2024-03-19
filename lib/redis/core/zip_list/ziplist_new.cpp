#include <vector>
#include <cstdint>
#include <cstring>
#include <limits>
#include <cassert>
#include <iostream>

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
    uint8_t content[];
    // vector<uint8_t> content; // ⚠️ 在实现时建议换成上面那个uint8_t[]

    ziplist_node(char *bytes, int len);
    ziplist_node(int64_t integer);
    ZipListResult mem2zlnode(size_t pos);     //底层存储到zlnode结构体
    ZipListResult zlnode2mem(ziplist_node zn);               //zlnode结构体到底层存储
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

public:

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

//用于测试，输出底层存储全部内容
void ziplist::output_store() {
    for(auto& it : this->store) {
        cout << static_cast<unsigned int>(it) << " ";
    }
}

/****************************************************/
/**
 * 将大小为bufSize的一段buf写到store中的指定位置
*/
void static insertEntry(vector<uint8_t>& store, size_t pos, const char* buf, size_t bufSize) {
    // 确保pos不会超出store的当前大小
    pos = min(pos, store.size());
    // 扩展store的大小以容纳新条目
    store.resize(store.size() + bufSize);
    // 如果是在vector的中间或开头插入，需要移动现有的元素来为新元素腾出空间
    if (pos < store.size() - bufSize) {
        move_backward(store.begin() + pos, store.end() - bufSize, store.end());
    }
    // 复制buf到store中的指定位置
    copy(buf, buf + bufSize, store.begin() + pos);
}

/* 
 * 以 encoding 指定的编码方式，将整数值 value 写入到 p 。
 */
static void zipSaveInteger(unsigned char *p, int64_t value, unsigned char encoding) {
    int16_t i16;
    int32_t i32;
    int64_t i64;

    //TODO 大小端序
    if (encoding == ZIP_INT_8B) {
        ((int8_t*)p)[0] = (int8_t)value;
    } else if (encoding == ZIP_INT_16B) {
        i16 = value;
        memcpy(p,&i16,sizeof(i16));
        // memrev16ifbe(p);
    } else if (encoding == ZIP_INT_24B) {
        i32 = value<<8;
        // memrev32ifbe(&i32);
        memcpy(p,((uint8_t*)&i32)+1,sizeof(i32)-sizeof(uint8_t));
    } else if (encoding == ZIP_INT_32B) {
        i32 = value;
        memcpy(p,&i32,sizeof(i32));
        // memrev32ifbe(p);
    } else if (encoding == ZIP_INT_64B) {
        i64 = value;
        memcpy(p,&i64,sizeof(i64));
        // memrev64ifbe(p);
    } else if (encoding >= ZIP_INT_IMM_MIN && encoding <= ZIP_INT_IMM_MAX) {
        /* Nothing to do, the value is stored in the encoding itself. */
    } else {
        assert(NULL);
    }
}

/* Convert a string into a long long. Returns 1 if the string could be parsed
 * into a (non-overflowing) long long, 0 otherwise. The value will be set to
 * the parsed value when appropriate. */
int string2ll(const char *s, size_t slen, long long *value) {
    const char *p = s;
    size_t plen = 0;
    int negative = 0;
    unsigned long long v;

    if (plen == slen)
        return 0;

    /* Special case: first and only digit is 0. */
    if (slen == 1 && p[0] == '0') {
        if (value != NULL) *value = 0;
        return 1;
    }

    if (p[0] == '-') {
        negative = 1;
        p++; plen++;

        /* Abort on only a negative sign. */
        if (plen == slen)
            return 0;
    }

    /* First digit should be 1-9, otherwise the string should just be 0. */
    if (p[0] >= '1' && p[0] <= '9') {
        v = p[0]-'0';
        p++; plen++;
    } else if (p[0] == '0' && slen == 1) {
        *value = 0;
        return 1;
    } else {
        return 0;
    }

    while (plen < slen && p[0] >= '0' && p[0] <= '9') {
        if (v > (ULLONG_MAX / 10)) /* Overflow. */
            return 0;
        v *= 10;

        if (v > (ULLONG_MAX - (p[0]-'0'))) /* Overflow. */
            return 0;
        v += p[0]-'0';

        p++; plen++;
    }

    /* Return if not all bytes were used. */
    if (plen < slen)
        return 0;

    if (negative) {
        if (v > ((unsigned long long)(-(LLONG_MIN+1))+1)) /* Overflow. */
            return 0;
        if (value != NULL) *value = -v;
    } else {
        if (v > LLONG_MAX) /* Overflow. */
            return 0;
        if (value != NULL) *value = v;
    }
    return 1;
}

/* 
 * 检查 entry 中指向的字符串能否被编码为整数。
 * 如果可以的话，
 * 将编码后的整数保存在指针 v 的值中，并将编码的方式保存在指针 encoding 的值中。
 * 注意，这里的 entry 和前面代表节点的 entry 不是一个意思。
 */
static int zipTryEncoding(uint8_t *entry, uint32_t entrylen, uint64_t *v, uint8_t *encoding) {
    //负责存储value
    long long value;

    // 忽略太长或太短的字符串
    if (entrylen >= 32 || entrylen == 0) 
        return Err;

    // 尝试转换
    if (string2ll((char*)entry,entrylen,&value)) {

        /* Great, the string can be encoded. Check what's the smallest
         * of our encoding types that can hold this value. */
        // 转换成功，以从小到大的顺序检查适合值 value 的编码方式
        if (value >= 0 && value <= 12) {
            *encoding = ZIP_INT_IMM_MIN+value;
        } else if (value >= INT8_MIN && value <= INT8_MAX) {
            *encoding = ZIP_INT_8B;
        } else if (value >= INT16_MIN && value <= INT16_MAX) {
            *encoding = ZIP_INT_16B;
        } else if (value >= INT24_MIN && value <= INT24_MAX) {
            *encoding = ZIP_INT_24B;
        } else if (value >= INT32_MIN && value <= INT32_MAX) {
            *encoding = ZIP_INT_32B;
        } else {
            *encoding = ZIP_INT_64B;
        }

        // 记录值到指针
        *v = value;

        // 返回转换成功标识
        return Ok;
    }

    // 转换失败
    return Err;
}



/* 
 * 以 encoding 指定的编码方式，读取并返回指针 p 中的整数值。
 */
static int64_t zipLoadInteger(unsigned char *p, unsigned char encoding) {
    int16_t i16;
    int32_t i32;
    int64_t i64, ret = 0;

    //TODO 大小端序
    if (encoding == ZIP_INT_8B) {
        ret = ((int8_t*)p)[0];
    } else if (encoding == ZIP_INT_16B) {
        memcpy(&i16,p,sizeof(i16));
        // memrev16ifbe(&i16);
        ret = i16;
    } else if (encoding == ZIP_INT_32B) {
        memcpy(&i32,p,sizeof(i32));
        // memrev32ifbe(&i32);
        ret = i32;
    } else if (encoding == ZIP_INT_24B) {
        i32 = 0;
        memcpy(((uint8_t*)&i32)+1,p,sizeof(i32)-sizeof(uint8_t));
        // memrev32ifbe(&i32);
        ret = i32>>8;
    } else if (encoding == ZIP_INT_64B) {
        memcpy(&i64,p,sizeof(i64));
        // memrev64ifbe(&i64);
        ret = i64;
    } else if (encoding >= ZIP_INT_IMM_MIN && encoding <= ZIP_INT_IMM_MAX) {
        ret = (encoding & ZIP_INT_IMM_MASK)-1;
    } else {
        assert(NULL);
    }

    return ret;
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
    char testPushChar[] = "hello"; 
    zp->push(testPushChar, sizeof(testPushChar));
    int s = 9, bi = 88;
    zp->push(s);
    zp->push(bi);
    zp->output_store();
    delete zp;
    return 0;
}