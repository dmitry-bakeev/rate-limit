package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

type NetworkStat struct {
	BlockTime time.Time
	Requests  []time.Time
}

type RateLimit map[string]NetworkStat

func (rl RateLimit) AddRequest(networkIP string, requestTime time.Time) {
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

func (rl RateLimit) CheckRequest(networkIP string, requestTime time.Time, a *App) bool {
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

func (rl RateLimit) DeleteNetworkStat(networkIP string) {
	_, ok := rl[networkIP]

	if !ok {
		return
	}

	delete(rl, networkIP)
}

type App struct {
	NetworkPrefix    int
	NumberOfRequests int
	UnitTime         time.Duration
	LimitTime        int
	WaitTime         int
	RateLimitMap     RateLimit
	Host             string
	Port             string
}

func (a *App) Init() {
	NETWORK_PREFIX := os.Getenv("NETWORK_PREFIX")
	NUMBER_OF_REQUESTS := os.Getenv("NUMBER_OF_REQUESTS")
	UNIT_TIME := os.Getenv("UNIT_TIME")
	LIMIT_TIME := os.Getenv("LIMIT_TIME")
	WAIT_TIME := os.Getenv("WAIT_TIME")
	HOST := os.Getenv("HOST")
	PORT := os.Getenv("PORT")

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

	if PORT == "" {
		PORT = "8000"
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

	a.Host = HOST

	a.Port = PORT
}

func (a *App) GetLimitTime() time.Duration {
	return a.UnitTime * time.Duration(a.LimitTime)
}

func (a *App) GetWaitTime() time.Duration {
	return a.UnitTime * time.Duration(a.WaitTime)
}

func GetNetworkIP(ip string, networkPrefix int) (string, error) {
	_, netIP, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ip, networkPrefix))
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("%s/%d", netIP.IP.String(), networkPrefix)
	return result, nil
}

func (a *App) WrapperHandler(fn func(w http.ResponseWriter, r *http.Request, a *App)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, a)
	}
}

func RootHandler(w http.ResponseWriter, r *http.Request, a *App) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Status Not Found\n")
		return
	}

	requestIP := r.Header.Get("X-Forwarded-For")
	if requestIP == "" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK!\n")
		return
	}
	requestTime := time.Now()

	netwprkCIDR, err := GetNetworkIP(requestIP, a.NetworkPrefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error\n")
		return
	}

	// check allow this request
	allow := a.RateLimitMap.CheckRequest(netwprkCIDR, requestTime, a)

	if !allow {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, "Too Many Requests\n")
		return
	}

	a.RateLimitMap.AddRequest(netwprkCIDR, requestTime)
	// check again if this request is equals a.NumberOfRequests then set BlockTime
	a.RateLimitMap.CheckRequest(netwprkCIDR, requestTime, a)

	fmt.Fprintf(w, "OK!\n")
}

func ResetHandler(w http.ResponseWriter, r *http.Request, a *App) {
	ip_param := r.URL.Query().Has("ip")

	if !ip_param {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad Request\n")
		return
	}

	ip := r.URL.Query().Get("ip")

	netwprkCIDR, err := GetNetworkIP(ip, a.NetworkPrefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error\n")
		return
	}

	a.RateLimitMap.DeleteNetworkStat(netwprkCIDR)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK!\n")
}

func main() {
	a := &App{}
	a.Init()

	http.HandleFunc("/", a.WrapperHandler(RootHandler))
	http.HandleFunc("/reset", a.WrapperHandler(ResetHandler))
	http.ListenAndServe(fmt.Sprintf("%s:%s", a.Host, a.Port), nil)
}
