# evansky: Media Renamer

CLI tool used to organize/rename media files, in order to be correctly detected by media server (e.g. [jellyfin](https://jellyfin.org/), [emby](https://emby.media/), [kodi](https://kodi.tv/)).

It does so by parsing each directory/file name using [middelink](https://github.com/middelink/go-parse-torrent-name)'s parser and match result against [TheMovieDatabase API](https://www.themoviedb.org/).

`evansky` does cache scan results in order to guarantee that applied directory/file renaming is the same as the one from the initial dry-run preview.

`evansky` follow naming convention as per the [Jellyfin documentation](https://jellyfin.org/docs/general/server/media/movies/).

## Requirement

* A TMDB API key: from [themoviedb.org/settings/api](https://www.themoviedb.org/settings/api)

## Run

```
$ export TMDB_API_KEY="your api key"
$ evansky rename /path/to/dir
INF [dry-run] renamed source="test1.1997.1080p.BluRay.x264.anoXmous" destination="test1 (1997)"
INF [dry-run] renamed source="test2 1977 1080p Bluray x265 10Bit AAC 2.0 - GetSchwifty.mkv" destination="test2 (1978)"
INF [dry-run] renamed source="test3 (2017) [1080p] [YTS.AM]" destination="test3 (2017)"
INF [dry-run] renamed source="test4 (1980) [1080p] [BluRay] [5.1] [YTS.MX]" destination="test4 (1980)"
INF [dry-run] renamed source="test5 (1982) [1080p]" destination="test5 (1982)"
INF [dry-run] renamed source="test6.1935.1080p.HDTV.x264-REGRET[rarbg]" destination="test6 (1935)"
INF [dry-run] renamed source="test7 (1973) [1080p] [BluRay] [YTS.MX]" destination="test7 (1973)"
INF [dry-run] renamed source="test8 (1987) [BluRay] [1080p] [YTS.AM]" destination="test8 (1987)"
INF [dry-run] renamed source="test9 (1984) [BluRay] [1080p] [YTS.AM]" destination="test9 (1984)"
INF [dry-run] renamed source="test10.1984.1080p.BluRay.H264.AAC-RARBG" destination="test10 (1984)"
INF [dry-run] renamed source="test11 (1958) [BluRay] [1080p] [YTS.LT]" destination="test11 (1958)"
INF [dry-run] renamed source="test12 (2001) [1080p]" destination="test12 (2001)"
INF [dry-run] renamed source="test13 (2001) - 1080p" destination="test13 (2001)"
INF [dry-run] renamed source="test14.2012.720p.BluRay.x264-LOST [PublicHD]" destination="test14 (2012)"
INF [dry-run] renamed source="test15 (2001) 1080p BluRay.x264 SUJAIDR" destination="test15 (2001)"
INF [dry-run] renamed source="test16.2011.LIMITED.1080p.BluRay.x264.anoXmous" destination="test16 (2011)"
INF [dry-run] renamed source="test17 (1996) 1080p BluRay x265 HEVC SUJAIDR" destination="test17 (1996)"
INF [dry-run] renamed source="test18.1999.director name" destination="test18 (1999)"
INF [dry-run] renamed source="test19.EXTENDED.KOREAN.1080p.BluRay.H264.AAC-VXT" destination="test19 (2016)"
INF [dry-run] renamed source="test20.1984.1080p.BluRay.x264.anoXmous" destination="test20 (1984)"
INF [dry-run] renamed source="test21.1975.Criterion.1080p.BluRay.HEVC.AAC-SARTRE" destination="test21 (1975)"
INF [dry-run] renamed source="test22 (director name, 1970).ru-eng.avi" destination="test22 (1970)"
INF [dry-run] renamed source="test23.2020.repack.1080p.web.hevc.x265.rmteam.mkv" destination="test23 (2020)"
INF [dry-run] renamed 23/23 file(s)
$ # Run the same command with --write to apply changes
```
