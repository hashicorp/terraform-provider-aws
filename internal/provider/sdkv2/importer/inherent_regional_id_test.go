// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package importer_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/internal/attribute"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var regionalInherentRegionSchema = map[string]*schema.Schema{
	"url": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"attr": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"region": attribute.Region(),
}

// lintignore:S013 // Identity Schemas cannot specify Computed, Optional, or Required
var regionalInherentRegionIdentitySchema = map[string]*schema.Schema{
	"url": {
		Type:              schema.TypeString,
		RequiredForImport: true,
	},
}

func parseID(value string) (inttypes.BaseIdentity, error) {
	re := regexache.MustCompile(`^https://a-service\.([a-z0-9-]+)\.[^/]+/([0-9]{12})/.+`)
	match := re.FindStringSubmatch(value)
	if match == nil {
		return inttypes.BaseIdentity{}, fmt.Errorf("could not parse import ID %q as SQS URL", value)
	}
	return inttypes.BaseIdentity{
		AccountID: match[2],
		Region:    match[1],
	}, nil
}

func createID(region, accountID, resource string) string {
	return fmt.Sprintf("https://a-service.%s.amazonaws.com/%s/%s", region, accountID, resource)
}

func TestRegionalInherentRegion_ImportID_Invalid(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, regionalInherentRegionSchema, map[string]any{})
	id := "invalid"
	rd.SetId(id)

	identity := inttypes.RegionalCustomInherentRegionIdentity("url", parseID,
		inttypes.WithIdentityDuplicateAttrs("id"),
	)

	err := importer.RegionalInherentRegion(context.Background(), rd, identity)
	if err != nil {
		if !strings.HasPrefix(err.Error(), fmt.Sprintf("parsing import ID %q", id)) {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalInherentRegion_ImportID_Invalid_WrongRegion(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, regionalInherentRegionSchema, map[string]any{
		"region": "another-region-1",
	})
	region := "a-region-1"
	accountID := "123456789012"
	id := createID(region, accountID, "res-abc123")
	rd.SetId(id)

	identity := inttypes.RegionalCustomInherentRegionIdentity("url", parseID,
		inttypes.WithIdentityDuplicateAttrs("id"),
	)

	err := importer.RegionalInherentRegion(context.Background(), rd, identity)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "the region passed for import") {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalInherentRegion_ImportID_Valid_DefaultRegion(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, regionalInherentRegionSchema, map[string]any{})
	region := "a-region-1"
	accountID := "123456789012"
	id := createID(region, accountID, "res-abc123")
	rd.SetId(id)

	identity := inttypes.RegionalCustomInherentRegionIdentity("url", parseID,
		inttypes.WithIdentityDuplicateAttrs("id"),
	)

	err := importer.RegionalInherentRegion(context.Background(), rd, identity)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := id, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := id, rd.Get("url"); e != a {
		t.Errorf("expected `url` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
	if e, a := "", rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestRegionalInherentRegion_ImportID_Valid_RegionOverride(t *testing.T) {
	t.Parallel()

	region := "a-region-1"
	rd := schema.TestResourceDataRaw(t, regionalInherentRegionSchema, map[string]any{
		"region": region,
	})
	accountID := "123456789012"
	id := createID(region, accountID, "res-abc123")
	rd.SetId(id)

	identity := inttypes.RegionalCustomInherentRegionIdentity("url", parseID,
		inttypes.WithIdentityDuplicateAttrs("id"),
	)

	err := importer.RegionalInherentRegion(context.Background(), rd, identity)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := id, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := id, rd.Get("url"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
	if e, a := "", rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestRegionalInherentRegion_Identity_Invalid_AttributeNotSet(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataWithIdentityRaw(t, regionalInherentRegionSchema, regionalInherentRegionIdentitySchema, map[string]string{})

	identity := inttypes.RegionalCustomInherentRegionIdentity("url", parseID,
		inttypes.WithIdentityDuplicateAttrs("id"),
	)

	err := importer.RegionalInherentRegion(context.Background(), rd, identity)
	if err != nil {
		if err.Error() != fmt.Sprintf("identity attribute %q is required", "url") {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalInherentRegion_Identity_Invalid(t *testing.T) {
	t.Parallel()

	id := "invalid"
	rd := schema.TestResourceDataWithIdentityRaw(t, regionalInherentRegionSchema, regionalInherentRegionIdentitySchema, map[string]string{
		"url": id,
	})

	identity := inttypes.RegionalCustomInherentRegionIdentity("url", parseID,
		inttypes.WithIdentityDuplicateAttrs("id"),
	)

	err := importer.RegionalInherentRegion(context.Background(), rd, identity)
	if err != nil {
		if !strings.HasPrefix(err.Error(), fmt.Sprintf("identity attribute %q: parsing %q: ", "url", id)) {
			t.Fatalf("Unexpected error: %s", err)
		}
	} else {
		t.Fatal("Expected error, got none")
	}
}

func TestRegionalInherentRegion_Identity_Valid(t *testing.T) {
	t.Parallel()

	region := "a-region-1"
	accountID := "123456789012"
	id := createID(region, accountID, "res-abc123")
	rd := schema.TestResourceDataWithIdentityRaw(t, regionalInherentRegionSchema, regionalInherentRegionIdentitySchema, map[string]string{
		"url": id,
	})

	identity := inttypes.RegionalCustomInherentRegionIdentity("url", parseID,
		inttypes.WithIdentityDuplicateAttrs("id"),
	)

	err := importer.RegionalInherentRegion(context.Background(), rd, identity)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := id, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := id, rd.Get("url"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
	if e, a := "", rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestRegionalInherentRegion_DuplicateAttrs_ImportID_Valid(t *testing.T) {
	t.Parallel()

	rd := schema.TestResourceDataRaw(t, regionalInherentRegionSchema, map[string]any{})
	region := "a-region-1"
	accountID := "123456789012"
	id := createID(region, accountID, "res-abc123")
	rd.SetId(id)

	identity := inttypes.RegionalCustomInherentRegionIdentity("url", parseID,
		inttypes.WithIdentityDuplicateAttrs("id", "attr"),
	)

	err := importer.RegionalInherentRegion(context.Background(), rd, identity)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := id, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := id, rd.Get("url"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
	if e, a := id, rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}

func TestRegionalInherentRegion_DuplicateAttrs_Identity_Valid(t *testing.T) {
	t.Parallel()

	region := "a-region-1"
	accountID := "123456789012"
	id := createID(region, accountID, "res-abc123")
	rd := schema.TestResourceDataWithIdentityRaw(t, regionalInherentRegionSchema, regionalInherentRegionIdentitySchema, map[string]string{
		"url": id,
	})

	identity := inttypes.RegionalCustomInherentRegionIdentity("url", parseID,
		inttypes.WithIdentityDuplicateAttrs("id", "attr"),
	)

	err := importer.RegionalInherentRegion(context.Background(), rd, identity)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if e, a := id, rd.Id(); e != a {
		t.Errorf("expected `id` to be %q, got %q", e, a)
	}
	if e, a := id, rd.Get("url"); e != a {
		t.Errorf("expected `arn` to be %q, got %q", e, a)
	}
	if e, a := region, rd.Get("region"); e != a {
		t.Errorf("expected `region` to be %q, got %q", e, a)
	}
	if e, a := id, rd.Get("attr"); e != a {
		t.Errorf("expected `attr` to be %q, got %q", e, a)
	}
}
