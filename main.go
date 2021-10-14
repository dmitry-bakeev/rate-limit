package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

type NetworkStat struct {
	BlockTime time.Time
	Requests  []time.Time
}

type RateLimit map[string]NetworkStat

func (rl RateLimit) addRequest(networkIP string, requestTime time.Time) {
	current, ok := rl[networkIP]

	if !ok {
		rl[networkIP] = NetworkStat{
			Requests: []time.Time{requestTime},
		}
		return
	}

	current.Requests = append(current.Requests, requestTime)
	rl[networkIP] = current
}

func (rl RateLimit) checkRequest(networkIP string, requestTime time.Time, a *App) bool {
	current, ok := rl[networkIP]

	if !ok {
		return true
	}

	// count requests for last time interval
	checkTime := requestTime.Add(-a.GetLimitTime())
	counter := 0
	for i := len(current.Requests) - 1; i >= 0; i-- {
		if current.Requests[i].Before(checkTime) {
			break
		}
		counter++
	}

	if current.BlockTime.After(requestTime) {
		return false
	}

	if counter < a.NumberOfRequests {
		return true
	}

	if counter == a.NumberOfRequests {
		current.BlockTime = requestTime.Add(a.GetWaitTime())
		rl[networkIP] = current
	}

	return false
}

type App struct {
	NetworkPrefix    int
	NumberOfRequests int
	UnitTime         time.Duration
	LimitTime        int
	WaitTime         int
	RateLimitMap     RateLimit
}

func (a *App) Init() {
	NETWORK_PREFIX := os.Getenv("NETWORK_PREFIX")
	NUMBER_OF_REQUESTS := os.Getenv("NUMBER_OF_REQUESTS")
	UNIT_TIME := os.Getenv("UNIT_TIME")
	LIMIT_TIME := os.Getenv("LIMIT_TIME")
	WAIT_TIME := os.Getenv("WAIT_TIME")

	if NETWORK_PREFIX == "" {
		NETWORK_PREFIX = "24"
	}

	if NUMBER_OF_REQUESTS == "" {
		NUMBER_OF_REQUESTS = "100"
	}

	if LIMIT_TIME == "" {
		LIMIT_TIME = "1"
	}

	if WAIT_TIME == "" {
		WAIT_TIME = "2"
	}

	var err error

	a.NetworkPrefix, err = strconv.Atoi(NETWORK_PREFIX)
	if err != nil {
		log.Fatalln(err)
	}

	a.NumberOfRequests, err = strconv.Atoi(NUMBER_OF_REQUESTS)
	if err != nil {
		log.Fatalln(err)
	}

	switch UNIT_TIME {
	case "Second":
		a.UnitTime = time.Second
	case "Hour":
		a.UnitTime = time.Hour
	default:
		a.UnitTime = time.Minute
	}

	a.LimitTime, err = strconv.Atoi(LIMIT_TIME)
	if err != nil {
		log.Fatalln(err)
	}

	a.WaitTime, err = strconv.Atoi(WAIT_TIME)
	if err != nil {
		log.Fatalln(err)
	}

	a.RateLimitMap = make(RateLimit)
}

func (a *App) GetLimitTime() time.Duration {
	return a.UnitTime * time.Duration(a.LimitTime)
}

func (a *App) GetWaitTime() time.Duration {
	return a.UnitTime * time.Duration(a.WaitTime)
}

func main() {}
