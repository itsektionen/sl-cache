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

var fetchingMutex sync.Mutex
var lastRequestTimes map[int]time.Time = make(map[int]time.Time)
var lastRequestTime time.Time

type SLJourney struct {
	Id              int     `json:"id"`
	State           string  `json:"state"`
	PredictionState *string `json:"prediction_state"`
	PassengerLevel  *string `json:"passenger_level"`
}
type SLStopArea struct {
	Id        int     `json:"id"`
	Name      string  `json:"name"`
	ShortName *string `json:"sname"`
	Type      *string `json:"type"`
}
type SLStopPoint struct {
	Id          int     `json:"id"`
	Name        *string `json:"name"`
	Designation *string `json:"designation"`
}
type SLLine struct {
	Id            int     `json:"id"`
	Designation   *string `json:"designation"`
	TransportMode *string `json:"transport_mode"`
	GroupOfLines  *string `json:"group_of_lines"`
}
type SLDepartureDeviation struct {
	ImportanceLevel int    `json:"importance_level"`
	Consequence     string `json:"consequence"`
	Message         string `json:"message"`
}

type SLDeparture struct {
	Direction     string  `json:"direction"`
	DirectionCode int     `json:"direction_code"`
	Via           *string `json:"via"`
	Destination   *string `json:"destination"`
	State         string  `json:"state"`
	Scheduled     string  `json:"scheduled"`
	ScheduledTime time.Time
	Expected      *string `json:"expected"`
	ExpectedTime  *time.Time
	Display       string                 `json:"display"`
	Journey       SLJourney              `json:"journey"`
	StopArea      SLStopArea             `json:"stop_area"`
	StopPoint     SLStopPoint            `json:"stop_point"`
	Line          SLLine                 `json:"line"`
	Deviations    []SLDepartureDeviation `json:"deviations"`
}

type SLStopDeviation struct {
	ImportanceLevel int   `json:"importance_level"`
	Consequence     int   `json:"consequence"`
	Message         int   `json:"message"`
	Scope           []any `json:"scope"`
}

type SLSiteDepartures struct {
	Departures     []SLDeparture     `json:"departures"`
	StopDeviations []SLStopDeviation `json:"stop_deviations"`
}

func throttleRequest(s int) {
	currentTime := time.Now()
	lrt, hasBeenRequested := lastRequestTimes[s]
	deltaTime := lastRequestTime.Sub(currentTime) + 5_000_000_000
	time.Sleep(deltaTime)
	currentTime = time.Now()
	lastRequestTime = currentTime
	if hasBeenRequested {
		deltaTime = lrt.Sub(currentTime) + 30_000_000_000
		if deltaTime > 0 {
			fmt.Printf("Excessive fetches to same URL, throttling for %d s\n", int(deltaTime.Seconds()))
			time.Sleep(deltaTime)
			currentTime = time.Now()
		}
	}
	lastRequestTimes[s] = currentTime
}

var emptySiteInfo = SLSiteDepartures{
	Departures:     make([]SLDeparture, 0),
	StopDeviations: make([]SLStopDeviation, 0),
}

func SLFetchSite(s int) (SLSiteDepartures, error) {
	u := "https://transport.integration.sl.se/v1/sites/" + strconv.Itoa(s) + "/departures"

	fetchingMutex.Lock()
	defer fetchingMutex.Unlock()

	throttleRequest(s)

	// println("Fetching " + u)

	resp, err := http.Get(u)
	if err != nil {
		return emptySiteInfo, fmt.Errorf("Failed to get URL (%s), err: %s", u, err.Error())
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return emptySiteInfo, fmt.Errorf("Failed to read the API response URL (%s), %s", u, err.Error())
	}

	if resp.StatusCode != 200 {
		return emptySiteInfo, fmt.Errorf("Unknown Response: %s, body: %s", resp.Status, string(body))
	}

	m := SLSiteDepartures{}

	err = json.Unmarshal(body, &m)
	if err != nil {
		return emptySiteInfo, fmt.Errorf("Received invalid json: %s, error: %s", string(body), err.Error())
	}

	timeZone, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		return emptySiteInfo, fmt.Errorf("Unable to read timezone Europe/Stockholm: %s", err.Error())
	}

	for k, v := range m.Departures {
		v.ScheduledTime, err = time.ParseInLocation("2006-01-02T15:04:05", v.Scheduled, timeZone)
		if err != nil {
			return emptySiteInfo, fmt.Errorf("Unable to parse time: %s, err: %s", v.Scheduled, err.Error())
		}
		if v.Expected != nil {
			expectedTime, err := time.ParseInLocation("2006-01-02T15:04:05", *v.Expected, timeZone)
			if err != nil {
				return emptySiteInfo, fmt.Errorf("Unable to parse time: %s, err: %s", *v.Expected, err.Error())
			}
			v.ExpectedTime = &expectedTime
		}

		m.Departures[k] = v
	}

	return m, err
}
