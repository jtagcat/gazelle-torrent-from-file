package main

import (
	"reflect"
	"testing"

	"github.com/jtagcat/whatapi"
)

func TestGetDirs(t *testing.T) {
	got, err := getDirs("testdata/root_files")
	if err != nil {
		t.Errorf("testable returned error: %v", err)
	}

	want := []dirMin{
		{0, "testdata/root_files/bar", "bar", 16, []whatapi.FileStruct{{NameF: "ping.txt", Size: 16}}},
		{0, "testdata/root_files/baz", "baz", 18, []whatapi.FileStruct{{NameF: "world.txt", Size: 18}}},
		{0, "testdata/root_files/foo", "foo", 18, []whatapi.FileStruct{{NameF: "hello.txt", Size: 18}}}}
	if reflect.DeepEqual(got, want) == false {
		t.Errorf("got %v want %v", got, want)
	}
}
