// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// The service invokes lifecycle hooks as POSTs to
// /aws/lambda-microvms/runtime/v1/<hook> on the image's configured hooks port.
func hook(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, r.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	for _, h := range []string{"run", "terminate", "suspend", "resume", "ready", "validate"} {
		http.HandleFunc("POST /aws/lambda-microvms/runtime/v1/"+h, hook)
	}
	http.HandleFunc("GET /", health)
	http.HandleFunc("GET /health", health)

	go func() {
		_ = http.ListenAndServe(":9000", nil)
	}()

	sleepDuration := time.Second * 5
	for {
		fmt.Println("hello") // nosemgrep:ci.calling-fmt.Print-and-variants
		time.Sleep(sleepDuration)
	}
}
