package seiscompml07

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

const seiscompURL = "http://seiscompml07.s3-website-ap-southeast-2.amazonaws.com/"

// Seiscomp the top level container for unmarshalling SeisCompML
//
// Reflection is used in parsing so if case doesn't match the names then have
// to name the corresponding element.  Tried changing case of the elements in the
// XML but it got problematic with namespaces.
type Seiscomp struct {
	EventParameters EventParameters `xml:"EventParameters"`
}

// EventParameters for unmarshalling SeisCompML
type EventParameters struct {
	Event Event    `xml:"event"`
	O     []Origin `xml:"origin"`
	P     []Pick   `xml:"pick"`
}

// Event for unmarshalling SeisCompML
type Event struct {
	PreferredOriginID    string `xml:"preferredOriginID"`
	PreferredMagnitudeID string `xml:"preferredMagnitudeID"`
	PreferredOrigin      *Origin
	PreferredMagnitude   *Magnitude
	Picks                map[string]*Pick
	Origins              map[string]*Origin
	Magnitudes           map[string]*Magnitude
	// Copy these from EventParameters so that the api will be the same as for
	// SeisCompML 1.2
	O []Origin
	M []Magnitude
	P []Pick
}

// Origin for unmarshalling SeisCompML
type Origin struct {
	PublicID string      `xml:"publicID,attr"`
	Time     TimeValue   `xml:"time"`
	Arrivals []Arrival   `xml:"arrival"`
	M        []Magnitude `xml:"magnitude"`
}

// Arrival for unmarshalling SeisCompML
type Arrival struct {
	PickID       string  `xml:"pickID"`
	Phase        string  `xml:"phase"`
	Azimuth      float64 `xml:"azimuth"`
	Distance     float64 `xml:"distance"`
	TimeResidual float64 `xml:"timeResidual"`
	TimeWeight   float64 `xml:"weight"`
	Pick         *Pick
}

// Pick for unmarshalling SeisCompML
type Pick struct {
	PublicID         string     `xml:"publicID,attr"`
	Time             TimeValue  `xml:"time"`
	WaveformID       WaveformID `xml:"waveformID"`
	PhaseHint        string     `xml:"phaseHint"`
	EvaluationMode   string     `xml:"evaluationMode"`
	EvaluationStatus string     `xml:"evaluationStatus"`
}

// WaveformID for unmarshalling SeisCompML
type WaveformID struct {
	NetworkCode  string `xml:"networkCode,attr"`
	StationCode  string `xml:"stationCode,attr"`
	LocationCode string `xml:"locationCode,attr"`
	ChannelCode  string `xml:"channelCode,attr"`
}

// Value for unmarshalling SeisCompML
type Value struct {
	Value       float64 `xml:"value"`
	Uncertainty float64 `xml:"uncertainty"`
}

// TimeValue for unmarshalling SeisCompML
type TimeValue struct {
	Value       time.Time `xml:"value"`
	Uncertainty float64   `xml:"uncertainty"`
}

// Mag for unmarshalling SeisCompML
type Mag struct {
	Value       float64 `xml:"value"`
	Uncertainty float64 `xml:"uncertainty"`
}

// Magnitude for unmarshalling SeisCompML
type Magnitude struct {
	PublicID     string `xml:"publicID,attr"`
	Mag          Mag    `xml:"magnitude"`
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

// PickMap remaps the Pick information in the SeisCompML to allow for user selectable output.
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

// ArrivalMap remaps the Arrival information in the SeisCompML to allow for user selectable output.
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

// init performs initialisation functions on the SeisCompML.  Should be called called after unmarshal.
func (q *Seiscomp) init() (err error) {

	if q.EventParameters.Event.PreferredOriginID == "" {
		err = errors.New(fmt.Sprint("Empty PreferredOriginID"))
		return err
	}

	if q.EventParameters.Event.PreferredMagnitudeID == "" {
		err = errors.New(fmt.Sprint("Empty PreferredMagnitudeID"))
		return err
	}

	if len(q.EventParameters.O) == 0 {
		err = errors.New(fmt.Sprint("Found no origins"))
		return err
	}

	q.EventParameters.Event.P = make([]Pick, len(q.EventParameters.P))
	copy(q.EventParameters.Event.P, q.EventParameters.P)

	q.EventParameters.Event.O = make([]Origin, len(q.EventParameters.O))
	copy(q.EventParameters.Event.O, q.EventParameters.O)

	q.EventParameters.Event.M = make([]Magnitude, 0)

	for _, origin := range q.EventParameters.Event.O {
		q.EventParameters.Event.M = append(q.EventParameters.Event.M, origin.M...)
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

// unmarshal unmarshalls the SeisCompML
func unmarshal(b []byte) (e Event, err error) {
	var q Seiscomp
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

// Fetcher reads eventids, fetches, unmarshals, and returns SeisCompML.
func fetcher(done <-chan struct{}, eventids <-chan string, c chan<- result) {
	client := &http.Client{}

	for publicid := range eventids {

		var b []byte
		var e Event

		r, err := client.Get(seiscompURL + publicid + ".xml")
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

// Get retrives SeisCompML for each EventID.  Errors are logged but not returned.
func Get(eventid []string) (seiscompml map[string]Event) {
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

	seiscompml = make(map[string]Event)
	var i = 0
	for r := range c {
		if r.err != nil {
			log.Println("Error fetching data for " + r.publicID)
			log.Println(r.err)
		} else {
			seiscompml[r.publicID] = r.event
			i++
		}
		if i == 50 {
			log.Printf("Downloaded %v quakes", len(seiscompml))
			i = 0
		}
	}
	return seiscompml
}
