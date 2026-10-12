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
	Mutex sync.Mutex
	sites map[int]SiteInfo
}
type APICache struct {
	Mutex sync.RWMutex
	Sites []byte
	ITTV  []byte
}

var slCache = SLCache{}
var ApiCache = APICache{}

func clearSLCache() {
	slCache.sites = make(map[int]SiteInfo, 0)
}
func updateAPICache() {
	ApiCache.Mutex.Lock()
	defer ApiCache.Mutex.Unlock()

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
}

func LoadBlankInfos() {
	clearSLCache()
	updateAPICache()
}

func fetchSite(conf SiteConfig) {
	slData, err := SLFetchSite(conf.Id)
	if err != nil {
		fmt.Printf("SL Fetching Error: " + err.Error())
		return
	}

	slCache.Mutex.Lock()
	defer slCache.Mutex.Unlock()

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
	for _, v := range GetConfig().Sites {
		fetchSite(v)
	}
	updateAPICache()
}

func RequestLoop() {
	for {
		fetchAllTargets()
	}
}
