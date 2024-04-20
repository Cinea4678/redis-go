#include "skip_list.h"

inline void SkipListNode::AddSibling(ZSetType value) {
    SkipListSiblingNode* last = this->nextSibling;
    // 若目前暂无Sibling，则直接插入并更新首节点的nextSibling指针
    if (!last) {
        this->nextSibling = new SkipListSiblingNode(value);
        return;
    }
    // 如果有Sibling，则按照链表push_back操作
    while (last->nextSibling) {
        last = last->nextSibling;
    }
    last->nextSibling = new SkipListSiblingNode(value);
    return;
}

void SkipListNode::print() const {
    cout << "[SLNode] s=" << score << ", v=" << value << endl;

    // 打印重复score的节点
    if (nextSibling) {
        SkipListSiblingNode* s = nextSibling;
        cout << "[";
        while (s) {
            cout << ", v=" << s->value;
            s = s->nextSibling;
        }
        cout << "]" << endl;
    }

    // for (int i = 1; i < SKIP_LIST_MAX_LEVEL; i++) {
    for (int i = 0; i < forward.size(); i++) {
        // 如果越界了，检查可能不为空，但根本不是forward节点
        if (!forward[i]) {
            break;
        }
        cout << "forward[" << i << "]: " << forward[i]->value << " "
             << forward[i]->score << " sp=" << span[i] << ", ";
    }
    if (backward) {
        cout << "backward: " << backward->score;
    }
    cout << endl;
}

// 随机生成节点层数
// 50%概率加层
// 少量高层节点快速跳过大部分低层节点
ZSetSizeType SkipList::randomLevel() const {
    int lvl = 1;
    while ((rand() & 1) && lvl < SKIP_LIST_MAX_LEVEL) {
        lvl++;
    }
    return lvl;
}

void SkipList::insert(double score, ZSetType value) {
    cout << "Inserting [" << score << "] " << value << endl;

    // newNodeLevel<level时：0~nL为新节点需更新的直接前驱；nL~l为“从新节点头顶跨过”的需要更新span的
    // newNodeLevel>=level时：0~l为新节点需更新的直接前驱，l~nL无需存储直接对header更新
    vector<SkipListNode*> update(level, nullptr);
    // 存储update[i]到header的距离
    vector<ZSetSizeType> updateRank(level, 0);

    SkipListNode* cur = header;
    // cur到header的距离
    ZSetSizeType dist = 0;

    // 从最高层开始，查找插入位置
    // 渐进式的查找，下一层查找起点在上一层终点
    for (int i = level - 1; i >= 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            dist += cur->span[i];
            cur = cur->forward[i];
        }
        // 如果level比newNodeLevel高，那么高出来的部分也要记录(需要更新跨度)
        update[i] = cur;
        updateRank[i] = dist;
    }
    // 经过上面的循环后，cur抵达小于score的最后一个Node(其第一层后继节点大于等于score)

    // 如果有重复score
    if (cur->forward[0] && cur->forward[0]->score == score) {
        // 如果判定有相等的，则先抵达score对应的Node
        cur = cur->forward[0];
        // 插入兄弟节点
        cur->AddSibling(value);

        // 更新跨度

        int curLevel = cur->forward.size();
        // 每一个level的span++
        for (int i = 0; i < curLevel; i++) {
            cur->span[i]++;
        }

        // 如果当前节点层数小于目前最大层数，则“从cur顶上跨过的forward”的跨度需要++
        for (int i = curLevel; i < level; i++) {
            if (update[i]->forward[i]) { // 如果forward不指向null
                update[i]->span[i]++;
            }
        }
        return;
    } // 直接return，无需更新其他forward

    // 如果无重复score
    // 此时cur位置：newnode应该在的位置的直接前驱节点
    // 新节点层数
    const ZSetType newNodeLevel = randomLevel();
    cout << "level = " << newNodeLevel << endl;
    SkipListNode* newNode = new SkipListNode(score, value, newNodeLevel);

    if (newNodeLevel > level) {
        // 如果新节点层数大于目前最大层数:
        // level以下的部分正常更新，多出来的部分指向header
        for (int i = 0; i < level; i++) {
            newNode->forward[i] = update[i]->forward[i];
            newNode->span[i] = dist - updateRank[i] + 1;

            update[i]->forward[i] = newNode;
        }

        for (int i = level; i < newNodeLevel; i++) {
            // newNode->forward[i] = nullptr, newNode->span[i] = 0; // 本来就是
            header->forward[i] = newNode;
            header->span[i] = dist + 1;
        }
        level = newNodeLevel;
    } else {
        // 如果新节点层数小于目前最大层数:
        // newNodeLevel以下的部分正常更新，nL~l的部分需要更新跨度
        for (int i = 0; i < newNodeLevel; i++) {
            newNode->forward[i] = update[i]->forward[i];
            newNode->span[i] = dist - updateRank[i] + 1;

            update[i]->forward[i] = newNode;
        }

        // 如果新节点层数小于目前最大层数，则“从newNode顶上跨过的forward”的跨度需要++
        for (int i = newNodeLevel; i < level; i++) {
            if (update[i]->forward[i]) { // 如果forward不指向null
                update[i]->span[i]++;
            }
        }
    }

    // 设置后向指针：由于修改了update存储的长度，所以不能还是单纯存update[0]了
    newNode->backward = update.size() > 0 ? update[0] : header;

    // 设置后继节点的后向指针
    if (newNode->forward[0]) {
        newNode->forward[0]->backward = newNode;
    } else {
        // 如果后继为null，则设置尾节点
        tail = newNode;
    }

    length++;
}

SkipListNode* SkipList::searchNode(double score) {
    SkipListNode* cur = header;

    for (int i = level - 1; i >= 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            cur = cur->forward[i];
        }
    }
    cur = cur->forward[0];

    // Score not found
    if (!cur || cur->score != score) {
        return nullptr;
    }
    // Score found
    return cur;
}

vector<SkipListNode*> SkipList::searchRangeNode(double lscore, double rscore) {
    vector<SkipListNode*> result;
    SkipListNode* cur = header;
    for (int i = level - 1; i >= 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < lscore) {
            cur = cur->forward[i];
        }
    }
    // cur抵达小于lscore的最后一个节点

    // 从>=lscore的第一个节点开始，按顺序找直系后继
    while (cur->forward[0] && cur->forward[0]->score < rscore) {
        cur = cur->forward[0];
        result.push_back(cur);
    }

    return result;
}

SkipListNode* SkipList::removeNode(double score) {
    vector<SkipListNode*> update(level, nullptr);
    SkipListNode* cur = header;

    // 寻找所有层的前驱节点，应该从1开始，而不是0
    for (int i = level; i > 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            cur = cur->forward[i];
        }
        update[i] = cur;
    }

    // 寻找同score的链表的起始节点
    cur = cur->forward[0];

    // 检查是否有这个score
    if (!cur || cur->score != score) {
        return nullptr;
    }

    // 循环删除所有兄弟节点
    SkipListSiblingNode* last = cur->nextSibling;
    while (last) {
        SkipListSiblingNode* nxt = last->nextSibling;
        delete last;
        last = nxt;
    }

    // 从第一层往上找（因为一定是较低层有前驱节点而较高层无
    // 此处是用判定update[i]->forward[i]是否为cur的方式来确认cur是否为某个节点的后继节点的
    for (int i = 0; i < level; i++) {
        if (update[i]->forward[i] != cur)
            break;
        update[i]->forward[i] = cur->forward[i];
    }

    delete cur;

    // 更新backward和tail指针
    // 如果对第一层而言，删除节点为最后一个节点，则更新tail
    if (update[0]->forward[0] == nullptr) {
        tail = update[0];
    }
    // 如果删除前存在直接后继，则更新其的backward指针
    if (update[0]->forward[0]) {
        update[0]->forward[0]->backward = update[1];
    }

    // 更新跳表层数
    while (level > 1 && header->forward[level] == nullptr) {
        level--;
    }

    return cur;
}

vector<ZSetType> SkipList::search(double score) {
    vector<ZSetType> result;
    SkipListNode* cur = searchNode(score);
    if (!cur) { // not found
        return result;
    }

    // 遍历
    result.push_back(cur->value);
    SkipListSiblingNode* last = cur->nextSibling;
    while (last) {
        result.push_back(last->value);
        last = last->nextSibling;
    }
    return result;
}

vector<pair<double, ZSetType>> SkipList::searchRange(double lscore,
                                                     double rscore) {
    vector<pair<double, ZSetType>> result;
    vector<SkipListNode*> r = searchRangeNode(lscore, rscore);
    if (r.empty()) { // not found
        return result;
    }

    // 对每一个节点遍历
    for (SkipListNode* cur : r) {
        result.push_back({cur->score, cur->value});
        SkipListSiblingNode* last = cur->nextSibling;
        while (last) {
            result.push_back({cur->score, last->value});
            last = last->nextSibling;
        }
    }
    return result;
}

vector<ZSetType> SkipList::remove(double score) {
    vector<ZSetType> result;
    SkipListNode* cur = removeNode(score);
    if (!cur) { // not found
        return result;
    }

    // 遍历
    result.push_back(cur->value);
    SkipListSiblingNode* last = cur->nextSibling;
    while (last) {
        result.push_back(last->value);
        last = last->nextSibling;
    }
    return result;
}

bool SkipList::remove(double score, ZSetType value) {
    vector<SkipListNode*> update(level, nullptr);
    SkipListNode* cur = header;

    // 寻找所有层的前驱节点，应该从1开始，而不是0
    for (int i = level; i > 0; i--) {
        while (cur->forward[i] && cur->forward[i]->score < score) {
            cur = cur->forward[i];
        }
        update[i] = cur;
    }

    // 寻找同score的链表的起始节点
    cur = cur->forward[0];

    // 检查是否有这个score
    if (!cur || cur->score != score) {
        return false;
    }

    // 如果首节点就是对应value，则需要判断其是否有兄弟节点
    //      如果有兄弟节点，则仅需将firstSibling晋升为Node(对cur节点的value、nextSibling进行赋值并删除firstSibling)
    //      如果无，则直接删除并连接
    // 如果首节点非对应value，则遍历Sibling后进行简单的链表删除(理论上在上层判断，此处不会有找不到的情况，不过还是处理了)

    // 如果首节点就是对应value
    if (cur->value == value) {
        SkipListSiblingNode* firstSibling = cur->nextSibling;
        if (firstSibling) { // 有兄弟节点
            SkipListNode* result = cur;
            cur->value = value;
            cur->nextSibling = firstSibling->nextSibling;

            delete firstSibling;
            return true;
        }
        // 无兄弟节点，则与只传入score的重载的操作一致
        for (int i = 0; i < level; i++) {
            if (update[i]->forward[i] != cur)
                break;
            update[i]->forward[i] = cur->forward[i];
        }
        delete cur;
        if (update[0]->forward[0] == nullptr) {
            tail = update[0];
        }
        if (update[0]->forward[0]) {
            update[0]->forward[0]->backward = update[1];
        }
        while (level > 1 && header->forward[level] == nullptr) {
            level--;
        }
        return true;
    }
    // 首节点非对应value

    // 遍历所有兄弟节点
    SkipListSiblingNode* sibling = cur->nextSibling;

    // 如果第一个兄弟节点就是对应value
    if (sibling->value == value) {
        cur->nextSibling = sibling->nextSibling;
        delete sibling;
    } else {
        SkipListSiblingNode* next = sibling->nextSibling;
        while (next) {
            if (next->value == value) {
                sibling->nextSibling = next->nextSibling;
                delete next;
            }
        }
    }
    return false;
}

void SkipList::print() const {
    cout << "{Skip List}-------------------------------" << endl;
    for (int i = 0; i < level; i++) {
        printLevel(i);
    }
    cout << "{Skip List}-------------------------------" << endl;
}

void SkipList::printLevel(ZSetSizeType lvl) const {
    cout << "[Skip List] level: " << lvl + 1 << endl;
    SkipListNode* cur = header->forward[lvl];

    while (cur) {
        cout << "(" << cur->score << " " << cur->value;
        SkipListSiblingNode* sibling = cur->nextSibling;
        while (sibling) {
            cout << ", " << sibling->value;
            sibling = sibling->nextSibling;
        }
        cout << ")->";
        cur = cur->forward[lvl];
    }
    cout << "null" << endl;
}