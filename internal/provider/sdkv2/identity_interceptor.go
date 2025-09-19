// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ crudInterceptor = identityInterceptor{}

type identityInterceptor struct {
	identitySpec *inttypes.Identity
}

func (r identityInterceptor) run(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
	var diags diag.Diagnostics
	awsClient := opts.c

	switch d, when, why := opts.d, opts.when, opts.why; when {
	case After:
		switch why {
		case Create, Read, Update:
			// For Update operations on resources with immutable identity,
			// still set identity if it has null values (e.g., after provider upgrade from pre-identity version)
			if why == Update && !(r.identitySpec.IsMutable && r.identitySpec.IsSetOnUpdate) {
				// Skip setting identity unless it has null values
				if !identityHasNullValues(d, r.identitySpec) {
					break
				}
			}
			if d.Id() == "" {
				break
			}
			identity, err := d.Identity()
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			for _, attr := range r.identitySpec.Attributes {
				switch attr.Name() {
				case names.AttrAccountID:
					if err := identity.Set(attr.Name(), awsClient.AccountID(ctx)); err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}

				case names.AttrRegion:
					if err := identity.Set(attr.Name(), awsClient.Region(ctx)); err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}

				default:
					val, ok := getAttributeOk(d, attr.ResourceAttributeName())
					if !ok {
						continue
					}
					if err := identity.Set(attr.Name(), val); err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}
				}
			}
		}
	}

	return diags
}

// identityHasNullValues checks if the resource's identity contains any null values.
// This typically occurs when a resource was created with a pre-identity version of the provider
// and then upgraded to a version that supports identity, but a failed update operation
// resulted in null identity values being written to state.
//
// Returns true only if identity data exists but contains null/empty values.
// Returns false if no identity data exists at all (which is expected for some resources).
func identityHasNullValues(d schemaResourceData, identitySpec *inttypes.Identity) bool {
	identity, err := d.Identity()
	if err != nil {
		// If we can't get identity at all, this is not the null values bug
		return false
	}

	// Check if identity data exists and has been initialized
	hasAnyIdentityData := false
	hasNullValues := false

	// Check each identity attribute
	for _, attr := range identitySpec.Attributes {
		value := identity.Get(attr.Name())
		if value != "" {
			// Found non-empty value, so identity data exists
			hasAnyIdentityData = true
		} else {
			// Found empty/null value
			hasNullValues = true
		}
	}

	// Only return true if identity data exists but some values are null
	// This indicates the specific bug scenario we're trying to fix
	return hasAnyIdentityData && hasNullValues
}

func getAttributeOk(d schemaResourceData, name string) (string, bool) {
	if name == "id" {
		return d.Id(), true
	}
	if v, ok := d.GetOk(name); !ok {
		return "", false
	} else {
		return v.(string), true
	}
}

func newIdentityInterceptor(identitySpec *inttypes.Identity) interceptorInvocation {
	interceptor := interceptorInvocation{
		when: After,
		why:  Create | Read,
		interceptor: identityInterceptor{
			identitySpec: identitySpec,
		},
	}

	if identitySpec.IsMutable && identitySpec.IsSetOnUpdate {
		interceptor.why |= Update
	}

	return interceptor
}

func newResourceIdentity(v inttypes.Identity) *schema.ResourceIdentity {
	return &schema.ResourceIdentity{
		SchemaFunc: func() map[string]*schema.Schema {
			return identity.NewIdentitySchema(v)
		},
	}
}

func newParameterizedIdentityImporter(identitySpec inttypes.Identity, importSpec *inttypes.SDKv2Import) *schema.ResourceImporter {
	if identitySpec.IsSingleParameter {
		if identitySpec.IsGlobalResource {
			return &schema.ResourceImporter{
				StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
					if err := importer.GlobalSingleParameterized(ctx, rd, identitySpec, meta.(importer.AWSClient)); err != nil {
						return nil, err
					}

					return []*schema.ResourceData{rd}, nil
				},
			}
		} else {
			return &schema.ResourceImporter{
				StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
					if err := importer.RegionalSingleParameterized(ctx, rd, identitySpec, meta.(importer.AWSClient)); err != nil {
						return nil, err
					}

					return []*schema.ResourceData{rd}, nil
				},
			}
		}
	} else {
		if identitySpec.IsGlobalResource {
			return &schema.ResourceImporter{
				StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
					if err := importer.GlobalMultipleParameterized(ctx, rd, identitySpec, importSpec, meta.(importer.AWSClient)); err != nil {
						return nil, err
					}

					return []*schema.ResourceData{rd}, nil
				},
			}
		} else {
			return &schema.ResourceImporter{
				StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
					if err := importer.RegionalMultipleParameterized(ctx, rd, identitySpec, importSpec, meta.(importer.AWSClient)); err != nil {
						return nil, err
					}

					return []*schema.ResourceData{rd}, nil
				},
			}
		}
	}
}

func arnIdentityResourceImporter(identity inttypes.Identity) *schema.ResourceImporter {
	if identity.IsGlobalResource {
		return &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.GlobalARN(ctx, rd, identity); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{rd}, nil
			},
		}
	} else {
		return &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.RegionalARN(ctx, rd, identity); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{rd}, nil
			},
		}
	}
}

func singletonIdentityResourceImporter(identity inttypes.Identity) *schema.ResourceImporter {
	if identity.IsGlobalResource {
		// Historically, we haven't validated *any* Import ID value for Global Singletons
		return &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.GlobalSingleton(ctx, rd, &identity, meta.(importer.AWSClient)); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{rd}, nil
			},
		}
	} else {
		return &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.RegionalSingleton(ctx, rd, &identity, meta.(importer.AWSClient)); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{rd}, nil
			},
		}
	}
}

func customResourceImporter(r *schema.Resource, identity *inttypes.Identity, importSpec *inttypes.SDKv2Import) {
	importF := r.Importer.StateContext

	r.Importer = &schema.ResourceImporter{
		StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
			ctx = importer.Context(ctx, identity, importSpec)

			return importF(ctx, rd, meta)
		},
	}
}
