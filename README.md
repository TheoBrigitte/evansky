# evansky: Media Renamer

CLI tool used to organize/rename media files, in order to be correctly detected by media server (e.g. jellyfin, emby, kodi).

It does so by parsing each directory/file name using [middelink](https://github.com/middelink/go-parse-torrent-name)'s parser and match result against [TheMovieDatabase search API](https://developers.themoviedb.org/3/search/search-movies).

`evansky` does cache scan results in order to guarantee that the directory/file being renamed are the one which where scanned. When changes occurs, directory need to be re-scanned.

`evansky` follow naming convention as per the [Jellyfin documentation](https://jellyfin.org/docs/general/server/media/movies.html) on naming.

## Requirement

* A TMDB API key: from [themoviedb.org/settings/api](https://www.themoviedb.org/settings/api)

## Run

```
$ evansky directory scan --apiKey my-api-key /path/to/dir
scanning /path/to/dir
scanned 23 file(s), found 23 result(s)

$ evansky directory show /path/to/dir
                                                        original              new
                                                        --------              ---
                           test1.1997.1080p.BluRay.x264.anoXmous     test1 (1997)
    test2 1977 1080p Bluray x265 10Bit AAC 2.0 - GetSchwifty.mkv     test2 (1978)
                                   test3 (2017) [1080p] [YTS.AM]     test3 (2017)
                    test4 (1980) [1080p] [BluRay] [5.1] [YTS.MX]     test4 (1980)
                                            test5 (1982) [1080p]     test5 (1982)
                        test6.1935.1080p.HDTV.x264-REGRET[rarbg]     test6 (1935)
                          test7 (1973) [1080p] [BluRay] [YTS.MX]     test7 (1973)
                          test8 (1987) [BluRay] [1080p] [YTS.AM]     test8 (1987)
                          test9 (1984) [BluRay] [1080p] [YTS.AM]     test9 (1984)
                         test10.1984.1080p.BluRay.H264.AAC-RARBG    test10 (1984)
                         test11 (1958) [BluRay] [1080p] [YTS.LT]    test11 (1958)
                                           test12 (2001) [1080p]    test12 (2001)
                                           test13 (2001) - 1080p    test13 (2001)
                    test14.2012.720p.BluRay.x264-LOST [PublicHD]    test14 (2012)
                         test15 (2001) 1080p BluRay.x264 SUJAIDR    test15 (2001)
                  test16.2011.LIMITED.1080p.BluRay.x264.anoXmous    test16 (2011)
                    test17 (1996) 1080p BluRay x265 HEVC SUJAIDR    test17 (1996)
                                       test18.1999.director name    test18 (1999)
                test19.EXTENDED.KOREAN.1080p.BluRay.H264.AAC-VXT    test19 (2016)
                          test20.1984.1080p.BluRay.x264.anoXmous    test20 (1984)
              test21.1975.Criterion.1080p.BluRay.HEVC.AAC-SARTRE    test21 (1975)
                         test22 (director name, 1970).ru-eng.avi    test22 (1970)
               test23.2020.repack.1080p.web.hevc.x265.rmteam.mkv    test23 (2020)

23/23 result(s)  100% complete

$ evansky directory rename --force /path/to/dir
> renaming
> renamed 23 file(s)
> cleaned cache /path/to/.cache/evansky/7ce101c7b750d72a018612aeaae80e69
```

## Clean

```
$ evansky cache clean -f
1 cache entries found
/path/to/.cache/evansky/7ce101c7b750d72a018612aeaae80e69 removed
```

## TODO

* fix/warn about duplicate target path
* add support for tv shows
* add support for music
