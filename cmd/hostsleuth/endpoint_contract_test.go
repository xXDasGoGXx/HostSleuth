package main

import (
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestContractAddressesFromQuerySupportsRepeatedAndCommaSeparatedValues(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/contract?ip=192.0.2.10&ip=2001:db8::1,192.0.2.11", nil)
	got := contractAddressesFromQuery(request)
	want := []string{"192.0.2.10", "2001:db8::1", "192.0.2.11"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}
