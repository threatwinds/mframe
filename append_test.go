package mframe_test

import (
	"testing"
	"time"

	"github.com/threatwinds/mframe"
)

func TestAppend(t *testing.T) {
	var cache mframe.DataFrame
	cache.Init(24 * time.Hour)

	kvs := []map[mframe.KeyName]interface{}{
		{"id": 1, "value": 1.0},
		{"id": 2, "value": 2.0},
		{"id": 3, "value": 3.0},
		{"id": 4, "value": 4.0},
		{"id": 5, "value": 5.0},
	}

	for _, v := range kvs {
		cache.Insert(v)
	}

	var cache2 mframe.DataFrame
	cache2.Init(24 * time.Hour)

	kvs2 := []map[mframe.KeyName]interface{}{
		{"id": 6, "value": 6.0},
		{"id": 7, "value": 7.0},
		{"id": 8, "value": 8.0},
		{"id": 9, "value": 9.0},
		{"id": 10, "value": 10.0},
	}

	for _, v := range kvs2 {
		cache2.Insert(v)
	}

	cache.Append(&cache2, "key")

	result := cache.Count()

	if result != 10 {
		t.Errorf("expected count 10, but got %v", result)
	}
}
