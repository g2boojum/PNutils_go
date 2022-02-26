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
const tepi int64 = 100000000 // 100 microsec in ps
const tcap int64 = 150000000 // 150 microsec in ps
const thresh = 150           // TTL threshold for real PNG pulses
const sec float64 = 1.0e12             // one second in ps

func nextPulse(r *csv.Reader) (int64, error) {
	for {
		items, err := r.Read()
		if err != nil {
			return 0, err
		}
		energy, _ := strconv.Atoi(items[3])
		if energy > thresh {
			timetag, _ := strconv.ParseInt(items[2], 10, 64)
			return timetag, err 
		}
	}
}

func nextGamma(r *csv.Reader, tprev int64, maxdt int64) (int64, int, error) {
	// Read the next gamma value from the CSV file.
	// Occasionally there's a glitch, where the time is dramatically wrong. 
	// In cases I've seen, it's always been about 17 seconds off, but we'll just 
	// look for cases where the time difference is more than a TTL pulse width
	var energy int
	var timetag int64
	for {
		items, err := r.Read()
		if err != nil {
			return 0, 0, err
		}
		timetag, _ = strconv.ParseInt(items[2], 10, 64)
		dt := timetag - tprev
		if dt < 0 {
			dt *= -1
		}
		if dt < maxdt {
			energy, _ = strconv.Atoi(items[3])
			break
		}
		fmt.Println("Bad time: ", timetag)
	}
	return timetag, energy, nil
}


func main() {
	var inel, epi, capt, total [num_channels]float64
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s pulsefile gammafile outfile\n", os.Args[0])
		os.Exit(1)
	}
	// TTL pulse initialization
	pulsefp, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer pulsefp.Close()
	rp := csv.NewReader(pulsefp)
	rp.Comma = ';'
	rp.ReuseRecord = true
	_, err = rp.Read() // skip header line
	if err != nil {
		log.Fatal(err)
	}
	// Start w/ a "current" TTL pulse time and the time of the subsequent pulse
	currttl, err := nextPulse(rp)
	if err != nil {
		log.Fatal(err)
	}
	nextttl, err := nextPulse(rp)
	if err != nil {
		log.Fatal(err)
	}
	maxdt := nextttl - currttl
	lastpulse := false
	// gamma initialization
	gammafp, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer gammafp.Close()
	rg := csv.NewReader(gammafp)
	rg.Comma = ';'
	rg.ReuseRecord = true
	_, err = rg.Read() // skip header line
	if err != nil {
		log.Fatal(err)
	}
	items, _ := rg.Read()
	tprev, _ := strconv.ParseInt(items[2], 10, 64)
	eprev, _ := strconv.Atoi(items[3])
	total[eprev] += 1
	tmax := 0.0
	var tcurr int64
	var ecurr int
	for {
		tcurr, ecurr, err = nextGamma(rg, tprev, maxdt)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		total[ecurr] += 1
		if tcurr < currttl {
			continue // PNG hasn't pulsed yet
		}
		if (!lastpulse) && (tcurr >= nextttl) {
			currttl = nextttl
			nextttl, err = nextPulse(rp)
			if err == io.EOF {
				lastpulse = true
				fmt.Println("Done with pulses. Last pulse at ", currttl)
			}
		}
		dt := tcurr - currttl
		if dt < tepi {
			inel[ecurr] += 1
		} else if dt < tcap {
			epi[ecurr] += 1
		} else {
			capt[ecurr] += 1
		}
		tprev = tcurr
		tmax = float64(tcurr)/sec
	}
	fmt.Println("tmax = ", tmax)
	fout, err := os.Create(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()
	w := bufio.NewWriter(fout)
	fmt.Fprintf(w, "channel,epi,inel,cap,total\n")
	for i, val := range total {
		fmt.Fprintf(w, "%v,%v,%v,%v,%v\n", 
		            i, epi[i]/tmax, inel[i]/tmax, capt[i]/tmax, val/tmax)
	}
}
