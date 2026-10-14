package main

import (
	"slices"
	"strconv"
	"time"
)

type ITTVTimeDelayable struct {
	Direction    string `json:"direction"`
	ScheduleDisp *int   `json:"scheduled"`
	ExpectedDisp *int   `json:"expected"`
	State        string `json:"state"` // Same as SL, but can also be "" if it doesnt exist
}

type ITTVLineDirection struct {
	Direction string            `json:"direction"`
	Next      ITTVTimeDelayable `json:"next"`
	NextNext  ITTVTimeDelayable `json:"nextnext"`
}

type ITTVLine struct {
	Name       string              `json:"name"`
	Type       string              `json:"type"`
	Group      string              `json:"group"`
	Directions []ITTVLineDirection `json:"dirs"`
}

type ITTVSite struct {
	Name         string     `json:"name"`
	Abbreviation string     `json:"abbreviation"`
	Lines        []ITTVLine `json:"lines"`
}

type ITTVInfo struct {
	Sites []ITTVSite `json:"sites"`
}

var emptyInfo = ITTVInfo{
	Sites: make([]ITTVSite, 0),
}

type slLineDepartures map[int][]SLDeparture
type slLineMap map[string]slLineDepartures

func formatDeltaTime(currentTime time.Time, targetTime *time.Time) *int {
	if targetTime == nil {
		return nil
	}
	delta := (*targetTime).Sub(currentTime)
	minutes := int(delta.Minutes())

	return &minutes
}

func emptyITTVTimeDelayable() ITTVTimeDelayable {
	return ITTVTimeDelayable{
		Direction:    "",
		ScheduleDisp: nil,
		ExpectedDisp: nil,
		State:        "",
	}
}

func generateITTVTimeDelayable(conf LineConfig, departure *SLDeparture, currentTime time.Time) ITTVTimeDelayable {
	if departure == nil {
		return emptyITTVTimeDelayable()
	}
	scheduled := formatDeltaTime(currentTime, &departure.ScheduledTime)
	expected := formatDeltaTime(currentTime, departure.ExpectedTime)
	if expected != nil && scheduled != nil {
		if *expected == *scheduled {
			expected = nil
		}
	}
	return ITTVTimeDelayable{
		Direction:    departure.Direction,
		ScheduleDisp: scheduled,
		ExpectedDisp: expected,
		State:        departure.State,
	}
}

func filterPast(departures []SLDeparture, cutoffTime time.Time) []SLDeparture {
	filtered := make([]SLDeparture, 0)
	for _, v := range departures {
		if v.ScheduledTime.Before(cutoffTime) {
			if v.ExpectedTime != nil {
				if v.ExpectedTime.Before(cutoffTime) {
					continue
				}
			} else {
				continue
			}
		}

		filtered = append(filtered, v)
	}
	return filtered
}

func generateITTVLineDirection(conf LineConfig, departures []SLDeparture, directionConfig DirectionConfig, currentTime time.Time) ITTVLineDirection {
	filtered := filterPast(departures, currentTime.Add(time.Minute))
	slices.SortFunc(filtered, func(i, j SLDeparture) int {
		ti := i.ScheduledTime
		tj := j.ScheduledTime
		if i.ExpectedTime != nil {
			ti = *i.ExpectedTime
		}
		if j.ExpectedTime != nil {
			tj = *j.ExpectedTime
		}
		return ti.UTC().Compare(tj)
	})

	if len(filtered) == 0 {
		return ITTVLineDirection{
			Direction: directionConfig.FallbackName,
			Next:      emptyITTVTimeDelayable(),
			NextNext:  emptyITTVTimeDelayable(),
		}
	}

	nextNext := emptyITTVTimeDelayable()
	if len(filtered) >= 2 {
		nextNext = generateITTVTimeDelayable(conf, &filtered[1], currentTime)
	}

	return ITTVLineDirection{
		Direction: filtered[0].Direction,
		Next:      generateITTVTimeDelayable(conf, &filtered[0], currentTime),
		NextNext:  nextNext,
	}
}

func generateITTVLinesCombined(conf SiteConfig, departures []SLDeparture) slLineMap {
	lines := make(slLineMap)

	for _, v := range conf.Lines {
		lines[v.Name] = make(slLineDepartures, 0)
		for _, d := range v.Directions {
			lines[v.Name][d.SLDirId] = make([]SLDeparture, 0)
		}
	}

	for _, v := range departures {
		id := strconv.Itoa(v.Line.Id)
		if v.Line.Designation != nil {
			id = *v.Line.Designation
		}
		line, exists := lines[id]
		if !exists {
			continue
		}

		direction, exists := line[v.DirectionCode]
		if !exists {
			continue
		}

		direction = append(direction, v)
		line[v.DirectionCode] = direction
		lines[id] = line
	}
	return lines
}

func generateITTVLine(lineConf LineConfig, line slLineDepartures) ITTVLine {
	directions := make([]ITTVLineDirection, 0)
	for _, d := range lineConf.Directions {
		directions = append(directions, generateITTVLineDirection(lineConf, line[d.SLDirId], d, time.Now()))
	}

	groupName := "Normal"
	if lineConf.Group != nil {
		groupName = *lineConf.Group
	}

	return ITTVLine{
		Group:      groupName,
		Type:       lineConf.Type,
		Name:       lineConf.Name,
		Directions: directions,
	}
}

func generateITTVSite(sl SiteInfo, conf SiteConfig) (ITTVSite, error) {
	site := ITTVSite{
		Name:         conf.Name,
		Abbreviation: conf.Abbreviation,
		Lines:        make([]ITTVLine, 0),
	}

	lines := generateITTVLinesCombined(conf, sl.Departures)

	for _, v := range conf.Lines {
		site.Lines = append(site.Lines, generateITTVLine(v, lines[v.Name]))
	}

	return site, nil
}

func GenerateITTVInfo(sl map[int]SiteInfo) (ITTVInfo, error) {
	conf := GetConfig()

	var info = ITTVInfo{
		Sites: make([]ITTVSite, 0),
	}

	for _, v := range conf.Sites {
		site, err := generateITTVSite(sl[v.Id], v)
		if err != nil {
			return emptyInfo, err
		}
		info.Sites = append(info.Sites, site)
	}

	return info, nil
}
