// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package model

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const (
	sdkmapGoldenDir        = "../sdkmap/testdata/golden"
	permissionsTestdata    = "../permissions/testdata"
	realPermissionsMapFile = "../../../permissions_map.json"
)

// mapFiles returns sdkmap's golden test outputs and the permissions test
// fixtures, plus the real permissions_map.json if it exists.
func mapFiles(t *testing.T) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(sdkmapGoldenDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("no sdkmap golden files in %s", sdkmapGoldenDir)
	}
	paths = append(paths,
		filepath.Join(permissionsTestdata, "sdkmap_input.json"),
		filepath.Join(permissionsTestdata, "permissions_map.json"),
	)
	if _, err := os.Stat(realPermissionsMapFile); err == nil {
		paths = append(paths, realPermissionsMapFile)
	}

	return paths
}

// TestMapRoundTrip checks that a file read and written back without changes
// is byte-for-byte identical.
func TestMapRoundTrip(t *testing.T) {
	t.Parallel()

	for _, path := range mapFiles(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			m, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			got, err := m.Encode()
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(got, want) {
				t.Error("re-encoded map differs from the original")
			}
		})
	}
}

// TestMapFuncRefs checks that FuncRefs visits every function reference, by
// counting the objects with an "sdk_calls" key in the untyped JSON.
func TestMapFuncRefs(t *testing.T) {
	t.Parallel()

	for _, path := range mapFiles(t) {
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

			m, err := Load(path)
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

func TestMapWriteReplacesFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "permissions_map.json")
	if err := os.WriteFile(path, []byte("old contents"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := &Map{Resources: []Resource{}, DataSources: []DataSource{}}
	if err := m.Write(path); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n  \"resources\": [],\n  \"dataSources\": []\n}\n"; string(got) != want {
		t.Errorf("file = %q, want %q", got, want)
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("directory has %d entries, want 1 (temporary file left behind?)", len(entries))
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
