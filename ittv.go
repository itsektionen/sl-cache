package main

import (
	"strconv"
)

type ITTVDeparture struct {
	Direction string `json:"direction"`
	Scheduled string `json:"scheduled"`
	Expected  string `json:"expected"`
	State     string `json:"state"`
}

type ITTVLine struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Group      string          `json:"group"`
	Departures []ITTVDeparture `json:"departures"`
}

type ITTVLinePair struct {
	DirA ITTVLine `json:"dir_a"`
	DirB ITTVLine `json:"dir_b"`
}

type ITTVSite struct {
	Name         string         `json:"name"`
	Abbreviation string         `json:"abbreviation"`
	Lines        []ITTVLinePair `json:"lines"`
}

type ITTVInfo struct {
	Sites []ITTVSite `json:"sites"`
}

var emptyInfo = ITTVInfo{
	Sites: make([]ITTVSite, 0),
}

func generateITTVDeparture(departure SLDeparture) ITTVDeparture {
	expected := departure.Scheduled
	if departure.Expected != nil {
		expected = *departure.Expected
	}

	return ITTVDeparture{
		Direction: departure.Direction,
		Scheduled: departure.Scheduled,
		Expected:  expected,
		State:     departure.State,
	}
}

func generateITTVSite(sl SiteInfo, conf SiteConfig) (ITTVSite, error) {
	site := ITTVSite{
		Name:         conf.Name,
		Abbreviation: conf.Abbreviation,
		Lines:        make([]ITTVLinePair, 0),
	}

	lines := make(map[string]ITTVLinePair)

	for _, v := range conf.Lines {
		groupName := "Normal"
		if v.Group != nil {
			groupName = *v.Group
		}
		lines[v.Name] = ITTVLinePair{
			DirA: ITTVLine{
				Name:       v.Name,
				Type:       v.Type,
				Group:      groupName,
				Departures: make([]ITTVDeparture, 0),
			},
			DirB: ITTVLine{
				Name:       v.Name,
				Type:       v.Type,
				Group:      groupName,
				Departures: make([]ITTVDeparture, 0),
			},
		}
	}

	for _, v := range sl.Departures {
		id := strconv.Itoa(v.Line.Id)
		if v.Line.Designation != nil {
			id = *v.Line.Designation
		}
		line, exists := lines[id]
		if !exists {
			continue
		}

		departure := generateITTVDeparture(v)

		if v.DirectionCode == 1 {
			line.DirA.Departures = append(line.DirA.Departures, departure)
		} else if v.DirectionCode == 2 {
			line.DirB.Departures = append(line.DirB.Departures, departure)
		} else {
			println("SL gave an invalid direction code")
		}
		lines[id] = line
	}

	for _, v := range conf.Lines {
		site.Lines = append(site.Lines, lines[v.Name])
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
