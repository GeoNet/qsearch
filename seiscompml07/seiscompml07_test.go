package seiscompml07

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestUnmarshal(t *testing.T) {
	xmlFile, err := os.Open("etc/2012p070732-sc3.xml")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)

	e, _ := unmarshal(b)

	ot, _ := time.Parse(time.RFC3339Nano, "2012-01-27T04:06:25.369465Z")
	if e.PreferredOrigin.Time.Value != ot {
		t.Error("PreferredOrigin.Time.Value expected 2012-01-27T04:06:25.369465Z, got ", e.PreferredOrigin.Time.Value)
	}
	// Test data only has one Arrival in it to simplify testing.
	// fmt.Println(e.PreferredOrigin.Arrivals[0].Azimuth)
	if e.PreferredOrigin.Arrivals[0].Phase != "P" {
		t.Error("Arrivals[0].Phase expected P, got ", e.PreferredOrigin.Arrivals[0].Phase)
	}
	if e.PreferredOrigin.Arrivals[0].Azimuth != 302.3194377 {
		t.Error("Arrivals[0].Azimuth expected 302.3194377, got ", e.PreferredOrigin.Arrivals[0].Azimuth)
	}
	if e.PreferredOrigin.Arrivals[0].Distance != 0.1515531055 {
		t.Error("Arrivals[0].Distance expected 0.1515531055, got ", e.PreferredOrigin.Arrivals[0].Distance)
	}
	if e.PreferredOrigin.Arrivals[0].TimeResidual != -3.801403636e-13 {
		t.Error("Arrivals[0].TimeResidual expected -3.801403636e-13, got ", e.PreferredOrigin.Arrivals[0].TimeResidual)
	}
	if e.PreferredOrigin.Arrivals[0].TimeWeight != 1.532535963 {
		t.Error("Arrivals[0].TimeWeight expected 1.532535963, got ", e.PreferredOrigin.Arrivals[0].TimeWeight)
	}
	if e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].PhaseHint != "P" {
		t.Error("Pick.PhaseHint expected 1.532535963, got ", e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].PhaseHint)
	}
	if e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].WaveformID.NetworkCode != "NZ" {
		t.Error("Pick.WaveformID.NetworkCode expected NZ, got ",
			e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].WaveformID.NetworkCode)
	}
	if e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].WaveformID.StationCode != "WVZ" {
		t.Error("Pick.WaveformID.StationCode expected WVZ, got ",
			e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].WaveformID.StationCode)
	}
	if e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].WaveformID.LocationCode != "10" {
		t.Error("Pick.WaveformID.LocationCode expected 10, got ",
			e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].WaveformID.LocationCode)
	}
	if e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].WaveformID.ChannelCode != "HHZ" {
		t.Error("Pick.WaveformID.ChannelCode expected HHZ, got ",
			e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].WaveformID.ChannelCode)
	}
	if e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].EvaluationMode != "automatic" {
		t.Error("Pick.EvaluationMode expected automatic, got ",
			e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].EvaluationMode)
	}

	pt, _ := time.Parse(time.RFC3339Nano, "2012-01-27T04:06:29.798393Z")
	if e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].Time.Value != pt {
		t.Error("Pick.Time.Value expected 2012-01-27T04:06:29.798393Z, got ", e.Picks["20120127.040629.79-AIC-NZ.WVZ.10.HHZ"].Time)
	}

	if e.PreferredMagnitude.Type != "M" {
		t.Error("e.PreferredMagnitude.Type expected M, got ", e.PreferredMagnitude.Type)
	}
	if e.PreferredMagnitude.Mag.Value != 2.652616042 {
		t.Error("e.PreferredMagnitude.Mag.Value expected 2.652616042, got ", e.PreferredMagnitude.Mag.Value)
	}
	if e.PreferredMagnitude.Mag.Uncertainty != 0 {
		t.Error("e.PreferredMagnitude.Mag.Uncertainty expected 0, got ", e.PreferredMagnitude.Mag.Uncertainty)
	}
	if e.PreferredMagnitude.StationCount != 7 {
		t.Error("e.PreferredMagnitude.StationCount expected 7, got ", e.PreferredMagnitude.StationCount)
	}
	if e.PreferredMagnitude.MethodID != "weighted average" {
		t.Error("e.PreferredMagnitude.MethodID expected weighted_average, got ", e.PreferredMagnitude.MethodID)
	}
}

func TestUnmarshalBad(t *testing.T) {
	xmlFile, err := os.Open("etc/2012p070732-missing-sc3.xml")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)

	_, err = unmarshal(b)
	if err == nil {
		t.Error("should have got an error")
	}
}

func TestUnmarshalEmpty(t *testing.T) {
	xmlFile, err := os.Open("etc/999.xml")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)

	_, err = unmarshal(b)
	if err == nil {
		t.Error("should have got an error")
	}
}
