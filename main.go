package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	frLocation *time.Location
	//  overrided during compilation
	Version = "development"
)

func main() {
	// Parse flags
	filePathFlag := flag.String("csv", "", "path to the Enedis hourly export file")
	detailsFlag := flag.Bool("monthly", false, "show the monthly details")
	versionFlag := flag.Bool("version", false, "show the version details")
	flag.Parse()
	// Prepare
	var err error
	if frLocation, err = time.LoadLocation("Europe/Paris"); err != nil {
		log.Fatalln(err)
	}
	if err = prepareDates(); err != nil {
		log.Fatalln(err)
	}
	// Act on flags
	if *versionFlag {
		fmt.Printf("Version:\t\t%s\n", Version)
		fmt.Printf("Last known data:\t%d %s %d\n", lastUpdate.Day(), lastUpdate.Month(), lastUpdate.Year())
		return
	}
	if *filePathFlag == "" {
		log.Fatalln("please set the -csv flag")
	}
	// Parse the file as enedis csv
	header, data, err := parseFile(*filePathFlag)
	if err != nil {
		log.Fatal(err)
	}
	// Issue warning if our data is outdated
	if header.End.After(lastUpdate) {
		fmt.Println()
		fmt.Printf("/!\\ the data set contains values that are beyong the internal data this program has. Please update the code.\n")
		fmt.Println()
	}
	// Compute
	compute(header, data, *detailsFlag)
}

func compute(header CSVHeader, data []point, montly bool) {
	var (
		pointAdjustedConso                       float64
		totalConso, monthConso                   float64
		totalBase, monthBase, pointBase          float64
		totalHC, monthHC, pointHC                float64
		totalTempo, monthTempo, pointTempo       float64
		totalZenFlex, monthZenFlex, pointZenFlex float64
		refMonth                                 time.Time
	)
	fmt.Printf("PRM:\t\t%s\n", header.PRMID)
	fmt.Printf("Start:\t\t%v\n", header.Start)
	fmt.Printf("End:\t\t%v\n", header.End)
	fmt.Printf("Stepping:\t%v\n", header.Step)
	for index, point := range data {
		// Adjust time for start and not end
		adjustedTime := point.Time.Add(header.Step * -1)
		if index == 0 {
			refMonth = adjustedTime
		}
		// Adjust average watt consumption to kWh
		if header.Step == steppingHour {
			pointAdjustedConso = point.Value
		} else if header.Step == steppingHalfHour {
			pointAdjustedConso = point.Value / 2
		} else {
			fmt.Printf("unexpected stepping: %s", header.Step)
			os.Exit(1)
		}
		// Compute price for current point
		pointBase = pointAdjustedConso * getBasePrice(adjustedTime)
		pointHC = pointAdjustedConso * getHCPrice(adjustedTime)
		pointTempo = pointAdjustedConso * getTempoPrice(adjustedTime)
		pointZenFlex = pointAdjustedConso * getZenFlexPrice(adjustedTime)
		// Handle months
		if refMonth.Year() == adjustedTime.Year() && refMonth.Month() == adjustedTime.Month() {
			monthBase += pointBase
			monthHC += pointHC
			monthTempo += pointTempo
			monthZenFlex += pointZenFlex
			monthConso += pointAdjustedConso
		} else {
			if montly {
				// Print total for previous month
				fmt.Println()
				fmt.Printf("* %s %d\n", refMonth.Month(), refMonth.Year())
				fmt.Printf("Consommation:\t~%0.2f kWh\n", monthConso)
				fmt.Printf("Option base:\t%0.2f€\n", monthBase)
				fmt.Printf("Option HC:\t%0.2f€\n", monthHC)
				fmt.Printf("Option Tempo:\t%0.2f€\n", monthTempo)
				fmt.Printf("Zen Flex:\t%0.2f€\n", monthZenFlex)
			}
			// New month, reset counters
			monthBase = pointBase
			monthHC = pointHC
			monthTempo = pointTempo
			monthZenFlex = pointZenFlex
			monthConso = pointAdjustedConso
			refMonth = adjustedTime
		}
		// Add to total
		totalConso += pointAdjustedConso
		totalBase += pointBase
		totalHC += pointHC
		totalTempo += pointTempo
		totalZenFlex += pointZenFlex
	}
	// Print total for old month
	if montly {
		fmt.Println()
		fmt.Printf("* %s %d\n", refMonth.Month(), refMonth.Year())
		fmt.Printf("Consommation:\t~%0.2f kWh\n", monthConso)
		fmt.Printf("Option base:\t%0.2f€\n", monthBase)
		fmt.Printf("Option HC:\t%0.2f€\n", monthHC)
		fmt.Printf("Option Tempo:\t%0.2f€\n", monthTempo)
		fmt.Printf("Zen Flex:\t%0.2f€\n", monthZenFlex)
	}
	// Print total
	fmt.Println()
	fmt.Println("* TOTAL")
	fmt.Printf("Consommation:\t~%0.2f kWh\n", totalConso)
	fmt.Printf("Option base:\t%0.2f€\n", totalBase)
	fmt.Printf("Option HC:\t%0.2f€\n", totalHC)
	fmt.Printf("Option Tempo:\t%0.2f€\n", totalTempo)
	fmt.Printf("Zen Flex:\t%0.2f€\n", totalZenFlex)
}
