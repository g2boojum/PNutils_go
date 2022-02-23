// Compute the total spectrum from a Compass spectrum csv
// Expected records are BOARD;CHANNEL;TIMETAG;ENERGY;ENERGYSHORT;FLAGS
// and we're just using TIMETAG and ENERGY here.
package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

const num_channels int = 4096

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
	var tmax, timetag int64
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
		total[energy] += 1
		tmax = timetag
	}
	tmax_s := float64(tmax) / 1e12
	fmt.Println("tmax = ", tmax_s)
	fout, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()
	w := bufio.NewWriter(fout)
	fmt.Fprintf(w, "channel,total\n")
	for i, val := range total {
		fmt.Fprintf(w, "%v,%v\n", i, float64(val)/tmax_s)
	}
}
