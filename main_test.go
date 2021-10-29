package main

import (
	"reflect"
	"testing"
)

func TestGetDirs(t *testing.T) {
	got := getDirs("testdata/root_files")
	want := []string{"bar", "baz", "foo"}

	if reflect.DeepEqual(got, want) == false {
		t.Errorf("got %q want %q", got, want)
	}
}
