package main

import (
	"errors"
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
		fmt.Printf("Version:\t\t%s\n", Version())
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
	// Issue warning if our data is outdated (tempo pricing is refused on daily-energy exports,
	// there is no point warning about the tempo calendar for them)
	if !header.Daily {
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
	}
	// Compute
	if err = compute(header, data, *detailsFlag); err != nil {
		log.Fatal(err)
	}
}

// tariffOption is a pricing option: a price function of the date time of the start of a
// period, and whether it needs the intraday consumption profile to be priced
type tariffOption struct {
	name          string
	price         func(time.Time) float64
	needsIntraday bool
}

var (
	optionBase  = tariffOption{name: "base", price: getBasePrice}
	optionHC    = tariffOption{name: "HC/HP", price: getHCPrice, needsIntraday: true}
	optionTempo = tariffOption{name: "Tempo", price: getTempoPrice, needsIntraday: true}
)

// errDailyIntradayRequired is returned when a pricing option needs the intraday consumption
// profile of a data set which only carries daily aggregates
var errDailyIntradayRequired = errors.New("daily-energy exports cannot be priced under HC/HP or Tempo: per-period consumption is not available, use the half-hourly load-curve export")

// monthlyPrice is the pricing of a calendar month under a pricing option
type monthlyPrice struct {
	Month time.Time // a date time within the month
	Conso float64   // kWh consumed over the month
	Price float64   // price of the month
}

// priceRecords prices the records of a data set under a pricing option and details the result
// by month. Each record is priced with the tariff of the start of the period it covers.
// Daily-energy exports carry no intraday consumption profile: the base option is exact for them
// but the HC/HP and Tempo options are refused rather than approximated.
func priceRecords(header FileHeader, data []point, option tariffOption) (total float64, months []monthlyPrice, err error) {
	if header.Daily && option.needsIntraday {
		err = errDailyIntradayRequired
		return
	}
	for _, record := range data {
		// Adjust time for start and not end
		adjustedTime := record.Time.Add(-record.Duration)
		price := record.Conso * option.price(adjustedTime)
		total += price
		// Handle months
		if len(months) == 0 || months[len(months)-1].Month.Month() != adjustedTime.Month() || months[len(months)-1].Month.Year() != adjustedTime.Year() {
			months = append(months, monthlyPrice{Month: adjustedTime})
		}
		months[len(months)-1].Conso += record.Conso
		months[len(months)-1].Price += price
	}
	return
}

func compute(header FileHeader, data []point, montly bool) error {
	// Price the data set under every option
	baseTotal, baseMonths, err := priceRecords(header, data, optionBase)
	if err != nil {
		return fmt.Errorf("failed to price the base option: %w", err)
	}
	var hcTotal, tempoTotal float64
	var hcMonths, tempoMonths []monthlyPrice
	var optionsErr error
	if hcTotal, hcMonths, err = priceRecords(header, data, optionHC); err != nil {
		optionsErr = err
	} else if tempoTotal, tempoMonths, err = priceRecords(header, data, optionTempo); err != nil {
		optionsErr = err
	}
	// Show the results
	fmt.Printf("PRM:\t\t%s\n", header.PRMID)
	fmt.Printf("Start:\t\t%v\n", header.Start)
	fmt.Printf("End:\t\t%v\n", header.End)
	if header.Step == noSteppingValue {
		fmt.Printf("Stepping:\tnot fixed (automatically detected for each point)\n")
	} else {
		fmt.Printf("Stepping:\t%v\n", header.Step)
	}
	if montly {
		for index, month := range baseMonths {
			fmt.Println()
			fmt.Printf("* %s %d\n", month.Month.Month(), month.Month.Year())
			fmt.Printf("Consommation:\t~%0.2f kWh\n", month.Conso)
			fmt.Printf("Option base:\t%0.2f€\n", month.Price)
			if optionsErr == nil {
				fmt.Printf("Option HC:\t%0.2f€\n", hcMonths[index].Price)
				fmt.Printf("Option Tempo:\t%0.2f€\n", tempoMonths[index].Price)
			}
		}
	}
	// Print total
	var totalConso float64
	for _, month := range baseMonths {
		totalConso += month.Conso
	}
	fmt.Println()
	fmt.Println("* TOTAL")
	fmt.Printf("Consommation:\t~%0.2f kWh\n", totalConso)
	fmt.Printf("Option base:\t%0.2f€\n", baseTotal)
	if optionsErr == nil {
		fmt.Printf("Option HC:\t%0.2f€\n", hcTotal)
		fmt.Printf("Option Tempo:\t%0.2f€\n", tempoTotal)
	}
	// HC/HP and Tempo could not be priced: refuse them instead of approximating them
	return optionsErr
}
