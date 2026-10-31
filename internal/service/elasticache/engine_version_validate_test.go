// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/google/go-cmp/cmp"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
)

func TestEngineVersionMatches(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		engine        string
		configVersion string
		apiVersion    string
		expected      bool
	}{
		"redis exact minor match": {
			engine:        tfelasticache.EngineRedis,
			configVersion: "7.1",
			apiVersion:    "7.1",
			expected:      true,
		},
		"redis exact minor mismatch": {
			engine:        tfelasticache.EngineRedis,
			configVersion: "7.1",
			apiVersion:    "7.0",
			expected:      false,
		},
		"redis major wildcard matches minor": {
			engine:        tfelasticache.EngineRedis,
			configVersion: "6.x",
			apiVersion:    "6.2",
			expected:      true,
		},
		"redis major wildcard different major": {
			engine:        tfelasticache.EngineRedis,
			configVersion: "6.x",
			apiVersion:    "7.0",
			expected:      false,
		},
		"redis pre-v6 patch match": {
			engine:        tfelasticache.EngineRedis,
			configVersion: "5.0.6",
			apiVersion:    "5.0.6",
			expected:      true,
		},
		"valkey exact match": {
			engine:        tfelasticache.EngineValkey,
			configVersion: "7.2",
			apiVersion:    "7.2",
			expected:      true,
		},
		"valkey does not honor wildcard": {
			engine:        tfelasticache.EngineValkey,
			configVersion: "7.x",
			apiVersion:    "7.2",
			expected:      false,
		},
		"memcached patch match": {
			engine:        tfelasticache.EngineMemcached,
			configVersion: "1.6.22",
			apiVersion:    "1.6.22",
			expected:      true,
		},
		"empty api version": {
			engine:        tfelasticache.EngineRedis,
			configVersion: "7.1",
			apiVersion:    "",
			expected:      false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tfelasticache.EngineVersionMatches(testcase.engine, testcase.configVersion, testcase.apiVersion); got != testcase.expected {
				t.Errorf("EngineVersionMatches(%q, %q, %q) = %t, want %t", testcase.engine, testcase.configVersion, testcase.apiVersion, got, testcase.expected)
			}
		})
	}
}

func TestFindAvailableCacheEngineVersion(t *testing.T) {
	t.Parallel()

	available := []awstypes.CacheEngineVersion{
		{EngineVersion: aws.String("6.0"), CacheParameterGroupFamily: aws.String("redis6.x")},
		{EngineVersion: aws.String("6.2"), CacheParameterGroupFamily: aws.String("redis6.x")},
		{EngineVersion: aws.String("7.0"), CacheParameterGroupFamily: aws.String("redis7")},
		{EngineVersion: aws.String("7.1"), CacheParameterGroupFamily: aws.String("redis7")},
	}

	testcases := map[string]struct {
		engine        string
		configVersion string
		wantFound     bool
		wantFamily    string
	}{
		"available exact": {
			engine:        tfelasticache.EngineRedis,
			configVersion: "7.0",
			wantFound:     true,
			wantFamily:    "redis7",
		},
		"available wildcard": {
			engine:        tfelasticache.EngineRedis,
			configVersion: "6.x",
			wantFound:     true,
			wantFamily:    "redis6.x",
		},
		"unavailable": {
			engine:        tfelasticache.EngineRedis,
			configVersion: "7.2",
			wantFound:     false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			matched, found := tfelasticache.FindAvailableCacheEngineVersion(testcase.engine, testcase.configVersion, available)
			if found != testcase.wantFound {
				t.Fatalf("FindAvailableCacheEngineVersion found = %t, want %t", found, testcase.wantFound)
			}
			if found {
				if got := aws.ToString(matched.CacheParameterGroupFamily); got != testcase.wantFamily {
					t.Errorf("matched family = %q, want %q", got, testcase.wantFamily)
				}
			}
		})
	}
}

func TestDefaultCacheParameterGroupFamily(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		name       string
		wantFamily string
		wantOK     bool
	}{
		"default redis 6.x": {
			name:       "default.redis6.x",
			wantFamily: "redis6.x",
			wantOK:     true,
		},
		"default redis 7": {
			name:       "default.redis7",
			wantFamily: "redis7",
			wantOK:     true,
		},
		"default redis 7 cluster mode": {
			name:       "default.redis7.cluster.on",
			wantFamily: "redis7",
			wantOK:     true,
		},
		"default valkey 8": {
			name:       "default.valkey8",
			wantFamily: "valkey8",
			wantOK:     true,
		},
		"default memcached": {
			name:       "default.memcached1.6",
			wantFamily: "memcached1.6",
			wantOK:     true,
		},
		"custom group": {
			name:   "my-custom-group",
			wantOK: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			family, ok := tfelasticache.DefaultCacheParameterGroupFamily(testcase.name)
			if ok != testcase.wantOK {
				t.Fatalf("DefaultCacheParameterGroupFamily(%q) ok = %t, want %t", testcase.name, ok, testcase.wantOK)
			}
			if ok && family != testcase.wantFamily {
				t.Errorf("DefaultCacheParameterGroupFamily(%q) = %q, want %q", testcase.name, family, testcase.wantFamily)
			}
		})
	}
}

func TestAvailableEngineVersionStrings(t *testing.T) {
	t.Parallel()

	available := []awstypes.CacheEngineVersion{
		{EngineVersion: aws.String("7.1")},
		{EngineVersion: aws.String("6.0")},
		{EngineVersion: aws.String("7.1")}, // duplicate
		{EngineVersion: aws.String("7.0")},
		{EngineVersion: aws.String("")}, // skipped
	}

	got := tfelasticache.AvailableEngineVersionStrings(available)
	want := []string{"6.0", "7.0", "7.1"}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected result (-want +got):\n%s", diff)
	}
}
