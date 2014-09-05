package quakeml12

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const quakeMLUrl = "http://quakeml.geonet.org.nz/quakeml/1.2/"

// Quakeml the top level container for unmarshalling QuakeML
//
// Reflection is used in parsing so if case doesn't match the names then have
// to name the corresponding element.  Tried changing case of the elements in the
// XML but it got problematic with namespaces.
type Quakeml struct {
	EventParameters EventParameters `xml:"eventParameters"`
}

// EventParameters for unmarshalling QuakeML
type EventParameters struct {
	Event Event `xml:"event"`
}

// Event for unmarshalling QuakeML
type Event struct {
	PreferredOriginID    string      `xml:"preferredOriginID"`
	PreferredMagnitudeID string      `xml:"preferredMagnitudeID"`
	O                    []Origin    `xml:"origin"`
	M                    []Magnitude `xml:"magnitude"`
	P                    []Pick      `xml:"pick"`
	Origins              map[string]*Origin
	Picks                map[string]*Pick
	Magnitudes           map[string]*Magnitude
	PreferredOrigin      *Origin
	PreferredMagnitude   *Magnitude
}

// Origin for unmarshalling QuakeML
type Origin struct {
	PublicID string    `xml:"publicID,attr"`
	Time     TimeValue `xml:"time"`
	Arrivals []Arrival `xml:"arrival"`
}

// Arrival for unmarshalling QuakeML
type Arrival struct {
	PickID       string  `xml:"pickID"`
	Phase        string  `xml:"phase"`
	Azimuth      float64 `xml:"azimuth"`
	Distance     float64 `xml:"distance"`
	TimeResidual float64 `xml:"timeResidual"`
	TimeWeight   float64 `xml:"timeWeight"`
	Pick         *Pick
}

// Pick for unmarshalling QuakeML
type Pick struct {
	PublicID         string     `xml:"publicID,attr"`
	Time             TimeValue  `xml:"time"`
	WaveformID       WaveformID `xml:"waveformID"`
	PhaseHint        string     `xml:"phaseHint"`
	EvaluationMode   string     `xml:"evaluationMode"`
	EvaluationStatus string     `xml:"evaluationStatus"`
}

// WaveformID for unmarshalling QuakeML
type WaveformID struct {
	NetworkCode  string `xml:"networkCode,attr"`
	StationCode  string `xml:"stationCode,attr"`
	LocationCode string `xml:"locationCode,attr"`
	ChannelCode  string `xml:"channelCode,attr"`
}

// Value for unmarshalling QuakeML
type Value struct {
	Value       float64 `xml:"value"`
	Uncertainty float64 `xml:"uncertainty"`
}

// TimeValue for unmarshalling QuakeML
type TimeValue struct {
	Value       time.Time `xml:"value"`
	Uncertainty float64   `xml:"uncertainty"`
}

// Mag for unmarshalling QuakeML
type Mag struct {
	Value       float64 `xml:"value"`
	Uncertainty float64 `xml:"uncertainty"`
}

// Magnitude for unmarshalling QuakeML
type Magnitude struct {
	PublicID     string `xml:"publicID,attr"`
	Mag          Mag    `xml:"mag"`
	Type         string `xml:"type"`
	MethodID     string `xml:"methodID"`
	StationCount int    `xml:"stationCount"`
}

// PickFormat describes the values that are in the map returned by PickMap.
// This can be used for query validation and documentation.
func PickFormat() (m map[string]string) {
	m = make(map[string]string)
	m["EventID"] = "e.g., 2014p072856.  This is the equivalent of the publicID attribute of Event."
	m["NetworkCode"] = "e.g., NZ"
	m["StationCode"] = "e.g., SNZO"
	m["ChannelCode"] = "e.g., HHZ"
	m["LocationCode"] = "e.g., 10"
	m["PhaseHint"] = "e.g., P"
	m["PhaseTime"] = "e.g., TODO"
	return m
}

// PickMap remaps the Pick information in the QuakeML to allow for user selectable output.
func (e *Event) PickMap() (m []map[string]string) {
	m = make([]map[string]string, len(e.Picks))

	i := 0
	for _, p := range e.Picks {
		pm := make(map[string]string)
		pm["NetworkCode"] = p.WaveformID.NetworkCode
		pm["StationCode"] = p.WaveformID.StationCode
		pm["ChannelCode"] = p.WaveformID.ChannelCode
		pm["LocationCode"] = p.WaveformID.LocationCode
		pm["PhaseHint"] = p.PhaseHint
		pm["PhaseTime"] = p.Time.Value.Format(time.RFC3339Nano)
		m[i] = pm
		i++
	}

	return m
}

// ArrivalFormat describes the values that are in the map returned by ArrivalMap.
// This can be used for query validation and documentation.
func ArrivalFormat() (m map[string]string) {
	m = make(map[string]string)
	m["EventID"] = "e.g., 2014p072856.  This is the equivalent of the publicID attribute of Event."
	m["NetworkCode"] = "e.g., NZ"
	m["StationCode"] = "e.g., SNZO"
	m["ChannelCode"] = "e.g., HHZ"
	m["LocationCode"] = "e.g., 10"
	m["Phase"] = "e.g., P"
	m["PhaseTime"] = "e.g., TODO"
	m["PhaseOriginOffset"] = "e.g., PhaseTime - OriginTime (s)"
	m["TimeResidual"] = "e.g., TODO"
	m["TimeWeight"] = "e.g., TODO"
	return m
}

// ArrivalMap remaps the Arrival information in the QuakeML to allow for user selectable output.
func (o *Origin) ArrivalMap() (m []map[string]string) {
	m = make([]map[string]string, len(o.Arrivals))

	i := 0
	for _, a := range o.Arrivals {
		am := make(map[string]string)
		am["NetworkCode"] = a.Pick.WaveformID.NetworkCode
		am["StationCode"] = a.Pick.WaveformID.StationCode
		am["ChannelCode"] = a.Pick.WaveformID.ChannelCode
		am["LocationCode"] = a.Pick.WaveformID.LocationCode
		am["Phase"] = a.Phase
		am["PhaseTime"] = a.Pick.Time.Value.Format(time.RFC3339Nano)
		am["PhaseOriginOffset"] = fmt.Sprintf("%f", a.Pick.Time.Value.Sub(o.Time.Value).Seconds())
		am["TimeResidual"] = fmt.Sprintf("%f", a.TimeResidual)
		am["TimeWeight"] = fmt.Sprintf("%f", a.TimeWeight)
		m[i] = am
		i++
	}

	return m
}

// init performs initialisation functions on the QuakeML.  Should be called called after unmarshal.
func (q *Quakeml) init() (err error) {

	if q.EventParameters.Event.PreferredOriginID == "" {
		err = errors.New(fmt.Sprint("Empty PreferredOriginID"))
		return err
	}

	if q.EventParameters.Event.PreferredMagnitudeID == "" {
		err = errors.New(fmt.Sprint("Empty PreferredMagnitudeID"))
		return err
	}

	if len(q.EventParameters.Event.O) == 0 {
		err = errors.New(fmt.Sprint("Found no origins"))
		return err
	}

	if len(q.EventParameters.Event.M) == 0 {
		err = errors.New(fmt.Sprint("Found no magnitudes"))
		return err
	}

	q.EventParameters.Event.Origins = make(map[string]*Origin)

	for i, origin := range q.EventParameters.Event.O {
		q.EventParameters.Event.Origins[origin.PublicID] = &q.EventParameters.Event.O[i]
	}

	q.EventParameters.Event.PreferredOrigin = q.EventParameters.Event.Origins[q.EventParameters.Event.PreferredOriginID]

	q.EventParameters.Event.Magnitudes = make(map[string]*Magnitude)

	for i, magnitude := range q.EventParameters.Event.M {
		q.EventParameters.Event.Magnitudes[magnitude.PublicID] = &q.EventParameters.Event.M[i]
	}

	q.EventParameters.Event.PreferredMagnitude = q.EventParameters.Event.Magnitudes[q.EventParameters.Event.PreferredMagnitudeID]

	q.EventParameters.Event.Picks = make(map[string]*Pick)

	for i, pick := range q.EventParameters.Event.P {
		q.EventParameters.Event.Picks[pick.PublicID] = &q.EventParameters.Event.P[i]
	}

	for i, a := range q.EventParameters.Event.PreferredOrigin.Arrivals {
		q.EventParameters.Event.PreferredOrigin.Arrivals[i].Pick = q.EventParameters.Event.Picks[a.PickID]
	}

	return
}

// unmarshal unmarshalls the quakeml
func unmarshal(b []byte) (e Event, err error) {
	var q Quakeml
	err = xml.Unmarshal(b, &q)
	if err != nil {
		return e, err
	}
	err = q.init()
	return q.EventParameters.Event, err
}

// result is used for passing variables on the processing pipeline
type result struct {
	event    Event
	publicID string
	err      error
}

// Fetcher reads eventids, fetches, unmarshals, and returns QuakeML.
func fetcher(done <-chan struct{}, eventids <-chan string, c chan<- result) {
	client := &http.Client{}

	for publicid := range eventids {

		var b []byte
		var e Event

		r, err := client.Get(quakeMLUrl + publicid)
		defer r.Body.Close()

		if err == nil {
			b, err = ioutil.ReadAll(r.Body)
		}

		if err == nil && r.StatusCode != 200 {
			err = errors.New(fmt.Sprintf("Non 200 response code: %d", r.StatusCode))
		}

		if err == nil {
			e, err = unmarshal(b)
		}

		select {
		case c <- result{e, publicid, err}:
		case <-done:
			return
		}
	}
}

// Get retrives QuakeML for each EventID.  Errors are logged but not returned.
func Get(eventid []string) (quakeml map[string]Event) {
	done := make(chan struct{})
	defer close(done)

	eventids := make(chan string)

	go func() {
		defer close(eventids)

		for _, e := range eventid {
			select {
			case eventids <- e:
			case <-done:
				return
			}
		}
	}()

	c := make(chan result)
	var wg sync.WaitGroup
	const numDownloaders = 15
	wg.Add(numDownloaders)
	for i := 0; i < numDownloaders; i++ {
		go func() {
			fetcher(done, eventids, c)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(c)
	}()

	quakeml = make(map[string]Event)
	var i = 0
	for r := range c {
		if r.err != nil {
			log.Println("Error fetching data for " + r.publicID)
			log.Println(r.err)
		} else {
			quakeml[r.publicID] = r.event
			i++
		}
		if i == 50 {
			log.Printf("Downloaded %v quakes", len(quakeml))
			i = 0
		}
	}
	return quakeml
}
