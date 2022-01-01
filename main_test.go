package main

//TODO: increase coverage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
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

func privdataProcessSingleDir(skip_trd_name_matching bool, trd_index int) error {
	ldirs, err := getDirs("testdata/privtest")
	if err != nil {
		return fmt.Errorf("1/2 getDirs returned error: %v", err)
	}

	wcd := tomlAPI()
	basedir := fmt.Sprintf("testdata/privtest_tmp_ProcessSingleDir/%d", trd_index)
	dl_loc := path.Join(basedir, "dl_loc")
	os.RemoveAll(basedir)
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

func privdataDownloadFromList(id int, trd_index int) error {
	wcd := tomlAPI()
	basedir := fmt.Sprintf("testdata/privtest_tmp_DownloadFromList/%d", trd_index)
	dl_loc := path.Join(basedir, "outdir")
	inputfile := path.Join(basedir, "input")
	os.RemoveAll(basedir)

	if err := os.MkdirAll(dl_loc, os.ModePerm); err != nil {
		return fmt.Errorf("prep: mkdir returned error: %v", err)
	}

	f, err := os.OpenFile(inputfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("prep: error opening inputfile: %v", err)
	}
	defer f.Close()
	// for _, sid := range id {
	if _, err = f.WriteString(strconv.Itoa(id) + "\n"); err != nil {
		return fmt.Errorf("prep: error writing to inputfile: %v", err)
	}
	// }
	var Boolerr bool
	log.DeferExitHandler(func() { Boolerr = true })

	downloadFromList(wcd, inputfile, dl_loc)

	privdata := initPrivdata()
	want := privdata.WantProcessSingleDir[trd_index]
	got, err := ioutil.ReadDir(dl_loc)
	if err != nil {
		return fmt.Errorf("check: readdir returned error: %v", err)
	}

	if got[0].Name() != want {
		return fmt.Errorf("want: %s, got %v", want, got)
	}

	if Boolerr {
		return fmt.Errorf("logging had errors")
	} else {
		return nil
	}
}

func TestPrivdataDownloadFromList196(t *testing.T) {
	err := privdataDownloadFromList(196, 0)
	if err != nil {
		t.Error(err)
	}
}
