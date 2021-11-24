package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	what "github.com/charles-haynes/whatapi"
	log "github.com/sirupsen/logrus"
	pflag "github.com/spf13/pflag"
)

// pflags usage inspiration from restic
//   is not that friendly towards tests,
//   no required/optional, checking
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

// list directories in localfs inside root_dir
func getDirs(root_dir string) (dirs []dirMin, err error) {

	err_walk := filepath.WalkDir(root_dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking subdirectory %v: %v", path, err)
		}
		if d.IsDir() {
			files, dirsize := []what.FileStruct{}, int64(-1096) // 1096: size of root dir
			err_fw := filepath.Walk(path, func(path2 string, f os.FileInfo, err error) error {
				files = append(files, what.FileStruct{NameF: f.Name(), Size: f.Size()})
				dirsize += f.Size()
				return nil
			})
			if err_fw != nil {
				return fmt.Errorf("error getting info for subdirectory %v: %v", path, err_fw)
			}

			files = files[1:] // 0th item would otherwise be parent dir
			dirs = append(dirs, dirMin{0, path, d.Name(), dirsize, files})
		}
		return nil
	})

	if err_walk != nil {
		return dirs, fmt.Errorf("error walking root %q: %v", root_dir, err)
	}

	dirs = dirs[1:] // 0th item would otherwise be root_dir
	return dirs, nil
}

func findMatch(local dirMin, remote []dirMin) (local_plus_id dirMin, err error) {
	var name_matches []dirMin
	for _, o := range remote {
		if local.name == o.name {
			name_matches = append(name_matches, o)
		}
	}
	if len(name_matches) == 0 {
		return dirMin{}, fmt.Errorf("matching: 1 name_match: no match found for %v", local.name)
	}

	var size_matches []dirMin
	for _, o := range name_matches {
		if local.size == o.size {
			size_matches = append(size_matches, o)
		}
	}
	if len(size_matches) == 0 {
		return dirMin{}, fmt.Errorf("matching: 2 size_match: no match found for %v", local.size)
	}

	sort.SliceStable(local.files, func(i, j int) bool { return local.files[i].NameF < local.files[j].NameF })
	var files_matches []dirMin
	for _, o := range name_matches {
		if len(local.files) == len(o.files) {
			sort.SliceStable(o.files, func(i, j int) bool { return o.files[i].NameF < o.files[j].NameF })
			if reflect.DeepEqual(local.files, o.files) {
				files_matches = append(files_matches, o)
			}
		}
	}

	log.Warnf("\n\nlocal: %v\nremote: %v\n\n", local.files, name_matches[0].files)
	switch len(files_matches) {
	default:
		return dirMin{}, fmt.Errorf("matching: 3 files_match: multiple matches found for %v", local.files)
	case 0:
		return dirMin{}, fmt.Errorf("matching: 3 files_match: no match found for %v", local.files)
	case 1:
		local.id = files_matches[0].id
		return local, nil
	}
}

// compare filecount
// add matches to slice
// if more than 1 items, use getAPIFileList:
//   compare source and api minDir-s (no id!) (files)
// when we have exactly one match, getAPIFileList if not already (?how ifnotalready)

//TODO: refactor warnf-s to give an error code, and return error;
//        that can be used by caller with filtering
