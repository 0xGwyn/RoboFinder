## Installation
```bash
go install github.com/0xgwyn/robofinder@latest
```
or
```bash
git clone https://github.com/0xGwyn/RoboFinder.git
cd RoboFinder
build -o $GOPATH/bin/robofinder main.go
```

# Flags
```yaml
   -u, -url string      Target URL for the `robots.txt` file
   -d, -delay float     Delay between requests in seconds (default: 0.5)
   -l, -limit int       Limit for the number of timestamps to retrieve. Use negative numbers to get the most recent entries (default: 10)
   -p, -paths           Display disallowed paths from the `robots.txt` file
   -sm, -sitemap        Display sitemap URLs from the `robots.txt` file
   -s, -silent          Only show essential output
   -v, -verbose         Show detailed debug-level messages
```

# Simple Example 
```
robofinder -u http://example.com -p -sm -d 1.0 -l 5 -v
```
