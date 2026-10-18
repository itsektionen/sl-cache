package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type SiteInfo struct {
	Name           string            `json:"name"`
	Abbreviation   string            `json:"abbreviation"`
	Departures     []SLDeparture     `json:"departures"`
	StopDeviations []SLStopDeviation `json:"stop_deviations"`
	FetchTime      int64             `json:"fetchtime"`
}

type SLCache struct {
	Mutex   sync.Mutex
	sites   map[int]SiteInfo
	IsValid bool
}
type APICache struct {
	Mutex   sync.RWMutex
	Sites   []byte
	ITTV    []byte
	IsValid bool
}

var slCache = SLCache{}
var ApiCache = APICache{}

func clearSLCache() {
	slCache.Mutex.Lock()
	defer slCache.Mutex.Unlock()

	slCache.sites = make(map[int]SiteInfo, 0)
	slCache.IsValid = false
}
func updateAPICache() {
	ApiCache.Mutex.Lock()
	defer ApiCache.Mutex.Unlock()
	slCache.Mutex.Lock()
	defer slCache.Mutex.Unlock()

	ApiCache.ITTV = make([]byte, 0)
	ApiCache.Sites = make([]byte, 0)
	ApiCache.IsValid = false

	if !slCache.IsValid {
		println("SL Cache is invalid")
		return
	}

	s, err := json.Marshal(slCache.sites)
	if err != nil {
		println("Error creating sites cache: " + err.Error())
		return
	}
	ApiCache.Sites = s

	ittv, err := GenerateITTVInfo(slCache.sites)
	if err != nil {
		println("Error generating ITTV info: " + err.Error())
		return
	}
	s, err = json.Marshal(ittv)
	if err != nil {
		println("Error marshaling the ittv data: " + err.Error())
		return
	}
	ApiCache.ITTV = s
	ApiCache.IsValid = true
}

func LoadBlankInfos() {
	clearSLCache()
	updateAPICache()
}

func fetchSite(conf SiteConfig) {
	slData, err := SLFetchSite(conf.Id)
	if err != nil {
		fmt.Printf("SL Fetching Error: %s\n", err.Error())
		slCache.IsValid = false
		return
	}

	siteInfo := slCache.sites[conf.Id]
	siteInfo.Abbreviation = conf.Abbreviation
	siteInfo.Name = conf.Name
	siteInfo.Departures = slData.Departures
	siteInfo.StopDeviations = slData.StopDeviations
	siteInfo.FetchTime = time.Now().Unix()
	slCache.sites[conf.Id] = siteInfo
}

func fetchAllTargets() {
	clearSLCache()
	slCache.Mutex.Lock()
	slCache.IsValid = true

	for _, v := range GetConfig().Sites {
		if slCache.IsValid {
			fetchSite(v)
		} else {
			println("Invalid cache, skipping fetches")
		}
	}
	slCache.Mutex.Unlock()
	updateAPICache()
}

func RequestLoop() {
	lastFetchTime := time.Now()
	for {
		fetchAllTargets()
		currentTime := time.Now()
		time.Sleep(lastFetchTime.Sub(currentTime) + 60_000_000_000)
		lastFetchTime = time.Now()
	}
}
