package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sync"

	"github.com/jtagcat/whatapi"
	pflag "github.com/spf13/pflag"
)

// inspiration from restic
var opts struct {
	root_dir        string
	torrentfile_dir string
	api_path        string
	api_user        string
	api_pass        string
	user_agent      string
}

func init() {
	pflag.StringVarP(&opts.root_dir, "input", "i", "", "root directory, in what .torrent-less directories are in")
	pflag.StringVarP(&opts.torrentfile_dir, "output", "o", "", "where .torrent files should be downloaded to")
	pflag.StringVarP(&opts.api_path, "host", "h", "https://orpheus.network/", "URL path to API (without ajax.php, trailing slash)")
	pflag.StringVarP(&opts.api_user, "user", "u", "", "username")
	pflag.StringVarP(&opts.api_pass, "pass", "p", "", "password")
	pflag.StringVarP(&opts.user_agent, "user-agent", "a", "gtff v0", "user agent")
	pflag.CommandLine.SortFlags = false
	pflag.Parse()
}

func initAPI(path string, user_agent string, user string, pass string) (client whatapi.Client) {
	wcd, err := whatapi.NewClient(path, user_agent)
	if err != nil {
		log.Fatalf("error initializing client: %q", err)
	}
	err = wcd.Login(user, pass)
	if err != nil {
		log.Fatalf("error logging in: %q", err)
	}
	return wcd
}

func main() {
	actionable_dirs := getDirs(opts.root_dir)

	var wg sync.WaitGroup
	for _, dir := range actionable_dirs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf(dir)
			//walkObjDir(dir)
		}()
	}
	wg.Wait() // not needed, as go would wait for groutine exits anyway

}

// list directories inside root_dir
func getDirs(root_dir string) []string {
	var dirs []string

	err := filepath.WalkDir(root_dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			dirs = append(dirs, filepath.Base(path))
		}
		return nil
	})

	if err != nil {
		log.Fatalf("error walking root %q: %v", root_dir, err)
	}

	dirs = dirs[1:] // 0th item would otherwise be root_dir
	return dirs
}

// walk files
//	// api search with file
//	// if 0 hits, log error, brake
//	// if 1 hit, downloadTorrentFile(), brake
//	// if > hit, keep results, go to next file
//	// find common matches between last (carried set) and current
// if out of files, log error
// TODO: API rate limit

// func searchAPI(wcd *whatapi.ClientStruct, searchterm string) (results slice, err error) { //([]response, error) {
// 	searchParams := url.Values{}
// 	searchParams.Set("filelist", searchterm)
// 	//TODO: init var repsonse slice; api call below should append to it
//
// 	page_current, pages_total := 0, 1
// 	for page_current < pages_total {
// 		page_current++
// 		searchParams.Set("page", string(page_current))
//
// 		r, err := wcd.SearchTorrents("", searchParams)
// 		if err != nil {
// 			return r, err // responses so far, and we had an err; //TODO: upstream handle the err to drop the data, and log a warn
// 		}
//
// 		if page_current != r.CurrentPage {
// 			return r, errors.New("wcd_pagination: API did not return the page we requested.")
// 		}
// 		pages_total = r.Pages // upstream possible bugbug: page count might increase mid-pagination
//
// 		// append r to results? appending above might conflict with previous results or CurrentPage
// 	}
// 	return nil, nil // first should be full results
// }
