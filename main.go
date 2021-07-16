package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bmizerany/perks/quantile"
)

type metrics struct {
	p50   float64
	p75   float64
	p90   float64
	p95   float64
	p99   float64
	count int
}

func (m metrics) String() string {
	return fmt.Sprintf("count: %d\tp50: %.1fms\tp75: %.1fms\tp90: %.1fms\tp95: %.1fms\tp99: %.1fms",
		m.count, m.p50, m.p75, m.p90, m.p95, m.p99)
}

func main() {
	ch := make(chan float64)
	go sendFloats(ch)

	// before entering the loop
	fmt.Print("\033[s") // save the cursor position

	// Compute the 50th, 90th, and 99th percentile.
	q := quantile.NewTargeted(0.50, 0.75, 0.90, 0.95, 0.99)
	for v := range ch {
		q.Insert(v)
		fmt.Print("\033[u\033[K") // restore the cursor position and clear the line
		fmt.Printf("%+v", metrics{
			p50:   q.Query(0.50),
			p75:   q.Query(0.75),
			p90:   q.Query(0.90),
			p95:   q.Query(0.95),
			p99:   q.Query(0.99),
			count: q.Count(),
		})
	}
}

func sendFloats(ch chan<- float64) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data := strings.Split(scanner.Text(), "\t")
		if len(data) > 7 {
			duration, err := time.ParseDuration(data[6])
			if err != nil {
				log.Fatal(err)
				return
			}
			if len(os.Args) == 2 {
				if strings.HasPrefix(strings.ToLower(data[7]), strings.ToLower(os.Args[1])) {
					ms := float64(duration) / float64(time.Millisecond)
					ch <- ms
				}
			} else {
				ms := float64(duration) / float64(time.Millisecond)
				ch <- ms
			}
		}
	}
	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}
	close(ch)
}
