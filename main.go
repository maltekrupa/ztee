package main

import (
	"./zutil"
	"flag"
	"log"
	"os"
	"net"
)

// Input Format Type (CSV or JSON)
type InFormat int

const (
	CSV InFormat = iota
	JSON
)

// Container for the input args
type RawArgs struct {
	OutputFileName string
	InputFormat    string
	Field string
	Column int
	SuccessOnly    bool
}

// Container for options with useful types
type Config struct  {
	OutputFile *os.File
	Format InFormat
	CsvIndex int
	JsonField string
	SuccessOnly bool
}

var args RawArgs
var config Config

func init() {
	// Bind the input args to variables
	flag.StringVar(&args.OutputFileName, "output-file", "", "Output file name for original data")
	flag.StringVar(&args.InputFormat, "format", "csv", "Input format (csv or json)")
	flag.IntVar(&args.Column, "column", 1, "Column number in csv")
	flag.StringVar(&args.Field, "field", "saddr", "Field name in json")
	flag.BoolVar(&args.SuccessOnly, "success-only", false, "Only pass successful ZMap results to stdout")
	flag.Parse()

	// Parse the input args and convert to usable types

	// Validate outputfile
	if args.OutputFileName == "" {
		log.Fatal("Required argument --output-file")
	} else {
		if file, err := os.Create(args.OutputFileName); err == nil {
			config.OutputFile = file
		} else {
			log.Fatal("Unable to open file: " + args.OutputFileName)
		}
	}

	// Validate input format
	switch args.InputFormat {
	case "csv":
		config.Format = CSV
	case "json":
		config.Format = JSON
	default:
		log.Fatal("--format must be one of [csv, json], given: " + args.InputFormat)
	}

	// Link up field names
	if args.Column <= 0 {
		log.Fatal("Invalid csv column")
	}
	config.CsvIndex = args.Column - 1
	config.JsonField = args.Field
	config.SuccessOnly = args.SuccessOnly
}

func GetExtractor() Extractor {
	switch config.Format {
	case CSV:
		return MakeCsvExtractor(config.CsvIndex)
	case JSON:
		return MakeJsonExtractor(config.JsonField)
	}
	// Should not be reached
	return nil
}

func main() {
	toChan, fromChan := zutil.NewNonblockingSendPair()
	ipChan := make(chan net.IP)
	rewriteChan := make(chan string)
	doneChan := make(chan int)

	extractor := GetExtractor()
	splitter := MakeLineSplitter(extractor)
	go WriteIP(ipChan, doneChan)
	go Rewrite(rewriteChan, config.OutputFile, doneChan)
	go splitter(fromChan, ipChan, rewriteChan)
	go GobbleInput(os.Stdin, toChan)
	<-doneChan
	<-doneChan
	return
}
