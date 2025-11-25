package main

import (
	"testing"
)

func TestRequestParser(t *testing.T) {
	fakeRequest := []byte("GET /hello HTTP/1.1\r\nHost: localhost:8080\r\n\r\n")
	requestHeader := NewRequestHeader(fakeRequest)

	wantType := "GET"
	wantRoute := "/hello"

	if wantType != requestHeader.getType() {
		t.Errorf(`RequestHeaders.getType() = %s, %s, want match for %s`, wantType, requestHeader.getType(), string(fakeRequest))
	}

	if wantRoute != requestHeader.getRoute() {
		t.Errorf(`RequestHeaders.getRoute() = %s, %s, want match for %s`, wantRoute, requestHeader.getRoute(), string(fakeRequest))
	}
}
