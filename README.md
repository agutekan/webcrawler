## Build instructions
1. Download/install [Go](https://golang.org)
2. Run `go get -u github.com/agutekan/webcrawler`
3. A compiled `webcrawler` binary will be in your `$GOBIN` directory (run `go env` to see where this is on your system if not previously configured.)

## Usage
```
Usage: ./webcrawler <URL> <Keyword> [depth]
```

```
> ./webcrawler https://apple.com/ genre 2
Starting crawler with URL:https://apple.com/, Keyword:genre, Depth:2
Crawled 79 pages. Found 3 pages with the term 'genre'
https://apple.com/apple-arcade/?itscg=10000&itsct=arc-0-apl_hp-Wylde_learn-apl-ref-220218 => 'this genre-bendi'
https://apple.com/apple-music/ => ', or genre.                          '
https://apple.com/apple-arcade/ => 'this genre-bendi'

>
```