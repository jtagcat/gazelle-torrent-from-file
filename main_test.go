package main

import (
	"fmt"
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
		{0, "testdata/root_files/bar", "bar", 4, []what.FileStruct{{NameF: "ping.txt", Size: 4}}},
		{0, "testdata/root_files/baz", "baz", 11, []what.FileStruct{{NameF: "space.txt", Size: 7}, {NameF: "world.txt", Size: 4}}},
		{0, "testdata/root_files/foo", "foo", 3, []what.FileStruct{{NameF: "hello.txt", Size: 3}}},
		{0, "testdata/root_files/bag", "bag", 26, []what.FileStruct{{NameF: "head.txt", Size: 10}, {NameF: "subthing/subsubfile.txt", Size: 16}}}}
	if reflect.DeepEqual(got, want) == false {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func singlePrivdataMatch(trd_index int, expected_id int) error {
	ldirs, err := getDirs("testdata/privtest")
	if err != nil {
		return fmt.Errorf("1/4 getDirs returned error: %v", err)
	}

	wcd := tomlAPI()

	sres, err := searchAPI(wcd, ldirs[trd_index].files[0].NameF)
	if err != nil {
		return fmt.Errorf("2/4 searchAPI returned error: %v", err)
	}
	rdirs, err := getAPIFilelist(wcd, sres)
	if err != nil {
		return fmt.Errorf("3/4 getAPIFilelist returned error: %v", err)
	}

	got, err := findMatch(ldirs[trd_index], rdirs, false)
	if err != nil {
		return fmt.Errorf("4/4 findMatch returned error: %v", err)
	}

	if got.id != expected_id {
		return fmt.Errorf("match returned id: %v, want: %v", got.id, expected_id)
	}
	return nil
}
func TestSinglePrivdataMatch0(t *testing.T) {
	err := singlePrivdataMatch(0, 196)
	if err != nil {
		t.Error(err)
	}

}
func TestSinglePrivdataMatch1(t *testing.T) {
	err := singlePrivdataMatch(1, 1200960)
	if err != nil {
		t.Error(err)
	}
}
