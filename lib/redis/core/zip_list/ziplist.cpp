#include <vector>
#include <cstdint>
#include <cstring>
using namespace std;

#define Ok 0
#define Err 1

typedef int ZipListResult;

// ⚠️ 压缩列表的节点不能照抄这段struct，应该照书上来（例如，书上说previous_entry_length的长度是可变的，它可能只占1个字节，也可能占5个字节
// ⚠️ 实现建议：在底层不存储这个struct。可以实现两个把这个struct和uint8_t[]互相转换的函数
struct ziplist_node
{
    int previous_entry_length;
    int encoding;
    int ba_length; // bit array length ⚠️ 这和书上的定义是不一样的，这里是为了方便写样例
    // uint8_t content[];
    vector<uint8_t> content; // ⚠️ 在实现时建议换成上面那个uint8_t[]

    ziplist_node(char *bytes, int len);
    ziplist_node(int64_t integer);
};

class ziplist
{
private:
    vector<ziplist_node> store; // ⚠️ 建议使用 vector<uint8_t> 而不是 uint8_t* 这种C风格的东西，可参考intset的用法

public:
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

/**
 * 样例实现
 * ⚠️ 下面的代码需要重新照着书按标准实现
 */

ziplist_node::ziplist_node(char *bytes, int len)
{
    encoding = 0x00000000;
    ba_length = len;
    content.resize(len);
    memcpy(content.data(), bytes, len);
}

ziplist_node::ziplist_node(int64_t integer)
{
    encoding = 0x11000000;
    content.resize(4);
    memcpy(content.data(), (void *)&integer, 4);
}

ziplist::ziplist()
{
}

ZipListResult ziplist::push(char *bytes, int len)
{
    store.emplace_back(bytes, len);
    return Ok;
}

ZipListResult ziplist::push(int64_t integer)
{
    store.emplace_back(integer);
    return Ok;
}

ZipListResult ziplist::insert(int pos, char *bytes, int len)
{
    auto iter = store.begin();
    advance(iter, pos);
    store.insert(iter, ziplist_node(bytes, len));
    return Ok;
}

ZipListResult ziplist::insert(int pos, int64_t integer)
{
    auto iter = store.begin();
    advance(iter, pos);
    store.insert(iter, ziplist_node(integer));
    return Ok;
}

ziplist_node *ziplist::index(int n)
{
    return &store[n];
}

ziplist_node *ziplist::find(char *bytes, int len)
{
    for (auto &node : store)
    {
        if (node.encoding & 0x11000000 == 0 && node.ba_length == len && memcmp(node.content.data(), bytes, len) == 0)
        {
            return &node;
        }
    }
    return nullptr;
}

ziplist_node *ziplist::find(int64_t integer)
{
    for (auto &node : store)
    {
        if (node.encoding & 0x11000000 > 0 && memcmp(node.content.data(), (void *)&integer, 4))
        {
            return &node;
        }
    }
    return nullptr;
}

ziplist_node *ziplist::next(ziplist_node *cur)
{
    int index = cur - store.data() - 1;
    if (index < 0)
    {
        return nullptr;
    }
    return &store[index];
}

ziplist_node *ziplist::prev(ziplist_node *cur)
{
    int index = cur - store.data() + 1;
    if (index >= store.size())
    {
        return nullptr;
    }
    return &store[index];
}
