package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

var (
	frLocation *time.Location
)

func main() {
	// Parse flags
	filePathFlag := flag.String("csv", "", "path to an Enedis export file (legacy flag: input files can also be passed as arguments)")
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
		fmt.Printf("Version:		%s\n", Version())
		fmt.Printf("Last known data:\t%d %s %d\n", lastUpdate.Day(), lastUpdate.Month(), lastUpdate.Year())
		return
	}
	if *filePathFlag == "" && flag.NArg() == 0 {
		log.Fatalln("please provide at least one Enedis export file (CSV or XLSX), either with the -csv flag or as argument(s)")
	}
	// Parse the input files as enedis exports (the legacy CSV and the XLSX of the web portal)
	paths := flag.Args()
	if *filePathFlag != "" {
		paths = append([]string{*filePathFlag}, paths...)
	}
	header, data, err := parseFiles(paths)
	if err != nil {
		log.Fatal(err)
	}
	// Warn about daily data approximation
	if header.Daily {
		fmt.Println()
		fmt.Println("/!\\ daily export: the intraday consumption profile is unknown so each day has been spread uniformly over its half-hours. The base option is exact but the HC/HP and Tempo amounts are approximations.")
		fmt.Println()
	}
	// Issue warning if our data is outdated
	if usingAPIData {
		if header.End.After(apiLastDecided) {
			fmt.Println()
			fmt.Printf("/!\\ the data set contains values that are beyond the last tempo day decided so far (%s): the color of the next days does not exist yet so they can not be priced.\n", apiLastDecided.Format(headerDateFormat))
			fmt.Println()
		}
	} else if header.End.After(lastUpdate) {
		fmt.Println()
		fmt.Printf("/!\\ the data set contains values that are beyond the internal data this program has. Please update the code.\n")
		fmt.Println()
	}
	// Compute
	compute(header, data, *detailsFlag)
}

func compute(header FileHeader, data []point, montly bool) {
	var (
		totalConso, monthConso             float64
		totalBase, monthBase, pointBase    float64
		totalHC, monthHC, pointHC          float64
		totalTempo, monthTempo, pointTempo float64
		refMonth                           time.Time
	)
	fmt.Printf("PRM:\t\t%s\n", header.PRMID)
	fmt.Printf("Start:\t\t%v\n", header.Start)
	fmt.Printf("End:\t\t%v\n", header.End)
	if header.Step == noSteppingValue {
		fmt.Printf("Stepping:\tnot fixed (automatically detected for each point)\n")
	} else {
		fmt.Printf("Stepping:\t%v\n", header.Step)
	}
	for index, point := range data {
		// Adjust time for start and not end
		adjustedTime := point.Time.Add(-point.Duration)
		if index == 0 {
			refMonth = adjustedTime
		}
		// Compute price for current point
		pointBase = point.Conso * getBasePrice(adjustedTime)
		pointHC = point.Conso * getHCPrice(adjustedTime)
		pointTempo = point.Conso * getTempoPrice(adjustedTime)
		// Handle months
		if refMonth.Year() == adjustedTime.Year() && refMonth.Month() == adjustedTime.Month() {
			monthBase += pointBase
			monthHC += pointHC
			monthTempo += pointTempo
			monthConso += point.Conso
		} else {
			if montly {
				// Print total for previous month
				fmt.Println()
				fmt.Printf("* %s %d\n", refMonth.Month(), refMonth.Year())
				fmt.Printf("Consommation:\t~%0.2f kWh\n", monthConso)
				fmt.Printf("Option base:\t%0.2f€\n", monthBase)
				fmt.Printf("Option HC:\t%0.2f€\n", monthHC)
				fmt.Printf("Option Tempo:\t%0.2f€\n", monthTempo)
			}
			// New month, reset counters
			monthBase = pointBase
			monthHC = pointHC
			monthTempo = pointTempo
			monthConso = point.Conso
			refMonth = adjustedTime
		}
		// Add to total
		totalConso += point.Conso
		totalBase += pointBase
		totalHC += pointHC
		totalTempo += pointTempo
	}
	// Print total for old month
	if montly {
		fmt.Println()
		fmt.Printf("* %s %d\n", refMonth.Month(), refMonth.Year())
		fmt.Printf("Consommation:\t~%0.2f kWh\n", monthConso)
		fmt.Printf("Option base:\t%0.2f€\n", monthBase)
		fmt.Printf("Option HC:\t%0.2f€\n", monthHC)
		fmt.Printf("Option Tempo:\t%0.2f€\n", monthTempo)
	}
	// Print total
	fmt.Println()
	fmt.Println("* TOTAL")
	fmt.Printf("Consommation:\t~%0.2f kWh\n", totalConso)
	fmt.Printf("Option base:\t%0.2f€\n", totalBase)
	fmt.Printf("Option HC:\t%0.2f€\n", totalHC)
	fmt.Printf("Option Tempo:\t%0.2f€\n", totalTempo)
}
