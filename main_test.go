package main

import (
	"reflect"
	"testing"
)

func TestGetDirs(t *testing.T) {
	got, err := getDirs("testdata/root_files")
	if err != nil {
		t.Errorf("testable returned error: %v", err)
	}

	want := []dirMin{
		{0, "testdata/root_files/bar", "bar", 4, []fileStruct{{Name: "ping.txt", Size: 4}}},
		{0, "testdata/root_files/baz", "baz", 11, []fileStruct{{Name: "space.txt", Size: 7}, {Name: "world.txt", Size: 4}}},
		{0, "testdata/root_files/foo", "foo", 3, []fileStruct{{Name: "hello.txt", Size: 3}}},
		{0, "testdata/root_files/bag", "bag", 26, []fileStruct{{Name: "head.txt", Size: 10}, {Name: "subthing/subsubfile.txt", Size: 16}}}}
	if reflect.DeepEqual(got, want) == false {
		t.Errorf("got %v\nwant %v", got, want)
	}
}
