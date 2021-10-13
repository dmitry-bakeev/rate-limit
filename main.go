package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

type App struct {
	NetworkPrefix    int
	NumberOfRequests int
	UnitTime         time.Duration
	LimitTime        int
	WaitTime         int
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
}

func (a *App) GetLimitTime() time.Duration {
	return a.UnitTime * time.Duration(a.LimitTime)
}

func (a *App) GetWaitTime() time.Duration {
	return a.UnitTime * time.Duration(a.WaitTime)
}

func main() {}
