package main

import (
	"flag"
	"fmt"
	"github.com/GeoNet/qsearch/quakeml12"
	"github.com/GeoNet/qsearch/wfs"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"
)

func main() {

	// Parse and validate command line flags.

	pickFormat := quakeml12.PickFormat()
	arrivalFormat := quakeml12.ArrivalFormat()
	eventFormat := wfs.EventFormat()

	eventid := flag.String("eventid", "", "a valid eventid for a GeoNet event e.g., --eventid 2012p070732.  If specifying eventid then start and end are not needed.")
	var start = flag.String("start", "", "start date time for the search in ISO8601 format to s precision e.g., 2014-02-22T04:06:25Z")
	var end = flag.String("end", "", "end date time for the search in ISO8601 format to s precision e.g., 2014-02-22T05:06:25Z")
	var poArrivals = flag.Bool("preferred-origin-arrivals", false,
		"output Arrival information for the PreferredOrigin.  An arrival-format must be specified.  An Arrival is a Pick associated with an Origin.")
	var arrivalsF = flag.String("arrivals-format", "",
		"output format selector for Arrival information.  Any combination and any order of the following values, separated by ',': "+formatString(arrivalFormat))
	var event = flag.Bool("event", false, "output event information.  An event-format must be specified.")
	var eventF = flag.String("event-format", "",
		"output format selector for event information.  Any combination and any order of the following values, separated by ',': "+formatString(eventFormat))
	var picks = flag.Bool("picks", false, "output Pick information for the Event.  A pick-format must be specified.")
	var picksF = flag.String("picks-format", "",
		"output format selector for Pick information.  Any combination and any order of the following values, separated by ',': "+formatString(pickFormat))
	var header = flag.Bool("header", false, "turns off the output of a header line.")
	var minUsedPhaseCount = flag.Int("min-used-phase-count", -999, "the minimum used phase count.  Comparison is >=")
	var minMagnitude = flag.Float64("min-magnitude", -999.9, "the minimum magnitude.  Comparison is >=")
	var bbox = flag.String("bbox", "", "search for quakes inside the bbox - a comma separated string of upper left and lower right bounday box coordinates for e.g., 174,-41,175,-42")

	flag.Parse()

	// Check that each output option has a format provided and that all the format parameters are legal keys.

	if *event && *eventF == "" {
		log.Fatal("--event selected but no --event-format provided.")
	}

	if *event {
		checkFormat(eventF, eventFormat)
	}

	if *picks && *picksF == "" {
		log.Fatal("--picks selected but no --picks-format provided.")
	}

	if *picks {
		checkFormat(picksF, pickFormat)
	}

	if *poArrivals && *arrivalsF == "" {
		log.Fatal("--arrivals selected but no --arrivals-format provided.")
	}

	if *poArrivals {
		checkFormat(arrivalsF, arrivalFormat)
	}

	var query wfs.Query

	if !(*start == "" || *end == "") {
		s, err := time.Parse(time.RFC3339, *start)
		if err != nil {
			log.Fatal(err)
		}
		e, err := time.Parse(time.RFC3339, *end)
		if err != nil {
			log.Fatal(err)
		}
		if s.After(e) {
			log.Fatal("start time is after end time.")
		}
		query = wfs.Query{Start: s, End: e, MinUsedPhaseCount: *minUsedPhaseCount, MinMagnitude: *minMagnitude, Bbox: *bbox}
	} else if *eventid != "" {
		pidr, _ := regexp.Compile("^[a-z0-9]+$")
		if !pidr.MatchString(*eventid) {
			log.Fatal("invalid eventid.")
		}
		query = wfs.Query{EventID: *eventid}
	}

	log.Printf("Searching the WFS")

	quakes, err := query.Get()
	if err != nil {
		log.Println("Error searching WFS.")
		log.Fatal(err)
	}

	log.Printf("Found %v quakes from the WFS.\n", len(quakes))

	// Fetch QuakeML information if it is required in the output

	var quakeml map[string]quakeml12.Event

	if *picks || *poArrivals {

		quakeml = make(map[string]quakeml12.Event)

		log.Printf("Searching for QuakeML.  This can take some time.\n")

		i := 0
		x := make([]string, len(quakes))
		for _, e := range quakes {
			x[i] = e["EventID"]
			i++
		}

		quakeml = quakeml12.Get(x)

		log.Printf("Found quake details for %v quakes.", len(quakeml))

		if len(quakes) > len(quakeml) {
			log.Printf("Failed to find details for %v quakes.  These might be in The Gap.\n", len(quakes)-len(quakeml))
			log.Println("Please see http://info.geonet.org.nz/display/appdata/The+Gap.")
		}
	}

	// Output.
	//
	// These all follow the same pattern.  The user supplies a list of ',' separated fields that they want to output
	// the values for.  This is split into a slice and then used to lookup the required values in a Map of the data.

	if *event {
		oF := strings.Split(*eventF, ",")
		o := make([]string, len(oF))
		if *header {
			fmt.Println(strings.Join(oF, ","))
		}
		for _, v := range quakes {
			for i, n := range oF {
				o[i] = v[n]
			}
			fmt.Println(strings.Join(o, ","))
		}
	}

	if *picks {
		oF := strings.Split(*picksF, ",")
		o := make([]string, len(oF))
		if *header {
			fmt.Println(strings.Join(oF, ","))
		}

		for eid, e := range quakeml {
			for _, v := range e.PickMap() {
				// Add the publicid from the WFS search, rather than the logical one from in the QuakeML.
				v["EventID"] = eid
				for i, n := range oF {
					o[i] = v[n]
				}
				fmt.Println(strings.Join(o, ","))
			}
		}
	}

	if *poArrivals {
		oF := strings.Split(*arrivalsF, ",")
		o := make([]string, len(oF))
		if *header {
			fmt.Println(strings.Join(oF, ","))
		}

		for eid, e := range quakeml {
			for _, v := range e.PreferredOrigin.ArrivalMap() {
				// Add the publicid from the WFS search, rather than the logical one from in the QuakeML.
				v["EventID"] = eid
				for i, n := range oF {
					o[i] = v[n]
				}
				fmt.Println(strings.Join(o, ","))
			}
		}
	}
}

// checkFormat checks that all comma separated strings in f have a key in the map.
// used to validate the user input format string.
func checkFormat(f *string, v map[string]string) {
	for _, s := range strings.Split(*f, ",") {
		if _, present := v[s]; !present {
			log.Fatal("Invalid format key: " + s)
		}
	}

}

// formatString returns a format string for docmentation.  This gives allowable entries in the user provided
// format string.
func formatString(v map[string]string) (s string) {
	st := make([]string, len(v))
	i := 0
	for f, _ := range v {
		st[i] = f
		i++
	}
	sort.Sort(sort.StringSlice(st))
	return strings.Join(st, ",")
}
