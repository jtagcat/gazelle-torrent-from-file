package main

import (
	"log"
	"net/url"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/jtagcat/whatapi"
)

type ConfEnv struct {
	Host string
	User string
	Pass string
}

func TestGetDirs(t *testing.T) {
	got := getDirs("testdata/root_files")
	want := []string{"bar", "baz", "foo"}

	if reflect.DeepEqual(got, want) == false {
		t.Errorf("got %q want %q", got, want)
	}
}

func tomlAPI() (client whatapi.Client) {
	var testenv ConfEnv
	if _, err := toml.DecodeFile("testenv.toml", &testenv); err != nil {
		log.Fatalf("Error reading sercrets file: %q", err)
	}

	wcd := initAPI(opts.api_path, "gotest", testenv.User, testenv.Pass)

	return wcd
}
func TestInitAPI(t *testing.T) {

	wcd := tomlAPI()

	got, err := wcd.GetTorrent(196, url.Values{})
	if err != nil {
		t.Errorf("Error getting data: %q", err)
	}

	files, err := got.Torrent.Files()
	if err != nil {
		t.Errorf("Error parsing files: %q", err)
	}

	want := "08. Bit - You Got Mail.flac"
	if files[7].NameF != want {
		t.Errorf("got %v in %v, want %q", files[7].NameF, files, want)
	}

}

func TestSearchAPI(t *testing.T) {
	wcd := tomlAPI()

	got, err := searchAPI(wcd, "08. Bit - You Got Mail.flac")
	if err != nil {
		t.Errorf("Error getting data: %q", err)
	}
	want := []searchMinResult{{196, 15, 561900175}}
	if reflect.DeepEqual(got, want) == false {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestPaginationSearchAPI(t *testing.T) {
	wcd := tomlAPI()

	got, err := searchAPI(wcd, "readme.txt")
	if err != nil {
		t.Errorf("Error getting data: %q", err)
	}

	want := 9293
	if got[0].TorrentID != want {
		t.Errorf("First pagination result, %v, of generic search readme.txt doesn't match the expected (%v).", got[0].TorrentID, want)
	}

}
