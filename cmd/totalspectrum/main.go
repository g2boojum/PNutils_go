// Compute the total spectrum from a Compass spectrum csv
// Expected records are BOARD;CHANNEL;TIMETAG;ENERGY;ENERGYSHORT;FLAGS
// and we're just using TIMETAG and ENERGY here.
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

const num_channels int = 4096
const num_out_channels int = 256
const dnsmpl_size int = num_channels / num_out_channels

func downsample(counts *[num_channels]uint64, newcounts *[num_out_channels]uint64) {
	for i := 0; i < num_out_channels; i++ {
		var sum uint64 = 0
		for j := i * dnsmpl_size; j < (i+1)*dnsmpl_size; j++ {
			sum += counts[j]
		}
		newcounts[i] = sum
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s infile outfile\n", os.Args[0])
		os.Exit(1)
	}
	fin, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer fin.Close()
	r := csv.NewReader(fin)
	r.Comma = ';'
	r.ReuseRecord = true
	isHeader := true
	var total [num_channels]uint64
	var tmax, timetag, firsttime int64
	var energy int
	for {
		items, rerr := r.Read()
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			log.Fatal(err)
		}
		if isHeader { // skip header line
			isHeader = false
			continue
		}
		energy, _ = strconv.Atoi(items[3])
		timetag, _ = strconv.ParseInt(items[2], 10, 64)
		if firsttime == 0 {
			firsttime = timetag
		}
		total[energy] += 1
		tmax = timetag
	}
	firsttime_s := float64(firsttime) / 1e12
	fmt.Println("firsttime = ", firsttime_s)
	tmax_s := float64(tmax) / 1e12
	fmt.Println("tmax = ", tmax_s)
	var totalds [num_out_channels]uint64
	downsample(&total, &totalds)
	fout, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()
	fmt.Fprintf(fout, "channel,total\n")
	for i, val := range totalds {
		fmt.Fprintf(fout, "%v,%v\n", i, float64(val)/(tmax_s-firsttime_s))
	}
}
