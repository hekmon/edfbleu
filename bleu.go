package main

import (
	"fmt"
	"time"
)

const (
	lastDataUpdate   = "02/04/2024"
	pricesDateFormat = "02/01/2006"
)

var (
	lastUpdate   time.Time
	prices2024   time.Time
	prices2023   time.Time
	prices022023 time.Time
	prices2022   time.Time
	prices2021   time.Time
)

func prepareDates() (err error) {
	// Last update
	lastUpdate, err = time.ParseInLocation(pricesDateFormat, lastDataUpdate, frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the last data update date: %w", err)
	}
	// setup prices start dates
	prices2024, err = time.ParseInLocation(pricesDateFormat, "01/02/2024", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new prices February 2024 date: %w", err)
	}
	prices2023, err = time.ParseInLocation(pricesDateFormat, "01/08/2023", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new prices August 2023 date: %w", err)
	}
	prices022023, err = time.ParseInLocation(pricesDateFormat, "01/02/2023", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new prices February 2023 date: %w", err)
	}
	prices2022, err = time.ParseInLocation(pricesDateFormat, "01/08/2022", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new prices august 2022 date: %w", err)
	}
	prices2021, err = time.ParseInLocation(pricesDateFormat, "01/08/2021", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new prices august 2021 date: %w", err)
	}
	// generate tempo days
	if err = generateTempoDays(); err != nil {
		return fmt.Errorf("failed to generate tempo days: %w", err)
	}
	return
}

func getBasePrice(datetime time.Time) float64 {
	if datetime.After(prices2024) {
		return 0.2516
	}
	if datetime.After(prices2023) {
		return 0.2276
	}
	if datetime.After(prices022023) {
		return 0.2062
	}
	if datetime.After(prices2022) {
		return 0.1740
	}
	if datetime.After(prices2021) {
		return 0.1740
	}
	return 0
}

// standard range 23H30-7H30
func getHCPrice(datetime time.Time) float64 {
	// are we within hc range ?
	var hc bool
	if datetime.Hour() == 23 && datetime.Minute() >= 30 {
		hc = true
	} else {
		if datetime.Hour() < 7 {
			hc = true
		} else if datetime.Hour() == 7 && datetime.Minute() < 30 {
			hc = true
		}
	}
	// return price
	if datetime.After(prices2024) {
		if hc {
			return 0.2068
		}
		return 0.2700
	}
	if datetime.After(prices2023) {
		if hc {
			return 0.1828
		}
		return 0.2460
	}
	if datetime.After(prices022023) {
		if hc {
			return 0.1615
		}
		return 0.2228
	}
	if datetime.After(prices2022) {
		if hc {
			return 0.1470
		}
		return 0.1841
	}
	if datetime.After(prices2021) {
		if hc {
			return 0.1470
		}
		return 0.1841
	}
	return 0
}

func getTempoPrice(datetime time.Time) float64 {
	// are we within hc range ?
	var hc bool
	if datetime.Hour() < 6 || datetime.Hour() >= 22 {
		hc = true
	}
	// return price
	if datetime.After(prices2024) {
		switch getTempoDayColor(datetime) {
		case tempoRed:
			if hc {
				return 0.1568
			}
			return 0.7562
		case tempoWhite:
			if hc {
				return 0.1486
			}
			return 0.1894
		case tempoBlue:
			if hc {
				return 0.1296
			}
			return 0.1609
		}
	}
	if datetime.After(prices2023) {
		switch getTempoDayColor(datetime) {
		case tempoRed:
			if hc {
				return 0.1328
			}
			return 0.7324
		case tempoWhite:
			if hc {
				return 0.1246
			}
			return 0.1654
		case tempoBlue:
			if hc {
				return 0.1056
			}
			return 0.1369
		}
	}
	if datetime.After(prices022023) {
		switch getTempoDayColor(datetime) {
		case tempoRed:
			if hc {
				return 0.1216
			}
			return 0.6712
		case tempoWhite:
			if hc {
				return 0.1140
			}
			return 0.1508
		case tempoBlue:
			if hc {
				return 0.0970
			}
			return 0.1249
		}
	}
	if datetime.After(prices2022) {
		switch getTempoDayColor(datetime) {
		case tempoRed:
			if hc {
				return 0.1222
			}
			return 0.5486
		case tempoWhite:
			if hc {
				return 0.1112
			}
			return 0.1653
		case tempoBlue:
			if hc {
				return 0.0862
			}
			return 0.1272
		}
	}
	if datetime.After(prices2021) {
		switch getTempoDayColor(datetime) {
		case tempoRed:
			if hc {
				return 0.1488
			}
			return 0.6371
		case tempoWhite:
			if hc {
				return 0.1392
			}
			return 0.1738
		case tempoBlue:
			if hc {
				return 0.1242
			}
			return 0.1531
		}
	}
	return 0
}
