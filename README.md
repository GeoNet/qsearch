# qsearch

qsearch - a command line tool to retrieve quake catalog information.

# Example Queries

Event information in a time window.  Include a header on the ouput:

```
qsearch --start 2014-02-22T04:06:25Z --end 2014-02-22T05:06:25Z --header --event --event-format EventID,Latitude,Longitude
```

Arrival information for the preferred origin for each event found in a time window.  No header line on the output:

```
qsearch --start 2014-02-24T04:06:25Z --end 2014-02-24T05:06:25Z --preferred-origin-arrivals --arrivals-format EventID,NetworkCode,StationCode,ChannelCode,LocationCode,Phase,PhaseTime,PhaseOriginOffset,TimeResidual,TimeWeight
```

Pick information for events found in a time window.

```
qsearch --start 2014-02-22T04:06:25Z --end 2014-02-22T05:06:25Z --picks --picks-format ChannelCode,EventID,LocationCode,NetworkCode,PhaseHint,PhaseTime,StationCode
```

Arrival information for a single event:

```
qsearch --eventid 2014p240753 --preferred-origin-arrivals --arrivals-format EventID,NetworkCode,StationCode,ChannelCode,LocationCode,Phase,PhaseTime,PhaseOriginOffset,TimeResidual,TimeWeight
```

Search for events with a minimum magnitude and phase count in a bounding box around Wellington:

```
qsearch --start 2000-02-22T04:06:25Z --end 2014-02-22T05:06:25Z --min-used-phase-count 30 --min-magnitude 4.8 --bbox 174,-41,175,-42 --event --event-format EventID,Latitude,Longitude,Magnitude 
```

# Usage

```
qsearch --help
```

## Search Criteria

### Single Event  

Use a valid GeoNet eventid.  No other search criteria can be combined with eventid. 

```
qsearch --eventid 2014p557808
```

### Advanced Search

All advanced queries must have a time range to search in.  Provide a start and end time to search in using ISO8601 format.  Comparison is >= and <= respectively  

```
qsearch  --start 2014-02-24T04:06:25Z --end 2014-02-24T05:06:25Z 
```

#### min-used-phase-count

Minimum number of phases used to locate and event.  Comparison is >= e.g.

```
qsearch --start 2010-02-22T04:06:25Z --end 2014-02-22T05:06:25Z ... --min-used-phase-count 60
```

#### min-magnitude

Min magnitude of the event.  Comparison is >= e.g.,

```
qsearch --start 2010-02-22T04:06:25Z --end 2014-02-22T05:06:25Z ... --min-magnitude 5.8
```

#### bbox

A bounding box to search for quakes in.  Provide upper left and lower right corners as a comma separated string e.g.,

```
qsearch --start 2010-02-22T04:06:25Z --end 2014-02-22T05:06:25Z --bbox 174,-41,175,-42
```


## Output

A range of outputs are possible.  All outputs are in CSV format.  An optional header line can be included.  

The output choice has a large input on search performance.  Outputting event data only required querying the WFS.  Other outputs require retrieving additional information from the full QuakeML which is a much slower process.

All outputs require the specification of an output format.  This specifies the presence and order of columns in the ouput.  Any combination and order of ouput columns can be used. 

### --header

Outputs a single header line naming the columns in the output.

```
qsearch --header ...
```

### event

Output event information.  An output format must be defined as well.  This is a comma separated line of output column names for the quake information. 

e.g., 

```
qsearch ... --event --event-format EventID,Latitude,Longitude
``` 

Any combination and order of column names can be selected from:

* AzimuthalGap
* Depth
* DepthType
* EarthModel
* EvaluationMethod
* EvaluationMode
* EvaluationStatus
* EventID
* EventType
* Latitude
* Longitude
* Magnitude
* MagnitudeStationCount
* MagnitudeType
* MagnitudeUncertainty
* MinimumDistance
* ModificationTime
* OriginError
* OriginTime
* UsedPhaseCount
* UsedStationCount

### preferred-origin-arrivals

Output arrival information for the preferred origin.  Arrivals are picks that have been associated with an origin.  An output format must be defined as well.  This is a comma separated line of output column names for the arrival information. 

e.g., 

```
qsearch ... --preferred-origin-arrivals --arrivals-format EventID,NetworkCode,StationCode,ChannelCode,LocationCode,Phase,PhaseTime
``` 

Any combination and order of column names can be selected from:

* ChannelCode
* EventID
* LocationCode
* NetworkCode
* Phase
* PhaseOriginOffset
* PhaseTime
* StationCode
* TimeResidual
* TimeWeight

### picks

Output pick information for the event.  An output format must be defined as well.  This is a comma separated line of output column names for the pick information. 

e.g.,

```
qsearch ... --picks --picks-format EventID,StationCode,PhaseHint,PhaseTime
```
Any combination and order of column names can be selected from:

* ChannelCode
* EventID
* LocationCode
* NetworkCode
* PhaseHint
* PhaseTime
* StationCode

# Sorting the Output

Sorting is an expensive operation and the output from this program is not sorted in anyway.  It can be piped through unix sort.  ISO8601 date times are lexographically sortable.  