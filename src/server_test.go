package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHello(t *testing.T) {
	w := httptest.NewRecorder()

	handleHello(w, nil)

	desiredStatus := http.StatusOK

	if w.Code != desiredStatus {
		t.Errorf("bad response code, expected %v but received %v\nbody: %s\n", desiredStatus, w.Code, w.Body.String())
	}

	expectedMessage := []byte("Hello, World!\n")

	if !bytes.Equal(expectedMessage, w.Body.Bytes()) {
		t.Errorf("bad return, got: %q,\n expected: %q", w.Body.String(), expectedMessage)
	}
}
