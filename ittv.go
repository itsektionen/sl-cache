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

type ITTVLineDirection struct {
	FallbackDirection string          `json:"fallback_dir"`
	Departures        []ITTVDeparture `json:"departures"`
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
		Lines:        make([]ITTVLine, 0),
	}

	lines := make(map[string]map[int]ITTVLineDirection)

	for _, v := range conf.Lines {
		line := make(map[int]ITTVLineDirection, 0)

		for _, d := range v.Directions {
			line[d.SLDirId] = ITTVLineDirection{
				FallbackDirection: d.FallbackName,
				Departures:        make([]ITTVDeparture, 0),
			}
		}
		lines[v.Name] = line
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

		direction, exists := line[v.DirectionCode]
		if !exists {
			continue
		}

		direction.Departures = append(direction.Departures, departure)
		line[v.DirectionCode] = direction
		lines[id] = line
	}

	for _, v := range conf.Lines {
		groupName := "Normal"
		if v.Group != nil {
			groupName = *v.Group
		}
		directions := make([]ITTVLineDirection, 0)
		for _, d := range v.Directions {
			directions = append(directions, lines[v.Name][d.SLDirId])
		}
		site.Lines = append(site.Lines, ITTVLine{
			Group:      groupName,
			Type:       v.Type,
			Name:       v.Name,
			Directions: directions,
		})
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
