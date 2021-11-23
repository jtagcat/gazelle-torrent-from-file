package main

import (
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

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

	// wcd := initAPI(opts.api_path, opts.user_agent, opts.api_user, opts.api_pass) // not in init(), because tests can't manipulate how/when it's called then
	// using a shared client for (#TODO:) rate-limiting

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

type searchMinResult struct {
	TorrentID  int
	FileCountF int
	Size       int64
}

func searchAPI(wcd whatapi.Client, searchterm string) (paginated_result []searchMinResult, err error) { //([]response, error) {
	searchParams := url.Values{}
	searchParams.Set("order_by", "time") // time added, unlikely to skip during pagination; sorting is funky (4y, 2y, **4y**, 1y, 6mo, etc)
	searchParams.Set("order_way", "asc") // older first
	searchParams.Set("filelist", searchterm)

	page_current, pages_total := 0, 1
	for page_current < pages_total { // pages_total updated with each request
		page_current++
		searchParams.Set("page", strconv.Itoa(page_current))

		r, search_err := wcd.SearchTorrents("", searchParams)
		if search_err != nil {
			return paginated_result, search_err // responses so far, and we had an err; //TODO: upstream handle the err to drop the data, and log a warn
		}
		if page_current != r.CurrentPage {
			return paginated_result, fmt.Errorf("wcd_pagination: We requested page %d, but API replied with page %d‽", page_current, r.CurrentPage)
		}

		pages_total = r.Pages // update totalpages on each request

		// TODO: do the returned groups return only matching torrents, or all within the group?
		//  There doesn't seem to be a way to exclude non-matches, if it were the case.
		for _, rr := range r.Results {
			for _, v := range rr.Torrents {
				paginated_result = append(paginated_result, searchMinResult{v.TorrentID, v.FileCountF, v.Size})
			}
		}
	}
	return paginated_result, nil
}

// func listfilesAPI
