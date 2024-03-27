package main

import "testing"

func TestBucketBurstThenBlock(t *testing.T) {
	b := NewBucket(0, 2) // no refill
	if !b.Allow("a") || !b.Allow("a") {
		t.Fatal("expected burst of 2")
	}
	if b.Allow("a") {
		t.Fatal("expected block after burst")
	}
}
