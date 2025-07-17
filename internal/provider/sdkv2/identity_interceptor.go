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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ crudInterceptor = identityInterceptor{}

type identityInterceptor struct {
	attributes []string
}

func (r identityInterceptor) run(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
	var diags diag.Diagnostics
	awsClient := opts.c

	switch d, when, why := opts.d, opts.when, opts.why; when {
	case After:
		switch why {
		case Create, Read:
			if d.Id() == "" {
				break
			}
			identity, err := d.Identity()
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			for _, attr := range r.attributes {
				switch attr {
				case names.AttrAccountID:
					if err := identity.Set(attr, awsClient.AccountID(ctx)); err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}

				case names.AttrRegion:
					if err := identity.Set(attr, awsClient.Region(ctx)); err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}

				default:
					val, ok := getAttributeOk(d, attr)
					if !ok {
						continue
					}
					if err := identity.Set(attr, val); err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}
				}
			}
		}
	}

	return diags
}

func getAttributeOk(d schemaResourceData, name string) (string, bool) {
	if name == "id" {
		return d.Id(), true
	}
	v, ok := d.GetOk(name)
	return v.(string), ok
}

func newIdentityInterceptor(attributes []inttypes.IdentityAttribute) interceptorInvocation {
	return interceptorInvocation{
		when: After,
		why:  Create | Read,
		interceptor: identityInterceptor{
			attributes: tfslices.ApplyToAll(attributes, func(v inttypes.IdentityAttribute) string {
				return v.Name
			}),
		},
	}
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
					if err := importer.GlobalSingleParameterized(ctx, rd, identitySpec.IdentityAttribute, meta.(importer.AWSClient)); err != nil {
						return nil, err
					}

					return []*schema.ResourceData{rd}, nil
				},
			}
		} else {
			return &schema.ResourceImporter{
				StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
					if err := importer.RegionalSingleParameterized(ctx, rd, identitySpec.IdentityAttribute, meta.(importer.AWSClient)); err != nil {
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
					if err := importer.GlobalMultipleParameterized(ctx, rd, identitySpec.Attributes, importSpec, meta.(importer.AWSClient)); err != nil {
						return nil, err
					}

					return []*schema.ResourceData{rd}, nil
				},
			}
		} else {
			return &schema.ResourceImporter{
				StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
					if err := importer.RegionalMultipleParameterized(ctx, rd, identitySpec.Attributes, importSpec, meta.(importer.AWSClient)); err != nil {
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
				if err := importer.GlobalARN(ctx, rd, identity.IdentityAttribute, identity.IdentityDuplicateAttrs); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{rd}, nil
			},
		}
	} else {
		return &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.RegionalARN(ctx, rd, identity.IdentityAttribute, identity.IdentityDuplicateAttrs); err != nil {
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
