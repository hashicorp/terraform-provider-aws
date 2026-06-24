// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/sync"
)

// batchReadsEnabled reports whether TF_AWS_ROUTE53_RECORD_BATCH_READS is set to
// a true value ("1", "true", "t", etc.). When enabled, reads are served from the
// zone-level cache and cache evictions are performed on write operations.
func batchReadsEnabled() bool {
	v, _ := strconv.ParseBool(os.Getenv("TF_AWS_ROUTE53_RECORD_BATCH_READS"))
	return v
}

// zoneRecordCache holds all cached ResourceRecordSets for a single hosted zone,
// keyed by the Terraform-style resource ID.
type zoneRecordCache struct {
	mu      sync.RWMutex
	records map[string]awstypes.ResourceRecordSet
}

// recordCacheZones maps cleaned hosted zone ID to its populated record cache.
var recordCacheZones tfsync.Map[string, *zoneRecordCache]

// recordCacheInitMu serializes the initial zone-wide listing per zone ID so that
// only one goroutine pays the cost of the full ListResourceRecordSets scan.
var recordCacheInitMu tfsync.Map[string, *sync.Mutex]

// recordCacheKey returns the cache map key for a record within a zone.
// The name is normalized via normalizeDomainName to consistently handle octal
// escape sequences (e.g. \052 for *), trailing dots, and casing — matching
// how the Route53 API and expandRecordName represent names internally.
func recordCacheKey(zoneID, name, rrType, setIdentifier string) string {
	parts := []string{
		zoneID,
		normalizeDomainName(name),
		rrType,
	}
	if setIdentifier != "" {
		parts = append(parts, setIdentifier)
	}
	return strings.Join(parts, "_")
}

// getOrLoadZoneRecordCache returns the record cache for zoneID, issuing a full
// ListResourceRecordSets scan if the zone has not been loaded yet. A per-zone
// mutex prevents duplicate scans under concurrent access.
func getOrLoadZoneRecordCache(ctx context.Context, conn *route53.Client, zoneID string) (*zoneRecordCache, error) {
	if v, ok := recordCacheZones.Load(zoneID); ok {
		return v, nil
	}

	mu, _ := recordCacheInitMu.LoadOrStore(zoneID, &sync.Mutex{})
	mu.Lock()
	defer mu.Unlock()

	// Prevent race conditions triggering multiple ListResourceRecordSets
	if v, ok := recordCacheZones.Load(zoneID); ok {
		return v, nil
	}

	cache := &zoneRecordCache{
		records: make(map[string]awstypes.ResourceRecordSet),
	}

	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}
	for rrs, err := range listRecords(ctx, conn, input) {
		if err != nil {
			return nil, err
		}
		key := recordCacheKey(zoneID, aws.ToString(rrs.Name), string(rrs.Type), aws.ToString(rrs.SetIdentifier))
		cache.records[key] = rrs
	}

	recordCacheZones.LoadOrStore(zoneID, cache)
	return cache, nil
}

// lookupInZoneRecordCache retrieves a single record from the zone cache.
func lookupInZoneRecordCache(cache *zoneRecordCache, key string) (awstypes.ResourceRecordSet, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	v, ok := cache.records[key]
	return v, ok
}

// storeInZoneRecordCache inserts or replaces a single record in the zone cache.
func storeInZoneRecordCache(cache *zoneRecordCache, key string, rrs awstypes.ResourceRecordSet) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.records[key] = rrs
}

// evictFromZoneRecordCache removes a single record from the zone cache. It is a
// no-op when the zone has not been cached or the key is not present.
// Is used during record update and delete operations to remove cache entries
// to ensure subsequent read operations call the AWS API instead of hittins stale cache
func evictFromZoneRecordCache(zoneID, key string) {
	if c, ok := recordCacheZones.Load(zoneID); ok {
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.records, key)
	}
}

// readRecordFromCache looks up a record in the zone cache, falling back to a
// direct API call on a cache miss. A miss result is stored back into the cache
// fallback behavior is to ensure resource data is refreshed from the API after a record is updated or deleted
func readRecordFromCache(ctx context.Context, conn *route53.Client, zoneID, name, rrType, setID string) (*awstypes.ResourceRecordSet, *string, error) {
	cache, err := getOrLoadZoneRecordCache(ctx, conn, zoneID)
	if err != nil {
		return nil, nil, err
	}

	key := recordCacheKey(zoneID, name, rrType, setID)
	rrs, ok := lookupInZoneRecordCache(cache, key)
	if !ok {
		record, fqdn, err := findResourceRecordSetByFourPartKey(ctx, conn, zoneID, name, rrType, setID)
		if err != nil {
			return nil, nil, err
		}
		storeInZoneRecordCache(cache, key, *record)
		return record, fqdn, nil
	}

	// Derive the FQDN string in the same form returned by findResourceRecordSetByFourPartKey:
	// normalized (no trailing dot, octal escape codes preserved).
	fqdnStr := normalizeDomainName(aws.ToString(rrs.Name))
	return &rrs, &fqdnStr, nil
}
