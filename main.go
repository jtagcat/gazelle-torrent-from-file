package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	what "github.com/charles-haynes/whatapi"
	log "github.com/sirupsen/logrus"
	pflag "github.com/spf13/pflag"
	retrywait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
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
	verbose                int
	textout                string
	mode_download_ids      bool
}

func init() {
	pflag.StringVarP(&opts.root_dir, "input", "i", "", "root directory, in what .torrent-less directories are in")
	pflag.BoolVarP(&opts.skip_trd_name_matching, "skip-trd-name-matching", "n", false, "skip torrent root directory name matching, where feasible")
	pflag.StringVarP(&opts.textout, "textout", "t", "", "text file to append IDs (\\n) to")
	pflag.BoolVarP(&opts.mode_download_ids, "download-text", "d", false, "download torrents from `-t` to `-o`")
	pflag.StringVarP(&opts.output, "output", "o", "", "where .torrent files should be downloaded to")
	pflag.StringVarP(&opts.moveto_success, "moveto-onsuccess", "s", "", "on success, move subdirectories of root to defined directory (optional)")
	pflag.StringVarP(&opts.moveto_failure, "moveto-onfailure", "f", "", "on failure (no match or other, generic error), move subdirectories of root to defined directory (optional)")
	pflag.StringVarP(&opts.api_path, "host", "h", "https://orpheus.network/", "URL path to API (without ajax.php, trailing slash)")
	pflag.StringVarP(&opts.api_user, "user", "u", "", "username")
	pflag.StringVarP(&opts.api_pass, "pass", "p", "", "password")
	pflag.StringVarP(&opts.user_agent, "user-agent", "a", "gtff v0", "user agent")
	pflag.CountVarP(&opts.verbose, "verbose", "v", "increase verbosity (up to 2)")
	pflag.CommandLine.SortFlags = false
	pflag.Parse()
	switch opts.verbose {
	case 1:
		log.SetLevel(log.InfoLevel)
	case 2:
		log.SetLevel(log.DebugLevel)
	}
}

func initAPI(path string, user_agent string, user string, pass string) (client what.Client) {
	wcd, err := what.NewClient(path, user_agent)
	if err != nil {
		log.Fatalf("error initializing client: %q", err)
	}
	err = wcd.Login(user, pass)
	if err != nil {
		log.Fatalf("error logging in: %q", err) // todo: check if user and pass are not empty (though for example, user might be fine to be empty)
	}
	return wcd
}

const programShortName = "gtff"

func initDir(dirpath string, dirname string, may_be_unset bool) {
	if dirpath == "" {
		if !may_be_unset {
			log.Fatalf("%s must be set", dirname)
		}
	} else { //TODO: check if path exists OR bettter error message (currently perms failed, not dir notexists)
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
	wcd := initAPI(opts.api_path, opts.user_agent, opts.api_user, opts.api_pass)

	if !opts.mode_download_ids {
		if opts.root_dir == "" {
			log.Fatal("root directory, input must be set")
		}

		// texthandle, err := os.OpenFile(opts.textout, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		// if err != nil {
		// 	panic(err)
		// }
		// defer texthandle.Close()

		//if opts.textout == "" {
		initDir(opts.output, "download location", false)
		//} else {
		//	initDir(opts.output, "download location", true)
		//}

		initDir(opts.moveto_success, "moveto onsuccess dir", true)
		initDir(opts.moveto_failure, "moveto onfailure dir", true)

		log.Info("init complete; getting local dirs")
		ldirs, err := getDirs(opts.root_dir)
		if err != nil {
			log.Fatalf("error reading source directories: %v", err)
		}

		processDirs(wcd, opts.skip_trd_name_matching, opts.output, opts.moveto_success, opts.moveto_failure, ldirs)
	} else { // mode_download_ids
		if opts.textout == "" {
			log.Fatal("text input (`-i`) must be set")
		}
		initDir(opts.output, "download location", false)

		downloadFromList(wcd, opts.textout, opts.output)
	}
}

func downloadFromList(wcd what.Client, listfile string, outdir string) {
	f, err := os.Open(listfile)
	if err != nil {
		log.Fatalf("error opening text input: %v", err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for line := 1; s.Scan(); line++ {
		id, err := strconv.Atoi(s.Text())
		if err != nil {
			log.Warnf("line %d: error converting %v to ID: %v", line, s.Text(), err)
		}
		log.Infof("downloading line %d, id %d", line, id)
		dlurl, err := wcd.CreateDownloadURL(id)
		if err != nil {
			log.Warnf("line %d, ID %d: error creating download URL: %v", line, id, err)
		}
		if err := downloadFile(outdir, dlurl); err != nil {
			log.Errorf("line %d, ID %d: error downloading torrent file: %v", line, id, err)
		}
		time.Sleep(time.Millisecond * 1100) // bad preemptive rate limiting
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
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

func processDirs(wcd what.Client, skip_trd_name_matching bool, dl_loc string, moveto_success string, moveto_failure string, ldirs []dirMin) {
	for _, ldir := range ldirs {
		log.Infof("processing trd %v", ldir.name)
		processSingleDir(wcd, skip_trd_name_matching, dl_loc, moveto_success, moveto_failure, ldir)
		time.Sleep(time.Second * 10) //TODO: very bad preemptive rate limiter
	}
}

func processSingleDir(wcd what.Client, skip_trd_name_matching bool, dl_loc string, moveto_success string, moveto_failure string, ldir dirMin) {
	ldir_with_id, merr := findDirMatch(wcd, skip_trd_name_matching, ldir)
	if merr != nil {
		if strings.Contains(merr.Error(), "matches found") { //TODO: implement more robust differentation for generic / zeromatch/multimatch errrors
			log.Warn(merr)
		} else {
			log.Error(merr)
		}
		processSingleDir_move(ldir.path, moveto_failure)
		return
	}

	log.Infof("match found, id %d", ldir_with_id.id)

	dlurl, err := wcd.CreateDownloadURL(ldir_with_id.id)
	if err != nil {
		log.Errorf("error creating download URL with ID %d, for %q: %v", ldir.id, ldir.name, err)
		processSingleDir_move(ldir.path, moveto_failure)
		return
	}
	if err := downloadFile(dl_loc, dlurl); err != nil {
		log.Errorf("error downloading torrent file with ID %d, for %q: %v", ldir.id, ldir.name, err)
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

func downloadFile(dl_loc string, url string) error { //TODO: retry mechanism
	return retry.OnError(retrywait.Backoff{
		Duration: 20 * time.Second,
		Steps:    4,
		Factor:   3,
		Jitter:   1},
		func(err error) bool {
			return true
		}, func() error {
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
			if filename == "" {
				// likely not a torrent file, but an (rate-limit) error response, parsing would be better
				return fmt.Errorf("no filename in response header")
			}

			// this is inside retry because defering would get complicated, and because we can.
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
		})
}
