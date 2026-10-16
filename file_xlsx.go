package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// XLSX exports of the Enedis web portal are read with excelize. Two exports are supported:
// the daily energy export and the half-hourly load curve export, both being detected from the
// column labels of their data sheet rather than from the file names.

const (
	// column labels of the data sheet of the daily Enedis export ("Export Consommation Quotidienne")
	xlsxDailyDateLabel  = "Date"
	xlsxDailyValueLabel = "Valeur (en kWh)"
	// column labels of the data sheet of the half-hourly Enedis export ("Puissance moyenne soutirée")
	xlsxStartLabel = "Début"
	xlsxEndLabel   = "Fin"
	xlsxValueLabel = "Valeur (en kW)"
	// data cell formats
	xlsxDailyDateFormat  = "02/01/2006"          // 31/12/2021
	xlsxRecordDateFormat = "02/01/2006 15:04:05" // 31/12/2021 00:30:00
	xlsxMissingValue     = "NA"                  // legend: "Données indisponibles"
	xlsxPRMLength        = 14                    // a PRM is a 14 digits number
	// daily exports deliver one record per day
	dailyRecordStep = 24 * time.Hour
)

// parseXLSXFile parses a XLSX export of the Enedis web portal: both the daily energy export and
// the half-hourly load curve export are supported, the format being detected from the column
// labels of the data sheet
func parseXLSXFile(path string) (header FileHeader, data []point, err error) {
	// Read the worksheets
	var sheets []xlsxSheet
	if sheets, err = readXLSXSheets(path); err != nil {
		err = fmt.Errorf("failed to read the XLSX file: %w", err)
		return
	}
	// Find the data sheet (an export can also carry a production sheet or legal notices)
	var sheet xlsxDataSheet
	if sheet, err = selectXLSXDataSheet(path, sheets); err != nil {
		return
	}
	// Parse the records
	header.PRMID = sheet.findPRM()
	if sheet.daily {
		header.Daily = true
		header.Step = dailyRecordStep
		data, err = sheet.parseDailyRecords()
	} else {
		header.Step, data, err = sheet.parseHalfHourlyRecords()
	}
	if err != nil {
		return
	}
	if len(data) == 0 {
		err = fmt.Errorf("no data record found in the data sheet")
		return
	}
	// The covered range is taken from the records themselves
	header.Start = data[0].Time.Add(-data[0].Duration)
	header.End = data[len(data)-1].Time
	return
}

// xlsxSheet is a worksheet of an XLSX export: its rows of cell values by column index
type xlsxSheet struct {
	rows []xlsxRow
}

// xlsxRow is a worksheet row: the cell values by spreadsheet column index (0 is the A column)
type xlsxRow struct {
	num   int
	cells map[int]string
}

// xlsxDataSheet locates the data records within a worksheet of an Enedis export
type xlsxDataSheet struct {
	rows      []xlsxRow
	daily     bool
	headerRow int // index of the row holding the column labels in rows
	// columns holding the data fields
	dateCol, startCol, endCol, valueCol int
}

// readXLSXSheets extracts the cell values of every worksheet of an XLSX file with excelize.
// The raw cell values are used so the values are read the way Enedis stores them: the display
// formats of the spreadsheet (like the French decimal comma) are not applied.
func readXLSXSheets(path string) (sheets []xlsxSheet, err error) {
	var workbook *excelize.File
	if workbook, err = excelize.OpenFile(path, excelize.Options{RawCellValue: true}); err != nil {
		return
	}
	defer workbook.Close()
	for _, name := range workbook.GetSheetList() {
		var rows [][]string
		if rows, err = workbook.GetRows(name); err != nil {
			err = fmt.Errorf("failed to read the worksheet %q: %w", name, err)
			return
		}
		sheet := xlsxSheet{rows: make([]xlsxRow, 0, len(rows))}
		for index, values := range rows {
			row := xlsxRow{num: index + 1, cells: make(map[int]string)}
			for column, value := range values {
				if value != "" {
					row.cells[column] = value
				}
			}
			sheet.rows = append(sheet.rows, row)
		}
		sheets = append(sheets, sheet)
	}
	if len(sheets) == 0 {
		err = fmt.Errorf("no worksheet in the XLSX file")
	}
	return
}

// selectXLSXDataSheet picks the worksheet holding the consumption records among the worksheets
// of an export: dual metering exports also carry a production sheet which can not be priced here
func selectXLSXDataSheet(path string, sheets []xlsxSheet) (sheet xlsxDataSheet, err error) {
	var candidates, consumption []xlsxDataSheet
	for _, candidate := range sheets {
		if detected := detectXLSXDataSheet(candidate); detected != nil {
			candidates = append(candidates, *detected)
			if detected.isConsumption() {
				consumption = append(consumption, *detected)
			}
		}
	}
	switch {
	case len(consumption) == 1:
		sheet = consumption[0]
	case len(consumption) == 0 && len(candidates) == 1:
		sheet = candidates[0]
	case len(candidates) == 0:
		err = fmt.Errorf("no supported data sheet (daily or half-hourly) found in %s", path)
	default:
		err = fmt.Errorf("several data sheets found in %s: keep only the consumption one (the production sheet can not be priced)", path)
	}
	return
}

// detectXLSXDataSheet returns the layout of the data of a worksheet, or nil if the worksheet
// does not carry one of the supported Enedis exports. The layout is detected from the column
// labels so it does not depend on the exact rows and columns the export uses.
func detectXLSXDataSheet(sheet xlsxSheet) *xlsxDataSheet {
	for index, row := range sheet.rows {
		labels := make(map[string]int) // column label -> column index
		for column, value := range row.cells {
			switch value {
			case xlsxDailyDateLabel, xlsxDailyValueLabel, xlsxStartLabel, xlsxEndLabel, xlsxValueLabel:
				labels[value] = column
			}
		}
		if dateCol, foundDate := labels[xlsxDailyDateLabel]; foundDate {
			if valueCol, foundValue := labels[xlsxDailyValueLabel]; foundValue {
				return &xlsxDataSheet{
					rows:      sheet.rows,
					daily:     true,
					headerRow: index,
					dateCol:   dateCol,
					valueCol:  valueCol,
				}
			}
		}
		if startCol, foundStart := labels[xlsxStartLabel]; foundStart {
			if endCol, foundEnd := labels[xlsxEndLabel]; foundEnd {
				if valueCol, foundValue := labels[xlsxValueLabel]; foundValue {
					return &xlsxDataSheet{
						rows:      sheet.rows,
						headerRow: index,
						startCol:  startCol,
						endCol:    endCol,
						valueCol:  valueCol,
					}
				}
			}
		}
	}
	return nil
}

// isConsumption tells if the data sheet holds consumption (soutirée) records rather than
// production (injectée) ones, from the metadata displayed above the column labels
func (sheet xlsxDataSheet) isConsumption() bool {
	for _, row := range sheet.rows[:sheet.headerRow] {
		for _, value := range row.cells {
			value = strings.ToLower(value)
			if strings.Contains(value, "consommation") || strings.Contains(value, "soutir") {
				return true
			}
		}
	}
	return false
}

// findPRM returns the PRM (a 14 digits number) from the metadata displayed above the column
// labels, or an empty string when it can not be found
func (sheet xlsxDataSheet) findPRM() string {
	for _, row := range sheet.rows[:sheet.headerRow] {
		columns := make([]int, 0, len(row.cells))
		for column := range row.cells {
			columns = append(columns, column)
		}
		sort.Ints(columns)
		for _, column := range columns {
			if isPRM(row.cells[column]) {
				return row.cells[column]
			}
		}
	}
	return ""
}

// isPRM tells if a cell value looks like a PRM (a 14 digits number)
func isPRM(value string) bool {
	if len(value) != xlsxPRMLength {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

// parseDailyRecords parses the daily records of a daily energy export: one record per day,
// covering the whole day (23h, 24h or 25h across daylight saving time changes). The intraday
// consumption profile is not part of the export: only the base option can price these records
// exactly, the HC/HP and Tempo options are refused (see priceRecords).
func (sheet xlsxDataSheet) parseDailyRecords() (data []point, err error) {
	for _, row := range sheet.rows[sheet.headerRow+1:] {
		date := row.cells[sheet.dateCol]
		if date == "" {
			continue // empty (formatted but valueless) rows
		}
		value := row.cells[sheet.valueCol]
		if value == "" || strings.EqualFold(value, xlsxMissingValue) {
			continue // unavailable data
		}
		var day time.Time
		if day, err = time.ParseInLocation(xlsxDailyDateFormat, date, frLocation); err != nil {
			err = fmt.Errorf("invalid daily record date %q on row %d: %w", date, row.num, err)
			return
		}
		var kwh float64
		if kwh, err = strconv.ParseFloat(value, 64); err != nil {
			err = fmt.Errorf("invalid daily record value %q on row %d: %w", value, row.num, err)
			return
		}
		end := day.AddDate(0, 0, 1) // next midnight, daylight saving time safe
		duration := end.Sub(day)
		data = append(data, point{
			Time:     end,
			Duration: duration,
			Value:    kwh / duration.Hours(), // kW average power over the day
			Conso:    kwh,
		})
	}
	return
}

// parseHalfHourlyRecords parses the records of a half-hourly load curve export. Each record
// covers the period of time between its start and end timestamps and carries the average power
// withdrawn over it. The returned stepping is the record duration when it is fixed.
func (sheet xlsxDataSheet) parseHalfHourlyRecords() (step time.Duration, data []point, err error) {
	step = noSteppingValue
	var uniformDuration time.Duration
	for _, row := range sheet.rows[sheet.headerRow+1:] {
		start := row.cells[sheet.startCol]
		end := row.cells[sheet.endCol]
		if start == "" && end == "" {
			continue // empty (formatted but valueless) rows
		}
		value := row.cells[sheet.valueCol]
		if value == "" || strings.EqualFold(value, xlsxMissingValue) {
			continue // unavailable data
		}
		var startTime, endTime time.Time
		if startTime, err = time.ParseInLocation(xlsxRecordDateFormat, start, frLocation); err != nil {
			err = fmt.Errorf("invalid record start %q on row %d: %w", start, row.num, err)
			return
		}
		if endTime, err = time.ParseInLocation(xlsxRecordDateFormat, end, frLocation); err != nil {
			err = fmt.Errorf("invalid record end %q on row %d: %w", end, row.num, err)
			return
		}
		if !endTime.After(startTime) {
			err = fmt.Errorf("invalid record on row %d: end %v is not after start %v", row.num, endTime, startTime)
			return
		}
		var kw float64
		if kw, err = strconv.ParseFloat(value, 64); err != nil {
			err = fmt.Errorf("invalid record value %q on row %d: %w", value, row.num, err)
			return
		}
		duration := endTime.Sub(startTime)
		if len(data) == 0 {
			uniformDuration = duration
		} else if uniformDuration != duration {
			uniformDuration = noSteppingValue // the time step changed during the period
		}
		data = append(data, point{
			Time:     endTime, // Enedis timestamps the end of the covered period
			Duration: duration,
			Value:    kw,
			Conso:    kw * duration.Hours(),
		})
	}
	if uniformDuration != noSteppingValue {
		step = uniformDuration
	}
	return
}
