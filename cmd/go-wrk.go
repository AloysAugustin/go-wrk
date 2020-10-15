package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/AloysAugustin/go-wrk/pkg/httptest"
	"github.com/montanaflynn/stats"
	"github.com/sirupsen/logrus"
)

func main() {
	conf := &httptest.TestConfig{}

	rate := flag.Int("rate", 100, "connection rate (req/s)")
	duration := flag.Int("duration", 10, "test duration (s)")
	urlFile := flag.String("url-file", "", "file storing urls to test, one per line")

	flag.Parse()

	if *urlFile != "" {
		err := loadUrlsFromFile(conf, *urlFile)
		if err != nil {
			logrus.Errorf("cannot load url file %s: %v", *urlFile, err)
		}
	}
	conf.URLs = append(conf.URLs, flag.Args()...)
	conf.ConnectionRate = float64(*rate)
	conf.Duration = time.Duration(*duration) * time.Second

	if len(conf.URLs) == 0 {
		flag.Usage()
		return
	}

	results := httptest.RunTest(conf)

	fmt.Println("Requests count: ", results.RequestCount)
	fmt.Println("Connect latencies:")
	analyzeResults(results.ConnectLatencies, results.RequestCount)
	fmt.Println("Request latencies:")
	analyzeResults(results.RequestLatencies, results.RequestCount)
}

func loadUrlsFromFile(conf *httptest.TestConfig, filename string) (err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	urls := strings.Split(string(data), "\n")
	for _, u := range urls {
		if u != "" {
			conf.URLs = append(conf.URLs, u)
		}
	}
	return nil
}

func analyzeResults(latencies []time.Duration, count int) {
	floats := make([]float64, count)
	for i := 0; i < count; i++ {
		floats[i] = float64(latencies[i])
	}
	mean, _ := stats.Mean(floats)
	fmt.Println("Average:     ", time.Duration(mean))
	stddev, _ := stats.StandardDeviation(floats)
	fmt.Println("Stddev:      ", time.Duration(stddev))
	max, _ := stats.Max(floats)
	fmt.Println("Max:         ", time.Duration(max))
	fmt.Println("Percentiles: ")
	p, _ := stats.Percentile(floats, 50)
	fmt.Println("     50%:    ", time.Duration(p))
	p, _ = stats.Percentile(floats, 90)
	fmt.Println("     90%:    ", time.Duration(p))
	p, _ = stats.Percentile(floats, 99)
	fmt.Println("     99%:    ", time.Duration(p))
	p, _ = stats.Percentile(floats, 99.9)
	fmt.Println("     99.9%:  ", time.Duration(p))
	p, _ = stats.Percentile(floats, 99.99)
	fmt.Println("     99.99%: ", time.Duration(p))
}
