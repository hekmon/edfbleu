package main

import (
	"fmt"
	"time"
)

const (
	lastDataUpdate   = "22/01/2023"
	pricesDateFormat = "02/01/2006"
)

var (
	lastUpdate time.Time
	prices2023 time.Time
	prices2022 time.Time
	prices2021 time.Time
)

func prepareDates() (err error) {
	// Last update
	lastUpdate, err = time.ParseInLocation(pricesDateFormat, lastDataUpdate, frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the last data update date: %w", err)
	}
	// setup prices start dates
	prices2023, err = time.ParseInLocation(pricesDateFormat, "01/01/2023", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new prices january 2023 date: %w", err)
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
	if datetime.After(prices2023) {
		return 0.1740
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
	if datetime.After(prices2023) {
		if hc {
			return 0.1470
		}
		return 0.1841
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
	if datetime.After(prices2023) {
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
