package ziplist

import (
	"fmt"
	"strconv"
	"testing"
)

// 测试创建新的压缩列表
func TestNewZiplist(t *testing.T) {
	zl := NewZiplist()
	if zl == nil {
		t.Error("NewZiplist failed to create a ziplist")
	}
}

// 测试向压缩列表中推送字节和整数
func TestZiplistPush(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1) // Assuming DeleteByPos correctly deletes the first node

	testBytes := []byte("test bytes")
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
	if val := zl.Index(1); val.isBytes() {
		fmt.Printf("GetStr: %s\n", string(val.GetByteArray()))
	} else {
		fmt.Printf("GetInteger: %d\n", val.GetInteger())
	}

}

// 测试节点的查找和导航功能
func TestZiplistNavigation(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1) // Cleanup after test

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

// 测试长度和blob长度功能
func TestZiplistLengths(t *testing.T) {
	zl := NewZiplist()
	defer zl.DeleteByPos(1) // Cleanup after test

	testByte1 := []byte("testByte 1")
	testByte2 := []byte("testByte 2")
	zl.PushInteger(42)
	zl.PushInteger(34)
	zl.PushBytes(testByte1)
	zl.PushInteger(50)
	zl.PushBytes(testByte2)

	//fmt.Println(zl.Len())

	if val := zl.Index(1); val.isBytes() {
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
		if node.isInteger() {
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
