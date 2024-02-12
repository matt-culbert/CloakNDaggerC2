package main

import (
	"math/rand"
	"testing"
	"time"
)

func TestSetIt(t *testing.T) {
	var want int32
	got, _ := SetIt("test", "3b5692ae3acd45d4bfd0243c6f9e0f72")
	want = 0

	if got != want {
		t.Errorf("got %q but wanted %q", got, want)
	}
}

func TestXor(t *testing.T) {
	ran := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng := rand.New(ran)
	randInt := rng.Intn(256) + 1

	have := "test"
	want := "test"

	firstPass := Xor([]byte(have), randInt)
	secondPass := Xor([]byte(firstPass), randInt)

	if secondPass != want {
		t.Errorf("got %q but wanted %q", secondPass, want)
	}

}
