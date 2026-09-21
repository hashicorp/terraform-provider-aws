// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"log"
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
// keyed by the record set resource ID.
type zoneRecordCache struct {
	once    sync.Once
	loadErr error
	mu      sync.RWMutex
	records map[string]awstypes.ResourceRecordSet
}

// recordCacheZones maps cleaned hosted zone ID to its populated record cache.
var recordCacheZones tfsync.Map[string, *zoneRecordCache]

// getOrLoadZoneRecordCache returns the existing zone record cache or loads it.
//
// If the cache is not yet loaded, a full ListResourceRecordSets scan is
// performed to populate the cache for future callers.
//
// The full scan, wrapped by once.Do, will execute only for the first caller
// to reach it. Subsequent callers will block until the function has completed,
// ensuring this zone's records are loaded before the outer function returns.
//
// A failure during loading will result in an empty or incomplete cache and a
// non-nil loadErr.
func getOrLoadZoneRecordCache(ctx context.Context, conn *route53.Client, zoneID string) (*zoneRecordCache, error) {
	v, _ := recordCacheZones.LoadOrStore(zoneID, &zoneRecordCache{records: make(map[string]awstypes.ResourceRecordSet)})

	v.once.Do(func() {
		v.load(ctx, conn, zoneID)
	})
	return v, v.loadErr
}

// load fetches all record sets in a hosted zone.
//
// This should only be called within zoneRecordCache.once.Do to ensure
// the List operation is performed once per Terraform operation.
func (c *zoneRecordCache) load(ctx context.Context, conn *route53.Client, zoneID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}
	for rrs, err := range listRecords(ctx, conn, input) {
		if err != nil {
			c.loadErr = err
			return
		}
		key := recordCacheKey(zoneID, aws.ToString(rrs.Name), string(rrs.Type), aws.ToString(rrs.SetIdentifier))
		c.records[key] = rrs
	}
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
func (c *zoneRecordCache) evict(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.records, key)
}

// readFromZoneRecordCache looks up a record in the zone cache.
//
// A cache miss falls back to a direct API call. The miss result is stored in
// the cache once retrieved.
func readFromZoneRecordCache(ctx context.Context, conn *route53.Client, zoneID, name, rrType, setID string) (*awstypes.ResourceRecordSet, error) {
	cache, err := getOrLoadZoneRecordCache(ctx, conn, zoneID)
	if err != nil {
		return nil, err
	}

	key := recordCacheKey(zoneID, name, rrType, setID)
	rrs, ok := cache.get(key)
	if !ok {
		record, err := findResourceRecordSetByFourPartKey(ctx, conn, zoneID, name, rrType, setID)
		if err != nil {
			return nil, err
		}
		cache.put(key, *record)
		return record, nil
	}

	return &rrs, nil
}

// evictFromZoneRecordCache removes a record in the zone cache.
//
// A failure to load the zone cache results in a no-op.
func evictFromZoneRecordCache(ctx context.Context, conn *route53.Client, zoneID, key string) {
	cache, err := getOrLoadZoneRecordCache(ctx, conn, zoneID)
	if err != nil {
		log.Printf("[WARN] Failed to load zone record cache for \"%s\". Key \"%s\" not evicted.", zoneID, key)
		return
	}

	cache.evict(key)
}

// recordCacheKey returns the cache map key for a record within a zone.
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
