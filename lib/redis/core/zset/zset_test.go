package zset

import (
	"fmt"
	"math"
	"testing"
)

func floatEquals(a, b float64) bool {
	epsilon := 1e6
	return math.Abs(a-b) < epsilon
}

func TestZSet(t *testing.T) {
	z := NewZSet()

	// 测试添加元素
	score, exist := z.ZSetAdd(10.0, "element1")
	if exist || !floatEquals(score, 10.0) {
		t.Errorf("Expected score 10.0, got %f", score)
	} else {
		fmt.Println("ZSetAdd Successfully")
	}

	// 测试获取分数
	gotScore, err := z.ZSetGetScore("element1")
	if err != nil || !floatEquals(gotScore, 10.0) {
		t.Errorf("Expected score 10.0, got %f, error: %v", gotScore, err)
	} else {
		fmt.Println("ZSetGetScore Successfully")
	}

	// 测试更新元素
	newScore, err := z.ZSetUpdate(20.0, "element1")
	if err != nil || !floatEquals(newScore, 20.0) {
		t.Errorf("Expected updated score 20.0, got %f, error: %v", newScore, err)
	} else {
		t.Logf("ZSetUpdate Successfully")
	}

	// 测试获取分数
	gotNewScore, err := z.ZSetGetScore("element1")
	if err != nil || !floatEquals(gotNewScore, 20.0) {
		t.Errorf("Expected score 20.0, got %f, error: %v", gotScore, err)
	} else {
		t.Logf("ZSetGetScore Successfully")
	}

	// 测试删除元素
	removeStatus := z.ZSetRemoveValue("element1")
	if removeStatus != ZSetOk {
		t.Errorf("Failed to remove element, got status %d", removeStatus)
	} else {
		t.Logf("ZSetRemoveValue Successfully")
	}

	// 再次获取分数确认元素已被删除
	gotScore, err = z.ZSetGetScore("element1")
	if err == nil || !floatEquals(gotScore, ZSetNotFound) {
		t.Errorf("Expected score to be not found, got %f", gotScore)
	} else {
		t.Logf("ZSetGetScore Successfully")
	}
}
