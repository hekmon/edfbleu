package main

import (
	"testing"
	"time"
)

// TestPriceBreakpoints20252026 checks the base, HC and HP prices right after each
// price breakpoint introduced by the 2025 and 2026 EDF Tarif Bleu revisions
// (the getters switch to a palier at its date 00:00 exclusive).
func TestPriceBreakpoints20252026(t *testing.T) {
	// parse the palier dates the same way prepareDates does
	var err error
	if prices0225, err = time.ParseInLocation(pricesDateFormat, "01/02/2025", frLocation); err != nil {
		t.Fatal(err)
	}
	if prices0825, err = time.ParseInLocation(pricesDateFormat, "01/08/2025", frLocation); err != nil {
		t.Fatal(err)
	}
	if prices0226, err = time.ParseInLocation(pricesDateFormat, "01/02/2026", frLocation); err != nil {
		t.Fatal(err)
	}
	if prices0826, err = time.ParseInLocation(pricesDateFormat, "01/08/2026", frLocation); err != nil {
		t.Fatal(err)
	}
	// each palier is checked at fixed times just after its start date: 02:00 falls
	// into the HC window (23H30-7H30) and 12:00 into the HP window
	testCases := []struct {
		palier time.Time
		base   float64
		hc     float64
		hp     float64
	}{
		{time.Date(2025, 2, 1, 0, 0, 0, 0, frLocation), 0.2016, 0.1696, 0.2146},
		{time.Date(2025, 8, 1, 0, 0, 0, 0, frLocation), 0.1952, 0.1635, 0.2081},
		{time.Date(2026, 2, 1, 0, 0, 0, 0, frLocation), 0.1940, 0.1579, 0.2065},
		{time.Date(2026, 8, 1, 0, 0, 0, 0, frLocation), 0.2001, 0.1589, 0.2142},
	}
	for _, testCase := range testCases {
		baseTime := testCase.palier.Add(12 * time.Hour)
		hcTime := testCase.palier.Add(2 * time.Hour)
		hpTime := testCase.palier.Add(12 * time.Hour)
		if got := getBasePrice(baseTime); got != testCase.base {
			t.Errorf("wrong base price at %s: %f instead of %f", baseTime, got, testCase.base)
		}
		if got := getHCPrice(hcTime); got != testCase.hc {
			t.Errorf("wrong HC price at %s: %f instead of %f", hcTime, got, testCase.hc)
		}
		if got := getHCPrice(hpTime); got != testCase.hp {
			t.Errorf("wrong HP price at %s: %f instead of %f", hpTime, got, testCase.hp)
		}
	}
}
