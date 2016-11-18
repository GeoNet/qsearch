package wfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestURL(t *testing.T) {
	s, _ := time.Parse(time.RFC3339, "2014-01-27T03:06:25Z")
	e, _ := time.Parse(time.RFC3339, "2014-01-27T04:06:25Z")

	q := Query{EventID: "2014p562279"}

	if !strings.HasSuffix(q.url(), "cql_filter=publicid=='2014p562279'") {
		t.Error("incorrect for eventid, got", q.url())
	}

	q = Query{Start: s, End: e, MinUsedPhaseCount: -999, MinMagnitude: -999.9, Bbox: ""}

	if !strings.HasSuffix(q.url(), "cql_filter=origintime>='2014-01-27T03:06:25'+AND+origintime<='2014-01-27T04:06:25'") {
		t.Error("incorrect for start and end times, got", q.url())
	}

	q = Query{EventID: "2014p562279", Start: s, End: e}

	if !strings.HasSuffix(q.url(), "cql_filter=publicid=='2014p562279'") {
		t.Error("incorrect for start and end times with eventid, got", q.url())
	}

	q = Query{Start: s, End: e, MinUsedPhaseCount: 60, MinMagnitude: -999.9, Bbox: ""}

	if !strings.HasSuffix(q.url(), "cql_filter=origintime>='2014-01-27T03:06:25'+AND+origintime<='2014-01-27T04:06:25'+AND+usedphasecount>=60") {
		t.Error("incorrect for min phase count, got", q.url())
	}

	q = Query{Start: s, End: e, MinUsedPhaseCount: -999, MinMagnitude: 6.1, Bbox: ""}

	if !strings.HasSuffix(q.url(), "cql_filter=origintime>='2014-01-27T03:06:25'+AND+origintime<='2014-01-27T04:06:25'+AND+magnitude>=6.1") {
		t.Error("incorrect for start and end times with magnitude, got", q.url())
	}

	q = Query{Start: s, End: e, MinUsedPhaseCount: 60, MinMagnitude: 6.1, Bbox: ""}

	if !strings.HasSuffix(q.url(), "cql_filter=origintime>='2014-01-27T03:06:25'+AND+origintime<='2014-01-27T04:06:25'+AND+usedphasecount>=60+AND+magnitude>=6.1") {
		t.Error("incorrect for min phase count with magnitude, got", q.url())
	}

	q = Query{Start: s, End: e, MinUsedPhaseCount: 60, MinMagnitude: 6.1, Bbox: "174,-41,175,-42"}

	if !strings.HasSuffix(q.url(), "cql_filter=origintime>='2014-01-27T03:06:25'+AND+origintime<='2014-01-27T04:06:25'+AND+usedphasecount>=60+AND+magnitude>=6.1+AND+BBOX(origin_geom,174,-41,175,-42)") {
		t.Error("incorrect for min phase count with magnitude and bbox, got", q.url())
	}
}

func TestGet(t *testing.T) {
	s, _ := time.Parse(time.RFC3339, "2014-01-27T03:06:25Z")
	e, _ := time.Parse(time.RFC3339, "2014-01-27T04:06:25Z")

	query := Query{
		Start: s,
		End:   e}

	fs, _ := query.Get()

	if len(fs) < 1 {
		t.Error("Didn't find any events ")
	}
}

func TestUnmarshal(t *testing.T) {
	f, err := os.Open("2014p549007.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	b, _ := ioutil.ReadAll(f)

	fs, _ := unmarshal(b)

	e := fs[0].Properties

	if e.MagnitudeStationCount != 13 {
		t.Error("e.MagnitudeStationCount expected 13, got ", e.MagnitudeStationCount)
	}
	if e.MagnitudeUncertainty != 0 {
		t.Error("e.MagnitudeUncertainty expected 0, got ", e.MagnitudeUncertainty)
	}
	if e.MagnitudeType != "M" {
		t.Error("e.MagnitudeType expected M, got ", e.MagnitudeType)
	}
	if e.AzimuthalGap != 206.88617 {
		t.Error("e.AzimuthalGap expected 206.88617, got ", e.AzimuthalGap)
	}
	if e.MinimumDistance != 0.38872472 {
		t.Error("e.MinimumDistance expected 0.38872472, got ", e.MinimumDistance)
	}
	if e.AzimuthalGap != 206.88617 {
		t.Error("e.MinimumDistance expected 206.88617, got ", e.MinimumDistance)
	}
	if e.UsedPhaseCount != 23 {
		t.Error("e.UsedPhaseCount expected 23, got ", e.UsedPhaseCount)
	}
	if e.UsedStationCount != 23 {
		t.Error("e.UsedStationCount expected 23, got ", e.UsedStationCount)
	}
	if e.PublicID != "2014p549333" {
		t.Error("e.PublicID expected 2014p549333, got ", e.PublicID)
	}
	if e.ModificationTime != "2014-07-23T06:07:12.232Z" {
		t.Error("e.ModificationTime expected 2014-07-23T06:07:12.232Z, got ", e.ModificationTime)
	}
	if e.OriginTime != "2014-07-23T06:04:43.625Z" {
		t.Error("e.OriginTime expected 2014-07-23T06:04:43.625Z, got ", e.OriginTime)
	}
	if e.Latitude != -39.648535 {
		t.Error("e.Latitude expected -39.648535, got ", e.Latitude)
	}
	if e.Longitude != 173.47803 {
		t.Error("e.Longitude expected 173.47803, got ", e.Longitude)
	}
	if e.Depth != 7.34375 {
		t.Error("e.Depth expected 7.34375, got ", e.Depth)
	}
	if e.Magnitude != 2.6416703 {
		t.Error("e.Magnitude expected 2.6416703, got ", e.Magnitude)
	}
	if e.EvaluationMethod != "NonLinLoc" {
		t.Error("e.EvaluationMethod expected NonLinLoc, got ", e.EvaluationMethod)
	}
	if e.EvaluationMode != "automatic" {
		t.Error("e.EvaluationMode expected automatic, got ", e.EvaluationMode)
	}
	if e.EarthModel != "nz3drx" {
		t.Error("e.EarthModel expected nz3drx, got ", e.EarthModel)
	}
	if e.OriginError != 0.48022989 {
		t.Error("e.OriginError expected 0.48022989, got ", e.OriginError)
	}
}
