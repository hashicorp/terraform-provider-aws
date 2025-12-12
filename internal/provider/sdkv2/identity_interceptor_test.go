// Copyright IBM Corp. 2014, 2025
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
			ExpectIdentity: true,
			Description:    "Immutable identity with all null attributes should get populated (bug fix scenario)",
		},
		"v6.0 SDK fix": {
			attrName: "name",
			identitySpec: regionalSingleParameterizedIdentitySpec("name",
				inttypes.WithV6_0SDKv2Fix(),
			),
			ExpectIdentity: true,
			Description:    "Mutable identity (v6.0 SDK fix) should always get populated on Update",
		},
		"identity fix": {
			attrName: "name",
			identitySpec: regionalSingleParameterizedIdentitySpec("name",
				inttypes.WithIdentityFix(),
			),
			ExpectIdentity: true,
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

func (c mockClient) TagPolicyConfig(ctx context.Context) *tftags.TagPolicyConfig {
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

func TestIdentityIsFullyNull(t *testing.T) {
	t.Parallel()

	identitySpec := &inttypes.Identity{
		Attributes: []inttypes.IdentityAttribute{
			inttypes.StringIdentityAttribute(names.AttrAccountID, false),
			inttypes.StringIdentityAttribute(names.AttrRegion, false),
			inttypes.StringIdentityAttribute(names.AttrBucket, true),
		},
	}

	testCases := map[string]struct {
		identityValues map[string]string
		expectNull     bool
		description    string
	}{
		"all_null": {
			identityValues: map[string]string{},
			expectNull:     true,
			description:    "All attributes null should return true",
		},
		"some_null": {
			identityValues: map[string]string{
				names.AttrAccountID: "123456789012",
				// region and bucket remain null
			},
			expectNull:  false,
			description: "Some attributes set should return false",
		},
		"all_set": {
			identityValues: map[string]string{
				names.AttrAccountID: "123456789012",
				names.AttrRegion:    "us-west-2", // lintignore:AWSAT003
				names.AttrBucket:    "test-bucket",
			},
			expectNull:  false,
			description: "All attributes set should return false",
		},
		"empty_string_values": {
			identityValues: map[string]string{
				names.AttrAccountID: "",
				names.AttrRegion:    "",
				names.AttrBucket:    "",
			},
			expectNull:  true,
			description: "Empty string values should be treated as null",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resourceSchema := map[string]*schema.Schema{
				names.AttrBucket: {Type: schema.TypeString, Required: true},
			}
			identitySchema := identity.NewIdentitySchema(*identitySpec)
			d := schema.TestResourceDataWithIdentityRaw(t, resourceSchema, identitySchema, nil)
			d.SetId("test-id")

			identity, err := d.Identity()
			if err != nil {
				t.Fatalf("unexpected error getting identity: %v", err)
			}
			for attrName, value := range tc.identityValues {
				if err := identity.Set(attrName, value); err != nil {
					t.Fatalf("unexpected error setting %s in identity: %v", attrName, err)
				}
			}

			result := identityIsFullyNull(d, identitySpec)
			if result != tc.expectNull {
				t.Errorf("%s: expected identityIsFullyNull to return %v, got %v",
					tc.description, tc.expectNull, result)
			}
		})
	}
}
