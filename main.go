package main

import (
	"fmt"
	"os"
	"path"
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

func findMatch(local dirMin, remote []dirMin, skip_trd_name_match bool) (local_plus_id dirMin, err error) {

	var size_matches []dirMin
	for _, o := range remote {
		if local.size == o.size {
			size_matches = append(size_matches, o)
		}
	}
	if len(size_matches) == 0 {
		return dirMin{}, findMatch_err_zeromatch{local.name, "totalsize", false, []int{}}
	}

	sort.SliceStable(local.files, func(i, j int) bool { return local.files[i].Name < local.files[j].Name })
	var files_matches []dirMin
	for _, o := range size_matches {
		if len(local.files) == len(o.files) {
			sort.SliceStable(o.files, func(i, j int) bool { return o.files[i].Name < o.files[j].Name })
			if reflect.DeepEqual(local.files, o.files) {
				files_matches = append(files_matches, o)
			}
		}
	}

	switch len(files_matches) {
	case 0:
		return dirMin{}, findMatch_err_zeromatch{local.name, "filelist", false, []int{}}
	case 1:
		if skip_trd_name_match {
			local.id = files_matches[0].id
			return local, nil
		}
		// default: still try to filter down matches, even if skip_trd_name_match
	}

	var name_matches []dirMin
	for _, o := range files_matches {
		if local.name == o.name {
			name_matches = append(name_matches, o)
		}
	}
	switch len(name_matches) {
	default:
		var multi_ids []int
		for _, o := range name_matches {
			multi_ids = append(multi_ids, o.id)
		}
		return dirMin{}, findMatch_err_multimatch{local.name, multi_ids}
	case 0:
		if skip_trd_name_match && len(files_matches) >= 1 {
			var lost_ids []int
			for _, o := range files_matches {
				lost_ids = append(lost_ids, o.id)
			}
			return dirMin{}, findMatch_err_zeromatch{local.name, "rtd_name", true, lost_ids}
		}
		return dirMin{}, findMatch_err_zeromatch{local.name, "rtd_name", false, []int{}}
	case 1:
		local.id = files_matches[0].id
		return local, nil
	}

}

type findMatch_err_zeromatch struct {
	lname                            string
	step                             string
	lost_match_due_to_rtd_filter     bool
	lost_match_due_to_rtd_filter_ids []int
}
type findMatch_err_multimatch struct {
	lname         string
	resulting_ids []int
}

func (e findMatch_err_zeromatch) Error() string {
	if e.lost_match_due_to_rtd_filter {
		return fmt.Sprintf("%v: 0 matches found with matcher %v; rtd filtering removed all remaining %v matches: %v", e.lname, e.step, len(e.lost_match_due_to_rtd_filter_ids), e.lost_match_due_to_rtd_filter_ids)
	} else {
		return fmt.Sprintf("%v: 0 matches found with matcher %v", e.lname, e.step)
	}
}
func (e findMatch_err_multimatch) Error() string {
	return fmt.Sprintf("%v: multiple matches found with IDs: %v", e.lname, e.resulting_ids)
}

// got, err := findMatch(ldirs[trd_index], rdirs, false)
// if err != nil {
// 	return fmt.Errorf("4/4 findMatch returned error: %v", err)
// }

//TODO: better naming to differentiate from findMatch()

// High-level function to get a match for a given torrent root directory.
// Error: (default nil)
// TODO: zeromatch No match
// TODO: manymatch Multiple matches
// (other errors possible)
// func getMatch(wcd what.Client, skip_trd_name_match bool, ldir dirMin) (match dirMin, err error) {
// 	var blacklisted_ids []int
// 	for _, f := range ldir.files { //TODO: early breaking
// 		sres, serr := searchAPI(wcd, f.Name)
// 		if serr != nil {
// 			return dirMin{}, fmt.Errorf("getMatch: 1 searchAPI for file %v errored: %v", f.Name, err)
// 		}
// 		rdirs, rerr := getAPIFilelist(wcd, sres)
// 		if rerr != nil {
// 			return dirMin{}, fmt.Errorf("getMatch: 2 getAPIFilelist for file %v errored: %v", f.Name, rerr)
// 		}
// 		match, err := findMatch(ldir, rdirs, skip_trd_name_match)
// 		if
// 	}
// }

// compare filecount
// add matches to slice
// if more than 1 items, use getAPIFileList:
//   compare source and api minDir-s (no id!) (files)
// when we have exactly one match, getAPIFileList if not already (?how ifnotalready)

//TODO: refactor warnf-s to give an error code, and return error;
//        that can be used by caller with filtering

//TODO: id ints, size int64s should actually be uints, since they can never be negative
