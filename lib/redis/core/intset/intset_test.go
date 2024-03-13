package intset

import (
	"testing"
	"time"
)

func TestIntset(t *testing.T) {
	st := time.Now()
	s := NewIntset()

	// Test Add
	if res := s.IntsetAdd(1); res != Ok {
		t.Errorf("IntsetAdd(1) failed, got %d, want %d", res, Ok)
	}

	// Test Find
	if res := s.IntsetFind(1); res != Ok {
		t.Errorf("IntsetFind(1) failed, got %d, want %d", res, Ok)
	}

	// Test Remove
	if res := s.IntsetRemove(1); res != Ok {
		t.Errorf("IntsetRemove(1) failed, got %d, want %d", res, Ok)
	}

	if res := s.IntsetFind(1); res != Err {
		t.Errorf("IntsetFind(1) failed, got %d, want %d", res, Err)
	}

	// Add multiple to test Length and Random
	s.IntsetAdd(2)
	s.IntsetAdd(3)

	// Test Len
	if res := s.IntsetLen(); res != 2 {
		t.Errorf("IntsetLen() failed, got %d, want %d", res, 2)
	}

	// Test Random (hard to predict, just ensure it runs)
	if res := s.IntsetRandom(); res != 2 && res != 3 {
		t.Errorf("IntsetRandom() failed, got %d, want 2 or 3", res)
	}

	// Test Get
	if res := s.IntsetGet(0); res != 2 && res != 3 {
		t.Errorf("IntsetGet(0) failed, got %d, want 2 or 3", res)
	}

	// Test BlobLen (implementation specific, assuming it's non-zero for non-empty set)
	if res := s.IntsetBlobLen(); res != 2 {
		t.Errorf("IntsetBlobLen() failed, got %d, want 2", res)
	}

	et := time.Now()
	t.Logf("TestIntset finished in %d us.", et.Sub(st).Microseconds())
}

func TestIntsetUpgrade(t *testing.T) {
	st := time.Now()
	s := NewIntset()

	s.IntsetAdd(1)
	if res := s.IntsetBlobLen(); res != 1 {
		t.Errorf("IntsetBlobLen() failed, got %d, want %d", res, 1)
	}

	s.IntsetAdd(256)
	if res := s.IntsetBlobLen(); res != 4 {
		t.Errorf("IntsetBlobLen() failed, got %d, want %d", res, 4)
	}

	s.IntsetAdd(65536)
	if res := s.IntsetBlobLen(); res != 12 {
		t.Errorf("IntsetBlobLen() failed, got %d, want %d", res, 12)
	}

	s.IntsetAdd(4_294_967_296)
	if res := s.IntsetBlobLen(); res != 32 {
		t.Errorf("IntsetBlobLen() failed, got %d, want %d", res, 32)
	}

	et := time.Now()
	t.Logf("TestIntsetUpgrade finished in %d us.", et.Sub(st).Microseconds())
}
