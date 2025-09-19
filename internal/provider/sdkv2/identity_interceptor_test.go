// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/internal/attribute"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestIdentityInterceptor(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "us-west-2" //lintignore:AWSAT003
	name := "a_name"

	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"type": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"region": attribute.Region(),
	}

	client := mockClient{
		accountID: accountID,
		region:    region,
	}

	testCases := map[string]struct {
		attrName     string
		identitySpec inttypes.Identity
	}{
		"same names": {
			attrName:     "name",
			identitySpec: regionalSingleParameterizedIdentitySpec("name"),
		},
		"name mapped": {
			attrName:     "resource_name",
			identitySpec: regionalSingleParameterizedIdentitySpecNameMapped("resource_name", "name"),
		},
	}

	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			invocation := newIdentityInterceptor(&tc.identitySpec)
			interceptor := invocation.interceptor.(identityInterceptor)

			identitySchema := identity.NewIdentitySchema(tc.identitySpec)

			d := schema.TestResourceDataWithIdentityRaw(t, resourceSchema, identitySchema, nil)
			d.SetId("some_id")
			d.Set("name", name)
			d.Set("region", region)
			d.Set("type", "some_type")

			opts := crudInterceptorOptions{
				c:    client,
				d:    d,
				when: After,
				why:  Create,
			}

			interceptor.run(ctx, opts)

			identity, err := d.Identity()
			if err != nil {
				t.Fatalf("unexpected error getting identity: %v", err)
			}

			if e, a := accountID, identity.Get(names.AttrAccountID); e != a {
				t.Errorf("expected account ID %q, got %q", e, a)
			}
			if e, a := region, identity.Get(names.AttrRegion); e != a {
				t.Errorf("expected region %q, got %q", e, a)
			}
			if e, a := name, identity.Get(tc.attrName); e != a {
				t.Errorf("expected %s %q, got %q", tc.attrName, e, a)
			}
		})
	}
}

func TestIdentityInterceptor_Read_Removed(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "us-west-2" //lintignore:AWSAT003
	name := "a_name"

	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"type": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"region": attribute.Region(),
	}

	identitySpec := regionalSingleParameterizedIdentitySpec("name")
	identitySchema := identity.NewIdentitySchema(identitySpec)

	invocation := newIdentityInterceptor(&identitySpec)
	interceptor := invocation.interceptor.(identityInterceptor)

	client := mockClient{
		accountID: accountID,
		region:    region,
	}

	ctx := t.Context()

	d := schema.TestResourceDataWithIdentityRaw(t, resourceSchema, identitySchema, nil)
	d.SetId("")
	d.Set("name", name)
	d.Set("region", region)
	d.Set("type", "some_type")

	opts := crudInterceptorOptions{
		c:    client,
		d:    d,
		when: After,
		why:  Read,
	}

	interceptor.run(ctx, opts)

	identity, err := d.Identity()
	if err != nil {
		t.Fatalf("unexpected error getting identity: %v", err)
	}

	if identity.Get(names.AttrAccountID) != "" {
		t.Errorf("expected no account ID, got %q", identity.Get(names.AttrAccountID))
	}
	if identity.Get(names.AttrRegion) != "" {
		t.Errorf("expected no region, got %q", identity.Get(names.AttrRegion))
	}
	if identity.Get("name") != "" {
		t.Errorf("expected no name, got %q", identity.Get("name"))
	}
}

func TestIdentityInterceptor_Update(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "us-west-2" //lintignore:AWSAT003
	name := "a_name"

	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"type": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"region": attribute.Region(),
	}

	client := mockClient{
		accountID: accountID,
		region:    region,
	}

	testCases := map[string]struct {
		attrName       string
		identitySpec   inttypes.Identity
		ExpectIdentity bool
		Description    string
	}{
		"not mutable - fresh resource": {
			attrName:       "name",
			identitySpec:   regionalSingleParameterizedIdentitySpec("name"),
			ExpectIdentity: true, // NOW EXPECTS IDENTITY because all attributes are null (bug scenario)
			Description:    "Immutable identity with all null attributes should get populated (bug fix scenario)",
		},
		"v6.0 SDK fix": {
			attrName: "name",
			identitySpec: regionalSingleParameterizedIdentitySpec("name",
				inttypes.WithV6_0SDKv2Fix(),
			),
			ExpectIdentity: true, // This makes identity mutable, so always expect it
			Description:    "Mutable identity (v6.0 SDK fix) should always get populated on Update",
		},
		"identity fix": {
			attrName: "name",
			identitySpec: regionalSingleParameterizedIdentitySpec("name",
				inttypes.WithIdentityFix(),
			),
			ExpectIdentity: true, // This makes identity mutable, so always expect it
			Description:    "Mutable identity (identity fix) should always get populated on Update",
		},
		"mutable": {
			attrName: "name",
			identitySpec: regionalSingleParameterizedIdentitySpec("name",
				inttypes.WithMutableIdentity(),
			),
			ExpectIdentity: true,
			Description:    "Explicitly mutable identity should always get populated on Update",
		},
	}

	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			invocation := newIdentityInterceptor(&tc.identitySpec)
			interceptor := invocation.interceptor.(identityInterceptor)

			identitySchema := identity.NewIdentitySchema(tc.identitySpec)

			d := schema.TestResourceDataWithIdentityRaw(t, resourceSchema, identitySchema, nil)
			d.SetId("some_id")
			d.Set("name", name)
			d.Set("region", region)
			d.Set("type", "some_type")

			opts := crudInterceptorOptions{
				c:    client,
				d:    d,
				when: After,
				why:  Update,
			}

			interceptor.run(ctx, opts)

			identity, err := d.Identity()
			if err != nil {
				t.Fatalf("unexpected error getting identity: %v", err)
			}

			if tc.ExpectIdentity {
				if e, a := accountID, identity.Get(names.AttrAccountID); e != a {
					t.Errorf("expected account ID %q, got %q", e, a)
				}
				if e, a := region, identity.Get(names.AttrRegion); e != a {
					t.Errorf("expected region %q, got %q", e, a)
				}
				if e, a := name, identity.Get(tc.attrName); e != a {
					t.Errorf("expected %s %q, got %q", tc.attrName, e, a)
				}
			} else {
				if identity.Get(names.AttrAccountID) != "" {
					t.Errorf("expected no account ID, got %q", identity.Get(names.AttrAccountID))
				}
				if identity.Get(names.AttrRegion) != "" {
					t.Errorf("expected no region, got %q", identity.Get(names.AttrRegion))
				}
				if identity.Get(tc.attrName) != "" {
					t.Errorf("expected no %s, got %q", tc.attrName, identity.Get(tc.attrName))
				}
			}
		})
	}
}

func TestIdentityInterceptor_Update_PartialNullValues(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "us-west-2" //lintignore:AWSAT003
	name := "a_name"

	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"type": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"region": attribute.Region(),
	}

	client := mockClient{
		accountID: accountID,
		region:    region,
	}

	ctx := t.Context()

	// Test immutable identity with some values already set (partial null scenario)
	identitySpec := regionalSingleParameterizedIdentitySpec("name")
	invocation := newIdentityInterceptor(&identitySpec)
	interceptor := invocation.interceptor.(identityInterceptor)

	identitySchema := identity.NewIdentitySchema(identitySpec)

	d := schema.TestResourceDataWithIdentityRaw(t, resourceSchema, identitySchema, nil)
	d.SetId("some_id")
	d.Set("name", name)
	d.Set("region", region)
	d.Set("type", "some_type")

	// Simulate partial identity values (some set, some null) by setting some identity attributes
	identity, err := d.Identity()
	if err != nil {
		t.Fatalf("unexpected error getting identity: %v", err)
	}

	// Set only account_id, leaving region and name null
	// This simulates a scenario where identity was partially set (not the full null bug scenario)
	if err := identity.Set(names.AttrAccountID, accountID); err != nil {
		t.Fatalf("unexpected error setting account_id in identity: %v", err)
	}

	opts := crudInterceptorOptions{
		c:    client,
		d:    d,
		when: After,
		why:  Update,
	}

	interceptor.run(ctx, opts)

	identity, err = d.Identity()
	if err != nil {
		t.Fatalf("unexpected error getting identity: %v", err)
	}

	// For partial null scenario, identity should NOT be updated on immutable identity
	// because not ALL values are null
	if e, a := accountID, identity.Get(names.AttrAccountID); e != a {
		t.Errorf("expected account ID to remain %q, got %q", e, a)
	}
	if identity.Get(names.AttrRegion) != "" {
		t.Errorf("expected region to remain empty, got %q", identity.Get(names.AttrRegion))
	}
	if identity.Get("name") != "" {
		t.Errorf("expected name to remain empty, got %q", identity.Get("name"))
	}
}

func regionalSingleParameterizedIdentitySpec(attrName string, opts ...inttypes.IdentityOptsFunc) inttypes.Identity {
	return inttypes.RegionalSingleParameterIdentity(attrName, opts...)
}

func regionalSingleParameterizedIdentitySpecNameMapped(identityAttrName, resourceAttrName string) inttypes.Identity {
	return inttypes.RegionalSingleParameterIdentityWithMappedName(identityAttrName, resourceAttrName)
}

type mockClient struct {
	accountID string
	region    string
}

func (c mockClient) AccountID(_ context.Context) string {
	return c.accountID
}

func (c mockClient) Region(_ context.Context) string {
	return c.region
}

func (c mockClient) DefaultTagsConfig(ctx context.Context) *tftags.DefaultConfig {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) IgnoreTagsConfig(ctx context.Context) *tftags.IgnoreConfig {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) Partition(context.Context) string {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) ServicePackage(_ context.Context, name string) conns.ServicePackage {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) ValidateInContextRegionInPartition(ctx context.Context) error {
	panic("not implemented") //lintignore:R009
}

func (c mockClient) AwsConfig(context.Context) aws.Config { // nosemgrep:ci.aws-in-func-name
	panic("not implemented") //lintignore:R009
}

// TestIdentityInterceptor_ProviderUpgradeBugScenario tests the specific bug reported
// where provider upgrade from pre-identity version to identity-enabled version causes
// "Missing Resource Identity After Update" error during terraform apply operations.
// See https://github.com/hashicorp/terraform-provider-aws/issues/44330
//
// This simulates the exact scenario reported in the GitHub issue where aws_s3_object
// (and similar resources) fail after provider upgrade when Update operations occur
// before the identity has been populated by a Read operation.
func TestIdentityInterceptor_ProviderUpgradeBugScenario(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "us-west-2" //lintignore:AWSAT003
	bucket := "test-bucket"
	key := "test-key"

	// Simulate S3 object-like resource schema (bucket + key identity)
	resourceSchema := map[string]*schema.Schema{
		names.AttrBucket: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		names.AttrKey: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		names.AttrContent: {
			Type:     schema.TypeString,
			Optional: true,
		},
		"region": attribute.Region(),
	}

	client := mockClient{
		accountID: accountID,
		region:    region,
	}

	ctx := context.Background()

	// Create identity spec similar to S3 object: regional with bucket and key
	identitySpec := s3ObjectLikeIdentitySpec()

	invocation := newIdentityInterceptor(&identitySpec)
	interceptor := invocation.interceptor.(identityInterceptor)

	identitySchema := identity.NewIdentitySchema(identitySpec)

	// Create resource data with identity, but simulate the "provider upgrade" scenario
	// by leaving the identity completely empty (all null values)
	d := schema.TestResourceDataWithIdentityRaw(t, resourceSchema, identitySchema, nil)
	d.SetId("test-bucket/test-key") // Resource exists in state
	d.Set(names.AttrBucket, bucket)
	d.Set(names.AttrKey, key)
	d.Set(names.AttrContent, "original-content")
	d.Set("region", region)

	// CRITICAL: Do NOT pre-populate identity - this simulates the upgrade scenario
	// where identity exists in schema but all attributes are null/empty
	identity, err := d.Identity()
	if err != nil {
		t.Fatalf("unexpected error getting identity: %v", err)
	}

	// Verify identity starts fully null (simulating pre-identity provider version)
	if identity.Get(names.AttrAccountID) != "" {
		t.Fatalf("expected account_id to start null, got %q", identity.Get(names.AttrAccountID))
	}
	if identity.Get(names.AttrRegion) != "" {
		t.Fatalf("expected region to start null, got %q", identity.Get(names.AttrRegion))
	}
	if identity.Get(names.AttrBucket) != "" {
		t.Fatalf("expected bucket to start null, got %q", identity.Get(names.AttrBucket))
	}
	if identity.Get(names.AttrKey) != "" {
		t.Fatalf("expected key to start null, got %q", identity.Get(names.AttrKey))
	}

	// Simulate the bug scenario: Update operation occurs (e.g., content change)
	// without a Read operation first to populate identity
	opts := crudInterceptorOptions{
		c:    client,
		d:    d,
		when: After,
		why:  Update, // This is where the bug occurred!
	}

	// Run the identity interceptor - this should NOW populate identity
	// instead of skipping it (which caused the original bug)
	diags := interceptor.run(ctx, opts)
	if diags.HasError() {
		t.Fatalf("unexpected error running interceptor: %v", diags)
	}

	// Verify the fix: identity should now be fully populated
	identity, err = d.Identity()
	if err != nil {
		t.Fatalf("unexpected error getting identity after interceptor: %v", err)
	}

	// These assertions verify the bug is fixed
	if e, a := accountID, identity.Get(names.AttrAccountID); e != a {
		t.Errorf("expected account ID %q, got %q (identity not populated - bug still exists!)", e, a)
	}
	if e, a := region, identity.Get(names.AttrRegion); e != a {
		t.Errorf("expected region %q, got %q (identity not populated - bug still exists!)", e, a)
	}
	if e, a := bucket, identity.Get(names.AttrBucket); e != a {
		t.Errorf("expected bucket %q, got %q (identity not populated - bug still exists!)", e, a)
	}
	if e, a := key, identity.Get(names.AttrKey); e != a {
		t.Errorf("expected key %q, got %q (identity not populated - bug still exists!)", e, a)
	}
}

// TestIdentityInterceptor_ProviderUpgradeBugScenario_PartiallyPopulated tests that
// we don't over-correct and populate identity when it's already partially set
// (which would be the normal case after the first Read operation post-upgrade)
func TestIdentityInterceptor_ProviderUpgradeBugScenario_partiallyPopulated(t *testing.T) {
	t.Parallel()

	accountID := "123456789012"
	region := "us-west-2" //lintignore:AWSAT003
	bucket := "test-bucket"
	key := "test-key"

	resourceSchema := map[string]*schema.Schema{
		names.AttrBucket: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		names.AttrKey: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		names.AttrContent: {
			Type:     schema.TypeString,
			Optional: true,
		},
		"region": attribute.Region(),
	}

	client := mockClient{
		accountID: accountID,
		region:    region,
	}

	ctx := context.Background()

	identitySpec := s3ObjectLikeIdentitySpec()

	invocation := newIdentityInterceptor(&identitySpec)
	interceptor := invocation.interceptor.(identityInterceptor)

	identitySchema := identity.NewIdentitySchema(identitySpec)

	d := schema.TestResourceDataWithIdentityRaw(t, resourceSchema, identitySchema, nil)
	d.SetId("test-bucket/test-key")
	d.Set(names.AttrBucket, bucket)
	d.Set(names.AttrKey, key)
	d.Set(names.AttrContent, "updated-content")
	d.Set("region", region)

	// Simulate partial population (e.g., after first Read post-upgrade)
	identity, err := d.Identity()
	if err != nil {
		t.Fatalf("unexpected error getting identity: %v", err)
	}

	// Set some but not all identity attributes
	if err := identity.Set(names.AttrAccountID, accountID); err != nil {
		t.Fatalf("unexpected error setting account_id in identity: %v", err)
	}
	if err := identity.Set(names.AttrRegion, region); err != nil {
		t.Fatalf("unexpected error setting region in identity: %v", err)
	}
	// Leave bucket and key null to simulate partial population

	// Run Update operation
	opts := crudInterceptorOptions{
		c:    client,
		d:    d,
		when: After,
		why:  Update,
	}

	diags := interceptor.run(ctx, opts)
	if diags.HasError() {
		t.Fatalf("unexpected error running interceptor: %v", diags)
	}

	// Verify identity is NOT re-populated for partial null scenario
	identity, err = d.Identity()
	if err != nil {
		t.Fatalf("unexpected error getting identity after interceptor: %v", err)
	}

	// Account ID and region should remain as set
	if e, a := accountID, identity.Get(names.AttrAccountID); e != a {
		t.Errorf("expected account ID to remain %q, got %q", e, a)
	}
	if e, a := region, identity.Get(names.AttrRegion); e != a {
		t.Errorf("expected region to remain %q, got %q", e, a)
	}

	// Bucket and key should remain null (not populated because not ALL were null)
	if identity.Get(names.AttrBucket) != "" {
		t.Errorf("expected bucket to remain null, got %q", identity.Get(names.AttrBucket))
	}
	if identity.Get(names.AttrKey) != "" {
		t.Errorf("expected key to remain null, got %q", identity.Get(names.AttrKey))
	}
}

// s3ObjectLikeIdentitySpec creates an identity spec similar to S3 objects
// with multiple parameters (bucket and key) for testing the provider upgrade scenario
func s3ObjectLikeIdentitySpec() inttypes.Identity {
	return inttypes.Identity{
		Attributes: []inttypes.IdentityAttribute{
			inttypes.StringIdentityAttribute(names.AttrAccountID, false),
			inttypes.StringIdentityAttribute(names.AttrRegion, false),
			inttypes.StringIdentityAttribute(names.AttrBucket, true),
			inttypes.StringIdentityAttribute(names.AttrKey, true),
		},
		IsSingleParameter: false,
		IsGlobalResource:  false,
		// Immutable identity (like real S3 object) - this is key to reproducing the bug
		IsMutable:     false,
		IsSetOnUpdate: false,
	}
}
