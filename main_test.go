package main_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/dmitry-bakeev/rate-limit"
)

var (
	now  = time.Now()
	cidr = "192.168.1.15/24"
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
		t.Errorf("func get wrong data: got %v want %v", testA, wantedA)
	}
}

func TestGetLimitTime(t *testing.T) {
	testA := main.App{}
	testA.Init()

	wanted := testA.UnitTime * time.Duration(testA.LimitTime)
	got := testA.GetLimitTime()

	if wanted != got {
		t.Errorf("func return wrong value: got %v want %v", got, wanted)
	}
}

func TestGetWaitTime(t *testing.T) {
	testA := main.App{}
	testA.Init()

	wanted := testA.UnitTime * time.Duration(testA.WaitTime)
	got := testA.GetWaitTime()

	if wanted != got {
		t.Errorf("func return wrong value: got %v want %v", got, wanted)
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
		t.Errorf("func add wrong data: got %v want %v", testRL, wantedRL)
	}
}

func TestCheckRequest(t *testing.T) {
	testA := main.App{}
	testA.Init()

	testA.RateLimitMap.AddRequest(cidr, now)

	result := testA.RateLimitMap.CheckRequest(cidr, now, &testA)

	if !result {
		t.Errorf("func return wrong result: got %v want %v", result, true)
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
		t.Errorf("func return wrong result: got %v want %v", result, false)
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
	networkIP, err := main.GetNetworkIP("192.168.1.15", 24)

	if err != nil {
		t.Errorf("func return error: got %v want %v", err, nil)
	}

	wanted := "192.168.1.0/24"

	if networkIP != wanted {
		t.Errorf("func return wrong value: got %v want %v", networkIP, wanted)
	}
}

func TestGetNetworkIPError(t *testing.T) {
	networkIP, err := main.GetNetworkIP("192.168.1.15", 33)

	if err == nil {
		t.Errorf("func does not return error")
	}

	wanted := ""

	if networkIP != wanted {
		t.Errorf("func return wrong value: got %v want %v", networkIP, wanted)
	}
}
