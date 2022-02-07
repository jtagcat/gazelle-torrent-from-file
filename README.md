Repo is archived. Feel free to mail to the one on gh profile.

***

match and download .torrent files for local directories with/from [gazelle](https://github.com/OPSnet/Gazelle/)

This is slapped and bodged together, don't expect anything.

## program logic
1. give (root) directory in which downloaded torrents are in
1. each subdir's (Torrent Root Directories) contents will be cross-referenced with API
   1. filenames within trd will be submitted to API, to search for torrents
1. trd will be compared to potential matches:
   1. total size
   1. file listing (names and sizes)
   1. (optional[trdname]) trd name
1. if single match is found:
   1. torrent file is downloaded
   1. (optional) trd is moved to a different root (onsuccess)
1. if single match is not found (or other error):
   1. warning is outputted
   1. (optional) trd is moved to a different root (onfailure)

[trdname]: with trd name matching disabled, and in the unlikely scenario that there are multiple matches left, gtff will still try to refine the selection to one match
