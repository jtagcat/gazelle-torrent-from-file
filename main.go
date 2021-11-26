package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

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
	files []fileStruct
}

type fileStruct struct {
	Name string
	Size int64
}

// list directories in localfs inside root_dir
func getDirs(root_dir string) (dirs []dirMin, err error) {
	dc, err := os.Open(root_dir) // dc: dirclient
	if err != nil {
		return []dirMin{}, fmt.Errorf("error opening root dir: %v", err)
	}
	dinfo, err := dc.Readdir(-1)
	dc.Close()
	if err != nil {
		return []dirMin{}, fmt.Errorf("error listing root dir: %v", err)
	}

	for _, trd := range dinfo { // torrent root directory
		if trd.IsDir() {
			dirpath := path.Join(root_dir, trd.Name())
			files, dirsize := []fileStruct{}, int64(0)

			err_fw := filepath.Walk(dirpath, func(fpath string, f os.FileInfo, err error) error {
				if !f.IsDir() {
					relpath, err := filepath.Rel(dirpath, fpath)
					if err != nil {
						return fmt.Errorf("error getting super relative path for file %v: %v", fpath, err)
					}
					files = append(files, fileStruct{Name: relpath, Size: f.Size()})
					dirsize += f.Size()
				}
				return nil
			})
			if err_fw != nil {
				return dirs, fmt.Errorf("error getting info for torrent root dir %v: %v", trd.Name(), err_fw)
			}
			dirs = append(dirs, dirMin{0, dirpath, trd.Name(), dirsize, files})
		}
	}
	return dirs, nil
}

//TODO: refactor warnf-s to give an error code, and return error;
//        that can be used by caller with filtering

//TODO: id ints, size int64s should actually be uints, since they can never be negative
