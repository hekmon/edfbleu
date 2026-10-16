package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Minimal XLSX (Office Open XML) reader built on the standard library so this program keeps
// having no external dependency: an XLSX file is a ZIP archive of XML parts, the ones of
// interest here being the worksheets (the cells) and the shared string table (the text values
// stored once and referenced by the cells). Only what is needed to read the Enedis exports is
// implemented.

const (
	xlsxSharedStringsPart = "xl/sharedStrings.xml"
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
	// daily exports deliver one record per day, spread over half-hours to be priced
	dailyRecordStep = 24 * time.Hour
	dailySpreadStep = 30 * time.Minute
)

// parseXLSXFile parses a XLSX export of the Enedis web portal: both the daily energy export and
// the half-hourly load curve export are supported, the format being detected from the column
// labels of the data sheet
func parseXLSXFile(path string) (header FileHeader, data []point, err error) {
	// Read the worksheets
	var sheets [][]xlsxRow
	if sheets, err = readXLSX(path); err != nil {
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

// xlsxRow is a worksheet row: the cell values by spreadsheet column letter ("B", "AC", ...)
type xlsxRow struct {
	num   int
	cells map[string]string
}

// xlsxDataSheet locates the data records within a worksheet of an Enedis export
type xlsxDataSheet struct {
	rows      []xlsxRow
	daily     bool
	headerRow int // index of the row holding the column labels in rows
	// columns holding the data fields
	dateCol, startCol, endCol, valueCol string
}

// selectXLSXDataSheet picks the worksheet holding the consumption records among the worksheets
// of an export: dual metering exports also carry a production sheet which can not be priced here
func selectXLSXDataSheet(path string, sheets [][]xlsxRow) (sheet xlsxDataSheet, err error) {
	var candidates, consumption []xlsxDataSheet
	for _, rows := range sheets {
		if candidate := detectXLSXDataSheet(rows); candidate != nil {
			candidates = append(candidates, *candidate)
			if candidate.isConsumption() {
				consumption = append(consumption, *candidate)
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
func detectXLSXDataSheet(rows []xlsxRow) *xlsxDataSheet {
	for index, row := range rows {
		labels := make(map[string]string) // column label -> column letters
		for column, value := range row.cells {
			switch value {
			case xlsxDailyDateLabel, xlsxDailyValueLabel, xlsxStartLabel, xlsxEndLabel, xlsxValueLabel:
				labels[value] = column
			}
		}
		if labels[xlsxDailyDateLabel] != "" && labels[xlsxDailyValueLabel] != "" {
			return &xlsxDataSheet{
				rows:      rows,
				daily:     true,
				headerRow: index,
				dateCol:   labels[xlsxDailyDateLabel],
				valueCol:  labels[xlsxDailyValueLabel],
			}
		}
		if labels[xlsxStartLabel] != "" && labels[xlsxEndLabel] != "" && labels[xlsxValueLabel] != "" {
			return &xlsxDataSheet{
				rows:      rows,
				headerRow: index,
				startCol:  labels[xlsxStartLabel],
				endCol:    labels[xlsxEndLabel],
				valueCol:  labels[xlsxValueLabel],
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
		for _, column := range sortedColumns(row.cells) {
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

// parseDailyRecords parses the daily records of a daily energy export and spreads each day
// uniformly over its half-hours: the intraday profile being unknown, this is an approximation
// (see the README), the base option being the only exact result
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
		data = append(data, spreadDay(day, kwh)...)
	}
	return
}

// spreadDay splits the daily consumption of a day uniformly over its half-hours so daily
// records can be priced with the half-hourly pricing pipeline. The day length is the real one
// (23h, 24h or 25h across daylight saving time changes) and its half-hours tile it exactly.
func spreadDay(day time.Time, kwh float64) (data []point) {
	end := day.AddDate(0, 0, 1) // next midnight, daylight saving time safe
	dayDuration := end.Sub(day)
	for start := day; start.Before(end); start = start.Add(dailySpreadStep) {
		sliceEnd := start.Add(dailySpreadStep)
		if sliceEnd.After(end) {
			sliceEnd = end
		}
		duration := sliceEnd.Sub(start)
		conso := kwh * float64(duration) / float64(dayDuration)
		data = append(data, point{
			Time:     sliceEnd,
			Duration: duration,
			Value:    conso / duration.Hours(), // kW average power over the slice
			Conso:    conso,
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

// readXLSX extracts the cell values of every worksheet of an XLSX file, as rows of values by
// column letters. Text values are resolved from the shared string table when there is one.
func readXLSX(path string) (sheets [][]xlsxRow, err error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return
	}
	defer archive.Close()
	// Shared string table (text values are stored once and referenced by the cells)
	var sharedStrings []string
	for _, part := range archive.File {
		if part.Name == xlsxSharedStringsPart {
			if sharedStrings, err = readXLSXSharedStrings(part); err != nil {
				err = fmt.Errorf("failed to read %s: %w", part.Name, err)
				return
			}
		}
	}
	// Worksheets
	for _, part := range archive.File {
		if !isXLSXWorksheet(part.Name) {
			continue
		}
		var rows []xlsxRow
		if rows, err = readXLSXWorksheet(part, sharedStrings); err != nil {
			err = fmt.Errorf("failed to read %s: %w", part.Name, err)
			return
		}
		sheets = append(sheets, rows)
	}
	if len(sheets) == 0 {
		err = fmt.Errorf("no worksheet in the XLSX archive")
	}
	return
}

// isXLSXWorksheet tells if a ZIP entry of an XLSX archive is a worksheet (chartsheets and
// relations are not)
func isXLSXWorksheet(name string) bool {
	return strings.HasPrefix(name, "xl/worksheets/") &&
		strings.HasSuffix(name, ".xml") &&
		!strings.Contains(name, "_rels")
}

// readXLSXSharedStrings reads the shared string table of an XLSX file
func readXLSXSharedStrings(part *zip.File) (sharedStrings []string, err error) {
	fd, err := part.Open()
	if err != nil {
		return
	}
	defer fd.Close()
	var table struct {
		Items []xlsxText `xml:"si"`
	}
	if err = xml.NewDecoder(fd).Decode(&table); err != nil {
		return
	}
	sharedStrings = make([]string, len(table.Items))
	for index, item := range table.Items {
		sharedStrings[index] = item.text
	}
	return
}

// readXLSXWorksheet reads the cells of a worksheet of an XLSX file
func readXLSXWorksheet(part *zip.File, sharedStrings []string) (rows []xlsxRow, err error) {
	fd, err := part.Open()
	if err != nil {
		return
	}
	defer fd.Close()
	var sheet struct {
		Rows []xlsxRowXML `xml:"sheetData>row"`
	}
	if err = xml.NewDecoder(fd).Decode(&sheet); err != nil {
		return
	}
	rows = make([]xlsxRow, 0, len(sheet.Rows))
	for index, rowXML := range sheet.Rows {
		row := xlsxRow{num: rowXML.Num, cells: make(map[string]string)}
		if row.num == 0 {
			row.num = index + 1 // missing row numbers are renumbered in order
		}
		for position, cellXML := range rowXML.Cells {
			var value string
			if value, err = cellXML.text(sharedStrings); err != nil {
				err = fmt.Errorf("cell %s: %w", cellXML.Ref, err)
				return
			}
			if value == "" {
				continue
			}
			row.cells[xlsxColumn(cellXML.Ref, position)] = value
		}
		rows = append(rows, row)
	}
	return
}

// xlsxRowXML is a worksheet row as stored in a worksheet part
type xlsxRowXML struct {
	Num   int           `xml:"r,attr"`
	Cells []xlsxCellXML `xml:"c"`
}

// xlsxCellXML is a worksheet cell as stored in a worksheet part
type xlsxCellXML struct {
	Ref    string   `xml:"r,attr"`
	Type   string   `xml:"t,attr"`
	Value  string   `xml:"v"`
	Inline xlsxText `xml:"is"`
}

// text returns the value of a cell as text: numbers are returned as-is and strings are
// resolved from the shared string table when needed
func (cell xlsxCellXML) text(sharedStrings []string) (string, error) {
	switch cell.Type {
	case "s": // shared string: the value holds the index in the shared string table
		index, err := strconv.Atoi(strings.TrimSpace(cell.Value))
		if err != nil || index < 0 || index >= len(sharedStrings) {
			return "", fmt.Errorf("invalid shared string reference %q", cell.Value)
		}
		return sharedStrings[index], nil
	case "inlineStr": // inline string: the text is stored in the cell itself
		return cell.Inline.text, nil
	default: // number (empty type) or formula result (str): the value holds the text
		return cell.Value, nil
	}
}

// xlsxText collects the text of a string element of an XLSX part: both plain strings
// (<t>text</t>) and rich strings made of several runs (<r><t>text</t></r>...) are supported
type xlsxText struct {
	text string
}

// UnmarshalXML walks through the element and concatenates the text of every <t> below it
func (t *xlsxText) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) (err error) {
	var text strings.Builder
	for {
		var token xml.Token
		if token, err = decoder.Token(); err != nil {
			return
		}
		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "t" {
				var part string
				if err = decoder.DecodeElement(&part, &element); err != nil {
					return
				}
				text.WriteString(part)
			}
			// any other element (rich text runs, run properties, ...) is walked through
		case xml.EndElement:
			if element.Name.Local == start.Name.Local {
				t.text = text.String()
				return
			}
		}
	}
}

// xlsxColumn returns the spreadsheet column letters of a cell reference (the "B12" reference
// is in the "B" column), falling back on the position of the cell within its row when the
// reference is missing
func xlsxColumn(ref string, position int) string {
	end := 0
	for end < len(ref) && isXLSXColumnLetter(ref[end]) {
		end++
	}
	if end == 0 {
		return xlsxColumnName(position)
	}
	return strings.ToUpper(ref[:end])
}

// isXLSXColumnLetter tells if a character can be part of a spreadsheet column name
func isXLSXColumnLetter(char byte) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')
}

// xlsxColumnName returns the spreadsheet column letters of a column position (0 is A, 26 is AA)
func xlsxColumnName(position int) string {
	name := ""
	for {
		name = string(rune('A'+position%26)) + name
		position = position/26 - 1
		if position < 0 {
			return name
		}
	}
}

// sortedColumns returns the column letters of a row in spreadsheet order
func sortedColumns(cells map[string]string) (columns []string) {
	for column := range cells {
		columns = append(columns, column)
	}
	for index := 1; index < len(columns); index++ {
		for cursor := index; cursor > 0 && columns[cursor] < columns[cursor-1]; cursor-- {
			columns[cursor], columns[cursor-1] = columns[cursor-1], columns[cursor]
		}
	}
	return
}
