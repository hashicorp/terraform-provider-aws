package awsops

import (
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

func testdataDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata")
}

func TestAnalyze_SDKResource(t *testing.T) {
	dir := filepath.Join(testdataDir(t), "sdkresource")
	results, err := analyzePackage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	res, ok := results["aws_test_thing"]
	if !ok {
		t.Fatalf("expected resource aws_test_thing, got keys: %v", mapKeys(results))
	}

	assertOps(t, "create", res.Create, []string{"CreateBucket", "PutBucketTagging"})
	assertOps(t, "read", res.Read, []string{"GetBucketLocation"})
	assertOps(t, "update", res.Update, []string{"PutBucketTagging"})
	assertOps(t, "delete", res.Delete, []string{"DeleteBucket"})
}

func TestAnalyze_FrameworkResource(t *testing.T) {
	dir := filepath.Join(testdataDir(t), "fwresource")
	results, err := analyzePackage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	res, ok := results["aws_test_widget"]
	if !ok {
		t.Fatalf("expected resource aws_test_widget, got keys: %v", mapKeys(results))
	}

	assertOps(t, "create", res.Create, []string{"CreateTable", "TagResource"})
	assertOps(t, "read", res.Read, []string{"DescribeTable"})
	assertOps(t, "update", res.Update, []string{"UpdateTable"})
	assertOps(t, "delete", res.Delete, []string{"DeleteTable"})
}

func assertOps(t *testing.T, method string, got, want []string) {
	t.Helper()
	sort.Strings(got)
	sort.Strings(want)

	if len(got) != len(want) {
		t.Errorf("%s: got %v, want %v", method, got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s[%d]: got %q, want %q", method, i, got[i], want[i])
		}
	}
}

func mapKeys(m map[string]ResourceOps) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
