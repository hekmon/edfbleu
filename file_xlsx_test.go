package main

import (
	"archive/zip"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	var err error
	if frLocation, err = time.LoadLocation("Europe/Paris"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

/*
	Minimal XLSX workbooks builder for the tests
*/

// testXLSXSheet is a worksheet of a test workbook: a list of rows of cells in reading order
type testXLSXSheet []testXLSXRow

// testXLSXRow is a worksheet row: a list of cells in column order
type testXLSXRow []testXLSXCell

// testXLSXCell is a worksheet cell: values are written as shared strings unless numeric is set
type testXLSXCell struct {
	column  string
	value   string
	numeric bool
}

// shared writes a text cell stored in the shared string table (as Enedis does for dates)
func shared(column, value string) testXLSXCell {
	return testXLSXCell{column: column, value: value}
}

// number writes a numeric cell (as Enedis does for the values)
func number(column, value string) testXLSXCell {
	return testXLSXCell{column: column, value: value, numeric: true}
}

// writeTestXLSX builds a minimal XLSX workbook file and returns its path. Text values are
// written as shared strings when inline is false, as inline strings otherwise.
func writeTestXLSX(t *testing.T, sheets []testXLSXSheet, inline bool) string {
	t.Helper()
	// Build the shared string table
	var sharedStrings []string
	sharedIndex := func(value string) int {
		for index, known := range sharedStrings {
			if known == value {
				return index
			}
		}
		sharedStrings = append(sharedStrings, value)
		return len(sharedStrings) - 1
	}
	// Build the worksheet parts
	var worksheets []string
	for _, sheet := range sheets {
		var rows []string
		for rowIndex, row := range sheet {
			var cells []string
			for _, cell := range row {
				ref := fmt.Sprintf("%s%d", cell.column, rowIndex+1)
				switch {
				case cell.numeric:
					cells = append(cells, fmt.Sprintf(`<c r="%s"><v>%s</v></c>`, ref, cell.value))
				case inline:
					cells = append(cells, fmt.Sprintf(`<c r="%s" t="inlineStr"><is><t>%s</t></is></c>`, ref, xmlEscape(cell.value)))
				default:
					cells = append(cells, fmt.Sprintf(`<c r="%s" t="s"><v>%d</v></c>`, ref, sharedIndex(cell.value)))
				}
			}
			rows = append(rows, fmt.Sprintf(`<row r="%d">%s</row>`, rowIndex+1, strings.Join(cells, "")))
		}
		worksheets = append(worksheets, fmt.Sprintf(
			`<?xml version="1.0" encoding="UTF-8"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>%s</sheetData></worksheet>`,
			strings.Join(rows, "")))
	}
	// Write the archive
	path := filepath.Join(t.TempDir(), "export.xlsx")
	fd, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()
	archive := zip.NewWriter(fd)
	// Package parts (content types, relationships and workbook index)
	writeZipEntry(t, archive, "[Content_Types].xml", fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`+
			`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`+
			`<Default Extension="xml" ContentType="application/xml"/>`+
			`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`+
			`<Override PartName="/xl/sharedStrings.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"/>%s</Types>`,
		strings.Join(contentTypesOverrides(len(worksheets)), "")))
	writeZipEntry(t, archive, "_rels/.rels",
		`<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`+
			`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`)
	var sheetsXML, workbookRels []string
	for index := range worksheets {
		sheetsXML = append(sheetsXML, fmt.Sprintf(`<sheet name="Sheet%d" sheetId="%d" r:id="rId%d"/>`, index+1, index+1, index+1))
		workbookRels = append(workbookRels, fmt.Sprintf(
			`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`,
			index+1, index+1))
	}
	workbookRels = append(workbookRels, fmt.Sprintf(
		`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>`,
		len(worksheets)+1))
	writeZipEntry(t, archive, "xl/workbook.xml", fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>%s</sheets></workbook>`,
		strings.Join(sheetsXML, "")))
	writeZipEntry(t, archive, "xl/_rels/workbook.xml.rels", fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">%s</Relationships>`,
		strings.Join(workbookRels, "")))
	// Shared string table (unless the text values are written as inline strings)
	if !inline {
		var items []string
		for _, value := range sharedStrings {
			items = append(items, `<si><t>`+xmlEscape(value)+`</t></si>`)
		}
		writeZipEntry(t, archive, "xl/sharedStrings.xml", fmt.Sprintf(
			`<?xml version="1.0" encoding="UTF-8"?><sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">%s</sst>`,
			strings.Join(items, "")))
	}
	for index, worksheet := range worksheets {
		writeZipEntry(t, archive, fmt.Sprintf("xl/worksheets/sheet%d.xml", index+1), worksheet)
	}
	if err = archive.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// contentTypesOverrides builds the content types of the worksheet parts
func contentTypesOverrides(count int) (overrides []string) {
	for index := 1; index <= count; index++ {
		overrides = append(overrides, fmt.Sprintf(
			`<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`,
			index))
	}
	return
}

func writeZipEntry(t *testing.T, archive *zip.Writer, name, content string) {
	t.Helper()
	fd, err := archive.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fd.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
}

// xmlEscape escapes the XML special characters of a text value
func xmlEscape(value string) string {
	for _, replacement := range [][2]string{{"&", "&amp;"}, {"<", "&lt;"}, {">", "&gt;"}} {
		value = strings.ReplaceAll(value, replacement[0], replacement[1])
	}
	return value
}

// testHomeSheet builds the legal notices worksheet of an Enedis export
func testHomeSheet() testXLSXSheet {
	return testXLSXSheet{
		{shared("B6", "Les données de ce fichier sont destinées à l'usage exclusif du titulaire du PRM.")},
		{shared("B7", "Les informations de ce document sont transmises à titre indicatif et n'ont pas de valeur contractuelle.")},
		{shared("B10", "Ce fichier est constitué des données de consommation et/ou de production quotidienne.")},
	}
}

// testDailySheet builds the data worksheet of a daily Enedis export (records are dates and
// daily consumption values, an empty value standing for an unavailable day)
func testDailySheet(comptage string, records ...testXLSXRow) testXLSXSheet {
	sheet := testXLSXSheet{
		{shared("B6", "Compte Client"), shared("E6", "Export des données de l'énergie")},
		{shared("B8", "Point Référence Mesure (PRM) : "), shared("C8", "50050741158553"), shared("E8", "Type de comptage : "), shared("F8", comptage)},
		{shared("E9", "Date de début : "), shared("F9", "29/03/2025"), shared("E10", "Date de fin : "), shared("F10", "31/03/2025")},
		{shared("B10", "NA"), shared("C10", "Données indisponibles")},
		{shared("B14", "Date"), shared("C14", "Valeur (en kWh)")},
	}
	return append(sheet, records...)
}

// testHalfHourlySheet builds the data worksheet of a half-hourly Enedis export (records are
// start and end timestamps and average power values, an empty value standing for an
// unavailable record)
func testHalfHourlySheet(records ...testXLSXRow) testXLSXSheet {
	sheet := testXLSXSheet{
		{shared("B6", "Compte Client"), shared("D6", "Export des données de la courbe de charge")},
		{shared("B8", "Point Référence Mesure (PRM) : "), shared("C8", "50050741158553"), shared("D8", "Grandeur physique : "), shared("E8", "Puissance Active")},
		{shared("D9", "Type de comptage : "), shared("E9", "Puissance moyenne soutirée")},
		{shared("D10", "Date de début : "), shared("E10", "01/01/2025"), shared("D11", "Date de fin : "), shared("E11", "02/01/2025")},
		{shared("B10", "NA"), shared("C10", "Données indisponibles")},
		{shared("C15", "Plage horaire")},
		{shared("C16", "Début"), shared("D16", "Fin"), shared("E16", "Valeur (en kW)")},
	}
	return append(sheet, records...)
}

/*
	Tests
*/

func TestParseXLSXDaily(t *testing.T) {
	path := writeTestXLSX(t, []testXLSXSheet{
		testHomeSheet(),
		testDailySheet("Consommation Quotidienne",
			testXLSXRow{shared("B", "29/03/2025"), number("C", "12.5")},
			testXLSXRow{shared("B", "30/03/2025"), number("C", "13.5")}, // daylight saving time change: 23h day
			testXLSXRow{shared("B", "31/03/2025"), number("C", "11")},
			testXLSXRow{shared("B", "01/04/2025")},                    // unavailable value
			testXLSXRow{shared("B", "02/04/2025"), shared("C", "NA")}, // unavailable data
		),
	}, false)
	header, data, err := parseXLSXFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Header
	if header.PRMID != "50050741158553" {
		t.Errorf("wrong PRM: %q", header.PRMID)
	}
	if !header.Daily {
		t.Error("daily export not detected as daily")
	}
	if header.Step != dailyRecordStep {
		t.Errorf("wrong stepping: %v", header.Step)
	}
	start := time.Date(2025, 3, 29, 0, 0, 0, 0, frLocation)
	if !header.Start.Equal(start) {
		t.Errorf("wrong start: %v instead of %v", header.Start, start)
	}
	end := time.Date(2025, 4, 1, 0, 0, 0, 0, frLocation)
	if !header.End.Equal(end) {
		t.Errorf("wrong end: %v instead of %v", header.End, end)
	}
	// Records: 2 days of 24h and one 23h day, spread over half-hours (the unavailable days are skipped)
	if len(data) != 48+46+48 {
		t.Fatalf("wrong record count: %d instead of %d", len(data), 48+46+48)
	}
	for _, record := range data {
		if record.Duration != dailySpreadStep {
			t.Fatalf("unexpected record duration: %v", record.Duration)
		}
	}
	var sum float64
	for _, record := range data {
		sum += record.Conso
	}
	expectedTotal := 12.5 + 13.5 + 11
	if math.Abs(sum-expectedTotal) > 1e-9 {
		t.Errorf("wrong consumption total: %f instead of %f", sum, expectedTotal)
	}
	// The first day must account for exactly 12.5 kWh over its 48 half-hours
	var firstDay float64
	for _, record := range data[:48] {
		firstDay += record.Conso
	}
	if math.Abs(firstDay-12.5) > 1e-9 {
		t.Errorf("wrong first day consumption: %f instead of 12.5", firstDay)
	}
	// The daylight saving time day must spread over the real length of the day (23h = 46 half-hours)
	var dstDay float64
	for _, record := range data[48 : 48+46] {
		dstDay += record.Conso
	}
	if math.Abs(dstDay-13.5) > 1e-9 {
		t.Errorf("wrong daylight saving time day consumption: %f instead of 13.5", dstDay)
	}
	// The records must tile the days exactly
	for index := 1; index < len(data); index++ {
		if !data[index].Time.Add(-data[index].Duration).Equal(data[index-1].Time) {
			t.Fatalf("hole or overlap between records %d and %d", index-1, index)
		}
	}
}

func TestParseXLSXHalfHourly(t *testing.T) {
	path := writeTestXLSX(t, []testXLSXSheet{
		testHomeSheet(),
		testHalfHourlySheet(
			testXLSXRow{shared("C", "01/01/2025 00:00:00"), shared("D", "01/01/2025 00:30:00"), number("E", "1.038")},
			testXLSXRow{shared("C", "01/01/2025 00:30:00"), shared("D", "01/01/2025 01:00:00"), number("E", "0.5")},
			testXLSXRow{shared("C", "01/01/2025 01:00:00"), shared("D", "01/01/2025 01:30:00")},                    // unavailable value
			testXLSXRow{shared("C", "01/01/2025 01:30:00"), shared("D", "01/01/2025 02:00:00"), shared("E", "NA")}, // unavailable data
			testXLSXRow{shared("C", "01/01/2025 02:00:00"), shared("D", "01/01/2025 03:00:00"), number("E", "0.4")},
		),
	}, false)
	header, data, err := parseXLSXFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if header.Daily {
		t.Error("half-hourly export detected as daily")
	}
	if header.PRMID != "50050741158553" {
		t.Errorf("wrong PRM: %q", header.PRMID)
	}
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, frLocation)
	if !header.Start.Equal(start) {
		t.Errorf("wrong start: %v instead of %v", header.Start, start)
	}
	end := time.Date(2025, 1, 1, 3, 0, 0, 0, frLocation)
	if !header.End.Equal(end) {
		t.Errorf("wrong end: %v instead of %v", header.End, end)
	}
	if len(data) != 3 {
		t.Fatalf("wrong record count: %d instead of 3", len(data))
	}
	// Each record covers its own period: 30min, 30min and 1h here (the stepping is not fixed anymore)
	wanted := []point{
		{Time: time.Date(2025, 1, 1, 0, 30, 0, 0, frLocation), Duration: 30 * time.Minute, Value: 1.038, Conso: 1.038 * 0.5},
		{Time: time.Date(2025, 1, 1, 1, 0, 0, 0, frLocation), Duration: 30 * time.Minute, Value: 0.5, Conso: 0.25},
		{Time: time.Date(2025, 1, 1, 3, 0, 0, 0, frLocation), Duration: time.Hour, Value: 0.4, Conso: 0.4},
	}
	for index, want := range wanted {
		got := data[index]
		if !got.Time.Equal(want.Time) || got.Duration != want.Duration {
			t.Errorf("record %d: got %v (%v) instead of %v (%v)", index, got.Time, got.Duration, want.Time, want.Duration)
		}
		if math.Abs(got.Conso-want.Conso) > 1e-9 {
			t.Errorf("record %d: got %f kWh instead of %f", index, got.Conso, want.Conso)
		}
	}
	// The record durations are not fixed (30min then 1h): the stepping can not be a single value
	if header.Step != noSteppingValue {
		t.Errorf("stepping should not be fixed when record durations vary, got %v", header.Step)
	}
}

func TestParseXLSXFixedStepping(t *testing.T) {
	// Uniform record durations give a fixed stepping
	path := writeTestXLSX(t, []testXLSXSheet{
		testHalfHourlySheet(
			testXLSXRow{shared("C", "01/01/2025 00:00:00"), shared("D", "01/01/2025 00:30:00"), number("E", "1")},
			testXLSXRow{shared("C", "01/01/2025 00:30:00"), shared("D", "01/01/2025 01:00:00"), number("E", "1")},
		),
	}, false)
	header, _, err := parseXLSXFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if header.Step != 30*time.Minute {
		t.Errorf("wrong stepping: %v", header.Step)
	}
}

func TestParseXLSXInlineStrings(t *testing.T) {
	// Same export as TestParseXLSXHalfHourly but with inline strings instead of shared ones
	path := writeTestXLSX(t, []testXLSXSheet{
		testHalfHourlySheet(
			testXLSXRow{shared("C", "01/01/2025 00:00:00"), shared("D", "01/01/2025 00:30:00"), number("E", "1.038")},
		),
	}, true)
	header, data, err := parseXLSXFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if header.PRMID != "50050741158553" || len(data) != 1 || math.Abs(data[0].Conso-1.038*0.5) > 1e-9 {
		t.Errorf("wrong parse of inline strings: PRM %q, %d records", header.PRMID, len(data))
	}
}

func TestParseXLSXNoDataSheet(t *testing.T) {
	path := writeTestXLSX(t, []testXLSXSheet{
		testHomeSheet(),
	}, false)
	if _, _, err := parseXLSXFile(path); err == nil {
		t.Error("expected an error on a workbook without data sheet")
	}
}

func TestParseXLSXDualMetering(t *testing.T) {
	// Dual metering exports also carry a production sheet: only the consumption one is priced
	production := testDailySheet("Production Quotidienne",
		testXLSXRow{shared("B", "29/03/2025"), number("C", "42")},
	)
	path := writeTestXLSX(t, []testXLSXSheet{
		testHomeSheet(),
		production,
		testDailySheet("Consommation Quotidienne",
			testXLSXRow{shared("B", "29/03/2025"), number("C", "12.5")},
		),
	}, false)
	_, data, err := parseXLSXFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var sum float64
	for _, record := range data {
		sum += record.Conso
	}
	if math.Abs(sum-12.5) > 1e-9 {
		t.Errorf("the consumption sheet was not selected: %f kWh instead of 12.5", sum)
	}
}

func TestMergeFiles(t *testing.T) {
	day := func(hour int, duration time.Duration) point {
		return point{
			Time:     time.Date(2025, 1, 1, hour, 0, 0, 0, frLocation),
			Duration: duration,
		}
	}
	base := parsedFile{
		path:   "first.xlsx",
		header: FileHeader{PRMID: "1", Step: 30 * time.Minute},
		data:   []point{day(2, 30*time.Minute), day(1, 30*time.Minute)}, // unsorted on purpose
	}
	// Sorting
	header, data, err := mergeFiles([]parsedFile{base, {
		path:   "second.xlsx",
		header: FileHeader{PRMID: "1", Step: 30 * time.Minute},
		data:   []point{day(3, 30*time.Minute)},
	}})
	if err != nil {
		t.Fatal(err)
	}
	for index := 1; index < len(data); index++ {
		if data[index].Time.Before(data[index-1].Time) {
			t.Fatal("merged records are not sorted")
		}
	}
	if header.Step != 30*time.Minute {
		t.Errorf("wrong merged stepping: %v", header.Step)
	}
	// Daily and half-hourly records can not be mixed
	if _, _, err = mergeFiles([]parsedFile{
		{path: "daily.xlsx", header: FileHeader{Daily: true}},
		{path: "halfhourly.xlsx", header: FileHeader{Daily: false}},
	}); err == nil {
		t.Error("expected an error when mixing daily and half-hourly exports")
	}
	// Records of different PRMs can not be mixed
	if _, _, err = mergeFiles([]parsedFile{
		{path: "first.xlsx", header: FileHeader{PRMID: "1"}},
		{path: "second.xlsx", header: FileHeader{PRMID: "2"}},
	}); err == nil {
		t.Error("expected an error when mixing different PRMs")
	}
	// Overlapping files are refused
	if _, _, err = mergeFiles([]parsedFile{
		{path: "first.xlsx", data: []point{day(1, 30*time.Minute)}},
		{path: "second.xlsx", data: []point{day(1, 30*time.Minute)}},
	}); err == nil {
		t.Error("expected an error when merging overlapping files")
	}
	// Empty inputs are refused
	if _, _, err = mergeFiles(nil); err == nil {
		t.Error("expected an error when merging no file")
	}
}

func TestParseCSVAndXLSXParity(t *testing.T) {
	// Same half-hourly data set as a legacy Enedis CSV export and as a XLSX one: both parsers
	// must return the same records (the CSV values are average powers in W)
	type record struct {
		start, end string
		kw         float64
	}
	records := []record{
		{"01/01/2025 00:00:00", "01/01/2025 00:30:00", 1.038},
		{"01/01/2025 00:30:00", "01/01/2025 01:00:00", 0.544},
		{"01/01/2025 01:00:00", "01/01/2025 01:30:00", 0.328},
	}
	// The XLSX export
	var rows testXLSXSheet
	for _, record := range records {
		rows = append(rows, testXLSXRow{
			shared("C", record.start), shared("D", record.end),
			number("E", fmt.Sprintf("%g", record.kw)),
		})
	}
	xlsxPath := writeTestXLSX(t, []testXLSXSheet{testHalfHourlySheet(rows...)}, false)
	// The legacy CSV export (timestamps at the end of the periods, values in W)
	var csv strings.Builder
	fmt.Fprintln(&csv, "Identifiant PRM;Type de donnees;Date de debut;Date de fin;Grandeur physique;Unite;Pas;Nature;Pas en minutes")
	fmt.Fprintln(&csv, "50050741158553;Consommation;01/01/2025;02/01/2025;Puissance moyenne soutirée;W;30;Brutes;30")
	fmt.Fprintln(&csv, "Horodate;Valeur")
	for _, record := range records {
		end, err := time.ParseInLocation(xlsxRecordDateFormat, record.end, frLocation)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&csv, "%s;%d\n", end.Format(dataDateFormat), int(record.kw*1000))
	}
	csvPath := filepath.Join(t.TempDir(), "export.csv")
	if err := os.WriteFile(csvPath, []byte(csv.String()), 0600); err != nil {
		t.Fatal(err)
	}
	// Compare both parses
	xlsxHeader, xlsxData, err := parseXLSXFile(xlsxPath)
	if err != nil {
		t.Fatal(err)
	}
	csvHeader, csvData, err := parseCSVFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	if csvHeader.PRMID != xlsxHeader.PRMID || csvHeader.Step != xlsxHeader.Step {
		t.Errorf("header mismatch: %v vs %v", csvHeader, xlsxHeader)
	}
	if len(csvData) != len(xlsxData) {
		t.Fatalf("record count mismatch: %d vs %d", len(csvData), len(xlsxData))
	}
	for index := range csvData {
		if !csvData[index].Time.Equal(xlsxData[index].Time) || csvData[index].Duration != xlsxData[index].Duration {
			t.Errorf("record %d period mismatch: %v (%v) vs %v (%v)", index,
				csvData[index].Time, csvData[index].Duration, xlsxData[index].Time, xlsxData[index].Duration)
		}
		if math.Abs(csvData[index].Conso-xlsxData[index].Conso) > 1e-9 {
			t.Errorf("record %d consumption mismatch: %f vs %f", index, csvData[index].Conso, xlsxData[index].Conso)
		}
	}
}
