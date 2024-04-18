package ziplist

import (
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

// 测试插入操作
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

	zl.PushInteger(42)

	if zl.Len() != 1 {
		t.Errorf("Len expected 1, got %d", zl.Len())
	}

	if zl.BlobLen() <= 0 {
		t.Errorf("BlobLen expected positive number, got %d", zl.BlobLen())
	}
}
