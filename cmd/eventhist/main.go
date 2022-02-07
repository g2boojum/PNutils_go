// Compute an event time histogram from a Compass csv file.
// The time separation between each even and the following event
// is binned on a logarithmic scale.
package main

import (
    "bufio"
    "fmt"
    "os"
    "log"
    "strings"
    "strconv"
    "math"
)

const num_bins int = 140
const maxtime float64 = 100.0 // seconds
const onesec float64 = 1.0e12 // one second in ps

func main () {
    if len(os.Args) != 3 {
        fmt.Printf("Usage: %s infile outfile\n", os.Args[0])
        os.Exit(1)
    }
    fin, err := os.Open(os.Args[1])
    if err != nil {
        log.Fatal(err)
    }
    scanner := bufio.NewScanner(fin)
    scanner.Scan() // skip header line
    scanner.Scan() // get first event
    ss := strings.Split(scanner.Text(), ";")
    timetag, _ := strconv.ParseInt(ss[2], 10, 64)
    fout, err := os.Create(os.Args[2])
    if err != nil {
        log.Fatal(err)
    }
    fmt.Fprintf(fout, "logtime,frac\n")
    var timebins[num_bins] int
    var tprev int64
    var delt float64
    var bin int
    var count int = 1
    for scanner.Scan() {
        tprev = timetag
        ss := strings.Split(scanner.Text(), ";")
        timetag, _ = strconv.ParseInt(ss[2], 10, 64)
        delt = float64(timetag - tprev)/onesec
        bin = int(10.0*math.Log10(delt) + 120)
        if bin > num_bins {
            fmt.Println("High bin: ", bin, ", delt = ", delt)
            continue
        }
        if bin < 0 {
            fmt.Println("Low bin: ", bin, ", delt = ", delt)
            fmt.Println("\t tprev = ", tprev, ", tcurr = ", timetag)
            continue
        }
        timebins[bin]++
        count++
    }
    fmt.Println("counts = ", count)
    for bin, val := range timebins {
        logt := float64(bin - 120)/10
        fmt.Fprintf(fout, "%v,%v\n", logt, float64(val)/float64(count-1))
    }
}
