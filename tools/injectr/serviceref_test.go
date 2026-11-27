// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// useServer starts an HTTPS test server serving responses, keyed by path, and
// points httpClient at it for the rest of the test. "HOST" in a response body
// is replaced with the server's host. Tests using it must not run in parallel.
func useServer(t *testing.T, timeout time.Duration, responses map[string]response) *httptest.Server {
	t.Helper()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, ok := responses[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(resp.status)
		fmt.Fprint(w, strings.ReplaceAll(resp.body, "HOST", r.Host))
	}))
	t.Cleanup(srv.Close)
	useClient(t, srv.Client(), timeout)

	return srv
}

// useClient points httpClient at client, with timeout, for the rest of the test.
func useClient(t *testing.T, client *http.Client, timeout time.Duration) {
	t.Helper()

	client.Timeout = timeout
	old := httpClient
	httpClient = client
	t.Cleanup(func() { httpClient = old })
}

type response struct {
	status int
	body   string
}

func TestValidateServiceURL(t *testing.T) {
	t.Parallel()

	const httpsURL = "https://servicereference.us-east-1.amazonaws.com/v1/s3/s3.json"

	testCases := map[string]struct {
		su      ServiceURL
		wantErr bool
	}{
		"valid":                {su: ServiceURL{Service: "s3", URL: httpsURL}},
		"valid with dash":      {su: ServiceURL{Service: "s3-object-lambda", URL: httpsURL}},
		"empty name":           {su: ServiceURL{URL: httpsURL}, wantErr: true},
		"dot":                  {su: ServiceURL{Service: ".", URL: httpsURL}, wantErr: true},
		"dot dot":              {su: ServiceURL{Service: "..", URL: httpsURL}, wantErr: true},
		"traversal":            {su: ServiceURL{Service: "../../../sdkmap", URL: httpsURL}, wantErr: true},
		"slash":                {su: ServiceURL{Service: "a/b", URL: httpsURL}, wantErr: true},
		"backslash":            {su: ServiceURL{Service: `a\b`, URL: httpsURL}, wantErr: true},
		"hidden file":          {su: ServiceURL{Service: ".s3", URL: httpsURL}, wantErr: true},
		"http URL":             {su: ServiceURL{Service: "s3", URL: "http://servicereference.us-east-1.amazonaws.com/"}, wantErr: true},
		"missing URL":          {su: ServiceURL{Service: "s3"}, wantErr: true},
		"unparseable URL":      {su: ServiceURL{Service: "s3", URL: "https://%zz"}, wantErr: true},
		"name with dot inside": {su: ServiceURL{Service: "a.b", URL: httpsURL}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := validateServiceURL(tc.su); (err != nil) != tc.wantErr {
				t.Errorf("err = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}

func TestGetServiceRefURLs(t *testing.T) {
	srv := useServer(t, 5*time.Second, map[string]response{
		"/ok":        {http.StatusOK, `[{"service": "s3", "url": "https://example.com/s3.json", "modified": 1}]`},
		"/error":     {http.StatusInternalServerError, `[]`},
		"/empty":     {http.StatusOK, `[]`},
		"/null":      {http.StatusOK, `null`},
		"/malformed": {http.StatusOK, `[`},
		"/traversal": {http.StatusOK, `[{"service": "../../../sdkmap", "url": "https://example.com/x.json"}]`},
		"/http-url":  {http.StatusOK, `[{"service": "s3", "url": "http://example.com/s3.json"}]`},
	})

	urls, err := GetServiceRefURLs(srv.URL + "/ok")
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 1 || urls[0].Service != "s3" {
		t.Errorf("urls = %v", urls)
	}

	for _, path := range []string{"/error", "/empty", "/null", "/malformed", "/traversal", "/http-url"} {
		if _, err := GetServiceRefURLs(srv.URL + path); err == nil {
			t.Errorf("%s: expected error", path)
		}
	}
}

func TestGetServiceRef(t *testing.T) {
	const oldContent = `{"Name": "s3", "Version": "old"}`

	srv := useServer(t, 5*time.Second, map[string]response{
		"/valid":     {http.StatusOK, `{"Name": "s3", "Version": "new"}`},
		"/error":     {http.StatusInternalServerError, `{"Name": "s3"}`},
		"/malformed": {http.StatusOK, `{`},
		"/nameless":  {http.StatusOK, `{}`},
	})

	testCases := map[string]struct {
		service, path string
		wantErr       bool
		wantContent   string
	}{
		"valid replaces file":      {service: "s3", path: "/valid", wantContent: `{"Name": "s3", "Version": "new"}`},
		"error keeps file":         {service: "s3", path: "/error", wantErr: true, wantContent: oldContent},
		"malformed keeps file":     {service: "s3", path: "/malformed", wantErr: true, wantContent: oldContent},
		"nameless keeps file":      {service: "s3", path: "/nameless", wantErr: true, wantContent: oldContent},
		"traversal writes nothing": {service: "../escape", path: "/valid", wantErr: true, wantContent: oldContent},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "service-data")
			path := filepath.Join(dir, "s3.json")
			writeFile(t, path, oldContent)

			err := GetServiceRef(ServiceURL{Service: tc.service, URL: srv.URL + tc.path}, dir)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %t", err, tc.wantErr)
			}

			if got := readFile(t, path); got != tc.wantContent {
				t.Errorf("s3.json = %s, want %s", got, tc.wantContent)
			}
			if _, err := os.Stat(filepath.Join(dir, "../escape.json")); err == nil {
				t.Error("file written outside the service data directory")
			}
			if entries, _ := os.ReadDir(dir); len(entries) != 1 {
				t.Errorf("service data directory has %d entries, want 1 (temporary file left behind?)", len(entries))
			}
		})
	}
}

func TestFetchTimeout(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send the headers and part of the body, then stall.
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "[")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)
	useClient(t, srv.Client(), 100*time.Millisecond)

	start := time.Now()
	if _, err := fetch(srv.URL); err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("fetch took %s, want it to time out", elapsed)
	}
}

func TestSyncServiceRefs(t *testing.T) {
	srv := useServer(t, 5*time.Second, map[string]response{
		"/index":   {http.StatusOK, `[{"service": "s3", "url": "https://HOST/s3.json", "modified": 2}]`},
		"/s3.json": {http.StatusOK, `{"Name": "s3"}`},
		"/error":   {http.StatusInternalServerError, `[]`},
		"/empty":   {http.StatusOK, `[]`},
	})

	t.Run("downloads and indexes", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "service-data")
		indexPath := filepath.Join(t.TempDir(), "service-urls-cache.json")

		if err := syncServiceRefs(srv.URL+"/index", dir, indexPath); err != nil {
			t.Fatal(err)
		}

		if got := readFile(t, filepath.Join(dir, "s3.json")); got != `{"Name": "s3"}` {
			t.Errorf("s3.json = %s", got)
		}
		index, err := LoadServiceIndex(indexPath)
		if err != nil {
			t.Fatal(err)
		}
		if index["s3"].Modified != 2 {
			t.Errorf("index = %v", index)
		}
	})

	// A failed or empty index must not prune cached files or replace the index.
	for _, path := range []string{"/error", "/empty"} {
		t.Run("bad index keeps cache "+path, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "service-data")
			indexPath := filepath.Join(t.TempDir(), "service-urls-cache.json")
			const cachedIndex = `{"s3": {"service": "s3", "url": "https://example.com/s3.json", "modified": 1}}`
			writeFile(t, filepath.Join(dir, "s3.json"), `{"Name": "s3"}`)
			writeFile(t, indexPath, cachedIndex)

			if err := syncServiceRefs(srv.URL+path, dir, indexPath); err == nil {
				t.Fatal("expected error")
			}

			if got := readFile(t, filepath.Join(dir, "s3.json")); got != `{"Name": "s3"}` {
				t.Errorf("s3.json = %s", got)
			}
			if got := readFile(t, indexPath); got != cachedIndex {
				t.Errorf("index = %s, want unchanged", got)
			}
		})
	}
}
