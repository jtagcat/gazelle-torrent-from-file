package main

import (
	"io/fs"
	"path/filepath"
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

	//wcd := initAPI(opts.api_path, opts.user_agent, opts.api_user, opts.api_pass) // not in init(), because tests can't manipulate how/when it's called then
	// using a shared client for (#TODO:) rate-limiting

	var wg sync.WaitGroup
	for _, dir := range actionable_dirs {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// within here: r, err := searchAPI(wcd, dir)
			log.Warnf(dir)
		}()
	}
	wg.Wait() // not needed, as go would wait for groutine exits anyway

}

type dirMin struct {
	id    int // 'optional'
	name  string
	size  int64
	files []whatapi.FileStruct
}

// list directories inside root_dir
func getDirs(root_dir string) []string {
	var dirs []string

	err := filepath.WalkDir(root_dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatalf("error walking subdirectory %q: %v", path, err)
		}
		if d.IsDir() { //TODO:
			filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {

				log.Fatalf("hello")
				return nil
			})
			// dirs = append(dirs, )
			dirs = append(dirs, d.Name())
			//dirs = append(dirs, filepath.Base(path))
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
