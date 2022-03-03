// Compute an event time histogram from a Compass csv file.
// The time separation between each even and the following event
// is binned on a logarithmic scale.
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
)

const num_bins int = 140
const maxtime float64 = 100.0 // seconds
const onesec float64 = 1.0e12 // one second in ps

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
	var items []string
	_, err = r.Read() // header line; ignore
	if err != nil {
		log.Fatal(err)
	}
	items, err = r.Read() // get first event
	if err != nil {
		log.Fatal(err)
	}
	tprev, _ := strconv.ParseInt(items[2], 10, 64)
	var timetag int64
	var timebins [num_bins]int
	var delt float64
	var bin int
	count := 1
	ferr, _ := os.Create("errors.txt")
	defer ferr.Close()
	badcount := 0
	for {
		items, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		timetag, _ = strconv.ParseInt(items[2], 10, 64)
		delt = float64(timetag-tprev) / onesec
		if delt < 0 {
			fmt.Fprintf(ferr, "dt < 0: t = %v, dt = %v\n", timetag, delt)
			fmt.Printf("dt < 0: t = %v, dt = %v\n", timetag, delt)
			badcount++
			tprev = timetag
			continue
		}
		bin = int(10.0*math.Log10(delt) + 120)
		if bin > num_bins {
			fmt.Fprintf(ferr, "dt exceeds %v seconds: t = %v, dt = %v, bin = %v\n", maxtime, timetag, delt, bin)
			fmt.Printf("dt exceeds %v seconds: t = %v, dt = %v, bin = %v\n", maxtime, timetag, delt, bin)
			badcount++
			tprev = timetag
			continue
		}
		timebins[bin]++
		count++
		tprev = timetag
	}
	fout, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()
	fmt.Fprintf(fout, "logtime,frac\n")
	fmt.Println("Total counts = ", count)
	fmt.Println("Bad counts = ", badcount)
	for bin, val := range timebins {
		logt := float64(bin-120) / 10
		fmt.Fprintf(fout, "%v,%v\n", logt, float64(val)/float64(count-1))
	}
}
