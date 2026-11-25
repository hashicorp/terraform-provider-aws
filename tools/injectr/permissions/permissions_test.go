// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package permissions

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/tools/injectr/model"
	"github.com/hashicorp/terraform-provider-aws/tools/injectr/serviceref"
)

var update = flag.Bool("update", false, "update testdata/permissions_map.json")

const (
	testServiceDataDir     = "../serviceref/testdata/service-data"
	testSDKMapFile         = "testdata/sdkmap_input.json"
	testPermissionsMapFile = "testdata/permissions_map.json"
	realServiceDataDir     = "../service-data"
	realPermissionsMapFile = "../../../permissions_map.json"
)

// TestGeneratePermissionsMap compares the output for the test fixtures to
// testdata/permissions_map.json. Run with -update to regenerate it, then
// review the diff.
func TestGeneratePermissionsMap(t *testing.T) {
	m, err := model.Load(testSDKMapFile)
	if err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "permissions_map.json")
	if err := GeneratePermissionsMap(testServiceDataDir, m, out); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	if *update {
		if err := os.WriteFile(testPermissionsMapFile, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	want, err := os.ReadFile(testPermissionsMapFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output differs from %s; run with -update and review the diff\n%s", testPermissionsMapFile, got)
	}
}

func TestAddPermissionsStats(t *testing.T) {
	t.Parallel()

	refs, err := serviceref.LoadServiceRefs(testServiceDataDir)
	if err != nil {
		t.Fatal(err)
	}
	m, err := model.Load(testSDKMapFile)
	if err != nil {
		t.Fatal(err)
	}

	stats := AddPermissions(m, NewOperationIndex(refs))

	if got, want := stats.FuncRefs, 8; got != want {
		t.Errorf("FuncRefs = %d, want %d", got, want)
	}
	if got, want := stats.SDKCalls, 14; got != want {
		t.Errorf("SDKCalls = %d, want %d", got, want)
	}
	if got, want := stats.Resolved, 10; got != want {
		t.Errorf("Resolved = %d, want %d", got, want)
	}
	wantUnresolved := map[model.UnresolvedReason]int{
		reasonNoIAMActions:           1,
		reasonOperationNotFound:      1,
		reasonServiceHasNoOperations: 2,
	}
	for reason, want := range wantUnresolved {
		if got := stats.Unresolved[reason]; got != want {
			t.Errorf("Unresolved[%s] = %d, want %d", reason, got, want)
		}
	}
}

// TestAddPermissionsRerun checks that running AddPermissions twice gives the
// same result, rather than accumulating unresolved calls.
func TestAddPermissionsRerun(t *testing.T) {
	t.Parallel()

	refs, err := serviceref.LoadServiceRefs(testServiceDataDir)
	if err != nil {
		t.Fatal(err)
	}
	index := NewOperationIndex(refs)
	m, err := model.Load(testSDKMapFile)
	if err != nil {
		t.Fatal(err)
	}

	AddPermissions(m, index)
	first, err := m.Encode()
	if err != nil {
		t.Fatal(err)
	}
	AddPermissions(m, index)
	second, err := m.Encode()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(first, second) {
		t.Error("second AddPermissions run changed the output")
	}
}

func BenchmarkAddPermissions(b *testing.B) {
	skipWithoutRealData(b)

	refs, err := serviceref.LoadServiceRefs(realServiceDataDir)
	if err != nil {
		b.Fatal(err)
	}
	index := NewOperationIndex(refs)
	m, err := model.Load(realPermissionsMapFile)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		AddPermissions(m, index)
	}
}

// BenchmarkGeneratePermissionsMap measures a full run on the real data,
// excluding the network sync.
func BenchmarkGeneratePermissionsMap(b *testing.B) {
	skipWithoutRealData(b)

	m, err := model.Load(realPermissionsMapFile)
	if err != nil {
		b.Fatal(err)
	}

	out := filepath.Join(b.TempDir(), "permissions_map.json")
	stdout := os.Stdout
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		b.Fatal(err)
	}
	defer devNull.Close()

	for b.Loop() {
		os.Stdout = devNull
		err := GeneratePermissionsMap(realServiceDataDir, m, out)
		os.Stdout = stdout
		if err != nil {
			b.Fatal(err)
		}
	}
}

// skipWithoutRealData skips tb unless the tool has been run, so that the real
// service-data and permissions_map.json exist.
func skipWithoutRealData(tb testing.TB) {
	tb.Helper()

	for _, path := range []string{realServiceDataDir, realPermissionsMapFile} {
		if _, err := os.Stat(path); err != nil {
			tb.Skipf("%s not found; run the tool first", path)
		}
	}
}
