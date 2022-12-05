package main

import "time"

const (
	pricesDateFormat = "02/01/2006"
)

var (
	prices2022 time.Time
	prices2021 time.Time
)

func init() {
	// get paris location
	parisLocation, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		panic(err)
	}
	// setup prices start dates
	prices2022, err = time.ParseInLocation(pricesDateFormat, "01/08/2022", parisLocation)
	if err != nil {
		panic(err)
	}
	prices2021, err = time.ParseInLocation(pricesDateFormat, "01/08/2021", parisLocation)
	if err != nil {
		panic(err)
	}
}

func getBasePrice(datetime time.Time) float64 {
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
