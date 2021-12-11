package main

import (
	"fmt"
	"reflect"
	"sort"

	what "github.com/charles-haynes/whatapi"
)

type errFindFileMatch struct {
	code int // golang doesn't have enums
	// 0: no err
	// 1: zeromatch
	// 2: multimatch
	step                     string
	zeromatch_trd_filter_ids []int // only non-default when code == 1 // IDs (all) filtered out in len(n)-th step
	multimatch_ids           []int // only non-default when code == 2 // IDs of matches left
}

// Finds match(es) for a local filepath
// CHECK IF err.code > 2 FOR RUNTIME ENUM CHECKING (see above)
// 	 panic(fmt.Sprintf("golang does not support enums; runtime enum checking: errFindFileMatch.code should never be %d", merr.code))
func findFileMatch(skip_trd_name_matching bool, local dirMin, remote []dirMin) (local_plus_id dirMin, err errFindFileMatch) {

	var size_matches []dirMin
	for _, o := range remote {
		if local.size == o.size {
			size_matches = append(size_matches, o)
		}
	}
	if size_matches == nil {
		return dirMin{}, errFindFileMatch{code: 1, step: "size"}
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
		return dirMin{}, errFindFileMatch{code: 1, step: "filelist"}
	case 1:
		if skip_trd_name_matching {
			local.id = files_matches[0].id
			return local, errFindFileMatch{}
		}
		// default: still try to filter down matches, even if skip_trd_name_matching
	}

	var trd_matches []dirMin
	for _, o := range files_matches {
		if local.name == o.name {
			trd_matches = append(trd_matches, o)
		}
	}
	switch len(trd_matches) {
	default:
		var multi_ids []int
		for _, o := range trd_matches {
			multi_ids = append(multi_ids, o.id)
		}
		return dirMin{}, errFindFileMatch{code: 2, step: "trd", multimatch_ids: multi_ids}
	case 0: //skip_trd: 2+ → 0; !skip_trd: 1+ → 0
		var lost_ids []int
		for _, o := range files_matches {
			lost_ids = append(lost_ids, o.id)
		}
		return dirMin{}, errFindFileMatch{code: 1, step: "trd", zeromatch_trd_filter_ids: lost_ids}
	case 1:
		local.id = files_matches[0].id
		return local, errFindFileMatch{}
	}

}

// High-level function to get a match for a given torrent root directory.
func findDirMatch(wcd what.Client, skip_trd_name_matching bool, ldir dirMin) (match dirMin, err error) {
	searchable := ldir.name
	var blacklisted_ids []int
	var merr errFindFileMatch // to return merr.multimatch_ids at end of func

	for _, f := range ldir.files {
		sres, serr := searchAPI(wcd, f.Name)
		if serr != nil {
			return dirMin{}, fmt.Errorf("%v: searchAPI for file %v errored: %v", searchable, f.Name, serr)
		}

		// remove blacklisted IDs (known no match)
		for _, b := range blacklisted_ids {
			for i, o := range sres {
				if o.id == b {
					sres = append(sres[:i], sres[i+1:]...)
					break
				}
			}
		}
		rdirs, rerr := getAPIFilelist(wcd, sres)
		if rerr != nil {
			return dirMin{}, fmt.Errorf("%v: getAPIFilelist for file %v errored: %v", searchable, f.Name, rerr)
		}

		match, merr := findFileMatch(skip_trd_name_matching, ldir, rdirs)
		switch merr.code {
		case 0: // single match
			return match, nil
		case 1: // zero matches
			if merr.zeromatch_trd_filter_ids == nil {
				return dirMin{}, fmt.Errorf("%v: 0 matches found", searchable)
			} else {
				return dirMin{}, fmt.Errorf("%v: 0 matches found; %v matches with IDs %v were dropped while comparing Torrent Root Directory names",
					searchable, len(merr.zeromatch_trd_filter_ids), merr.zeromatch_trd_filter_ids)
			}
		case 2: // 2+ matches
			var searched_ids []int
			for _, o := range rdirs {
				searched_ids = append(searched_ids, o.id)
			}

			// remove merr.multimatch_ids from searched_ids
			for _, m := range merr.multimatch_ids {
				for i, s := range searched_ids {
					if s == m {
						searched_ids = append(searched_ids[:i], searched_ids[i+1:]...)
						break
					}
				}
			}
			blacklisted_ids = append(blacklisted_ids, searched_ids...)
			// (loop)
		default:
			panic(fmt.Sprintf("golang does not support enums; runtime enum checking: errFindFileMatch.code should never be %d", merr.code))
		}
	}
	return dirMin{}, fmt.Errorf("%v: multiple matches found with IDs: %v", searchable, merr.multimatch_ids)
}
