package main

import (
	"fmt"
	"time"
)

const (
	lastDataUpdate   = "17/10/2026"
	pricesDateFormat = "02/01/2006"
)

var (
	lastUpdate   time.Time
	prices0826   time.Time
	prices0226   time.Time
	prices0825   time.Time
	prices0225   time.Time
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
	prices0826, err = time.ParseInLocation(pricesDateFormat, "01/08/2026", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new prices August 2026 date: %w", err)
	}
	prices0226, err = time.ParseInLocation(pricesDateFormat, "01/02/2026", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new prices February 2026 date: %w", err)
	}
	prices0825, err = time.ParseInLocation(pricesDateFormat, "01/08/2025", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new prices August 2025 date: %w", err)
	}
	prices0225, err = time.ParseInLocation(pricesDateFormat, "01/02/2025", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new prices February 2025 date: %w", err)
	}
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
	// then try to refresh them from the API (compiled ones are kept as fallback)
	updateTempoDaysFromAPI()
	return
}

func getBasePrice(datetime time.Time) float64 {
	if datetime.After(prices0826) {
		return 0.2001
	}
	if datetime.After(prices0226) {
		return 0.1940
	}
	if datetime.After(prices0825) {
		return 0.1952
	}
	if datetime.After(prices0225) {
		return 0.2016
	}
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
	if datetime.After(prices0826) {
		if hc {
			return 0.1589
		}
		return 0.2142
	}
	if datetime.After(prices0226) {
		if hc {
			return 0.1579
		}
		return 0.2065
	}
	if datetime.After(prices0825) {
		if hc {
			return 0.1635
		}
		return 0.2081
	}
	if datetime.After(prices0225) {
		if hc {
			return 0.1696
		}
		return 0.2146
	}
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
	if datetime.After(prices0826) {
		switch getTempoDayColor(datetime) {
		case tempoRed:
			if hc {
				return 0.1615
			}
			return 0.7295
		case tempoWhite:
			if hc {
				return 0.1536
			}
			return 0.1921
		case tempoBlue:
			if hc {
				return 0.1356
			}
			return 0.1654
		}
	}
	if datetime.After(prices0226) {
		switch getTempoDayColor(datetime) {
		case tempoRed:
			if hc {
				return 0.1575
			}
			return 0.7060
		case tempoWhite:
			if hc {
				return 0.1499
			}
			return 0.1871
		case tempoBlue:
			if hc {
				return 0.1325
			}
			return 0.1612
		}
	}
	if datetime.After(prices0825) {
		switch getTempoDayColor(datetime) {
		case tempoRed:
			if hc {
				return 0.1460
			}
			return 0.6468
		case tempoWhite:
			if hc {
				return 0.1391
			}
			return 0.1730
		case tempoBlue:
			if hc {
				return 0.1232
			}
			return 0.1494
		}
	}
	if datetime.After(prices0225) {
		switch getTempoDayColor(datetime) {
		case tempoRed:
			if hc {
				return 0.1518
			}
			return 0.6586
		case tempoWhite:
			if hc {
				return 0.1447
			}
			return 0.1792
		case tempoBlue:
			if hc {
				return 0.1288
			}
			return 0.1552
		}
	}
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
