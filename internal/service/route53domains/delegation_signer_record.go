// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package route53domains

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_route53domains_delegation_signer_record", name="Delegation Signer Record")
func newDelegationSignerRecordResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &delegationSignerRecordResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type delegationSignerRecordResource struct {
	framework.ResourceWithModel[delegationSignerRecordResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *delegationSignerRecordResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
						},
						"flags": schema.Int64Attribute{
							Required: true,
						},
						names.AttrPublicKey: schema.StringAttribute{
							Required: true,
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					// GetDomainDetail does not return the key's flags or public key, so signing_attributes
					// cannot be refreshed from the API. Replacement is decided by comparing the DS digest
					// of the planned key with the digest of the associated key, which also allows an
					// imported record to be completed in place instead of being destroyed and recreated.
					listplanmodifier.RequiresReplaceIf(
						signingAttributesRequiresReplaceIf,
						"If the value of this attribute changes the DS record digest, Terraform will destroy and recreate the resource.",
						"If the value of this attribute changes the DS record digest, Terraform will destroy and recreate the resource.",
					),
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

func (r *delegationSignerRecordResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data delegationSignerRecordResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53DomainsClient(ctx)

	domainName := data.DomainName.ValueString()
	digests, diags := data.dsDigests(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &route53domains.AssociateDelegationSignerToDomainInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.AssociateDelegationSignerToDomain(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Route 53 Domains Delegation Signer Record", err.Error())

		return
	}

	if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		detail := err.Error()
		// Registries reject a DS record that is already associated with the domain.
		if dnssecKey, findErr := findDNSSECKeyByDigests(ctx, conn, domainName, digests); findErr == nil {
			detail = fmt.Sprintf("%s. A DS record with digest %s is already associated with domain %s; import it using the ID %q", detail, aws.ToString(dnssecKey.Digest), domainName, domainName+","+aws.ToString(dnssecKey.Digest))
		}
		response.Diagnostics.AddError("waiting for Route 53 Domains Delegation Signer Record create", detail)

		return
	}

	// GetDomainDetail does not return the key's flags or public key, so the associated key
	// is identified by its DS digest, computed locally from the signing attributes.
	dnssecKey, err := tfresource.RetryWhenNotFound(ctx, dnssecKeyPropagationTimeout, func(ctx context.Context) (*awstypes.DnssecKey, error) {
		return findDNSSECKeyByDigests(ctx, conn, domainName, digests)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 Domains Domain (%s) DNSSEC key", domainName), err.Error())

		return
	}

	// Set values for unknowns.
	// The DS digest is stable across registrars and API changes, unlike DnssecKey.Id.
	data.DNSSECKeyID = fwflex.StringToFramework(ctx, dnssecKey.Digest)
	id, err := data.setID()
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("flattening resource ID Route 53 Domains Domain (%s) DNSSEC key", domainName), err.Error())
		return
	}
	data.ID = types.StringValue(id)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *delegationSignerRecordResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data delegationSignerRecordResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().Route53DomainsClient(ctx)

	dnssecKey, err := findDNSSECKeyByTwoPartKey(ctx, conn, data.DomainName.ValueString(), data.DNSSECKeyID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 Domains Domain (%s) DNSSEC key", data.DomainName.ValueString()), err.Error())

		return
	}

	// Normalize the key ID to the DS digest. State written by earlier provider versions
	// holds the registry-assigned DnssecKey.Id instead, whose format varies by registrar.
	data.DNSSECKeyID = fwflex.StringToFramework(ctx, dnssecKey.Digest)
	id, err := data.setID()
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("flattening resource ID Route 53 Domains Domain (%s) DNSSEC key", data.DomainName.ValueString()), err.Error())
		return
	}
	data.ID = types.StringValue(id)

	// signing_attributes is left untouched: GetDomainDetail does not return the key's flags or
	// public key, and the DS digest already proves that the associated key matches them.

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *delegationSignerRecordResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new delegationSignerRecordResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	// The only in-place change is to signing_attributes with an unchanged DS digest
	// (e.g. completing the configuration after import), which requires no API call.
	new.DNSSECKeyID = old.DNSSECKeyID
	new.ID = old.ID

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *delegationSignerRecordResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data delegationSignerRecordResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53DomainsClient(ctx)

	// DisassociateDelegationSignerFromDomain requires the registry-assigned DnssecKey.Id.
	dnssecKey, err := findDNSSECKeyByTwoPartKey(ctx, conn, data.DomainName.ValueString(), data.DNSSECKeyID.ValueString())

	if retry.NotFound(err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 Domains Domain (%s) DNSSEC key", data.DomainName.ValueString()), err.Error())

		return
	}

	output, err := conn.DisassociateDelegationSignerFromDomain(ctx, &route53domains.DisassociateDelegationSignerFromDomainInput{
		DomainName: data.DomainName.ValueStringPointer(),
		Id:         dnssecKey.Id,
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Route 53 Domains Delegation Signer Record (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		if !errs.Contains(err, "The DNSSEC you specified is not found on domain") {
			response.Diagnostics.AddError("waiting for Route 53 Domains Delegation Signer Record delete", err.Error())

			return
		}
	}
}

// signingAttributesRequiresReplaceIf forces replacement only when the planned signing
// attributes produce a different DS digest than the associated key's.
func signingAttributesRequiresReplaceIf(ctx context.Context, request planmodifier.ListRequest, response *listplanmodifier.RequiresReplaceIfFuncResponse) {
	response.RequiresReplace = true

	var plan, state delegationSignerRecordResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if plan.SigningAttributes.IsUnknown() || state.DNSSECKeyID.IsNull() || state.DNSSECKeyID.IsUnknown() {
		return
	}

	// domain_name itself requires replacement, so the digest is computed for the associated key's domain.
	plan.DomainName = state.DomainName
	digests, diags := plan.dsDigests(ctx)
	if diags.HasError() {
		// Invalid or unknown signing attributes; keep the default of replacing the resource.
		return
	}

	response.RequiresReplace = !dsDigestMatches(digests, state.DNSSECKeyID.ValueString())
}

type delegationSignerRecordResourceModel struct {
	DNSSECKeyID       types.String                                                                  `tfsdk:"dnssec_key_id"`
	DomainName        types.String                                                                  `tfsdk:"domain_name"`
	ID                types.String                                                                  `tfsdk:"id"`
	SigningAttributes fwtypes.ListNestedObjectValueOf[delegationSignerRecordSigningAttributesModel] `tfsdk:"signing_attributes"`
	Timeouts          timeouts.Value                                                                `tfsdk:"timeouts"`
}

type delegationSignerRecordSigningAttributesModel struct {
	Algorithm types.Int64  `tfsdk:"algorithm"`
	Flags     types.Int64  `tfsdk:"flags"`
	PublicKey types.String `tfsdk:"public_key"`
}

const (
	delegationSignerRecordResourceIDPartCount = 2

	// Time to wait for a newly associated key to be returned by GetDomainDetail.
	dnssecKeyPropagationTimeout = 2 * time.Minute
)

func (data *delegationSignerRecordResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, delegationSignerRecordResourceIDPartCount, false)

	if err != nil {
		return err
	}

	data.DNSSECKeyID = types.StringValue(parts[1])
	data.DomainName = types.StringValue(parts[0])

	return nil
}

func (data *delegationSignerRecordResourceModel) setID() (string, error) {
	parts := []string{
		data.DomainName.ValueString(),
		data.DNSSECKeyID.ValueString(),
	}

	return flex.FlattenResourceId(parts, delegationSignerRecordResourceIDPartCount, false)
}

// dsDigests computes the DS record digests (by digest type) of the configured signing attributes.
func (data *delegationSignerRecordResourceModel) dsDigests(ctx context.Context) (map[int32]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	signingAttributes, d := data.SigningAttributes.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	if signingAttributes == nil || data.DomainName.IsUnknown() || signingAttributes.Algorithm.IsUnknown() || signingAttributes.Flags.IsUnknown() || signingAttributes.PublicKey.IsUnknown() {
		diags.AddError("computing Route 53 Domains Delegation Signer Record digest", "signing attributes are not known")
		return nil, diags
	}

	digests, err := dsDigests(data.DomainName.ValueString(), signingAttributes.Flags.ValueInt64(), signingAttributes.Algorithm.ValueInt64(), signingAttributes.PublicKey.ValueString())
	if err != nil {
		diags.AddError("computing Route 53 Domains Delegation Signer Record digest", err.Error())
		return nil, diags
	}

	return digests, diags
}

// findDNSSECKeyByTwoPartKey finds a domain's DNSSEC key by its DS digest.
// For state written by earlier provider versions, the key ID may instead be the
// registry-assigned DnssecKey.Id (e.g. a number, or "DS:<keytag>-<algorithm>-<digesttype>-<digest>").
func findDNSSECKeyByTwoPartKey(ctx context.Context, conn *route53domains.Client, domainName, keyID string) (*awstypes.DnssecKey, error) {
	output, err := findDomainDetailByName(ctx, conn, domainName)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.DnssecKeys, func(v awstypes.DnssecKey) bool {
		digest := aws.ToString(v.Digest)

		return strings.EqualFold(digest, keyID) || aws.ToString(v.Id) == keyID || (digest != "" && strings.HasSuffix(strings.ToUpper(keyID), "-"+strings.ToUpper(digest)))
	}))
}

// findDNSSECKeyByDigests finds a domain's DNSSEC key whose DS digest matches one of digests.
func findDNSSECKeyByDigests(ctx context.Context, conn *route53domains.Client, domainName string, digests map[int32]string) (*awstypes.DnssecKey, error) {
	output, err := findDomainDetailByName(ctx, conn, domainName)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.DnssecKeys, func(v awstypes.DnssecKey) bool {
		return dsDigestMatches(digests, aws.ToString(v.Digest))
	}))
}
