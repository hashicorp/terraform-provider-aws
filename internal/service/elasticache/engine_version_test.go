package elasticache

import (
	"fmt"
	"math"
	"testing"
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

type differ struct {
	id        string
	old, new  string
	hasChange bool // force HasChange() to return true
	forceNew  bool
}

func (d *differ) Id() string {
	return d.id
}

func (d *differ) HasChange(key string) bool {
	return d.hasChange || d.old != d.new
}

func (d *differ) GetChange(key string) (interface{}, interface{}) {
	return d.old, d.new
}

func (d *differ) ForceNew(key string) error {
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
			diff := &differ{}
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
