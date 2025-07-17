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
	region := "us-west-2"
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

	invocation := newIdentityInterceptor(identitySpec.Attributes)
	interceptor := invocation.interceptor.(identityInterceptor)

	client := mockClient{
		accountID: accountID,
		region:    region,
	}

	testCases := map[string]struct {
		id             string
		expectIdentity bool
	}{
		"with id": {
			id:             "some_id",
			expectIdentity: true,
		},
		"without id": {
			id:             "",
			expectIdentity: false,
		},
	}

	for tname, tc := range testCases {
		t.Run(tname, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			d := schema.TestResourceDataWithIdentityRaw(t, resourceSchema, identitySchema, nil)
			d.SetId(tc.id)
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

			if tc.expectIdentity {
				if e, a := accountID, identity.Get(names.AttrAccountID); e != a {
					t.Errorf("expected account ID %q, got %q", e, a)
				}
				if e, a := region, identity.Get(names.AttrRegion); e != a {
					t.Errorf("expected region %q, got %q", e, a)
				}
				if e, a := name, identity.Get("name"); e != a {
					t.Errorf("expected name %q, got %q", e, a)
				}
			} else {
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
		})
	}
}

func regionalSingleParameterizedIdentitySpec(attrName string) inttypes.Identity {
	return inttypes.Identity{
		IsGlobalResource:  true,
		IdentityAttribute: attrName,
		Attributes: []inttypes.IdentityAttribute{
			{
				Name:     "account_id",
				Required: false,
			},
			{
				Name:     "region",
				Required: false,
			},
			{
				Name:     attrName,
				Required: true,
			},
		},
	}
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
	panic("not implemented")
}

func (c mockClient) IgnoreTagsConfig(ctx context.Context) *tftags.IgnoreConfig {
	panic("not implemented")
}

func (c mockClient) Partition(context.Context) string {
	panic("not implemented")
}

func (c mockClient) ServicePackage(_ context.Context, name string) conns.ServicePackage {
	panic("not implemented")
}

func (c mockClient) ValidateInContextRegionInPartition(ctx context.Context) error {
	panic("not implemented")
}

func (c mockClient) AwsConfig(context.Context) aws.Config {
	panic("not implemented")
}
