package main

import (
	"fmt"
	"time"
)

type tempoDay string

const (
	tempoRed   tempoDay = "rouge"
	tempoWhite tempoDay = "blanc"
	tempoBlue  tempoDay = "bleu"
)

var (
	redDays   []time.Time
	whiteDays []time.Time
)

// https://www.services-rte.com/fr/visualisez-les-donnees-publiees-par-rte/calendrier-des-offres-de-fourniture-de-type-tempo.html
// https://particulier.edf.fr/fr/accueil/gestion-contrat/options/tempo.html#/selection-bp
// maybe later: https://data.rte-france.com/catalog/-/api/doc/user-guide/Tempo+Like+Supply+Contract/1.1
func generateTempoDays() (err error) {
	// setup red days
	var day time.Time
	redDates := []string{
		"02/03/2023",
		"01/03/2023",
		"10/02/2023",
		"09/02/2023",
		"08/02/2023",
		"07/02/2023",
		"06/02/2023",
		"31/01/2023",
		"30/01/2023",
		"27/01/2023",
		"26/01/2023",
		"25/01/2023",
		"24/01/2023",
		"23/01/2023",
		"20/01/2023",
		"19/01/2023",
		"18/01/2023",
		"17/01/2023",
		"14/12/2022",
		"13/12/2022",
		"12/12/2022",
		"08/12/2022",
		"27/01/2022",
		"26/01/2022",
		"25/01/2022",
		"24/01/2022",
		"21/01/2022",
		"20/01/2022",
		"19/01/2022",
		"18/01/2022",
		"17/01/2022",
		"14/01/2022",
		"13/01/2022",
		"12/01/2022",
		"11/01/2022",
		"10/01/2022",
		"06/01/2022",
		"22/12/2021",
		"21/12/2021",
		"20/12/2021",
		"15/12/2021",
		"14/12/2021",
		"13/12/2021",
		"29/11/2021",
	}
	for index, redDate := range redDates {
		day, err = time.ParseInLocation(pricesDateFormat, redDate, frLocation)
		if err != nil {
			return fmt.Errorf("failed to parse tempo red day at index %d: %s", index, err)
		}
		redDays = append(redDays, day)
	}
	// setup white days
	whiteDates := []string{
		"12/06/2023",
		"09/06/2023",
		"08/06/2023",
		"07/06/2023",
		"26/04/2023",
		"25/04/2023",
		"13/04/2023",
		"07/04/2023",
		"06/04/2023",
		"05/04/2023",
		"04/04/2023",
		"07/03/2023",
		"06/03/2023",
		"03/03/2023",
		"28/02/2023",
		"27/02/2023",
		"24/02/2023",
		"23/02/2023",
		"22/02/2023",
		"16/02/2023",
		"15/02/2023",
		"14/02/2023",
		"13/02/2023",
		"11/02/2023",
		"04/02/2023",
		"03/02/2023",
		"02/02/2023",
		"01/02/2023",
		"28/01/2023",
		"21/01/2023",
		"16/01/2023",
		"17/12/2022",
		"16/12/2022",
		"15/12/2022",
		"10/12/2022",
		"09/12/2022",
		"07/12/2022",
		"06/12/2022",
		"05/12/2022",
		"02/12/2022",
		"01/12/2022",
		"30/11/2022",
		"31/05/2022",
		"30/05/2022",
		"24/05/2022",
		"14/04/2022",
		"06/04/2022",
		"05/04/2022",
		"04/04/2022",
		"08/03/2022",
		"07/03/2022",
		"03/03/2022",
		"02/03/2022",
		"01/03/2022",
		"28/02/2022",
		"25/02/2022",
		"23/02/2022",
		"11/02/2022",
		"10/02/2022",
		"09/02/2022",
		"08/02/2022",
		"07/02/2022",
		"03/02/2022",
		"02/02/2022",
		"31/01/2022",
		"28/01/2022",
		"22/01/2022",
		"15/01/2022",
		"07/01/2022",
		"05/01/2022",
		"17/12/2021",
		"16/12/2021",
		"11/12/2021",
		"09/12/2021",
		"08/12/2021",
		"07/12/2021",
		"06/12/2021",
		"03/12/2021",
		"02/12/2021",
		"01/12/2021",
		"30/11/2021",
		"26/11/2021",
		"25/11/2021",
		"24/11/2021",
		"23/11/2021",
	}
	for index, whiteDate := range whiteDates {
		day, err = time.ParseInLocation(pricesDateFormat, whiteDate, frLocation)
		if err != nil {
			return fmt.Errorf("failed to parse tempo white day at index %d: %s", index, err)
		}
		whiteDays = append(whiteDays, day)
	}
	return
}

func getTempoDayColor(datetime time.Time) tempoDay {
	// adjust date for early HC
	if datetime.Hour() < 6 {
		// we need to take the color of the previous day
		datetime = datetime.Add(-1 * 7 * time.Hour)
	}
	// search red days
	for _, redDay := range redDays {
		if redDay.Year() == datetime.Year() && redDay.Month() == datetime.Month() && redDay.Day() == datetime.Day() {
			return tempoRed
		}
	}
	// search white days
	for _, whiteDay := range whiteDays {
		if whiteDay.Year() == datetime.Year() && whiteDay.Month() == datetime.Month() && whiteDay.Day() == datetime.Day() {
			return tempoWhite
		}
	}
	return tempoBlue
}
