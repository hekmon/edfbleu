package main

import (
	"time"
)

func getBasePrice(datetime time.Time) float64 {
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
