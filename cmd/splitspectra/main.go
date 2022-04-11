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
const num_out_channels int = 256 // downsampled channels
const dnsmpl_size int = num_channels / num_out_channels
const tepi int64 = 100000000 // 100 microsec in ps
const tcap int64 = 150000000 // 150 microsec in ps
const thresh = 150           // TTL threshold for real PNG pulses
const sec float64 = 1.0e12   // one second in ps

type ttlbuf struct {
	tcurr      int64
	tnext      int64
	reachedEOF bool
}

func nextPulse(r *csv.Reader, tb *ttlbuf) {
	// Read the next ttl value from the CSV file.
	// Some values are bogus, which we catch because the "energy" is
	// below the threshold, and we just skip those. Also, occasionally
	// there's a glitch, where the time is dramatically wrong.
	// The signature is a time that is _before_ the previous time, but what
	// that indicates is the previous time was bad. That's why I'm using
	// a buffer that stores the current and the next ttl data.
	if tb.reachedEOF {
		// no more to read, and the buffer has been used
		return
	}
	for {
		items, err := r.Read()
		if err == io.EOF {
			// no more data to read
			tb.reachedEOF = true
			tb.tcurr, tb.tnext = tb.tnext, 0
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		energy, _ := strconv.Atoi(items[3])
		if energy < thresh {
			continue
		}
		timetag, _ := strconv.ParseInt(items[2], 10, 64)
		if timetag-tb.tnext < 0 {
			// the tnext time is bad
			log.Println("Bad ttl time: ", tb.tnext)
			tb.tnext = timetag
			continue
		}
		tb.tcurr, tb.tnext = tb.tnext, timetag
		if tb.tcurr == 0 {
			// buffer not yet full
			continue
		}
		return
	}
}

type gambuf struct {
	tcurr      int64
	ecurr      int
	tnext      int64
	enext      int
	reachedEOF bool
}

func nextGamma(r *csv.Reader, gb *gambuf) error {
	// Read the next gamma value from the CSV file.
	// Occasionally there's a glitch, where the time is dramatically wrong.
	// The signature is a time that is _before_ the previous time, but what
	// that indicates is the previous time was bad. That's why I'm using
	// a buffer that stores the current and next gamma data.
	if gb.reachedEOF {
		// no more to read, and the buffer has been used
		return io.EOF
	}
	var energy int
	var timetag int64
	for {
		items, err := r.Read()
		if err == io.EOF {
			// no more data to read
			gb.reachedEOF = true
			// I'm assuming there was at least enough data to fill the buffer,
			// so one value left
			gb.tcurr, gb.ecurr = gb.tnext, gb.enext
			gb.tnext, gb.enext = 0, 0
			return nil
		}
		if err != nil {
			log.Fatal(err)
		}
		timetag, _ = strconv.ParseInt(items[2], 10, 64)
		energy, _ = strconv.Atoi(items[3])
		if timetag-gb.tnext < 0 {
			// the data corresponding to tnext is bad
			log.Println("Bad gamma time: ", gb.tnext)
			gb.tnext, gb.enext = timetag, energy
			continue
		}
		gb.tcurr, gb.ecurr = gb.tnext, gb.enext
		gb.tnext, gb.enext = timetag, energy
		if gb.tcurr == 0 {
			// buffer not yet full
			continue
		}
		return nil
	}
}

func downsample(counts *[num_channels]float64, newcounts *[num_out_channels]float64) {
	for i := 0; i < num_out_channels; i++ {
		var sum float64 = 0
		for j := i * dnsmpl_size; j < (i+1)*dnsmpl_size; j++ {
			sum += counts[j]
		}
		newcounts[i] = sum
	}
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
	tb := &ttlbuf{}
	nextPulse(rp, tb) // populate the ttl buffer
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
	gb := &gambuf{}
	tmax := 0.0
	for {
		err = nextGamma(rg, gb)
		if err == io.EOF {
			// done w/ gamma data
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if gb.tcurr < tb.tcurr {
			// before the first pulse, so reject
			continue
		}
		if (!tb.reachedEOF) && (gb.tcurr >= tb.tnext) {
			// need to advance the TTL pulse buffer
			for {
				nextPulse(rp, tb)
				if tb.reachedEOF || (gb.tcurr < tb.tnext) {
					break
				}
			}
		}
		dt := gb.tcurr - tb.tcurr
		if dt < tepi {
			inel[gb.ecurr] += 1
		} else if dt < tcap {
			epi[gb.ecurr] += 1
		} else {
			capt[gb.ecurr] += 1
		}
		total[gb.ecurr] += 1
		tmax = float64(gb.tcurr) / sec
	}
	fmt.Println("tmax = ", tmax)
	fout, err := os.Create(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()
	var epids, inelds, captds, totalds [num_out_channels]float64
	downsample(&total, &totalds)
	downsample(&epi, &epids)
	downsample(&inel, &inelds)
	downsample(&capt, &captds)
	fmt.Fprintf(fout, "channel,epi,inel,cap,total\n")
	for i, val := range totalds {
		fmt.Fprintf(fout, "%v,%v,%v,%v,%v\n",
			i, epids[i]/tmax, inelds[i]/tmax, captds[i]/tmax, val/tmax)
	}
}
