package main

import (
	"io"
	"time"
)
import "fmt"

type PrinterHolder struct {
	Printers []Printer
}

func (p PrinterHolder) doPrint(f func(p Printer)) {
	for _, printer := range p.Printers {
		f(printer)
	}
}

type Printer interface {
	PrintHeader()
	PrintPreLine(conLevel uint)
	PrintLine(conLevel uint, rps float64, avgDur time.Duration, succRate float64, cuEstimate uint64, dist map[string]int)
	PrintEvent(event string)
	PrintFooter(maxCU uint64)
}

type CsvPrinter struct {
	Out io.Writer
}

func (p CsvPrinter) PrintHeader() {
	_, err := fmt.Fprintf(p.Out, "concurrency, rps, avg, succRate, cuEstimate\n")
	if err != nil {
		panic(err)
	}
}

func (p CsvPrinter) PrintPreLine(conLevel uint) {}

func (p CsvPrinter) PrintLine(conLevel uint, rps float64, avgDur time.Duration, succRate float64, cuEstimate uint64, dist map[string]int) {
	_, err := fmt.Fprintf(p.Out, "%d, %f, %s, %f, %d\n", conLevel, rps, avgDur, succRate, cuEstimate)
	if err != nil {
		panic(err)
	}
}

func (p CsvPrinter) PrintEvent(event string) {}

func (p CsvPrinter) PrintFooter(maxCU uint64) {}

type TextPrinter struct {
	Out io.Writer
}

func (p TextPrinter) PrintHeader() {}

func (p TextPrinter) PrintPreLine(conLevel uint) {
	fmt.Fprintf(p.Out, "-------------------------------------\n")
	fmt.Fprintf(p.Out, "LOAD LEVEL: %d\n", conLevel)
	fmt.Fprintf(p.Out, "-------------------------------------\n")
}

func (p TextPrinter) PrintLine(conLevel uint, rps float64, avgDur time.Duration, succRate float64, cuEstimate uint64, dist map[string]int) {
	fmt.Fprintf(p.Out, "Concurrency level: %d\n", conLevel)
	fmt.Fprintf(p.Out, "RPS: %f\n", rps)
	fmt.Fprintf(p.Out, "Avg: %s\n", avgDur)
	fmt.Fprintf(p.Out, "SuccRate: %f\n", succRate)
	fmt.Fprintf(p.Out, "CU Estimate: %d\n", cuEstimate)
	fmt.Fprintf(p.Out, "Error Distribution: %v\n", dist)
}

func (p TextPrinter) PrintEvent(event string) {
	fmt.Fprintf(p.Out, "\n")
	fmt.Fprintf(p.Out, "=====================================\n")
	fmt.Fprintf(p.Out, event+"\n")
}

func (p TextPrinter) PrintFooter(maxCU uint64) {
	fmt.Fprintf(p.Out, "=====================================\n")
	fmt.Fprintf(p.Out, "MAX CU ESTIMATION: %d\n", maxCU)
}
