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

}
