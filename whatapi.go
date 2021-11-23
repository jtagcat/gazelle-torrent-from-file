package main

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/jtagcat/whatapi"
)

type searchMinResult struct {
	id        int
	filecount int
	size      int64
}

func searchAPI(wcd whatapi.Client, searchterm string) (paginated_result []searchMinResult, err error) { //([]response, error) {
	searchParams := url.Values{}
	searchParams.Set("order_by", "time") // time added, unlikely to skip during pagination; sorting is funky (4y, 2y, **4y**, 1y, 6mo, etc)
	searchParams.Set("order_way", "asc") // older first
	searchParams.Set("filelist", searchterm)

	page_current, pages_total := 0, 1
	for page_current < pages_total { // pages_total updated with each request
		page_current++
		searchParams.Set("page", strconv.Itoa(page_current))

		r, search_err := wcd.SearchTorrents("", searchParams)
		if search_err != nil {
			return paginated_result, fmt.Errorf("wcd_pagination: Error searching for torrents with filename %v: %v", searchterm, search_err) // responses so far, and we had an err; //TODO: upstream handle the err to drop the data, and log a warn
		}
		if page_current != r.CurrentPage {
			return paginated_result, fmt.Errorf("wcd_pagination: We requested page %d, but API replied with page %d‽", page_current, r.CurrentPage)
		}

		pages_total = r.Pages // update totalpages on each request

		// TODO: do the returned groups return only matching torrents, or all within the group?
		//  There doesn't seem to be a way to exclude non-matches, if it were the case.
		for _, rr := range r.Results {
			for _, v := range rr.Torrents {
				paginated_result = append(paginated_result, searchMinResult{v.TorrentID, v.FileCountF, v.Size})
			}
		}
	}
	return paginated_result, nil
}

type searchResult struct {
	torrent searchMinResult
	files   []whatapi.FileStruct
}

func getAPIFilelist(wcd whatapi.Client, rootobjs []searchMinResult) (completedResult []searchResult, err error) {

	for _, o := range rootobjs { // to single torrent
		r, err := wcd.GetTorrent(o.id, url.Values{})
		if err != nil {
			return completedResult, fmt.Errorf("wcd_gettorrent: Error getting torrent of id %v: %v", o.id, err)
		}

		parsedfiles, pars_err := r.Torrent.Files()
		if pars_err != nil {
			return completedResult, fmt.Errorf("wcd_gettorrent: Error parsing file list of torrent with id %v: %v", o.id, pars_err)
		}
		completedResult = append(completedResult, searchResult{searchMinResult{o.id, o.filecount, o.size}, parsedfiles})
	}

	return completedResult, nil
}

//func getMatch(wcd whatapi.Client)
