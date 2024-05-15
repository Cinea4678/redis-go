#include<iostream>
#include<vector>
#include <algorithm>

using namespace std;

void insertEntry(std::vector<uint8_t>& store, size_t pos, const char* buf, size_t bufSize) {
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

int main() {
    vector<int> val(10);
    val.resize(15);

    for(int i = 1; i<=15; i++) {
        val[i-1] = i;
    }

    for(auto c :val) {
        cout<<c<<' ';
    }
    cout<<endl<<val.size()<<endl;

    // vector<int>::iterator it = val.begin();
    for(auto it = val.begin(); it!=val.end(); it++) {
        cout<<*it<<' ';
    }
    val.push_back(111);
    
    cout<<val.front()<<endl;    //获取第一个元素
    cout<<val.back()<<endl;     //获取最后一个元素

    //也可以通过下标获取
    cout<<val[3]<<endl;
    //at操作和下标操作一样
    cout<<val.at(3)<<endl;

    //迭代器
    auto it = val.begin();
    //插入元素
    val.insert(it+2, 222);
    cout<<val[2]<<endl;
    //删除元素
    it = val.begin();   //最好在插入或删除元素之后，重新获取一下迭代器
    val.erase(it);      //erase也可以删除一定范围内的值val.earse(val.begin(), val.end())
    for(auto it = val.begin(); it!=val.end(); it++) {
        cout<<*it<<' ';
    }
    cout<<endl;
    val.clear();//清空

    /*
    *现在，我有一个需求。我用C++的vector<uint8_t> store
    *来模拟压缩列表的底层存储结构，如果我想实现一个压缩列表的插入操作，
    *那么首先我需要先构建出一块char buf[]来暂时存储当前的压缩列表节点，
    *那么，我需要如何运用resize等操作，将这块char buf的内容写入到vector数组store中
    */

    vector<uint8_t> store = {1, 2, 3, 4, 5}; // 假设这是已有的存储
    char buf[] = {6, 7, 8}; // 要插入的新条目

    // 插入操作，将buf插入到store的末尾
    insertEntry(store, 2, buf, sizeof(buf)/sizeof(buf[0]));

    // 打印结果
    for (auto byte : store) {
        cout << (int)byte << " ";
    }
    cout << endl;

    return 0;

}
