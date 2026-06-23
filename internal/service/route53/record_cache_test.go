// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func TestRecordCacheKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		zoneID        string
		recordName    string
		rrType        string
		setIdentifier string
		want          string
	}{
		{
			name:       "basic record",
			zoneID:     "Z123456",
			recordName: "www.example.com",
			rrType:     "A",
			want:       "Z123456_www.example.com_A",
		},
		{
			name:       "trailing dot stripped",
			zoneID:     "Z123456",
			recordName: "www.example.com.",
			rrType:     "CNAME",
			want:       "Z123456_www.example.com_CNAME",
		},
		{
			name:       "uppercase normalized to lowercase",
			zoneID:     "Z123456",
			recordName: "WWW.EXAMPLE.COM",
			rrType:     "A",
			want:       "Z123456_www.example.com_A",
		},
		{
			// State stores "*" but the Route53 API returns "\052" (octal for '*').
			// Both must produce the same key so cache hits work for wildcard records.
			name:       "wildcard asterisk normalizes to octal",
			zoneID:     "Z123456",
			recordName: "*.example.com",
			rrType:     "A",
			want:       "Z123456_\\052.example.com_A",
		},
		{
			name:       "wildcard octal from API response matches",
			zoneID:     "Z123456",
			recordName: `\052.example.com.`,
			rrType:     "A",
			want:       "Z123456_\\052.example.com_A",
		},
		{
			name:          "record with set_identifier",
			zoneID:        "Z123456",
			recordName:    "www.example.com",
			rrType:        "A",
			setIdentifier: "primary",
			want:          "Z123456_www.example.com_A_primary",
		},
		{
			name:          "empty set_identifier omitted from key",
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
			got := recordCacheKey(tc.zoneID, tc.recordName, tc.rrType, tc.setIdentifier)
			if got != tc.want {
				t.Errorf("recordCacheKey(%q, %q, %q, %q) = %q, want %q",
					tc.zoneID, tc.recordName, tc.rrType, tc.setIdentifier, got, tc.want)
			}
		})
	}
}

func TestRecordCacheKeyWildcardSymmetry(t *testing.T) {
	t.Parallel()

	// The key built from state ("*") and from the API ("\052") must be identical.
	fromState := recordCacheKey("Z123456", "*.example.com", "A", "")
	fromAPI := recordCacheKey("Z123456", `\052.example.com.`, "A", "")
	if fromState != fromAPI {
		t.Errorf("wildcard key mismatch: state=%q api=%q", fromState, fromAPI)
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

	if _, ok := lookupInZoneRecordCache(cache, key); ok {
		t.Fatal("expected cache miss on empty cache")
	}

	storeInZoneRecordCache(cache, key, rrs)

	got, ok := lookupInZoneRecordCache(cache, key)
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
	key := zoneID + "_www.example.com_A"
	cache := &zoneRecordCache{
		records: make(map[string]awstypes.ResourceRecordSet),
	}
	storeInZoneRecordCache(cache, key, awstypes.ResourceRecordSet{
		Name: aws.String("www.example.com."),
		Type: awstypes.RRTypeA,
	})

	// Register the cache in the global map so evictFromZoneRecordCache can find it.
	recordCacheZones.LoadOrStore(zoneID, cache)

	evictFromZoneRecordCache(zoneID, key)

	if _, ok := lookupInZoneRecordCache(cache, key); ok {
		t.Fatal("expected cache miss after eviction")
	}
}

func TestZoneRecordCacheConcurrentReads(t *testing.T) {
	t.Parallel()

	cache := &zoneRecordCache{
		records: make(map[string]awstypes.ResourceRecordSet),
	}
	key := "Z123456_www.example.com_A"
	storeInZoneRecordCache(cache, key, awstypes.ResourceRecordSet{
		Name: aws.String("www.example.com."),
		Type: awstypes.RRTypeA,
	})

	// Multiple concurrent RLock reads must not block each other.
	const readers = 10
	doneCh := make(chan struct{}, readers)
	for range readers {
		go func() {
			lookupInZoneRecordCache(cache, key)
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
