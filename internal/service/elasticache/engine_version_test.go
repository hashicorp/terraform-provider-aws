package elasticache

import (
	"fmt"
	"math"
	"regexp"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestValidMemcachedVersionString(t *testing.T) {
	testcases := []struct {
		version string
		valid   bool
	}{
		{
			version: "1.2.3",
			valid:   true,
		},
		{
			version: "10.20.30",
			valid:   true,
		},
		{
			version: "1.2.",
			valid:   false,
		},
		{
			version: "1.2",
			valid:   false,
		},
		{
			version: "1.",
			valid:   false,
		},
		{
			version: "1",
			valid:   false,
		},
		{
			version: "1.2.x",
			valid:   false,
		},
		{
			version: "1.x",
			valid:   false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.version, func(t *testing.T) {
			warnings, errors := validMemcachedVersionString(testcase.version, "key")

			if l := len(warnings); l != 0 {
				t.Errorf("expected no warnings, got %d", l)
			}

			if testcase.valid {
				if l := len(errors); l != 0 {
					t.Errorf("expected no errors, got %d: %v", l, errors)
				}
			} else {
				if l := len(errors); l == 0 {
					t.Error("expected one error, got none")
				} else if l > 1 {
					t.Errorf("expected one error, got %d: %v", l, errors)
				}
			}
		})
	}
}

func TestValidRedisVersionString(t *testing.T) {
	testcases := []struct {
		version string
		valid   bool
	}{
		{
			version: "5.4.3",
			valid:   true,
		},
		{
			version: "5.4.",
			valid:   false,
		},
		{
			version: "5.4",
			valid:   false,
		},
		{
			version: "5.",
			valid:   false,
		},
		{
			version: "5",
			valid:   false,
		},
		{
			version: "5.4.x",
			valid:   false,
		},
		{
			version: "5.x",
			valid:   false,
		},
		{
			version: "6.x",
			valid:   true,
		},
		{
			version: "6.2",
			valid:   true,
		},
		{
			version: "6.5.0",
			valid:   false,
		},
		{
			version: "6.5.",
			valid:   false,
		},
		{
			version: "6.",
			valid:   false,
		},
		{
			version: "6",
			valid:   false,
		},
		{
			version: "6.y",
			valid:   false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.version, func(t *testing.T) {
			warnings, errors := validRedisVersionString(testcase.version, "key")

			if l := len(warnings); l != 0 {
				t.Errorf("expected no warnings, got %d", l)
			}

			if testcase.valid {
				if l := len(errors); l != 0 {
					t.Errorf("expected no errors, got %d: %v", l, errors)
				}
			} else {
				if l := len(errors); l == 0 {
					t.Error("expected one error, got none")
				} else if l > 1 {
					t.Errorf("expected one error, got %d: %v", l, errors)
				}
			}
		})
	}
}

func TestValidateClusterEngineVersion(t *testing.T) {
	testcases := []struct {
		engine  string
		version string
		valid   bool
	}{
		// Empty engine value is Memcached
		{
			engine:  "",
			version: "1.2.3",
			valid:   true,
		},
		{
			engine:  "",
			version: "6.x",
			valid:   false,
		},
		{
			engine:  "",
			version: "6.0",
			valid:   false,
		},

		{
			engine:  engineMemcached,
			version: "1.2.3",
			valid:   true,
		},
		{
			engine:  engineMemcached,
			version: "6.x",
			valid:   false,
		},
		{
			engine:  engineMemcached,
			version: "6.0",
			valid:   false,
		},

		{
			engine:  engineRedis,
			version: "1.2.3",
			valid:   true,
		},
		{
			engine:  engineRedis,
			version: "6.x",
			valid:   true,
		},
		{
			engine:  engineRedis,
			version: "6.0",
			valid:   true,
		},
	}

	for _, testcase := range testcases {
		t.Run(fmt.Sprintf("%s %s", testcase.engine, testcase.version), func(t *testing.T) {
			err := validateClusterEngineVersion(testcase.engine, testcase.version)

			if testcase.valid {
				if err != nil {
					t.Errorf("expected no error, got %s", err)
				}
			} else {
				if err == nil {
					t.Error("expected an error, got none")
				}
			}
		})
	}
}

type mockGetChangeDiffer struct {
	old, new string
}

func (d *mockGetChangeDiffer) GetChange(key string) (interface{}, interface{}) {
	return d.old, d.new
}

func TestCustomizeDiffEngineVersionIsDowngrade(t *testing.T) {
	testcases := map[string]struct {
		old, new string
		expected bool
	}{
		"no change": {
			old:      "1.2.3",
			new:      "1.2.3",
			expected: false,
		},

		"upgrade minor versions": {
			old:      "1.2.3",
			new:      "1.3.5",
			expected: false,
		},

		"upgrade major versions": {
			old:      "1.2.3",
			new:      "2.4.6",
			expected: false,
		},

		// "upgrade major 6.x": {
		// 	old:            "5.0.6",
		// 	new:            "6.x",
		// 	expectForceNew: false,
		// },

		// "upgrade major 6.digit": {
		// 	old:            "5.0.6",
		// 	new:            "6.0",
		// 	expectForceNew: false,
		// },

		"downgrade minor versions": {
			old:      "1.3.5",
			new:      "1.2.3",
			expected: true,
		},

		"downgrade major versions": {
			old:      "2.4.6",
			new:      "1.2.3",
			expected: true,
		},

		"downgrade from major 6.x": {
			old:      "6.x",
			new:      "5.0.6",
			expected: true,
		},

		"downgrade major 6.digit": {
			old:      "6.2",
			new:      "6.0",
			expected: true,
		},

		"switch major 6.digit to 6.x": {
			old:      "6.2",
			new:      "6.x",
			expected: false,
		},

		"downgrade from major 7.x to 6.x": {
			old:      "7.x",
			new:      "6.x",
			expected: true,
		},

		"downgrade from major 7.digit to 6.x": {
			old:      "7.2",
			new:      "6.x",
			expected: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			diff := &mockGetChangeDiffer{
				old: testcase.old,
				new: testcase.new,
			}

			actual, err := engineVersionIsDowngrade(diff)

			if err != nil {
				t.Fatalf("no error expected, got %s", err)
			}

			if testcase.expected != actual {
				t.Errorf("expected %t, got %t", testcase.expected, actual)
			}
		})
	}
}

type mockForceNewDiffer struct {
	id        string
	old, new  string
	hasChange bool // force HasChange() to return true
	forceNew  bool
}

func (d *mockForceNewDiffer) Id() string {
	return d.id
}

func (d *mockForceNewDiffer) HasChange(key string) bool {
	return d.hasChange || d.old != d.new
}

func (d *mockForceNewDiffer) GetChange(key string) (interface{}, interface{}) {
	return d.old, d.new
}

func (d *mockForceNewDiffer) ForceNew(key string) error {
	d.forceNew = true

	return nil
}

func TestCustomizeDiffEngineVersionForceNewOnDowngrade(t *testing.T) {
	testcases := map[string]struct {
		isNew          bool
		old, new       string
		hasChange      bool // force HasChange() to return true
		expectForceNew bool
	}{
		"new resource": {
			isNew:          true,
			expectForceNew: false,
		},

		"no change": {
			old:            "1.2.3",
			new:            "1.2.3",
			expectForceNew: false,
		},

		"spurious change": {
			old:            "1.2.3",
			new:            "1.2.3",
			hasChange:      true,
			expectForceNew: false,
		},

		"upgrade minor versions": {
			old:            "1.2.3",
			new:            "1.3.5",
			expectForceNew: false,
		},

		"upgrade major versions": {
			old:            "1.2.3",
			new:            "2.4.6",
			expectForceNew: false,
		},

		// "upgrade major 6.x": {
		// 	old:            "5.0.6",
		// 	new:            "6.x",
		// 	expectForceNew: false,
		// },

		// "upgrade major 6.digit": {
		// 	old:            "5.0.6",
		// 	new:            "6.0",
		// 	expectForceNew: false,
		// },

		"downgrade minor versions": {
			old:            "1.3.5",
			new:            "1.2.3",
			expectForceNew: true,
		},

		"downgrade major versions": {
			old:            "2.4.6",
			new:            "1.2.3",
			expectForceNew: true,
		},

		"downgrade from major 6.x": {
			old:            "6.x",
			new:            "5.0.6",
			expectForceNew: true,
		},

		"downgrade major 6.digit": {
			old:            "6.2",
			new:            "6.0",
			expectForceNew: true,
		},

		"switch major 6.digit to 6.x": {
			old:            "6.2",
			new:            "6.x",
			expectForceNew: false,
		},

		"downgrade from major 7.x to 6.x": {
			old:            "7.x",
			new:            "6.x",
			expectForceNew: true,
		},

		"downgrade from major 7.digit to 6.x": {
			old:            "7.2",
			new:            "6.x",
			expectForceNew: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			diff := &mockForceNewDiffer{}
			if !testcase.isNew {
				diff.id = "some id"
				diff.old = testcase.old
				diff.new = testcase.new
			}
			diff.hasChange = testcase.hasChange

			err := engineVersionForceNewOnDowngrade(diff)

			if err != nil {
				t.Fatalf("no error expected, got %s", err)
			}

			if testcase.expectForceNew {
				if !diff.forceNew {
					t.Error("expected ForceNew")
				}
			} else {
				if diff.forceNew {
					t.Error("unexpected ForceNew")
				}
			}
		})
	}
}

func TestNormalizeEngineVersion(t *testing.T) {
	testcases := []struct {
		version    string
		normalized string
		valid      bool
	}{
		{
			version:    "5.4.3",
			normalized: "5.4.3",
			valid:      true,
		},
		{
			version:    "6.2",
			normalized: "6.2.0",
			valid:      true,
		},
		{
			version:    "6.x",
			normalized: fmt.Sprintf("6.%d.0", math.MaxInt),
			valid:      true,
		},
		{
			version: "5.x",
			valid:   false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.version, func(t *testing.T) {
			version, err := normalizeEngineVersion(testcase.version)

			if testcase.valid {
				if err != nil {
					t.Fatalf("expected no error, got %s", err)
				}
				if a, e := version.String(), testcase.normalized; a != e {
					t.Errorf("expected %q, got %q", e, a)
				}
			} else {
				if err == nil {
					t.Error("expected an error, got none")
				}
			}
		})
	}
}

func TestVersionDiff(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected versionDiff
	}{
		{"1.2.3", "1.2.3", versionDiff{0, 0, 0}},
		{"1.2.3", "1.1.7", versionDiff{0, 1, 0}},
		{"1.2.3", "1.4.5", versionDiff{0, -1, 0}},
		{"2.0.0", "1.2.3", versionDiff{1, 0, 0}},
		{"1.2.3", "2.0.0", versionDiff{-1, 0, 0}},
		{"1.2.3", "1.2.1", versionDiff{0, 0, 1}},
		{"1.2.3", "1.2.4", versionDiff{0, 0, -1}},
	}

	for _, tc := range cases {
		v1, err := version.NewVersion(tc.v1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		v2, err := version.NewVersion(tc.v2)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := diffVersion(v1, v2)
		expected := tc.expected
		if actual != expected {
			t.Fatalf(
				"%s <=> %s\nexpected: %d\nactual: %d",
				tc.v1, tc.v2,
				expected, actual)
		}
	}
}

type mockDiff struct {
	old, new  string
	hasChange bool // force HasChange() to return true
}

func (d mockDiff) HasChange() bool {
	return d.hasChange || d.old != d.new
}

func (d mockDiff) GetChange() (interface{}, interface{}) {
	return d.old, d.new
}

type mockChangesDiffer struct {
	id     string
	values map[string]mockDiff
}

func (d *mockChangesDiffer) Id() string {
	return d.id
}

func (d *mockChangesDiffer) HasChange(key string) bool {
	return d.values[key].HasChange()
}

func (d *mockChangesDiffer) GetChange(key string) (interface{}, interface{}) {
	return d.values[key].GetChange()
}

func TestParamGroupNameRequiresMajorVersionUpgrade(t *testing.T) {
	testcases := map[string]struct {
		isNew                  bool
		paramOld, paramNew     string
		paramHasChange         bool
		versionOld, versionNew string
		versionHasChange       bool
		expectError            *regexp.Regexp
	}{
		"new resource, no param group set": {
			isNew:    true,
			paramOld: "",
			paramNew: "",
		},

		"new resource, param group spurious diff": {
			isNew:          true,
			paramOld:       "",
			paramNew:       "",
			paramHasChange: true,
		},

		"new resource, set param group, no version set": {
			isNew:       true,
			paramOld:    "old",
			paramNew:    "",
			expectError: regexp.MustCompile(`cannot change parameter group name without upgrading major engine version`),
		},

		// new resource with version changes can only be verified at apply-time

		"update, no param group change": {
			paramOld: "no-change",
			paramNew: "no-change",
		},

		"update, param group spurious diff": {
			paramOld:       "no-change",
			paramNew:       "no-change",
			paramHasChange: true,
		},

		"update, param group change, no version change": {
			paramOld:    "old",
			paramNew:    "new",
			versionOld:  "6.0",
			versionNew:  "6.0",
			expectError: regexp.MustCompile(`cannot change parameter group name without upgrading major engine version`),
		},

		"update, param group change, version spurious diff": {
			paramOld:         "old",
			paramNew:         "new",
			versionOld:       "6.0",
			versionNew:       "6.0",
			versionHasChange: true,
			expectError:      regexp.MustCompile(`cannot change parameter group name without upgrading major engine version`),
		},

		"update, param group change, minor version change": {
			paramOld:    "old",
			paramNew:    "new",
			versionOld:  "6.0",
			versionNew:  "6.2",
			expectError: regexp.MustCompile(`cannot change parameter group name on minor engine version upgrade, upgrading from 6\.0\.[[:digit:]]+ to 6\.2\.[[:digit:]]+`),
		},

		"update, param group change, major version change": {
			paramOld:   "old",
			paramNew:   "new",
			versionOld: "5.0.6",
			versionNew: "6.2",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			diff := &mockChangesDiffer{
				values: map[string]mockDiff{
					"parameter_group_name": {
						old:       testcase.paramOld,
						new:       testcase.paramNew,
						hasChange: testcase.paramHasChange,
					},
					"engine_version": {
						old:       testcase.versionOld,
						new:       testcase.versionNew,
						hasChange: testcase.versionHasChange,
					},
				},
			}
			if !testcase.isNew {
				diff.id = "some id"
			}

			err := paramGroupNameRequiresMajorVersionUpgrade(diff)

			if testcase.expectError == nil {
				if err != nil {
					t.Fatalf("no error expected, got %s", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				if !testcase.expectError.MatchString(err.Error()) {
					t.Fatalf("unexpected error: %q", err.Error())
				}
			}

		})
	}
}
