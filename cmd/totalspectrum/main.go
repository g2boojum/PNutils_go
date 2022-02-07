// Compute the total spectrum from a Compass spectrum csv
package main

import (
    "bufio"
    "fmt"
    "os"
    "log"
    "strings"
    "strconv"
)

const num_channels int = 4096

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
    fout, err := os.Create(os.Args[2])
    if err != nil {
        log.Fatal(err)
    }
    fmt.Fprintf(fout, "channel,total\n")
    var total[num_channels] uint64
    var tmax int64
    var timetag int64
    var energy int
    for scanner.Scan() {
        ss := strings.Split(scanner.Text(), ";")
        energy, _ = strconv.Atoi(ss[3])
        timetag, _ = strconv.ParseInt(ss[2], 10, 64)
        total[energy] += 1
        tmax = timetag
    }
    tmax_s := float64(tmax)/1e12
    fmt.Println("tmax = ", tmax_s)
    for i, val := range total {
        fmt.Fprintf(fout, "%v,%v\n", i, float64(val)/tmax_s)
    }
}

    

