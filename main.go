package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/INFURA/go-ethlibs/node"
	ethspam "github.com/p2p-org/ethspam/lib"
	"github.com/bojand/ghz/runner"
	"github.com/golang/protobuf/proto"
	"github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
	"io"
	"math/rand"
	"os"
	"github.com/p2p-org/drpc-provider-estimator/dshackle"
	"github.com/p2p-org/drpc-provider-estimator/gas"
	"time"
)

type Options struct {
	Host             string `long:"target" short:"t" description:"target host" required:"true"`
	Chain            int    `long:"chain" short:"c" description:"chain id" default:"100"`
	StepDuration     uint64 `long:"step-duration" short:"d" description:"step duration in minutes" default:"3"`
	SourceHost       string `long:"source" short:"s" description:"source eth host" default:"https://eth.drpc.org"`
	StopOnEvents     bool   `long:"stop-on-events" short:"e" description:"stop on events"`
	CsvOutput        string `long:"csv-output" short:"o" description:"csv output file"`
	Mode             string `long:"mode" short:"m" description:"mode. Can be spam or prepared" default:"spam"`
	SpamProfile      string `long:"spam-profile" short:"p" description:"spam profile"`
	PreparedRequests string `long:"prepared-requests" short:"r" description:"prepared requests folder"`
	PreparedCU       uint64 `long:"prepared-cu" short:"u" description:"prepared request cu cost" default:"0"`
	RequestLabel     string `long:"request-label" short:"l" description:"request label for dshackle"`
	LoadLevels       string `long:"load-levels" short:"a" description:"load levels"`
	Insecure         bool   `long:"insecure" short:"i" description:"certificate"`
}

var DEFAULT_PROFILE = map[string]int64{
	"eth_getCode":               100,
	"eth_getLogs":               250,
	"eth_getTransactionByHash":  250,
	"eth_blockNumber":           350,
	"eth_getTransactionCount":   400,
	"eth_getBlockByNumber":      400,
	"eth_getBalance":            550,
	"eth_getTransactionReceipt": 600,
	"eth_call":                  2000,
}

func main() {
	options := Options{}
	_, err := flags.Parse(&options)
	if err != nil {
		return
	}

	var load []uint

	if options.LoadLevels != "" {
		err := json.Unmarshal([]byte("["+options.LoadLevels+"]"), &load)
		if err != nil {
			exit(1, "error during fetching loading load level: %v", err)
		}
	} else {
		load = []uint{10, 50, 100, 500, 1000, 5000, 10000}
	}

	prevRps := 0.0
	prevMean := time.Hour * 1
	var maxCu uint64 = 0

	var printers PrinterHolder
	printers.Printers = append(printers.Printers, &TextPrinter{
		Out: os.Stdout,
	})
	if options.CsvOutput != "" {
		f, err := os.Create(options.CsvOutput)
		if err != nil {
			exit(1, "error during creating output: %v", err)
		}
		defer f.Close()
		printers.Printers = append(printers.Printers, &CsvPrinter{
			Out: f,
		})
	}

	printers.doPrint(func(printer Printer) {
		printer.PrintHeader()
	})

	var dataFuncProvider func(cuCount *uint64) func(mtd *desc.MethodDescriptor, callData *runner.CallData) []byte

	fmt.Println(options.Mode)
	if options.Mode == "spam" {
		var profile map[string]int64
		if options.SpamProfile == "" {
			profile = DEFAULT_PROFILE
		} else {
			profile = map[string]int64{}
			profileRaw, _ := os.ReadFile(options.SpamProfile)
			err := yaml.Unmarshal(profileRaw, profile)
			if err != nil {
				exit(1, "error during fetching spam profile: %v", err)
			}
		}
		fmt.Println(profile)
		dataFuncProvider = func(cuCount *uint64) func(mtd *desc.MethodDescriptor, callData *runner.CallData) []byte {
			return NewEthSpamBinaryDataFunc(profile, options.SourceHost, dshackle.ChainRef(options.Chain), cuCount, context.Background())
		}
	} else if options.Mode == "prepared" {
		files, err := os.ReadDir(options.PreparedRequests)
		if err != nil {
			exit(1, "error during fetching requests: %v", err)
		}
		reqs := make([]*dshackle.NativeCallRequest, 0)
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			name := options.PreparedRequests + "/" + file.Name()
			raw, err := os.ReadFile(name)
			if err != nil {
				panic(err)
			}
			jsonrpc := JsonRpcRequest{}
			err = json.Unmarshal(raw, &jsonrpc)
			fmt.Printf("loading  %s\n", name)
			if err != nil {
				panic(err)
			}
			params, _ := json.Marshal(jsonrpc.Params)
			req := &dshackle.NativeCallRequest{
				Chain: dshackle.ChainRef(options.Chain),
				Items: []*dshackle.NativeCallItem{
					{
						Id:      uint32(jsonrpc.Id),
						Method:  jsonrpc.Method,
						Payload: params,
					},
				},
			}

			if options.RequestLabel != "" {
				selector := dshackle.Selector{}
				_ = protojson.Unmarshal([]byte(options.RequestLabel), &selector)
				req.Selector = &selector
			}

			fmt.Printf("request added to load: %s from %s\n", jsonrpc.Method, name)
			reqs = append(reqs, req)
		}
		dataFuncProvider = func(cuCount *uint64) func(mtd *desc.MethodDescriptor, callData *runner.CallData) []byte {
			pointer := 0
			return func(mtd *desc.MethodDescriptor, callData *runner.CallData) []byte {
				req := reqs[pointer]
				pointer = (pointer + 1) % len(reqs)
				data, _ := proto.Marshal(req)
				if options.PreparedCU > 0 {
					*cuCount += options.PreparedCU
				} else {
					*cuCount += gas.CountGas(req.Items[0].Method)
				}
				return data
			}
		}
	} else {
		exit(1, "unknown mode")
	}

	for _, l := range load {

		printers.doPrint(func(printer Printer) {
			printer.PrintPreLine(l)
		})

		var curCu uint64 = 0

		ghzOpts := make([]runner.Option, 0)

		ghzOpts = append(ghzOpts,
			runner.WithBinaryDataFunc(dataFuncProvider(&curCu)),
			runner.WithEnableCompression(true),
			runner.WithConnections(10),
			runner.WithTotalRequests(500),
			runner.WithConcurrency(l),
			runner.WithInsecure(options.Insecure),
			runner.WithRunDuration(time.Duration(options.StepDuration)*time.Minute),
		)

		report, err := runner.Run(
			"emerald.Blockchain.NativeCall",
			options.Host,
			ghzOpts...,
		)

		if err != nil {
			exit(1, "error during test execution: %v", err)
		}

		succRate := calcErrorRate(report.StatusCodeDist)

		printers.doPrint(func(printer Printer) {
			printer.PrintLine(l, report.Rps, report.Average, succRate, curCu/options.StepDuration, report.ErrorDist)
		})

		if curCu > maxCu {
			maxCu = curCu
		}

		if prevRps > report.Rps {
			printers.doPrint(func(printer Printer) {
				printer.PrintEvent("RPS DECREASED")
			})

			if options.StopOnEvents {
				break
			}
		}

		if prevMean*100 < report.Average {
			printers.doPrint(func(printer Printer) {
				printer.PrintEvent("LATENCY INCREASED")
			})

			if options.StopOnEvents {
				break
			}
		}

		if succRate < 0.85 {
			printers.doPrint(func(printer Printer) {
				printer.PrintEvent("ERROR RATE INCREASED")
			})

			if options.StopOnEvents {
				break
			}
		}

		prevRps = report.Rps
		prevMean = report.Average
	}

	maxCu = maxCu / options.StepDuration

	printers.doPrint(func(printer Printer) {
		printer.PrintFooter(maxCu)
	})
}

func calcErrorRate(dist map[string]int) float64 {
	total := 0.0
	nok := 0.0
	for k, v := range dist {
		if k == "Canceled" {
			continue
		}
		total += float64(v)
		if k != "OK" {
			nok += float64(v)
		}
	}
	return 1 - (nok / total)
}

func NewEthSpamBinaryDataFunc(queryParams map[string]int64, parentHost string, chain dshackle.ChainRef, cuCount *uint64, ctx context.Context) func(mtd *desc.MethodDescriptor, callData *runner.CallData) []byte {
	generator, err := ethspam.MakeQueriesGenerator(queryParams)
	if err != nil {
		exit(1, "failed to install defaults: %s", err)
	}

	client, err := node.NewClient(ctx, parentHost)
	if err != nil {
		exit(1, "failed to make a new client: %s", err)
	}
	mkState := ethspam.StateProducer{
		Client: client,
	}

	stateChannel := make(chan ethspam.State, 1)

	randSrc := rand.NewSource(time.Now().UnixNano())
	go func() {
		state := ethspam.LiveState{
			IdGen:   &ethspam.IdGenerator{},
			RandSrc: randSrc,
		}
		for {
			newState, err := mkState.Refresh(&state)
			if err != nil {
				// It can happen in some testnets that most of the blocks
				// are empty(no transaction included), don't refresh the
				// QueriesGenerator state without new inclusion.
				if err == ethspam.ErrEmptyBlock {
					select {
					case <-time.After(5 * time.Second):
					case <-ctx.Done():
						return
					}
					continue
				}
				fmt.Printf("failed to refresh state: %s", err)
				<-time.After(1 * time.Second)
				continue
			}
			select {
			case stateChannel <- newState:
			case <-ctx.Done():
				return
			}

			select {
			case <-time.After(15 * time.Second):
			case <-ctx.Done():
			}
		}
	}()

	state := <-stateChannel

	queries := make(chan ethspam.QueryContent, 1000)

	go func() {
		for {
			// Update state when a new one is emitted
			select {
			case state = <-stateChannel:
			case <-ctx.Done():
				return
			default:
			}
			if q, err := generator.Query(state); err == io.EOF {
				return
			} else if err != nil {
				exit(2, "failed to write generated query: %s", err)
			} else {
				queries <- q
			}
		}
	}()

	return func(mtd *desc.MethodDescriptor, callData *runner.CallData) []byte {
		raw, ok := <-queries
		if !ok {
			panic("no more queries")
		}
		req := dshackle.NativeCallRequest{
			Chain: chain,
			Items: []*dshackle.NativeCallItem{
				{
					Id:      uint32(raw.Id),
					Method:  raw.Method,
					Payload: []byte(raw.Params),
				},
			},
		}
		data, err := proto.Marshal(&req)
		if err != nil {
			panic(err)
		}
		*cuCount += gas.CountGas(raw.Method)
		return data
	}
}

func exit(code int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(code)
}

type JsonRpcRequest struct {
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
	Params  []interface{} `json:"params"`
	Method  string        `json:"method"`
}
