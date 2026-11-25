// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const (
	sdkmapGoldenDir = "../sdkmap/testdata/golden"
	realSDKMapFile  = "../../sdkmap.json"
)

// sdkMapFiles returns sdkmap's golden test outputs, plus the real sdkmap.json
// if it exists.
func sdkMapFiles(t *testing.T) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(sdkmapGoldenDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("no sdkmap golden files in %s", sdkmapGoldenDir)
	}
	if _, err := os.Stat(realSDKMapFile); err == nil {
		paths = append(paths, realSDKMapFile)
	}

	return paths
}

// TestPermissionsMapRoundTrip checks that the PermissionsMap types mirror
// sdkmap's output: a file read and written back without changes is
// byte-for-byte identical.
func TestPermissionsMapRoundTrip(t *testing.T) {
	t.Parallel()

	for _, path := range sdkMapFiles(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			m, err := LoadSDKMap(path)
			if err != nil {
				t.Fatal(err)
			}
			got, err := m.encode()
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(got, want) {
				t.Error("re-encoded sdkmap differs from the original; are the PermissionsMap types in sync with tools/sdkmap?")
			}
		})
	}
}

// TestPermissionsMapFuncRefs checks that FuncRefs visits every function reference, by
// counting the objects with an "sdk_calls" key in the untyped JSON.
func TestPermissionsMapFuncRefs(t *testing.T) {
	t.Parallel()

	for _, path := range sdkMapFiles(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var raw any
			if err := json.Unmarshal(b, &raw); err != nil {
				t.Fatal(err)
			}
			want := countFuncRefs(raw)

			m, err := LoadSDKMap(path)
			if err != nil {
				t.Fatal(err)
			}
			got := 0
			for range m.FuncRefs() {
				got++
			}

			if got != want {
				t.Errorf("FuncRefs yielded %d, want %d", got, want)
			}
		})
	}
}

func countFuncRefs(v any) int {
	n := 0
	switch v := v.(type) {
	case map[string]any:
		if _, ok := v["sdk_calls"]; ok {
			n++
		}
		for _, x := range v {
			n += countFuncRefs(x)
		}
	case []any:
		for _, x := range v {
			n += countFuncRefs(x)
		}
	}
	return n
}
