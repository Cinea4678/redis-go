#include <vector>
#include <cstdint>
using namespace std;

#define Ok 0
#define Err 1

typedef int ZipListResult;

// ⚠️ 压缩列表的节点不能照抄这段struct，应该照书上来（例如，书上说previous_entry_length的长度是可变的，它可能只占1个字节，也可能占5个字节
struct ziplist_node
{
    int previous_entry_length;
    int encoding;
    int content[];
};

class ziplist
{
private:
    vector<ziplist_node> store;

public:
    ziplist();
    ZipListResult push(char *bytes, int len);
    ZipListResult push(int64_t integer);

    ZipListResult insert(int pos, char *bytes, int len);
    ZipListResult insert(int pos, int64_t integer);

    // 返回压缩列表给定索引上的节点。
    ziplist_node &index(int n);

    // 查找具有指定值的节点
    ziplist_node &find(char *bytes, int len);
    ziplist_node &find(int64_t integer);

    // 返回指定节点的下一个节点
}