package main

import (
	"fmt"
	"reflect"
	"sort"
)

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
