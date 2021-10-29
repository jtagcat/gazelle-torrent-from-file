package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

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

func main() {
	// flags: required: rootdir, apikey, apipath, torrentFileDir

	// getDirs(FLAGHERE)

	// async for each dir:
	//	// walk files
	//	//	// api search with file
	//	//	// if 0 hits, log error, brake
	//	//	// if 1 hit, downloadTorrentFile(), brake

	//	//	// if > hit, keep results, go to next file
	//	//	// find common matches between last (carried set) and current
	//	// if out of files, log error

}
