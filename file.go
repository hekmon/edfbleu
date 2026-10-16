package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	enedisCustomCSVSep = ';'
	headerFields       = 9
	headerDateFormat   = "02/01/2006"                // 31/12/2021
	dataDateFormat     = "2006-01-02T15:04:05-07:00" // 2021-12-31T00:30:00+01:00
	noSteppingValue    = -1
	dataStartLine      = 4
	defaultStepping    = 30 * time.Minute
)

// FileHeader summarizes the content of a parsed input file
type FileHeader struct {
	PRMID string
	Start time.Time
	End   time.Time
	Step  time.Duration // time between records (noSteppingValue when it is not fixed)
	Daily bool          // records are daily aggregates without intraday details
}

// point is a consumption record: the average power and the energy consumed over a period of
// time ending at Time and lasting Duration (Enedis timestamps the end of the periods)
type point struct {
	Time     time.Time
	Duration time.Duration
	Value    float64 // kW average power over the period
	Conso    float64 // kWh consumed over the period
}

// parsedFile associates an input file with its parsed header and records
type parsedFile struct {
	path   string
	header FileHeader
	data   []point
}

// parseFile parses an Enedis export file: the legacy CSV export and the XLSX exports of the
// current Enedis web portal (daily and half-hourly) are supported
func parseFile(path string) (header FileHeader, data []point, err error) {
	if strings.EqualFold(filepath.Ext(path), ".xlsx") {
		return parseXLSXFile(path)
	}
	return parseCSVFile(path)
}

// parseFiles parses several Enedis export files and merges them into a single data set
func parseFiles(paths []string) (header FileHeader, data []point, err error) {
	if len(paths) == 0 {
		err = fmt.Errorf("no input file provided")
		return
	}
	files := make([]parsedFile, 0, len(paths))
	for _, path := range paths {
		var file parsedFile
		file.path = path
		if file.header, file.data, err = parseFile(path); err != nil {
			err = fmt.Errorf("failed to parse %s: %w", path, err)
			return
		}
		files = append(files, file)
	}
	return mergeFiles(files)
}

// mergeFiles merges the records of several parsed input files into a single data set sorted
// chronologically. All the files must cover the same PRM and use the same kind of records.
func mergeFiles(files []parsedFile) (header FileHeader, data []point, err error) {
	if len(files) == 0 {
		err = fmt.Errorf("no input file to merge")
		return
	}
	// Check that the files can be merged
	header = files[0].header
	header.Start, header.End = files[0].header.Start, files[0].header.End
	for _, file := range files[1:] {
		if file.header.Daily != header.Daily {
			err = fmt.Errorf("can not mix daily and half-hourly exports (%s and %s)", files[0].path, file.path)
			return
		}
		if file.header.PRMID != "" && header.PRMID != "" && file.header.PRMID != header.PRMID {
			err = fmt.Errorf("can not mix exports from different PRMs (%s: %s and %s: %s)", files[0].path, header.PRMID, file.path, file.header.PRMID)
			return
		}
		if header.PRMID == "" {
			header.PRMID = file.header.PRMID
		}
		if file.header.Step != header.Step {
			header.Step = noSteppingValue
		}
		if file.header.Start.Before(header.Start) {
			header.Start = file.header.Start
		}
		if file.header.End.After(header.End) {
			header.End = file.header.End
		}
	}
	// Gather all the records and sort them chronologically
	for _, file := range files {
		data = append(data, file.data...)
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Time.Before(data[j].Time)
	})
	// Check for overlapping files and warn about missing records
	for index := 1; index < len(data); index++ {
		previous, current := data[index-1], data[index]
		if current.Time == previous.Time {
			err = fmt.Errorf("overlapping input files: several records end at %v", current.Time)
			return
		}
		gap := current.Time.Add(-current.Duration).Sub(previous.Time)
		if gap > 0 && gap >= previous.Duration {
			fmt.Printf("/!\\ missing data between %v and %v: the consumption of this period is not accounted for\n", previous.Time, current.Time.Add(-current.Duration))
		}
	}
	if len(data) == 0 {
		err = fmt.Errorf("no data record found in the input files")
	}
	return
}

func parseCSVFile(path string) (header FileHeader, data []point, err error) {
	// Prepare the CSV reader
	fd, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("failed to open file: %w", err)
		return
	}
	defer fd.Close()
	cr := csv.NewReader(fd)
	cr.Comma = enedisCustomCSVSep
	cr.ReuseRecord = true
	// Parse the header
	if header, err = parseCSVHeader(cr); err != nil {
		err = fmt.Errorf("failed to parse header: %w", err)
		return
	}
	// Parse data
	if data, err = parseCSVData(cr, header.Step); err != nil {
		err = fmt.Errorf("failed to parse data: %w", err)
		return
	}
	return
}

func parseCSVHeader(cr *csv.Reader) (header FileHeader, err error) {
	var records []string
	// line 1
	records, err = cr.Read()
	if err != nil {
		err = fmt.Errorf("failed to read first line: %w", err)
		return
	}
	if len(records) != headerFields {
		err = fmt.Errorf("invalid headers, expecting %d got %d: %s",
			headerFields, len(records), strings.Join(records, ", "))
		return
	}
	// line 2
	records, err = cr.Read()
	if err != nil {
		err = fmt.Errorf("failed to read second line: %w", err)
		return
	}
	header.PRMID = records[0]
	header.Start, err = time.ParseInLocation(headerDateFormat, records[2], frLocation)
	if err != nil {
		err = fmt.Errorf("failed to parse start date from second line: %w", err)
		return
	}
	header.End, err = time.ParseInLocation(headerDateFormat, records[3], frLocation)
	if err != nil {
		err = fmt.Errorf("failed to parse end date from second line: %w", err)
		return
	}
	step_str := records[8]
	if step_str != "" {
		var step int
		if step, err = strconv.Atoi(step_str); err != nil {
			err = fmt.Errorf("non integer stepping found: %s", err)
			return
		}
		header.Step = time.Duration(step) * time.Minute
	} else {
		// step empty
		header.Step = noSteppingValue
	}
	return
}

func parseCSVData(cr *csv.Reader, stepping time.Duration) (data []point, err error) {
	// nb of records changes for data
	cr.FieldsPerRecord = -1 // variable, some Enedis exports append trailing empty fields to data lines
	// remove data header
	_, err = cr.Read()
	if err != nil {
		err = fmt.Errorf("failed to read the third line: %w", err)
		return
	}
	// Process data lines
	var (
		records        []string
		line           int
		recordTime     time.Time
		recordValue    int
		recordDuration time.Duration
		computedkWh    float64
	)
	if stepping == noSteppingValue {
		data = make([]point, 0, 365*24*(time.Hour/defaultStepping))
	} else {
		data = make([]point, 0, 365*24*(time.Hour/stepping))
	}
	for line = dataStartLine; ; line++ {
		// read line
		records, err = cr.Read()
		if err != nil {
			err = fmt.Errorf("failed to parse line: %w", err)
			break
		}
		// tolerate trailing empty fields appended by some Enedis exports
		for len(records) > 2 && records[len(records)-1] == "" {
			records = records[:len(records)-1]
		}
		if len(records) != 2 {
			err = fmt.Errorf("invalid data line, expecting 2 fields got %d: %s", len(records), strings.Join(records, ";"))
			break
		}
		// skip recorp if empty
		if records[1] == "" {
			continue
		}
		// parse line
		if recordTime, err = time.Parse(dataDateFormat, records[0]); err != nil {
			err = fmt.Errorf("failed to parse record date time: %w", err)
			break
		}
		if recordValue, err = strconv.Atoi(records[1]); err != nil {
			err = fmt.Errorf("failed to parse record value: %w", err)
			break
		}
		// check
		if recordTime.Second() != 0 {
			err = fmt.Errorf("seconds should always be 00: %v", recordTime)
			break
		}
		// Determine stepping if needed
		if stepping == noSteppingValue {
			if line > dataStartLine {
				// get previous point and compute stepping
				prevPoint := data[len(data)-1]
				recordDuration = recordTime.Sub(prevPoint.Time)
				computedkWh = float64(recordValue) / 1000 / float64(time.Hour/recordDuration)
				if line == dataStartLine+1 {
					// compute the first point kWh using the same stepping
					firstPoint := data[len(data)-1]
					firstPoint.Conso = firstPoint.Value / 1000 / float64(time.Hour/recordDuration)
					firstPoint.Duration = recordDuration
					data[len(data)-1] = firstPoint
				}
			}
			// else first point, can not compute stepping without a previous point, waiting for second point to retro compute first point
		} else {
			recordDuration = stepping
			computedkWh = float64(recordValue) / 1000 / float64(time.Hour/recordDuration)
		}
		// save value
		data = append(data, point{
			Time:     recordTime.In(frLocation),   // make sure every date time in this program is in the same loc
			Duration: recordDuration,              // 0 until retro computed for the first point of an auto stepped file
			Value:    float64(recordValue) / 1000, // convert W to kW
			Conso:    computedkWh,
		})
	}
	if errors.Is(err, io.EOF) {
		err = nil
	}
	if err != nil {
		err = fmt.Errorf("error at line %d: %w", line, err)
	}
	return
}
