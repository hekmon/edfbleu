package main

import (
	"fmt"
	"log"
)

func main() {
	// Get the CSV reader
	header, data, err := parseFile(reffile)
	if err != nil {
		log.Fatal(err)
	}
	// Compute
	var (
		base  float64
		hc    float64
		tempo float64
	)
	for _, point := range data {
		base += point.Value * getBasePrice(point.Time)
		hc += point.Value * getHCPrice(point.Time)
		tempo += point.Value * getTempoPrice(point.Time)
	}
	// Show Results
	fmt.Printf("PRM:\t\t%s\n", header.ID)
	fmt.Printf("Start:\t\t%v\n", header.Start)
	fmt.Printf("End:\t\t%v\n", header.End)
	fmt.Println()
	fmt.Printf("Option base:\t%0.2f€\n", base)
	fmt.Printf("Option HC:\t%0.2f€\n", hc)
	fmt.Printf("Option Tempo:\t%0.2f€\n", tempo)
}
