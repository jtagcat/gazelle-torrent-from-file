package main

import (
	"fmt"
	"testing"
)

func privdataSingleFindFileMatch(trd_index int, expected_id int) error {
	ldirs, err := getDirs("testdata/privtest")
	if err != nil {
		return fmt.Errorf("1/4 getDirs returned error: %v", err)
	}

	wcd := tomlAPI()

	sres, err := searchAPI(wcd, ldirs[trd_index].files[0].Name)
	if err != nil {
		return fmt.Errorf("2/4 searchAPI returned error: %v", err)
	}
	rdirs, err := getAPIFilelist(wcd, sres)
	if err != nil {
		return fmt.Errorf("3/4 getAPIFilelist returned error: %v", err)
	}

	got, erry := findFileMatch(false, ldirs[trd_index], rdirs)
	if erry.code != 0 {
		return fmt.Errorf("4/4 findFileMatch returned error: %v", erry)
	}

	if got.id != expected_id {
		return fmt.Errorf("match returned id: %v, want: %v", got.id, expected_id)
	}
	return nil
}
func TestPrivdataSingleFindFileMatch0(t *testing.T) {
	err := privdataSingleFindFileMatch(0, 196)
	if err != nil {
		t.Error(err)
	}

}
func TestPrivdataSingleFindFileMatch1(t *testing.T) {
	err := privdataSingleFindFileMatch(1, 1200960)
	if err != nil {
		t.Error(err)
	}
}

func privdataSingleFindDirMatch(skip_trd_name_matching bool, trd_index int, expected_id int) error {
	ldirs, err := getDirs("testdata/privtest")
	if err != nil {
		return fmt.Errorf("1/2 getDirs returned error: %v", err)
	}

	wcd := tomlAPI()

	got, err := findDirMatch(wcd, skip_trd_name_matching, ldirs[trd_index])
	if err != nil {
		return err
	}

	if got.id != expected_id {
		return fmt.Errorf("match returned id: %v, want: %v", got.id, expected_id)
	}
	return nil
}

func TestPrivdataSingleFindDirMatch0(t *testing.T) {
	err := privdataSingleFindDirMatch(false, 0, 196)
	if err != nil {
		t.Error(err)
	}

}
func TestPrivdataSingleFindDirMatch1(t *testing.T) {
	err := privdataSingleFindDirMatch(false, 1, 1200960)
	if err != nil {
		t.Error(err)
	}
}
