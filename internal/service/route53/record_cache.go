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

const recordBatchReadEnvVar = "TF_AWS_ROUTE53_RECORD_BATCH_READS"

// batchReadsEnabled reports whether TF_AWS_ROUTE53_RECORD_BATCH_READS is set to
// a true value.
//
// When enabled, reads are served from the zone-level cache and cache evictions
// are performed on write operations.
func batchReadsEnabled() bool {
	v, _ := strconv.ParseBool(os.Getenv(recordBatchReadEnvVar))
	return v
}

// zoneRecordCache holds all cached ResourceRecordSets for a single hosted zone,
// keyed by zoneRecordCache.key.
type zoneRecordCache struct {
	mu       sync.RWMutex
	loaded   bool
	zoneName string
	records  map[string]awstypes.ResourceRecordSet
}

// recordCacheZones maps cleaned hosted zone ID to its populated record cache.
var recordCacheZones tfsync.Map[string, *zoneRecordCache]

// getOrLoadZoneRecordCache returns the record cache for zoneID, performing a
// full ListResourceRecordSets scan if the zone has not been loaded yet.
//
// Concurrent callers for the same zone serialize on that zone's mutex, so only
// the first pays for the scan. Scans of different zones proceed in parallel. A
// failed scan is not cached: partial results are discarded and the next caller
// retries.
//
// The scan runs on the calling goroutine's context, so a caller whose context
// is cancelled mid-scan surfaces that cancellation to the callers blocked
// behind it.
func getOrLoadZoneRecordCache(ctx context.Context, conn *route53.Client, zoneID string) (*zoneRecordCache, error) {
	c, _ := recordCacheZones.LoadOrStore(zoneID, &zoneRecordCache{records: make(map[string]awstypes.ResourceRecordSet)})

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loaded {
		return c, nil
	}

	if err := c.load(ctx, conn, zoneID); err != nil {
		clear(c.records)
		return nil, err
	}
	c.loaded = true

	return c, nil
}

// load fetches all record sets in a hosted zone into the cache.
//
// The zone name is resolved first because keys are built from fully qualified
// record names, and configuration may name a record relative to its zone.
//
// The caller must hold c.mu for writing.
func (c *zoneRecordCache) load(ctx context.Context, conn *route53.Client, zoneID string) error {
	zone, err := findHostedZoneByID(ctx, conn, zoneID)
	if err != nil {
		return err
	}
	c.zoneName = aws.ToString(zone.HostedZone.Name)

	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}
	for rrs, err := range listRecords(ctx, conn, input) {
		if err != nil {
			return err
		}
		key := c.key(zoneID, aws.ToString(rrs.Name), string(rrs.Type), aws.ToString(rrs.SetIdentifier))
		c.records[key] = rrs
	}

	return nil
}

// key returns the cache map key for a record within this zone.
//
// The name is expanded against the zone name so that a record configured
// relative to its zone ("www") produces the same key as the fully qualified
// name the API returns ("www.example.com."). expandRecordName is a no-op for
// names that already carry the suffix, so the same call serves both sides.
func (c *zoneRecordCache) key(zoneID, name, rrType, setIdentifier string) string {
	parts := []string{
		zoneID,
		expandRecordName(name, c.zoneName),
		rrType,
	}
	if setIdentifier != "" {
		parts = append(parts, setIdentifier)
	}
	return strings.Join(parts, "_")
}

// get retrieves a single record from the zone cache.
func (c *zoneRecordCache) get(key string) (awstypes.ResourceRecordSet, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.records[key]
	return v, ok
}

// put inserts or replaces a single record in the zone cache.
func (c *zoneRecordCache) put(key string, rrs awstypes.ResourceRecordSet) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records[key] = rrs
}

// evict removes a single record from the zone cache.
func (c *zoneRecordCache) evict(zoneID, name, rrType, setIdentifier string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.records, c.key(zoneID, name, rrType, setIdentifier))
}

// readFromZoneRecordCache looks up a record in the zone cache.
//
// A cache miss falls back to a direct API call. The miss result is stored in
// the cache once retrieved.
func readFromZoneRecordCache(ctx context.Context, conn *route53.Client, zoneID, name, rrType, setIdentifier string) (*awstypes.ResourceRecordSet, error) {
	cache, err := getOrLoadZoneRecordCache(ctx, conn, zoneID)
	if err != nil {
		return nil, err
	}

	key := cache.key(zoneID, name, rrType, setIdentifier)
	rrs, ok := cache.get(key)
	if !ok {
		record, err := findResourceRecordSetByFourPartKey(ctx, conn, zoneID, name, rrType, setIdentifier)
		if err != nil {
			return nil, err
		}
		cache.put(key, *record)
		return record, nil
	}

	return &rrs, nil
}

// evictFromZoneRecordCache removes a record from the zone cache.
//
// An uncached zone is a no-op: any later scan starts after this write and
// already reflects it. Otherwise evict blocks on the zone mutex until an
// in-flight scan completes.
func evictFromZoneRecordCache(zoneID, name, rrType, setIdentifier string) {
	if c, ok := recordCacheZones.Load(zoneID); ok {
		c.evict(zoneID, name, rrType, setIdentifier)
	}
}
