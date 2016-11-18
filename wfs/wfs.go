package wfs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const wfsUrl = "http://wfs.geonet.org.nz/geonet/ows?service=WFS&version=1.0.0&request=GetFeature&typeName=geonet:quake_search_v1&outputFormat=json"

// Query parameters for querying the WFS.
type Query struct {
	EventID           string
	Start             time.Time
	End               time.Time
	MinUsedPhaseCount int
	MinMagnitude      float64
	Bbox              string
}

// Features is the top level container for unmarshalling the JSON returned from the WFS.
type Features struct {
	Features []Feature
}

// Feature for unmarshalling the JSON returned from the WFS.
type Feature struct {
	Properties Properties
}

// Properties for unmarshalling the JSON returned from the WFS.
type Properties struct {
	PublicID              string
	EventType             string
	OriginTime            string
	ModificationTime      string
	Latitude              float64
	Longitude             float64
	Depth                 float64
	Magnitude             float64
	EvaluationMethod      string
	EvaluationStatus      string
	EvaluationMode        string
	EarthModel            string
	DepthType             string
	OriginError           float64
	UsedPhaseCount        int
	UsedStationCount      int
	MinimumDistance       float64
	AzimuthalGap          float64
	MagnitudeType         string
	MagnitudeUncertainty  float64
	MagnitudeStationCount int
}

// EventFormat describes the values that are in the map returned by Quakes.
// This can be used for query validation and documentation.
func EventFormat() (format map[string]string) {
	format = make(map[string]string)
	format["EventID"] = "todo"
	format["EventType"] = "todo"
	format["OriginTime"] = "todo"
	format["ModificationTime"] = "todo"
	format["Latitude"] = "todo"
	format["Longitude"] = "todo"
	format["Depth"] = "todo"
	format["Magnitude"] = "todo"
	format["EvaluationMethod"] = "todo"
	format["EvaluationStatus"] = "todo"
	format["EvaluationMode"] = "todo"
	format["EarthModel"] = "todo"
	format["DepthType"] = "todo"
	format["OriginError"] = "todo"
	format["UsedPhaseCount"] = "todo"
	format["UsedStationCount"] = "todo"
	format["MinimumDistance"] = "todo"
	format["AzimuthalGap"] = "todo"
	format["MagnitudeType"] = "todo"
	format["MagnitudeUncertainty"] = "todo"
	format["MagnitudeStationCount"] = "todo"
	return
}

// Get searchs the WFS for quakes based on the query.  Refer to EventFormat for the
// structure of the returned map.
func (q *Query) Get() (quakes []map[string]string, err error) {

	f, err := q.search()
	if err != nil {
		return nil, err
	}

	quakes = make([]map[string]string, len(f))

	var i = 0
	for p, ft := range f {
		e := make(map[string]string)
		e["EventID"] = p
		e["EventType"] = ft.Properties.EventType
		e["OriginTime"] = ft.Properties.OriginTime
		e["ModificationTime"] = ft.Properties.ModificationTime
		e["Latitude"] = fmt.Sprintf("%v", ft.Properties.Latitude)
		e["Longitude"] = fmt.Sprintf("%v", ft.Properties.Longitude)
		e["Depth"] = fmt.Sprintf("%f", ft.Properties.Depth)
		e["Magnitude"] = fmt.Sprintf("%v", ft.Properties.Magnitude)
		e["EvaluationMethod"] = ft.Properties.EvaluationMethod
		e["EvaluationStatus"] = ft.Properties.EvaluationStatus
		e["EvaluationMode"] = ft.Properties.EvaluationMode
		e["EarthModel"] = ft.Properties.EarthModel
		e["DepthType"] = ft.Properties.DepthType
		e["OriginError"] = fmt.Sprintf("%v", ft.Properties.OriginError)
		e["UsedPhaseCount"] = fmt.Sprintf("%v", ft.Properties.UsedPhaseCount)
		e["UsedStationCount"] = fmt.Sprintf("%v", ft.Properties.UsedStationCount)
		e["MinimumDistance"] = fmt.Sprintf("%v", ft.Properties.MinimumDistance)
		e["AzimuthalGap"] = fmt.Sprintf("%v", ft.Properties.AzimuthalGap)
		e["MagnitudeType"] = ft.Properties.MagnitudeType
		e["MagnitudeUncertainty"] = fmt.Sprintf("%v", ft.Properties.MagnitudeUncertainty)
		e["MagnitudeStationCount"] = fmt.Sprintf("%v", ft.Properties.MagnitudeStationCount)
		quakes[i] = e
		i++
	}
	return quakes, nil
}

// Unmarshal unmarshalls the JSON returned from the WFS.
func unmarshal(b []byte) (fs []Feature, err error) {
	var f Features

	err = json.Unmarshal(b, &f)

	return f.Features, err
}

// Url converts the query to a WFS search URL.
func (q *Query) url() string {
	var s string

	if q.EventID != "" {
		s = fmt.Sprintf("&cql_filter=publicid=='%s'", q.EventID)
	} else {
		s = fmt.Sprintf("&cql_filter=origintime>='%s'+AND+origintime<='%s'",
			q.Start.Format("2006-01-02T15:04:05"),
			q.End.Format("2006-01-02T15:04:05"))
		if q.MinUsedPhaseCount != -999 {
			s = fmt.Sprintf("%s+AND+usedphasecount>=%v", s, q.MinUsedPhaseCount)
		}
		if q.MinMagnitude != -999.9 {
			s = fmt.Sprintf("%s+AND+magnitude>=%v", s, q.MinMagnitude)
		}
		if q.Bbox != "" {
			s = fmt.Sprintf("%s+AND+BBOX(origin_geom,%v)", s, q.Bbox)
		}
	}

	return fmt.Sprintf("%s%s", wfsUrl, s)
}

// result is used for passing variables on the processing pipeline
type result struct {
	features []Feature
	err      error
}

// fetcher queries the WFS and unmarshalls and returns the resulting JSON.
func fetcher(done <-chan struct{}, urls <-chan string, c chan<- result) {
	client := &http.Client{}

	for url := range urls {

		var b []byte
		var fs []Feature

		r, err := client.Get(url)

		if err == nil {
			b, err = ioutil.ReadAll(r.Body)
			r.Body.Close()
		}

		log.Print(url)

		if err == nil && r.StatusCode != 200 {
			err = errors.New(fmt.Sprintf("Non 200 response code: %d", r.StatusCode))
		}

		if err == nil {
			fs, err = unmarshal(b)
		}

		select {
		case c <- result{fs, err}:
		case <-done:
			return
		}
	}
}

// writeURLs converts the query to WFS search URLs.  The query is
// chunked into years.  The years overlap on 1 s so there is a small chance
// of duplicate events being returned.
func (q *Query) writeUrls(done <-chan struct{}) <-chan string {
	urls := make(chan string)
	a := *q

	go func() {
		defer close(urls)

		// Break the query down into year chunks.
		if (q.End.Year() - q.Start.Year()) > 0 {
			a.End = a.Start
			for i := 0; i < (q.End.Year() - q.Start.Year()); i++ {
				a.End = a.End.AddDate(1, 0, 0)
				select {
				case urls <- a.url():
				case <-done:
					return
				}
				a.Start = a.Start.AddDate(1, 0, 0)
			}
		}

		a.End = q.End

		select {
		case urls <- a.url():
		case <-done:
			return
		}
		return
	}()

	return urls
}

// search runs a pipeline to query the WFS.
func (q *Query) search() (res map[string]Feature, err error) {

	done := make(chan struct{})
	defer close(done)

	urls := q.writeUrls(done)

	c := make(chan result)
	var wg sync.WaitGroup
	const numDownloaders = 10
	wg.Add(numDownloaders)
	for i := 0; i < numDownloaders; i++ {
		go func() {
			fetcher(done, urls, c)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(c)
	}()

	// small poss that time can in wfs query overlap in same second so
	// use map for the results.
	res = make(map[string]Feature)
	quakes := 0

	for r := range c {
		if r.err != nil {
			return nil, r.err
		}
		quakes = quakes + len(r.features)
		log.Printf("Downloaded %v quakes", quakes)
		for _, feature := range r.features {
			res[feature.Properties.PublicID] = feature
		}
	}
	return res, nil
}
