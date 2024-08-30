package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Options struct {
	URL     string
	Delay   float64
	Limit   int
	Paths   bool
	Sitemap bool
	Silent  bool
	Verbose bool
}

func printMessage(message string, options Options) {
	if !options.Silent {
		fmt.Println(message)
	}
}

func printVerbose(message string, options Options) {
	if options.Verbose && !options.Silent {
		fmt.Println(message)
	}
}

func checkOptions(options Options) {
	if options.URL == "" {
		fmt.Println("[-] Enter URL in following format: scheme://domain.tld")
		os.Exit(1)
	}

	if !strings.HasPrefix(strings.ToLower(options.URL), "http") {
		fmt.Println("[-] Enter URL with its scheme. (http|https)")
		os.Exit(1)
	}

	if strings.Count(options.URL, "/") > 3 {
		fmt.Println("[-] Only enter domain name, not full URL\nFormat: scheme://domain.tld")
		os.Exit(1)
	}

	if options.Silent && options.Verbose {
		fmt.Println("[-] Cannot use -q with -v at the same time.")
		os.Exit(1)
	}

	if !options.Paths && !options.Sitemap {
		fmt.Println("[-] Either use -p or -sm in order to extract data.")
		os.Exit(1)
	}
}

func parseArgs() Options {
	var options Options
	flag.StringVar(&options.URL, "u", "", "Target URL")
	flag.Float64Var(&options.Delay, "d", 0.5, "Amount of delay between each request\n\tDefault: 0.5(s)")
	flag.IntVar(&options.Limit, "l", 10, "Limit for timestamps (negative numbers can be used for the last recent results)\n\tDefault: 10")
	flag.BoolVar(&options.Paths, "p", false, "Show robots.txt paths")
	flag.BoolVar(&options.Sitemap, "sm", false, "Show robots.txt sitemaps")
	flag.BoolVar(&options.Silent, "s", false, "Silent output messages.")
	flag.BoolVar(&options.Verbose, "v", false, "Show debug level messages.")
	flag.Parse()
	return options
}

func main() {
	options := parseArgs()
	checkOptions(options)

	baseDomain := options.URL
	if baseDomain[len(baseDomain)-1] == '/' {
		baseDomain = baseDomain[:len(baseDomain)-1]
	}
	domain := baseDomain + "/robots.txt"
	limit := options.Limit

	timestampsURL := fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=%s&output=json&filter=statuscode:200&fl=timestamp,original&collapse=digest&limit=%d", domain, limit)

	// Disable SSL verification
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	resp, err := http.Get(timestampsURL)
	if err != nil {
		printVerbose(fmt.Sprintf("[DEBUG] Request error: %v", err), options)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		printVerbose(fmt.Sprintf("[DEBUG] Error reading response: %v", err), options)
		os.Exit(1)
	}

	var response [][]string
	if err := json.Unmarshal(body, &response); err != nil {
		printVerbose(fmt.Sprintf("[DEBUG] Error unmarshalling response: %v", err), options)
		os.Exit(1)
	}

	timestamps := make([]string, 0)
	for _, item := range response[1:] {
		timestamp := item[0]
		address := item[1]
		requestAddress := fmt.Sprintf("http://web.archive.org/web/%sif_/%s", timestamp, address)
		timestamps = append(timestamps, requestAddress)
	}

	printMessage(fmt.Sprintf("[+] Found [%d] Timestamps.", len(timestamps)), options)

	addressRegex := regexp.MustCompile(`(?i)allow\s?:\s?(.*)`)
	sitemapRegex := regexp.MustCompile(`(?i)(sitemap|site-map)\s?:\s?(.*)`)

	targetSitemaps := make(map[string]bool)
	targetURLs := make(map[string]bool)

	printMessage("[+] Fetching timestamps", options)
	for _, timestamp := range timestamps {
		printVerbose(fmt.Sprintf("[DEBUG] Getting New Timestamp Data.\n\t%s", timestamp), options)
		time.Sleep(time.Duration(options.Delay * float64(time.Second)))

		resp, err := http.Get(timestamp)
		if err != nil {
			printVerbose(fmt.Sprintf("[DEBUG] Error On Request: %v", err), options)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			printVerbose(fmt.Sprintf("[DEBUG] Error reading response: %v", err), options)
			continue
		}

		lines := strings.Split(string(body), "\n")
		for _, line := range lines {
			addressMatches := addressRegex.FindStringSubmatch(line)
			sitemapMatches := sitemapRegex.FindStringSubmatch(line)

			if len(addressMatches) > 0 {
				matchedAddress := addressMatches[1]
				if len(matchedAddress) > 0 {
					if matchedAddress[0] != '/' {
						matchedAddress = "/" + matchedAddress
					}
				}

				finalURL := baseDomain + matchedAddress
				if options.Paths && !targetURLs[finalURL] {
					fmt.Println(finalURL)
				}
				targetURLs[finalURL] = true
			}

			if len(sitemapMatches) > 0 {
				finalSitemap := sitemapMatches[2]
				if options.Sitemap && !targetSitemaps[finalSitemap] {
					fmt.Println(finalSitemap)
				}
				targetSitemaps[finalSitemap] = true
			}
		}
	}

	if options.Paths && !options.Silent && len(targetURLs) == 0 {
		printMessage("[-] No URL was found from robots.txt", options)
	}

	if options.Sitemap && !options.Silent && len(targetSitemaps) == 0 {
		printMessage("[-] No sitemap was found from robots.txt", options)
	}
}
