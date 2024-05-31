package zset

import (
	"fmt"
	"math"
	"testing"
	"time"
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
	gotScore, exist := z.ZSetGetScore("element1")
	if !exist || !floatEquals(gotScore, 10.0) {
		t.Errorf("Expected score 10.0, got %f, exist: %v", gotScore, exist)
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
	gotNewScore, exist := z.ZSetGetScore("element1")
	if !exist || !floatEquals(gotNewScore, 20.0) {
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
	gotScore, exist = z.ZSetGetScore("element1")
	if !exist || !floatEquals(gotScore, ZSetNotFound) {
		t.Errorf("Expected score to be not found, got %f", gotScore)
	} else {
		t.Logf("ZSetGetScore Successfully")
	}
}

func TestZSet_AddingDuplicate(t *testing.T) {
	z := NewZSet()
	_, _ = z.ZSetAdd(1.0, "item")
	score, exist := z.ZSetAdd(1.0, "item")

	if !exist || !floatEquals(score, 1.0) {
		t.Errorf("Expected item to exist with score 1.0, got %f, exist: %t", score, exist)
	}
}

func TestZSet_RemoveNonExistent(t *testing.T) {
	z := NewZSet()
	status := z.ZSetRemoveValue("nonexistent")

	if status != ZSetErr {
		t.Errorf("Expected error status when removing nonexistent item, got %d", status)
	}
}

// func TestZSet_ConcurrentAccess(t *testing.T) {
// 	z := NewZSet()
// 	wg := sync.WaitGroup{}
// 	wg.Add(2)

// 	go func() {
// 		defer wg.Done()
// 		for i := 0; i < 100; i++ {
// 			_, _ = z.ZSetAdd(float64(i), fmt.Sprintf("item%d", i))
// 		}
// 	}()

// 	go func() {
// 		defer wg.Done()
// 		for i := 0; i < 100; i++ {
// 			_, err := z.ZSetGetScore(fmt.Sprintf("item%d", i))
// 			if err != nil && !errors.Is(err, errors.New("[zs.ZSetGetScore] can't get score, value not found")) {
// 				t.Errorf("Error getting score for item%d: %v", i, err)
// 			}
// 		}
// 	}()

// 	wg.Wait()
// }

func TestZSet_Performance(t *testing.T) {
	z := NewZSet()
	start := time.Now()
	for i := 0; i < 10000; i++ {
		_, _ = z.ZSetAdd(float64(i), fmt.Sprintf("item%d", i))
	}
	duration := time.Since(start)
	t.Logf("Inserted 10000 items in %v", duration)
}
