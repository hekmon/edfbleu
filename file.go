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
	enedisCustomCSVSep     = ';'
	headerFields           = 9
	headerDateFormat       = "02/01/2006"                // 31/12/2021
	dataDateFormat         = "2006-01-02T15:04:05-07:00" // 2021-12-31T00:30:00+01:00
	enedisHourlyExportStep = 30 * time.Minute
)

type CSVHeader struct {
	PRMID string
	Start time.Time
	End   time.Time
}

type point struct {
	Time  time.Time // 30min end
	Value float64   // kWh
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
	if data, err = parseData(cr); err != nil {
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
	parisLocation, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		err = fmt.Errorf("failed to get paris timezone: %w", err)
		return
	}
	header.PRMID = records[0]
	header.Start, err = time.ParseInLocation(headerDateFormat, records[2], parisLocation)
	if err != nil {
		err = fmt.Errorf("failed to parse start date from second line: %w", err)
		return
	}
	header.End, err = time.ParseInLocation(headerDateFormat, records[3], parisLocation)
	if err != nil {
		err = fmt.Errorf("failed to parse end date from second line: %w", err)
		return
	}
	return
}

func parseData(cr *csv.Reader) (data []point, err error) {
	// nb of records changes for data
	cr.FieldsPerRecord = 2
	// remove data header
	_, err = cr.Read()
	if err != nil {
		err = fmt.Errorf("failed to read the third line: %w", err)
		return
	}
	// Process data
	var (
		records     []string
		line        int
		recordTime  time.Time
		recordValue int
	)
	data = make([]point, 0, 365*24*2) // most people will analyse a full year (make more sense for tempo)
	for line = 4; ; line++ {
		// read line
		records, err = cr.Read()
		if err != nil {
			err = fmt.Errorf("failed to parse line: %w", err)
			break
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
		// checks
		if recordTime.Minute() != 30 && recordTime.Minute() != 0 {
			err = fmt.Errorf("minutes should always be 00 or 30: %v", recordTime)
			break
		}
		if recordTime.Second() != 0 {
			err = fmt.Errorf("seconds should always be 00: %v", recordTime)
			break
		}
		// save value
		data = append(data, point{
			Time:  recordTime,
			Value: float64(recordValue) / 1000, // convert Wh to kWh
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
