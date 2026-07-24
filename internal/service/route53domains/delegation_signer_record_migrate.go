// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// delegationSignerRecordSchemaV0 mirrors the schema as it existed before the
// digest-keying fix (#47928). The attribute set is identical to v1; the
// version bump exists solely to give us a hook to rewrite the value stored in
// dnssec_key_id (and the composite id) from the prior
// "DS:keytag-algorithm-digesttype-digest" form down to the bare digest that
// GetDomainDetail actually returns.
func delegationSignerRecordSchemaV0(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"dnssec_key_id": framework.IDAttribute(),
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"signing_attributes": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[delegationSignerRecordSigningAttributesModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"algorithm": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						"flags": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						names.AttrPublicKey: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

// delegationSignerRecordResourceModelV0 has the same shape as the current
// model. The migration only mutates string values.
type delegationSignerRecordResourceModelV0 struct {
	DNSSECKeyID       types.String                                                                  `tfsdk:"dnssec_key_id"`
	DomainName        types.String                                                                  `tfsdk:"domain_name"`
	ID                types.String                                                                  `tfsdk:"id"`
	SigningAttributes fwtypes.ListNestedObjectValueOf[delegationSignerRecordSigningAttributesModel] `tfsdk:"signing_attributes"`
	Timeouts          timeouts.Value                                                                `tfsdk:"timeouts"`
}

// digestFromLegacyDNSSECKeyID extracts the bare digest from a prior-form
// dnssec_key_id value. The Route 53 Domains API documents the DS record id as
// "DS:keytag-algorithm-digesttype-digest". We split on "-" and take the
// trailing segment after stripping the "DS:" prefix.
//
// If the input is empty or does not have the "DS:" prefix it is returned
// unchanged so that the upgrader is idempotent and safe to run against state
// that was already imported with the digest form.
func digestFromLegacyDNSSECKeyID(v string) string {
	if !strings.HasPrefix(v, "DS:") {
		return v
	}
	trimmed := strings.TrimPrefix(v, "DS:")
	parts := strings.Split(trimmed, "-")
	if len(parts) < 4 {
		// Unexpected shape - leave the original alone rather than corrupt
		// state. A subsequent Read will surface the mismatch loudly.
		return v
	}
	return parts[len(parts)-1]
}

func upgradeDelegationSignerRecordStateFromV0(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var dataV0 delegationSignerRecordResourceModelV0
	response.Diagnostics.Append(request.State.Get(ctx, &dataV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	upgraded := delegationSignerRecordResourceModel{
		DNSSECKeyID:       types.StringValue(digestFromLegacyDNSSECKeyID(dataV0.DNSSECKeyID.ValueString())),
		DomainName:        dataV0.DomainName,
		SigningAttributes: dataV0.SigningAttributes,
		Timeouts:          dataV0.Timeouts,
	}

	// Rebuild the composite id from the (possibly rewritten) DomainName and
	// DNSSECKeyID so that the id stored in state matches what setID() would
	// have produced for a fresh Create.
	id, err := upgraded.setID()
	if err != nil {
		response.Diagnostics.AddError("upgrading Route 53 Domains Delegation Signer Record state", err.Error())
		return
	}
	upgraded.ID = types.StringValue(id)

	response.Diagnostics.Append(response.State.Set(ctx, upgraded)...)
}
