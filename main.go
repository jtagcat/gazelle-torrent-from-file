package main

import (
	"fmt"
	"io/fs"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	what "github.com/jtagcat/whatapi"
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

func initAPI(path string, user_agent string, user string, pass string) (client what.Client) {
	wcd, err := what.NewClient(path, user_agent)
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
	/* 	actionable_dirs := getDirs(opts.root_dir)

	   	//wcd := initAPI(opts.api_path, opts.user_agent, opts.api_user, opts.api_pass) // not in init(), because tests can't manipulate how/when it's called then

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
	*/
}

type dirMin struct {
	id    int    // 0 for source, localfs
	path  string // source/localfs-only
	name  string // this is seperate from path bc we might only have this;
	size  int64  //   in some cases a bit insignificantly inefficient in memory
	files []what.FileStruct
}

// list directories inside root_dir
func getDirs(root_dir string) (dirs []dirMin, err error) {

	err_walk := filepath.WalkDir(root_dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking subdirectory %v: %v", path, err)
		}
		if d.IsDir() { //TODO:
			do, err_inf := d.Info()
			if err_inf != nil {
				return fmt.Errorf("error getting info for subdirectory %v: %v", path, err_inf)
			}

			var files []what.FileStruct
			err_fw := filepath.WalkDir(path, func(path2 string, f fs.DirEntry, err error) error {
				fo, err_fi := d.Info()
				if err_fi != nil {
					return fmt.Errorf("error getting info for %v: %v", path2, err_fi)
				}
				files = append(files, what.FileStruct{NameF: f.Name(), Size: fo.Size()})
				return nil
			})
			if err_fw != nil {
				return fmt.Errorf("error getting info for subdirectory %v: %v", path, err_fw)
			}

			files = files[1:] // 0th item would otherwise be parent dir
			dirs = append(dirs, dirMin{0, path, d.Name(), do.Size(), files})
		}
		return nil
	})

	if err_walk != nil {
		return dirs, fmt.Errorf("error walking root %q: %v", root_dir, err)
	}

	dirs = dirs[1:] // 0th item would otherwise be root_dir
	return dirs, nil
}

// walk files
//	// api search with file
//	// if 0 hits, log error, brake
//	// if 1 hit, downloadTorrentFile(), brake
//	// if > hit, keep results, go to next file
//	// find common matches between last (carried set) and current
// if out of files, log error
// TODO: API rate limit; can we do it via shared client?
