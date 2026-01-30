// Copyright IBM Corp. 2014, 2026
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
			if why == Update && !(r.identitySpec.IsMutable && r.identitySpec.IsSetOnUpdate) && !identityIsFullyNull(d, r.identitySpec) {
				break
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
	case OnError:
		switch why {
		case Update:
			if identityIsFullyNull(d, r.identitySpec) {
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
	}

	return diags
}

// identityIsFullyNull returns true if a resource supports identity and
// all attributes are set to null values
func identityIsFullyNull(d schemaResourceData, identitySpec *inttypes.Identity) bool {
	identity, err := d.Identity()
	if err != nil {
		return false
	}

	for _, attr := range identitySpec.Attributes {
		value := identity.Get(attr.Name())
		if value != "" {
			return false
		}
	}

	return true
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
		when: After | OnError,
		why:  Create | Read | Update,
		interceptor: identityInterceptor{
			identitySpec: identitySpec,
		},
	}

	return interceptor
}

func newResourceIdentity(v inttypes.Identity) *schema.ResourceIdentity {
	return &schema.ResourceIdentity{
		Version:           v.Version(),
		IdentityUpgraders: v.SDKv2IdentityUpgraders(),
		SchemaFunc: func() map[string]*schema.Schema {
			return identity.NewIdentitySchema(v)
		},
	}
}

func newParameterizedIdentityImporter(identitySpec inttypes.Identity, importSpec inttypes.SDKv2Import) *schema.ResourceImporter {
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
				if err := importer.GlobalSingleton(ctx, rd, identity, meta.(importer.AWSClient)); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{rd}, nil
			},
		}
	} else {
		return &schema.ResourceImporter{
			StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if err := importer.RegionalSingleton(ctx, rd, identity, meta.(importer.AWSClient)); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{rd}, nil
			},
		}
	}
}

func customInherentRegionResourceImporter(identity inttypes.Identity) *schema.ResourceImporter {
	// Not supported for Global resources. This is validated in validateResourceSchemas().
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
			if err := importer.RegionalInherentRegion(ctx, rd, identity); err != nil {
				return nil, err
			}

			return []*schema.ResourceData{rd}, nil
		},
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
