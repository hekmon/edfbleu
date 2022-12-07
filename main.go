package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {
	// Take the file path
	filePathFlag := flag.String("csv", "", "path to the Enedis hourly export file")
	detailsFlag := flag.Bool("monthly", false, "show the montly details")
	flag.Parse()
	if *filePathFlag == "" {
		log.Fatal("please set the -csv flag")
	}
	// Parse the file as enedis csv
	header, data, err := parseFile(*filePathFlag)
	if err != nil {
		log.Fatal(err)
	}
	// Issue warning if our data is outdated
	if header.End.After(lastUpdate) {
		fmt.Println()
		fmt.Printf("/!\\ the data set contains values that are beyong the internal data this program has. Please update the code.")
		fmt.Println()
	}
	// Compute
	var (
		totalBase, monthBase, pointBase    float64
		totalHC, monthHC, pointHC          float64
		totalTempo, monthTempo, pointTempo float64
		refMonth                           time.Time
	)
	fmt.Printf("PRM:\t\t%s\n", header.ID)
	fmt.Printf("Start:\t\t%v\n", header.Start)
	fmt.Printf("End:\t\t%v\n", header.End)
	for index, point := range data {
		// Adjust time for start and not end
		adjustedTime := point.Time.Add(30 * time.Minute * -1)
		if index == 0 {
			refMonth = adjustedTime
		}
		// Compute price for current point
		pointBase = point.Value * getBasePrice(adjustedTime)
		pointHC = point.Value * getHCPrice(adjustedTime)
		pointTempo = point.Value * getTempoPrice(adjustedTime)
		// Handle months
		if refMonth.Year() == adjustedTime.Year() && refMonth.Month() == adjustedTime.Month() {
			monthBase += pointBase
			monthHC += pointHC
			monthTempo += pointTempo
		} else {
			if !refMonth.IsZero() && *detailsFlag {
				// Print total for previous month
				fmt.Println()
				fmt.Printf("* %s %d\n", refMonth.Month(), refMonth.Year())
				fmt.Printf("Option base:\t%0.2f€\n", monthBase)
				fmt.Printf("Option HC:\t%0.2f€\n", monthHC)
				fmt.Printf("Option Tempo:\t%0.2f€\n", monthTempo)
			}
			// New month, reset counters
			monthBase = pointBase
			monthHC = pointHC
			monthTempo = pointTempo
			refMonth = adjustedTime
		}
		// Add to total
		totalBase += pointBase
		totalHC += pointHC
		totalTempo += pointTempo
	}
	// Print total for old month
	if *detailsFlag {
		fmt.Println()
		fmt.Printf("* %s %d\n", refMonth.Month(), refMonth.Year())
		fmt.Printf("Option base:\t%0.2f€\n", monthBase)
		fmt.Printf("Option HC:\t%0.2f€\n", monthHC)
		fmt.Printf("Option Tempo:\t%0.2f€\n", monthTempo)
	}
	// Print total
	fmt.Println()
	fmt.Println("* TOTAL")
	fmt.Printf("Option base:\t%0.2f€\n", totalBase)
	fmt.Printf("Option HC:\t%0.2f€\n", totalHC)
	fmt.Printf("Option Tempo:\t%0.2f€\n", totalTempo)
}
