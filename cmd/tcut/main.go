// Cut a Compass data file when the timetag is greater than
// a certain number of seconds.
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s infile outfile time_in_s\n", os.Args[0])
		os.Exit(1)
	}
	var tcut int64
	tcut, _ = strconv.ParseInt(os.Args[3], 10, 64)
	tcutns := tcut * 1e12
	fin, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer fin.Close()
	r := csv.NewReader(fin)
	r.Comma = ';'
	r.ReuseRecord = true
	fout, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()
	w := csv.NewWriter(fout)
	w.Comma = r.Comma
	items, rerr := r.Read() // read header line
	if rerr != nil {
		log.Fatal(rerr)
	}
	w.Write(items) // write header line
	var timetag, tmax int64
	for {
		items, rerr = r.Read()
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			log.Fatal(err)
		}
		timetag, _ = strconv.ParseInt(items[2], 10, 64)
		if timetag > tcutns {
			break
		}
		w.Write(items)
		tmax = timetag
	}
	tmax_s := float64(tmax) / 1e12
	fmt.Println("tmax = ", tmax_s)
	w.Flush()
	if err = w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
}
