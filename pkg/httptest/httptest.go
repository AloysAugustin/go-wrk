package httptest

import (
	"bufio"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type TestConfig struct {
	URLs           []string
	ConnectionRate float64
	Duration       time.Duration
}

type TestResults struct {
	RequestCount     int
	RequestIndices   []int
	ConnectLatencies []time.Duration
	RequestLatencies []time.Duration

	lock sync.Mutex
}

func RunTest(conf *TestConfig) *TestResults {
	rand.Seed(time.Now().UnixNano())
	expectedRequests := int(2.0 * conf.ConnectionRate * conf.Duration.Seconds())
	requestDelay := time.Nanosecond * time.Duration(1.0e9/conf.ConnectionRate)
	lateCount := 0

	results := &TestResults{
		RequestCount:     0,
		RequestIndices:   make([]int, expectedRequests),
		ConnectLatencies: make([]time.Duration, expectedRequests),
		RequestLatencies: make([]time.Duration, expectedRequests),
	}

	nextTime := time.Now()
	endTime := nextTime.Add(conf.Duration)

	for nextTime.Before(endTime) {
		go makeOneRequest(conf, results)

		nextTime = nextTime.Add(requestDelay)
		sleepTime := nextTime.Sub(time.Now())
		if sleepTime < 0 {
			lateCount++
		} else if sleepTime > 20*time.Microsecond {
			<-time.After(sleepTime)
		}
	}

	results.lock.Lock() // Prevent unfinished requests from modifying results
	if lateCount > 0 {
		logrus.Warnf("Late requests: %d", lateCount)
	}
	return results
}

func makeOneRequest(conf *TestConfig, results *TestResults) {
	id := rand.Intn(len(conf.URLs))
	uri := conf.URLs[id]

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		logrus.Errorf("Cannot create http request for %s: %v", uri, err)
		return
	}
	// Force connection closing at the end
	req.Close = true

	startTime := time.Now()
	conn, err := net.Dial("tcp", req.URL.Host+":80")
	connTime := time.Now()
	if err != nil {
		logrus.Errorf("Cannot connect to %s: %v", req.URL.Host, err)
		return
	}
	req.Write(conn)
	buf := bufio.NewReader(conn)
	resp, err := http.ReadResponse(buf, req)
	respTime := time.Now()
	if resp.StatusCode != 200 {
		logrus.Errorf("HTTP error: %s", resp.Status)
		return
	}

	connDuration := connTime.Sub(startTime)
	reqDuration := respTime.Sub(connTime)

	results.lock.Lock()
	defer results.lock.Unlock()

	results.RequestIndices[results.RequestCount] = id
	results.ConnectLatencies[results.RequestCount] = connDuration
	results.RequestLatencies[results.RequestCount] = reqDuration
	results.RequestCount++
}
