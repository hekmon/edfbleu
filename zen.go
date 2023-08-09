package main

import (
	"time"
)

func getZenFlexPrice(datetime time.Time) float64 {
	// are we within hc range ?
	var hc bool
	if datetime.Hour() < 8 || datetime.Hour() >= 20 {
		hc = true
	} else {
		if datetime.Hour() >= 13 {
			if datetime.Hour() < 18 {
			      hc = true
		        }
		}
	}
	// return price
	if datetime.After(zenPrices2023) {
		switch getZenFlexDayColor(datetime) {
		case sobriete:
			if hc {
				return 0.2228
			}
			return 0.6712
		case eco:
			if hc {
				return 0.1295
			}
			return 0.2228
		}
	}
	if datetime.After(zenPrices092023) {
		switch getZenFlexDayColor(datetime) {
		case sobriete:
			if hc {
				return 0.2460
			}
			return 0.7324
		case eco:
			if hc {
				return 0.1464
			}
			return 0.2460
		}
	}
	return 0
}
