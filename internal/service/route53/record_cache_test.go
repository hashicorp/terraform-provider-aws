// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func TestZoneRecordCacheKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		zoneName      string
		zoneID        string
		recordName    string
		rrType        string
		setIdentifier string
		want          string
	}{
		{
			name:       "basic record",
			zoneName:   "example.com.",
			zoneID:     "Z123456",
			recordName: "www.example.com",
			rrType:     "A",
			want:       "Z123456_www.example.com_A",
		},
		{
			name:       "trailing dot stripped",
			zoneName:   "example.com.",
			zoneID:     "Z123456",
			recordName: "www.example.com.",
			rrType:     "CNAME",
			want:       "Z123456_www.example.com_CNAME",
		},
		{
			name:       "uppercase normalized to lowercase",
			zoneName:   "example.com.",
			zoneID:     "Z123456",
			recordName: "WWW.EXAMPLE.COM",
			rrType:     "A",
			want:       "Z123456_www.example.com_A",
		},
		{
			// Configuration may name a record relative to its zone. The key must
			// match the one built from the fully qualified name the API returns,
			// or every lookup misses and falls back to a per-record API call.
			name:       "relative name expanded against zone",
			zoneName:   "example.com.",
			zoneID:     "Z123456",
			recordName: "www",
			rrType:     "A",
			want:       "Z123456_www.example.com_A",
		},
		{
			name:       "relative multi-label name expanded against zone",
			zoneName:   "example.com.",
			zoneID:     "Z123456",
			recordName: "a.b",
			rrType:     "A",
			want:       "Z123456_a.b.example.com_A",
		},
		{
			name:       "empty name resolves to zone apex",
			zoneName:   "example.com.",
			zoneID:     "Z123456",
			recordName: "",
			rrType:     "A",
			want:       "Z123456_example.com_A",
		},
		{
			// State stores "*" but the Route53 API returns "\052" (octal for '*').
			// Both must produce the same key so cache hits work for wildcard records.
			name:       "wildcard asterisk normalizes to octal",
			zoneName:   "example.com.",
			zoneID:     "Z123456",
			recordName: "*.example.com",
			rrType:     "A",
			want:       "Z123456_\\052.example.com_A",
		},
		{
			name:       "wildcard octal from API response matches",
			zoneName:   "example.com.",
			zoneID:     "Z123456",
			recordName: `\052.example.com.`,
			rrType:     "A",
			want:       "Z123456_\\052.example.com_A",
		},
		{
			name:       "relative wildcard expanded against zone",
			zoneName:   "example.com.",
			zoneID:     "Z123456",
			recordName: "*",
			rrType:     "A",
			want:       "Z123456_\\052.example.com_A",
		},
		{
			name:          "record with set_identifier",
			zoneName:      "example.com.",
			zoneID:        "Z123456",
			recordName:    "www.example.com",
			rrType:        "A",
			setIdentifier: "primary",
			want:          "Z123456_www.example.com_A_primary",
		},
		{
			name:          "empty set_identifier omitted from key",
			zoneName:      "example.com.",
			zoneID:        "Z123456",
			recordName:    "www.example.com",
			rrType:        "A",
			setIdentifier: "",
			want:          "Z123456_www.example.com_A",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &zoneRecordCache{zoneName: tc.zoneName}
			got := c.key(tc.zoneID, tc.recordName, tc.rrType, tc.setIdentifier)
			if got != tc.want {
				t.Errorf("key(%q, %q, %q, %q) = %q, want %q",
					tc.zoneID, tc.recordName, tc.rrType, tc.setIdentifier, got, tc.want)
			}
		})
	}
}

func TestZoneRecordCacheKeySymmetry(t *testing.T) {
	t.Parallel()

	// Each pair is (name as written in configuration, name as returned by the
	// API). Both must produce the same key or the record is never served from
	// the cache.
	cases := []struct {
		name      string
		fromState string
		fromAPI   string
	}{
		{
			name:      "wildcard",
			fromState: "*.example.com",
			fromAPI:   `\052.example.com.`,
		},
		{
			name:      "relative",
			fromState: "www",
			fromAPI:   "www.example.com.",
		},
		{
			name:      "relative wildcard",
			fromState: "*",
			fromAPI:   `\052.example.com.`,
		},
		{
			name:      "apex",
			fromState: "",
			fromAPI:   "example.com.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &zoneRecordCache{zoneName: "example.com."}
			state := c.key("Z123456", tc.fromState, "A", "")
			api := c.key("Z123456", tc.fromAPI, "A", "")
			if state != api {
				t.Errorf("key mismatch: state=%q api=%q", state, api)
			}
		})
	}
}

func TestZoneRecordCacheStoreAndLookup(t *testing.T) {
	t.Parallel()

	cache := &zoneRecordCache{
		records: make(map[string]awstypes.ResourceRecordSet),
	}
	key := "Z123456_www.example.com_A"
	rrs := awstypes.ResourceRecordSet{
		Name: aws.String("www.example.com."),
		Type: awstypes.RRTypeA,
	}

	if _, ok := cache.get(key); ok {
		t.Fatal("expected cache miss on empty cache")
	}

	cache.put(key, rrs)

	got, ok := cache.get(key)
	if !ok {
		t.Fatal("expected cache hit after store")
	}
	if aws.ToString(got.Name) != aws.ToString(rrs.Name) {
		t.Errorf("got Name %q, want %q", aws.ToString(got.Name), aws.ToString(rrs.Name))
	}
}

func TestZoneRecordCacheEvict(t *testing.T) {
	t.Parallel()

	zoneID := "Z_evict_test"
	cache, _ := recordCacheZones.LoadOrStore(zoneID, &zoneRecordCache{
		loaded:   true,
		zoneName: "example.com.",
		records:  make(map[string]awstypes.ResourceRecordSet),
	})

	// Stored under the fully qualified key the zone scan would produce, then
	// evicted using the relative name a configuration would supply.
	key := cache.key(zoneID, "www.example.com.", "A", "")
	cache.put(key, awstypes.ResourceRecordSet{
		Name: aws.String("www.example.com."),
		Type: awstypes.RRTypeA,
	})

	evictFromZoneRecordCache(zoneID, "www", "A", "")

	if _, ok := cache.get(key); ok {
		t.Fatal("expected cache miss after eviction")
	}
}

func TestZoneRecordCacheConcurrentReads(t *testing.T) {
	t.Parallel()

	cache := &zoneRecordCache{
		records: make(map[string]awstypes.ResourceRecordSet),
	}
	key := "Z123456_www.example.com_A"
	cache.put(key, awstypes.ResourceRecordSet{
		Name: aws.String("www.example.com."),
		Type: awstypes.RRTypeA,
	})

	// Multiple concurrent RLock reads must not block each other.
	const readers = 10
	doneCh := make(chan struct{}, readers)
	for range readers {
		go func() {
			cache.get(key)
			doneCh <- struct{}{}
		}()
	}

	for range readers {
		select {
		case <-doneCh:
		case <-time.After(50 * time.Millisecond):
			t.Fatal("concurrent reads blocked unexpectedly")
		}
	}
}
