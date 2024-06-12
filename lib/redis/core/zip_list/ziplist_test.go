package ziplist

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// 测试创建新的压缩列表
func TestNewZiplist(t *testing.T) {
	zl := NewZiplist()
	if zl == nil {
		t.Error("NewZiplist failed to create a ziplist")
	}
}

func TestNewZiplist_(t *testing.T) {
	for i := 0; i < 100; i++ { // 多次测试创建新的压缩列表
		zl := NewZiplist()
		if zl == nil {
			t.Error("NewZiplist failed to create a ziplist")
		}
		if zl.Len() != 0 {
			t.Error("Newly created ziplist should be empty")
		}
	}
}

// 测试向压缩列表中推送字节和整数
func TestZiplistPush(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1) // Assuming DeleteByPos correctly deletes the first node

	testBytes := []byte("getAllKeys bytes")
	testInt := int64(12345)

	// 测试推送字节
	if res := zl.PushBytes(testBytes); res != 0 {
		t.Errorf("PushBytes returned non-zero result: %d", res)
	}

	// 测试推送整数
	if res := zl.PushInteger(testInt); res != 0 {
		t.Errorf("PushInteger returned non-zero result: %d", res)
	}
}

func TestZiplistPush_(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1) // Assuming DeleteByPos correctly deletes the first node

	var tests = []struct {
		input []byte
		want  int
	}{
		{[]byte("getAllKeys bytes"), 0},
		{[]byte("another getAllKeys"), 0},
		{[]byte("more data"), 0},
	}

	for _, tt := range tests {
		if res := zl.PushBytes(tt.input); res != tt.want {
			t.Errorf("PushBytes(%q) = %d, want %d", tt.input, res, tt.want)
		}
	}

	testInts := []int64{12345, 67890, 101112}
	for _, ti := range testInts {
		if res := zl.PushInteger(ti); res != 0 {
			t.Errorf("PushInteger returned non-zero result: %d for input %d", res, ti)
		}
	}
}

// 测试插入操作和对整数还是字符串节点的判断
func TestZiplistInsert(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1) // Assuming DeleteByPos correctly deletes the first node

	testBytes := []byte("insert bytes")
	testInt := int64(67890)
	pos := 1

	// 插入字节
	if res := zl.InsertBytes(pos, testBytes); res != 0 {
		t.Errorf("InsertBytes returned non-zero result: %d", res)
	}

	// 插入整数
	if res := zl.InsertInteger(pos, testInt); res != 0 {
		t.Errorf("InsertInteger returned non-zero result: %d", res)
	}

	//if node := zl.Index(2); node != nil {
	//	if val := node.GetByteArray(); val == nil || len(val) == 0 {
	//		fmt.Printf("GetInteger: %d\n", node.GetInteger())
	//	} else {
	//		fmt.Printf("GetStr: %s\n", string(val))
	//	}
	//} else {
	//	t.Errorf("node is nil")
	//}

	// 判断某个节点是数字节点还是字符串节点的操作
	if val := zl.Index(1); val.IsBytes() {
		fmt.Printf("GetStr: %s\n", string(val.GetByteArray()))
	} else {
		fmt.Printf("GetInteger: %d\n", val.GetInteger())
	}

}

func TestZiplistInsert_(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1)

	var byteTests = []struct {
		input []byte
		pos   int
		want  int
	}{
		{[]byte("insert bytes"), 1, 0},
		{[]byte("insert at start"), 1, 0},
		{[]byte("insert at middle"), 2, 0},
	}

	for _, bt := range byteTests {
		if res := zl.InsertBytes(bt.pos, bt.input); res != bt.want {
			t.Errorf("InsertBytes at pos %d returned non-zero result: %d", bt.pos, res)
		}
	}

	var intTests = []struct {
		input int64
		pos   int
		want  int
	}{
		{67890, 1, 0},
		{100000, 1, 0},
		{200000, 2, 0},
	}

	for _, it := range intTests {
		if res := zl.InsertInteger(it.pos, it.input); res != it.want {
			t.Errorf("InsertInteger at pos %d returned non-zero result: %d for input %d", it.pos, res, it.input)
		}
	}
}

// 测试节点的查找和导航功能
func TestZiplistNavigation(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1) // Cleanup after getAllKeys

	testValue := int64(42)
	zl.PushInteger(testValue)

	node := zl.FindInteger(testValue)
	if node == nil {
		t.Fatalf("FindInteger did not find the node with value %d", testValue)
	}

	if val := node.GetInteger(); val != testValue {
		t.Errorf("GetInteger expected %d, got %d", testValue, val)
	} else {
		//GetInteger: 42
		fmt.Printf("GetInteger: %d\n", val)
	}

	bytes := node.GetByteArray()
	if bytes == nil {
		fmt.Println("GetByteArray returned nil")
	} else {
		//GetByteArray: []
		fmt.Printf("GetByteArray: %v\n", bytes)
	}

	nextNode := zl.Next(node)
	if nextNode != nil {
		t.Error("Next expected to be nil, got non-nil")
	}

	prevNode := zl.Prev(node)
	if prevNode != nil {
		t.Error("Prev expected to be nil, got non-nil")
	}
}

func TestZiplistNavigation_(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1)

	values := []int64{42, 55, 66, 77, 88}
	for _, v := range values {
		zl.PushInteger(v)
		node := zl.FindInteger(v)
		if node == nil || node.GetInteger() != v {
			t.Errorf("FindInteger did not find the node with value %d or returned incorrect value", v)
		}
	}

	for _ = range values {
		if val := zl.Index(1); val.GetInteger() != 42 { // always check first insertion
			t.Errorf("GetInteger expected 42, got %d", val.GetInteger())
			break
		}
	}

	if zl.Len() != len(values) {
		t.Errorf("Expected ziplist length %d, got %d", len(values), zl.Len())
	}
}

// 测试长度和blob长度功能
func TestZiplistLengths(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1) // Cleanup after getAllKeys

	testByte1 := []byte("testByte 1")
	testByte2 := []byte("testByte 2")
	zl.PushInteger(42)
	zl.PushInteger(34)
	zl.PushBytes(testByte1)
	zl.PushInteger(50)
	zl.PushBytes(testByte2)

	//fmt.Println(zl.Len())

	if val := zl.Index(1); val.IsBytes() {
		fmt.Printf("GetStr: %s\n", string(val.GetByteArray()))
	} else {
		fmt.Printf("GetInteger: %d\n", val.GetInteger())
	}

	var startPos int = 1
	var endPos int = zl.Len()

	for i := startPos; i <= endPos; i++ {
		node := zl.Index(i)
		if node == nil {
			break
		}
		if node.IsInteger() {
			// 输出格式化的整数内容，包括序号和内容
			fmt.Printf("%d: %s\n", i, strconv.Itoa(int(node.GetInteger())))
		} else {
			// 输出格式化的字节数组内容，作为字符串，包括序号和内容
			fmt.Printf("%d: %s\n", i, string(node.GetByteArray()))
		}
	}
	//if zl.Len() != 1 {
	//	t.Errorf("Len expected 1, got %d", zl.Len())
	//}
	//
	//if zl.BlobLen() <= 0 {
	//	t.Errorf("BlobLen expected positive number, got %d", zl.BlobLen())
	//}
}

func TestZiplistLengths_(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1)

	testData := []struct {
		bytes  []byte
		intVal int64
	}{
		{[]byte("testByte 1"), 42},
		{[]byte("testByte 2"), 34},
	}

	for _, td := range testData {
		zl.PushBytes(td.bytes)
		zl.PushInteger(td.intVal)
	}

	if zl.Len() != len(testData)*2 {
		t.Errorf("Expected ziplist length %d, got %d", len(testData)*2, zl.Len())
	}

	for i, data := range testData {
		if val := zl.Index(i*2 + 1); !val.IsBytes() || string(val.GetByteArray()) != string(data.bytes) {
			t.Errorf("Expected byte array %s at index %d, got %s", string(data.bytes), i*2+1, string(val.GetByteArray()))
		}
	}
}

// 测试压缩列表的最大容量和溢出行为
func TestZiplistMaxCapacity(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1) // Cleanup after all inserts

	const maxEntries = 1000
	for i := 0; i < maxEntries; i++ {
		if res := zl.PushInteger(int64(i)); res != 0 {
			t.Fatalf("Failed to push integer %d, received non-zero result: %d", i, res)
		}
	}

	// 尝试超出最大容量
	if res := zl.PushInteger(1001); res != 0 {
		t.Error("PushInteger should fail or handle overflow gracefully when exceeding max capacity")
	}
}

// 测试空ziplist的行为
func TestEmptyZiplist(t *testing.T) {
	zl := NewZiplist()
	if zl.Len() != 0 {
		t.Error("Newly created ziplist should be empty")
	}

	// 尝试从空ziplist删除元素
	if res := zl.DeleteByPos(1); res == 0 {
		t.Error("DeleteByPos on empty ziplist should fail")
	}

	// 尝试访问空ziplist的元素
	if node := zl.Index(1); node != nil {
		t.Error("Index on empty ziplist should return nil")
	}
}

func TestEmptyZiplist_(t *testing.T) {
	zl := NewZiplist()
	if zl.Len() != 0 {
		t.Error("Newly created ziplist should be empty")
	}

	// 尝试从空ziplist删除元素
	if res := zl.DeleteByPos(1); res == 0 {
		t.Error("DeleteByPos on empty ziplist should fail")
	}

	// 尝试访问空ziplist的元素
	if node := zl.Index(1); node != nil {
		t.Error("Index on empty ziplist should return nil")
	}

	// 尝试插入后立即删除元素
	zl.PushInteger(123)
	if res := zl.DeleteByPos(1); res != 0 {
		t.Error("DeleteByPos failed to delete the only existing node")
	}

	if zl.Len() != 0 {
		t.Error("Ziplist should be empty after deleting the only node")
	}
}

// 大规模插入测试
func TestMassInsertion(t *testing.T) {
	zl := NewZiplist()
	//defer zl.DeleteByPos(1) // 清理

	const count = 5000
	for i := 0; i < count; i++ {
		if res := zl.PushInteger(int64(i)); res != 0 {
			t.Errorf("Failed to insert integer %d, result: %d", i, res)
		}
	}

	if zl.Len() != count {
		t.Errorf("Expected ziplist length %d, got %d", count, zl.Len())
	}
}

func TestMassInsertion_(t *testing.T) {
	zl := NewZiplist()
	const count = 5000

	// 插入大量整数
	for i := 0; i < count; i++ {
		if res := zl.PushInteger(int64(i)); res != 0 {
			t.Errorf("Failed to insert integer %d, result: %d", i, res)
		}
	}
	if zl.Len() != count {
		t.Errorf("Expected ziplist length %d for integers, got %d", count, zl.Len())
	}

	// 清空列表以进行字符串测试
	zl = NewZiplist()

	// 插入大量字符串
	for i := 0; i < count; i++ {
		str := fmt.Sprintf("string%d", i)
		if res := zl.PushBytes([]byte(str)); res != 0 {
			t.Errorf("Failed to insert string %s, result: %d", str, res)
		}
	}
	if zl.Len() != count {
		t.Errorf("Expected ziplist length %d for strings, got %d", count, zl.Len())
	}

	// 随机删除元素以测试列表的稳定性
	rand.Seed(time.Now().Unix())
	for i := 0; i < count/10; i++ { // 删除总数的10%
		pos := rand.Intn(zl.Len()) + 1
		if res := zl.DeleteByPos(pos); res != 0 {
			t.Errorf("Failed to delete at position %d, result: %d", pos, res)
		}
	}

	// 检查剩余长度
	expectedLength := count - (count / 10)
	if zl.Len() != expectedLength {
		t.Errorf("Expected ziplist length %d after deletions, got %d", expectedLength, zl.Len())
	}
}
