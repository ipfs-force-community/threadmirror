package xscraper

import (
	"math/rand/v2"
	"net/http"
	"sync"
)

type RealHeader struct {
	SecChUa                 string
	SecChUaMobile           string
	SecUaPlatform           string
	UpgradeInsecureRequests string
	UserAgent               string
	Accept                  string
	SecFetchSite            string
	SecFetchMode            string
	SecFetchUser            string
	SecFetchDest            string
	AcceptLanguage          string
}

func (rh *RealHeader) Apply(req *http.Request) {
	if req.Header.Get("Sec-Ch-Ua") == "" && rh.SecChUa != "" {
		req.Header.Set("Sec-Ch-Ua", rh.SecChUa)
	}
	if req.Header.Get("Sec-Ch-Ua-Mobile") == "" && rh.SecChUaMobile != "" {
		req.Header.Set("Sec-Ch-Ua-Mobile", rh.SecChUaMobile)
	}
	if req.Header.Get("Sec-Ch-Ua-Platform") == "" && rh.SecUaPlatform != "" {
		req.Header.Set("Sec-Ch-Ua-Platform", rh.SecUaPlatform)
	}
	if req.Header.Get("Upgrade-Insecure-Requests") == "" && rh.UpgradeInsecureRequests != "" {
		req.Header.Set("Upgrade-Insecure-Requests", rh.UpgradeInsecureRequests)
	}
	if req.Header.Get("User-Agent") == "" && rh.UserAgent != "" {
		req.Header.Set("User-Agent", rh.UserAgent)
	}
	if req.Header.Get("Accept") == "" && rh.Accept != "" {
		req.Header.Set("Accept", rh.Accept)
	}
	if req.Header.Get("Sec-Fetch-Site") == "" && rh.SecFetchSite != "" {
		req.Header.Set("Sec-Fetch-Site", rh.SecFetchSite)
	}
	if req.Header.Get("Sec-Fetch-Mode") == "" && rh.SecFetchMode != "" {
		req.Header.Set("Sec-Fetch-Mode", rh.SecFetchMode)
	}
	if req.Header.Get("Sec-Fetch-User") == "" && rh.SecFetchUser != "" {
		req.Header.Set("Sec-Fetch-User", rh.SecFetchUser)
	}
	if req.Header.Get("Sec-Fetch-Dest") == "" && rh.SecFetchDest != "" {
		req.Header.Set("Sec-Fetch-Dest", rh.SecFetchDest)
	}
	if req.Header.Get("Accept-Language") == "" && rh.AcceptLanguage != "" {
		req.Header.Set("Accept-Language", rh.AcceptLanguage)
	}
}

var (
	realHeaders = []RealHeader{
		// Chrome
		{
			SecChUa:                 `"Google Chrome";v="137", "Chromium";v="137", "Not/A)Brand";v="24"`,
			SecChUaMobile:           "?0",
			SecUaPlatform:           `"Linux"`,
			UpgradeInsecureRequests: "1",
			UserAgent:               "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36",
			Accept:                  "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			SecFetchSite:            "none",
			SecFetchMode:            "navigate",
			SecFetchUser:            "?1",
			SecFetchDest:            "document",
			AcceptLanguage:          "en-US,en;q=0.9",
		},
		// Chrome Fetch/XHR
		{
			SecChUa:        `"Google Chrome";v="137", "Chromium";v="137", "Not/A)Brand";v="24"`,
			SecChUaMobile:  "?0",
			SecUaPlatform:  `"Linux"`,
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36",
			Accept:         "*/*",
			SecFetchSite:   "same-origin",
			SecFetchMode:   "cors",
			SecFetchDest:   "empty",
			AcceptLanguage: "en-US,en;q=0.9",
		},
		// Edge
		{
			SecChUa:                 `"Microsoft Edge";v="137", "Chromium";v="137", "Not/A)Brand";v="24"`,
			SecChUaMobile:           "?0",
			SecUaPlatform:           `"Linux"`,
			UpgradeInsecureRequests: "1",
			UserAgent:               "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0",
			Accept:                  "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			SecFetchSite:            "none",
			SecFetchMode:            "navigate",
			SecFetchUser:            "?1",
			SecFetchDest:            "document",
			AcceptLanguage:          "en-US,en;q=0.9",
		},
		// Edge Fetch/XHR
		{
			SecChUa:        `"Microsoft Edge";v="137", "Chromium";v="137", "Not/A)Brand";v="24"`,
			SecChUaMobile:  "?0",
			SecUaPlatform:  `"Linux"`,
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0",
			Accept:         "*/*",
			SecFetchSite:   "same-origin",
			SecFetchMode:   "cors",
			SecFetchDest:   "empty",
			AcceptLanguage: "en-US,en;q=0.9",
		},
		// Firefox
		{
			UserAgent:               "Mozilla/5.0 (X11; Linux x86_64; rv:139.0) Gecko/20100101 Firefox/139.0",
			Accept:                  "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			AcceptLanguage:          "en-US,en;q=0.5",
			UpgradeInsecureRequests: "1",
			SecFetchDest:            "document",
			SecFetchMode:            "navigate",
			SecFetchSite:            "none",
			SecFetchUser:            "?1",
		},
		// Firefox Fetch
		{
			UserAgent:      "Mozilla/5.0 (X11; Linux x86_64; rv:139.0) Gecko/20100101 Firefox/139.0",
			Accept:         "*/*",
			AcceptLanguage: "en-US,en;q=0.5",
			SecFetchDest:   "empty",
			SecFetchMode:   "cors",
			SecFetchSite:   "same-origin",
		},
	}
	mu sync.Mutex
)

func RandRealHeader() RealHeader {
	mu.Lock()
	defer mu.Unlock()

	return realHeaders[rand.IntN(len(realHeaders))]
}
