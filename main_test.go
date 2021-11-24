package main

import (
	"reflect"
	"testing"

	what "github.com/charles-haynes/whatapi"
)

func TestGetDirs(t *testing.T) {
	got, err := getDirs("testdata/root_files")
	if err != nil {
		t.Errorf("testable returned error: %v", err)
	}

	want := []dirMin{
		{0, "testdata/root_files/bar", "bar", 16, []what.FileStruct{{NameF: "ping.txt", Size: 16}}},
		{0, "testdata/root_files/baz", "baz", 18, []what.FileStruct{{NameF: "world.txt", Size: 18}}},
		{0, "testdata/root_files/foo", "foo", 18, []what.FileStruct{{NameF: "hello.txt", Size: 18}}}}
	if reflect.DeepEqual(got, want) == false {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestFullPrivDataFindMatch(t *testing.T) {
	ldirs, err := getDirs("testdata/privtest")
	if err != nil {
		t.Errorf("1/4 getDirs returned error: %v", err)
	}

	wcd := tomlAPI()
	sres, err := searchAPI(wcd, ldirs[0].files[0].NameF)
	if err != nil {
		t.Errorf("2/4 searchAPI returned error: %v", err)
	}
	rdirs, err := getAPIFilelist(wcd, sres)
	if err != nil {
		t.Errorf("3/4 getAPIFilelist returned error: %v", err)
	}

	got, err := findMatch(ldirs[0], rdirs)
	if err != nil {
		t.Errorf("4/4 findMatch returned error: %v", err)
	}
	want := 196
	if got.id != want {
		t.Errorf("match returned id: %v, want: %v", got.id, want)
	}
}
