package main

import (
	"fmt"
	"time"
)

type zenFlexDay string

const (
	sobriete zenFlexDay = "sobriete"
	eco      zenFlexDay = "eco"
)

var (
	sobrieteDays []time.Time
)

// TODO check API to get Zen Flex Days
func generateZenFlexDays() (err error) {
	// setup sobriete days
	// WARNING: The following list is an example from Tempo Red days
	// TODO: update dates with real value when EDF will publish calendar
	// 20 "sobriete" days between October 15th and April 15th
	var day time.Time
	sobrieteDates := []string{
		"02/03/2023",
		"01/03/2023",
		"10/02/2023",
		"09/02/2023",
		"08/02/2023",
		"07/02/2023",
		"06/02/2023",
		"31/01/2023",
		"27/01/2023",
		"26/01/2023",
		"25/01/2023",
		"23/01/2023",
		"20/01/2023",
		"19/01/2023",
		"18/01/2023",
		"17/01/2023",
		"14/12/2022",
		"13/12/2022",
		"12/12/2022",
		"08/12/2022",
	}
	for index, sobrieteDate := range sobrieteDates {
		day, err = time.ParseInLocation(pricesDateFormat, sobrieteDate, frLocation)
		if err != nil {
			return fmt.Errorf("failed to parse zen flex sobriete day at index %d: %s", index, err)
		}
		sobrieteDays = append(sobrieteDays, day)
	}
	return
}

func getZenFlexDayColor(datetime time.Time) zenFlexDay {
	// adjust date for early HC
	if datetime.Hour() < 8 {
		// we need to take the color of the previous day
		datetime = datetime.Add(-1 * 9 * time.Hour)
	}
	// search sobriete days
	for _, sobrieteDay := range sobrieteDays {
		if sobrieteDay.Year() == datetime.Year() && sobrieteDay.Month() == datetime.Month() && sobrieteDay.Day() == datetime.Day() {
			return sobriete
		}
	}
	return eco
}
