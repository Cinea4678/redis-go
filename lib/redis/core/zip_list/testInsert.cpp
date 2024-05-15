#include <vector>
#include <iostream>
#include <algorithm> // for std::copy

void insertIntoCompressedList(std::vector<uint8_t>& store, const std::vector<uint8_t>& new_node, size_t position) {
    // 检查插入位置是否有效
    if (position > store.size()) {
        std::cerr << "Invalid position" << std::endl;
        return;
    }

    // 扩展store的大小以容纳新节点
    store.resize(store.size() + new_node.size());

    // 将从position开始的旧元素向后移动new_node.size()个位置
    std::move_backward(store.begin() + position, store.end() - new_node.size(), store.end());

    // 复制new_node到store的指定位置
    std::copy(new_node.begin(), new_node.end(), store.begin() + position);
}

int main() {
    // 示例使用
    std::vector<uint8_t> store = {1, 2, 3, 4, 5}; // 假设的压缩列表
    std::vector<uint8_t> new_node = {10, 11, 12}; // 要插入的新节点

    size_t insertPosition = 2; // 插入位置

    insertIntoCompressedList(store, new_node, insertPosition);

    // 打印结果
    for (auto elem : store) {
        std::cout << (int)elem << " ";
    }
    std::cout << std::endl;

    return 0;
}
