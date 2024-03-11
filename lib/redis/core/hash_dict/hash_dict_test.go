package hash_dict

import "testing"

func TestHashDict(t *testing.T) {
	dict := NewDict()

	// 测试添加功能
	if dict.DictAdd("key1", "value1") != DictOk {
		t.Error("Expected DictOk from DictAdd when adding a new key")
	}

	// 测试重复添加相同的键
	if dict.DictAdd("key1", "value1") != DictErr {
		t.Error("Expected DictErr from DictAdd when adding a duplicate key")
	}

	// 测试查找功能
	if val := dict.DictFind("key1"); val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	// 测试更新功能
	if dict.DictUpdate("key1", "value2") != DictOk {
		t.Error("Expected DictOk from DictUpdate when key exists")
	}

	// 测试更新不存在的键
	if dict.DictUpdate("key2", "value2") != DictErr {
		t.Error("Expected DictErr from DictUpdate when key does not exist")
	}

	// 测试插入或更新功能
	dict.DictInsertOrUpdate("key1", "value3")
	if val := dict.DictFind("key1"); val != "value3" {
		t.Errorf("Expected 'value3', got %v", val)
	}
	dict.DictInsertOrUpdate("key3", "value3")
	if val := dict.DictFind("key3"); val != "value3" {
		t.Errorf("Expected 'value3', got %v", val)
	}

	// 测试删除功能
	if dict.DictRemove("key1") != DictOk {
		t.Error("Expected DictOk from DictRemove when key exists")
	}

	// 测试删除不存在的键
	if dict.DictRemove("key2") != DictErr {
		t.Error("Expected DictErr from DictRemove when key does not exist")
	}

	// 验证删除后键是否不存在
	if val := dict.DictFind("key1"); val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}
