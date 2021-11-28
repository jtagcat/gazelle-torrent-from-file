package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
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
	root_dir               string
	output                 string
	moveto_success         string
	moveto_failure         string
	api_path               string
	api_user               string
	api_pass               string
	user_agent             string
	skip_trd_name_matching bool
}

func init() {
	pflag.StringVarP(&opts.root_dir, "input", "i", "", "root directory, in what .torrent-less directories are in")
	pflag.BoolVarP(&opts.skip_trd_name_matching, "skip-trd-name-matching", "n", false, "skip torrent root directory name matching")
	pflag.StringVarP(&opts.output, "output", "o", "", "where .torrent files should be downloaded to")
	pflag.StringVarP(&opts.moveto_success, "moveto-onsuccess", "s", "", "on success, move subdirectories of root to defined directory (optional)")
	pflag.StringVarP(&opts.moveto_failure, "moveto-onfailure", "f", "", "on failure, move subdirectories of root to defined directory (optional)")
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

const programShortName = "gtff"

func initDir(dirpath string, dirname string, may_be_unset bool) {
	if dirpath == "" {
		if !may_be_unset {
			log.Fatalf("%s must be set", dirname)
		}
	} else {
		filepath := path.Join(dirpath, programShortName+"_permtest")
		if werr := ioutil.WriteFile(filepath, []byte("delete me, testing writability"), os.ModePerm); werr != nil {
			log.Fatalf("error writing permtest file to %s: %v", dirname, werr)
		}
		if derr := os.Remove(filepath); derr != nil {
			log.Fatalf("error removing permtest file from %s: %v", dirname, derr)
		}
	}
}

func main() {
	if opts.root_dir == "" {
		log.Fatal("root directory, input must be set")
	}
	initDir(opts.output, "download location", false)
	initDir(opts.moveto_success, "moveto onsuccess dir", true)
	initDir(opts.moveto_failure, "moveto onfailure dir", true)
	wcd := initAPI(opts.api_path, opts.user_agent, opts.api_user, opts.api_pass)

	ldirs, err := getDirs(opts.root_dir)
	if err != nil {
		log.Fatalf("error reading source directories: %v", err)
	}

	processDirs(wcd, opts.skip_trd_name_matching, opts.output, opts.moveto_success, opts.moveto_failure, ldirs)
}

func processDirs(wcd what.Client, skip_trd_name_matching bool, dl_loc string, moveto_success string, moveto_failure string, ldirs []dirMin) {
	for _, ldir := range ldirs {
		processSingleDir(wcd, skip_trd_name_matching, dl_loc, moveto_success, moveto_failure, ldir)
	}
}

func processSingleDir(wcd what.Client, skip_trd_name_matching bool, dl_loc string, moveto_success string, moveto_failure string, ldir dirMin) {
	ldir_with_id, merr := findDirMatch(wcd, skip_trd_name_matching, ldir)
	if merr != nil {
		log.Warn(merr)
		processSingleDir_move(ldir.path, moveto_failure)
		return
	}

	dlurl, err := wcd.CreateDownloadURL(ldir_with_id.id)
	if err != nil {
		log.Warnf("error creating download URL for %q: %v", ldir.name, err)
		processSingleDir_move(ldir.path, moveto_failure)
		return
	}
	if err := downloadFile(dl_loc, dlurl); err != nil {
		log.Warnf("error downloading torrent file for %q: %v", ldir.name, err)
		processSingleDir_move(ldir.path, moveto_failure)
		return
	}

	processSingleDir_move(ldir.path, moveto_success)
}

func processSingleDir_move(frompath string, destdir string) {
	if destdir != "" {
		if rerr := os.Rename(frompath, path.Join(destdir, path.Base(frompath))); rerr != nil {
			log.Fatalf("error moving %q to %q: %v", path.Base(frompath), destdir, rerr)
		}
	}
}

func downloadFile(dl_loc string, url string) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response code: %s", r.Status)
	}

	var filename string
	for _, h := range r.Header["Content-Disposition"] {
		_, cdheader, err := mime.ParseMediaType(h)
		if err != nil {
			return err
		}
		if cdheader["filename"] != "" {
			filename = cdheader["filename"]
			break
		}
	}
	f, err := os.Create(path.Join(dl_loc, filename))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r.Body)
	if err != nil {
		return err
	}
	return nil
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
