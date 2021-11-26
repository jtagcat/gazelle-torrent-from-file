package main

import (
	"fmt"
	"testing"
)

func singlePrivdataMatch(trd_index int, expected_id int) error {
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
