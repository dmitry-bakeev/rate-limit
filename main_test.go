package main_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/dmitry-bakeev/rate-limit"
)

var (
	now  = time.Now()
	cidr = "192.168.1.15/24"
	ip   = "192.168.1.15"
)

func TestInit(t *testing.T) {
	wantedA := main.App{
		NetworkPrefix:    24,
		NumberOfRequests: 100,
		UnitTime:         time.Minute,
		LimitTime:        1,
		WaitTime:         2,
		RateLimitMap:     make(main.RateLimit),
		Port:             "8000",
	}

	testA := main.App{}
	testA.Init()

	equal := reflect.DeepEqual(wantedA, testA)
	if !equal {
		t.Errorf("func gets wrong data: got %v want %v", testA, wantedA)
	}
}

func TestGetLimitTime(t *testing.T) {
	testA := main.App{}
	testA.Init()

	wanted := testA.UnitTime * time.Duration(testA.LimitTime)
	got := testA.GetLimitTime()

	if wanted != got {
		t.Errorf("func returns wrong value: got %v want %v", got, wanted)
	}
}

func TestGetWaitTime(t *testing.T) {
	testA := main.App{}
	testA.Init()

	wanted := testA.UnitTime * time.Duration(testA.WaitTime)
	got := testA.GetWaitTime()

	if wanted != got {
		t.Errorf("func returns wrong value: got %v want %v", got, wanted)
	}
}

func TestAddRequest(t *testing.T) {
	testRL := make(main.RateLimit)

	wantedRL := make(main.RateLimit)
	wantedRL[cidr] = main.NetworkStat{
		Requests: []time.Time{now},
	}
	testRL.AddRequest(cidr, now)
	equal := reflect.DeepEqual(testRL, wantedRL)
	if !equal {
		t.Errorf("func adds wrong data: got %v want %v", testRL, wantedRL)
	}
}

func TestCheckRequest(t *testing.T) {
	testA := main.App{}
	testA.Init()

	testA.RateLimitMap.AddRequest(cidr, now)

	result := testA.RateLimitMap.CheckRequest(cidr, now, &testA)

	if !result {
		t.Errorf("func returns wrong result: got %v want %v", result, true)
	}
}

func TestCheckRequestDecline(t *testing.T) {
	testA := main.App{}
	testA.Init()

	for i := 0; i < testA.NumberOfRequests+10; i++ {
		testA.RateLimitMap.AddRequest(cidr, now)
	}

	result := testA.RateLimitMap.CheckRequest(cidr, now, &testA)

	if result {
		t.Errorf("func returns wrong result: got %v want %v", result, false)
	}
}

func TestDeleteNetworkStat(t *testing.T) {
	testA := main.App{}
	testA.Init()

	testA.RateLimitMap.AddRequest(cidr, now)

	testA.RateLimitMap.DeleteNetworkStat(cidr)

	_, ok := testA.RateLimitMap[cidr]

	if ok {
		t.Errorf("func does not work: got %v want %v", ok, false)
	}
}

func TestGetNetworkIP(t *testing.T) {
	networkIP, err := main.GetNetworkIP(ip, 24)

	if err != nil {
		t.Errorf("func returns error: got %v want %v", err, nil)
	}

	wanted := "192.168.1.0/24"

	if networkIP != wanted {
		t.Errorf("func returns wrong value: got %v want %v", networkIP, wanted)
	}
}

func TestGetNetworkIPError(t *testing.T) {
	networkIP, err := main.GetNetworkIP(ip, 33)

	if err == nil {
		t.Errorf("func does not return error")
	}

	wanted := ""

	if networkIP != wanted {
		t.Errorf("func returns wrong value: got %v want %v", networkIP, wanted)
	}
}

func TestRootHandler(t *testing.T) {
	testA := main.App{}
	testA.Init()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testA.WrapperHandler(main.RootHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returns wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	wanted := "OK!\n"

	if rr.Body.String() != wanted {
		t.Errorf("handler returns wrong data: got %v want %v", rr.Body.String(), wanted)
	}
}

func TestRootHandlerInternalError(t *testing.T) {
	testA := main.App{}
	testA.Init()

	req, err := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-Forwarded-For", "1.2.3")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testA.WrapperHandler(main.RootHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("handler returns wrong status code: got %v want %v", rr.Code, http.StatusInternalServerError)
	}

	wanted := "Internal Server Error\n"

	if rr.Body.String() != wanted {
		t.Errorf("handler returns wrong data: got %v want %v", rr.Body.String(), wanted)
	}
}

func TestRootHandlerManyRequests(t *testing.T) {
	testA := main.App{}
	testA.Init()

	req, err := http.NewRequest("GET", "/", nil)
	req.Header.Add("X-Forwarded-For", ip)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testA.WrapperHandler(main.RootHandler))

	for i := 0; i < testA.NumberOfRequests; i++ {
		handler.ServeHTTP(rr, req)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("handler returns wrong status code: got %v want %v", rr.Code, http.StatusTooManyRequests)
	}

	wanted := "Too Many Requests\n"

	if rr.Body.String() != wanted {
		t.Errorf("handler returns wrong data: got %v want %v", rr.Body.String(), wanted)
	}
}

func TestResetHandler(t *testing.T) {
	testA := main.App{}
	testA.Init()

	req, err := http.NewRequest("GET", "/reset", nil)
	if err != nil {
		t.Fatal(err)
	}

	query := req.URL.Query()
	query.Add("ip", ip)
	req.URL.RawQuery = query.Encode()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testA.WrapperHandler(main.ResetHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returns wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	wanted := "OK!\n"

	if rr.Body.String() != wanted {
		t.Errorf("handler returns wrong data: got %v want %v", rr.Body.String(), wanted)
	}
}

func TestResetHandlerBadRequest(t *testing.T) {
	testA := main.App{}
	testA.Init()

	req, err := http.NewRequest("GET", "/reset", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testA.WrapperHandler(main.ResetHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("handler returns wrong status code: got %v want %v", rr.Code, http.StatusBadRequest)
	}

	wanted := "Bad Request\n"

	if rr.Body.String() != wanted {
		t.Errorf("handler returns wrong data: got %v want %v", rr.Body.String(), wanted)
	}
}

func TestResetHandlerInternalError(t *testing.T) {
	testA := main.App{}
	testA.Init()

	req, err := http.NewRequest("GET", "/reset", nil)
	if err != nil {
		t.Fatal(err)
	}

	query := req.URL.Query()
	query.Add("ip", "1.2.3")
	req.URL.RawQuery = query.Encode()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testA.WrapperHandler(main.ResetHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("handler returns wrong status code: got %v want %v", rr.Code, http.StatusInternalServerError)
	}

	wanted := "Internal Server Error\n"

	if rr.Body.String() != wanted {
		t.Errorf("handler returns wrong data: got %v want %v", rr.Body.String(), wanted)
	}
}
