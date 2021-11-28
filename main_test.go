package main

//TODO: increase coverage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
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

type privdataEnv struct {
	WantProcessSingleDir []string
}

func initPrivdata() (privenv privdataEnv) {
	if _, err := toml.DecodeFile("testdata/privtest/data.toml", &privenv); err != nil {
		log.Fatalf("Error reading sercrets file: %q", err)
	}
	return privenv
}

func privdataProcessSingleDir(skip_trd_name_matching bool, trd_index int) (err error) {
	ldirs, err := getDirs("testdata/privtest")
	if err != nil {
		return fmt.Errorf("1/2 getDirs returned error: %v", err)
	}

	wcd := tomlAPI()
	outdir := fmt.Sprintf("testdata/privtest_tmp/%d", trd_index)
	dl_loc := path.Join(outdir, "dl_loc")
	os.RemoveAll(outdir)
	if err := os.MkdirAll(dl_loc, os.ModePerm); err != nil {
		return fmt.Errorf("prep: mkdir returned error: %v", err)
	}

	processSingleDir(wcd, skip_trd_name_matching, dl_loc, "", "", ldirs[trd_index])

	privdata := initPrivdata()
	want := privdata.WantProcessSingleDir[trd_index]
	got, err := ioutil.ReadDir(dl_loc)
	if err != nil {
		return fmt.Errorf("check: readdir returned error: %v", err)
	}
	for _, f := range got {
		if f.Name() == want {
			if f.Size() <= 1000 {
				return fmt.Errorf("check: file %q (%s) is too small", f.Size(), f.Name())
			}
			return nil
		}
	}

	return fmt.Errorf("want: %s, got %v", want, got)
}

func TestPrivdataProcessSingleDir0(t *testing.T) {
	err := privdataProcessSingleDir(false, 0)
	if err != nil {
		t.Error(err)
	}
}

//TODO: add test for processDirs
//TODO: add test for processSingleDir_move
