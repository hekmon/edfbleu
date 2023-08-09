package main

import (
	"fmt"
	"time"
)

const (
	lastDataUpdate   = "09/08/2023"
	pricesDateFormat = "02/01/2006"
)

var (
	lastUpdate      time.Time
	zenPrices092023 time.Time
	zenPrices2023   time.Time
	prices2023      time.Time
	prices022023    time.Time
	prices2022      time.Time
	prices2021      time.Time
)

func prepareDates() (err error) {
	// Last update
	lastUpdate, err = time.ParseInLocation(pricesDateFormat, lastDataUpdate, frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the last data update date: %w", err)
	}
	// setup prices start dates
	zenPrices092023, err = time.ParseInLocation(pricesDateFormat, "14/09/2023", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new Zen prices September 14th 2023 date: %w", err)
	}
	zenPrices2023, err = time.ParseInLocation(pricesDateFormat, "01/08/2023", frLocation)
	if err != nil {
		return fmt.Errorf("failed to parse the new Zen prices August 2023 date: %w", err)
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
