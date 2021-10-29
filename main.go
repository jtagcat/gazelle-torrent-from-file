package main

import (
	"fmt"
	"io/fs"
	"path/filepath"

	pflag "github.com/spf13/pflag"
)

// inspiration from restic
var opts struct {
	root_dir        string
	torrentfile_dir string
	apipath         string
	apikey          string
}

func init() {
	pflag.StringVarP(&opts.root_dir, "input", "i", "", "root directory, in what .torrent-less directories are in")
	pflag.StringVarP(&opts.torrentfile_dir, "output", "o", "", "where .torrent files should be downloaded to")
	pflag.StringVarP(&opts.apipath, "host", "h", "", "URL path to api, ex `https://foo.bar/ajax.php`")
	pflag.StringVarP(&opts.apikey, "key", "k", "", "api key")
	pflag.CommandLine.SortFlags = false
	pflag.Parse()
}

func main() {
	getDirs(opts.root_dir) //

	// async for each dir:
	//	// walk files
	//	//	// api search with file
	//	//	// if 0 hits, log error, brake
	//	//	// if 1 hit, downloadTorrentFile(), brake

	//	//	// if > hit, keep results, go to next file
	//	//	// find common matches between last (carried set) and current
	//	// if out of files, log error

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
		panic(fmt.Sprintf("error walking path %q: %v", root_dir, err))
	}

	dirs = dirs[1:] // 0th item would otherwise be root_dir
	return dirs
}
