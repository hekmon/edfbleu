package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
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

type CSVHeader struct {
	PRMID string
	Start time.Time
	End   time.Time
	Step  time.Duration
}

type point struct {
	Time  time.Time
	Value float64 // kW
	Conso float64 // kWh
}

func parseFile(path string) (header CSVHeader, data []point, err error) {
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
	if header, err = parseHeader(cr); err != nil {
		err = fmt.Errorf("failed to parse header: %w", err)
		return
	}
	// Parse data
	if data, err = parseData(cr, header.Step); err != nil {
		err = fmt.Errorf("failed to parse data: %w", err)
		return
	}
	return
}

func parseHeader(cr *csv.Reader) (header CSVHeader, err error) {
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

func parseData(cr *csv.Reader, stepping time.Duration) (data []point, err error) {
	// nb of records changes for data
	cr.FieldsPerRecord = 2
	// remove data header
	_, err = cr.Read()
	if err != nil {
		err = fmt.Errorf("failed to read the third line: %w", err)
		return
	}
	// Process data lines
	var (
		records     []string
		line        int
		recordTime  time.Time
		recordValue int
		computedkWh float64
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
				stepping := recordTime.Sub(prevPoint.Time)
				computedkWh = float64(recordValue) / 1000 / float64(time.Hour/stepping)
				if line == dataStartLine+1 {
					// compute the first point kWh using the same stepping
					firstPoint := data[len(data)-1]
					firstPoint.Conso = firstPoint.Value / 1000 / float64(time.Hour/stepping)
					data[len(data)-1] = firstPoint
				}
			}
			// else first point, can not compute stepping without a previous point, waiting for second point to retro compute first point
		} else {
			computedkWh = float64(recordValue) / 1000 / float64(time.Hour/stepping)
		}
		// save value
		data = append(data, point{
			Time:  recordTime.In(frLocation),   // make sure every date time in this program is in the same loc
			Value: float64(recordValue) / 1000, // convert W to kW
			Conso: computedkWh,
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
