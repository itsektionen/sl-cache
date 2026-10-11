package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type SiteInfo struct {
	Name         string         `json:"name"`
	Abbreviation string         `json:"abbreviation"`
	Departures   map[string]any `json:"departures"`
	FetchTime    int64          `json:"fetchtime"`
}

var fetchingMutex sync.Mutex
var dataCacheMutex sync.RWMutex
var lastRequestTimes map[int]time.Time = make(map[int]time.Time)
var lastRequestTime time.Time
var dataCache map[int]SiteInfo = make(map[int]SiteInfo)
var sitesCache string = "{}"

func updateSitesCache() {
	s, err := json.Marshal(dataCache)
	if err != nil {
		println("Error creating sites cache: " + err.Error())
		return
	}
	sitesCache = string(s)
}

func throttleRequest(s int) {
	currentTime := time.Now()
	lrt, hasBeenRequested := lastRequestTimes[s]
	deltaTime := lastRequestTime.Nanosecond() + 10_000_000_000 - currentTime.Nanosecond()
	time.Sleep(time.Duration(deltaTime))
	if hasBeenRequested {
		deltaTime = lrt.Nanosecond() + 120_000_000_000 - currentTime.Nanosecond()
		currentTime = time.Now()
		time.Sleep(time.Duration(deltaTime))
	}
	lastRequestTimes[s] = currentTime
}

func fetchSite(s int) {
	u := "https://transport.integration.sl.se/v1/sites/" + strconv.Itoa(s) + "/departures"

	fetchingMutex.Lock()
	defer fetchingMutex.Unlock()

	throttleRequest(s)

	resp, err := http.Get(u)
	if err != nil {
		fmt.Println("Failed to get: " + u + ", error: " + err.Error())
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Invalid response when fetching url: " + u + ", error: " + err.Error())
		return
	}

	if resp.StatusCode != 200 {
		fmt.Println("Unknown Response: " + resp.Status + " - " + string(body))
		return
	}

	m := make(map[string]any)

	err = json.Unmarshal(body, &m)
	if err != nil {
		fmt.Println("Received invalid json: " + string(body) + " | Error: " + err.Error())
		return
	}

	dataCacheMutex.Lock()
	defer dataCacheMutex.Unlock()

	siteInfo := dataCache[s]
	siteInfo.Departures = m
	siteInfo.FetchTime = time.Now().Unix()
	dataCache[s] = siteInfo

	updateSitesCache()
}

func LoadBlankInfos() {
	for k, v := range GetConfig().Sites {
		siteId, err := strconv.Atoi(k)
		if err != nil {
			println("Failed to parse int for site id: ", k)
		} else {
			emptyData := make(map[string]any)
			emptyData["departures"] = []string{}
			dataCache[siteId] = SiteInfo{
				Name:         v.Name,
				Abbreviation: v.Abbreviation,
				Departures:   emptyData,
				FetchTime:    time.Now().Unix(),
			}
		}
	}
}

func fetchAllTargets() {
	for k, _ := range GetConfig().Sites {
		siteId, err := strconv.Atoi(k)
		if err != nil {
			println("Failed to parse int for site id: ", k)
		} else {
			fetchSite(siteId)
		}
	}
}

func GetSite(s int) (SiteInfo, bool) {
	dataCacheMutex.RLock()
	defer dataCacheMutex.RUnlock()
	c, e := dataCache[s]
	return c, e
}
func GetSites() string {
	dataCacheMutex.RLock()
	defer dataCacheMutex.RUnlock()
	return sitesCache
}

func RequestLoop() {
	for {
		fetchAllTargets()
	}
}
