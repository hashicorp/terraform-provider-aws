// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Delegation Signer Record")
func newDelegationSignerRecordResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &delegationSignerRecordResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type delegationSignerRecordResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *delegationSignerRecordResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_route53domains_delegation_signer_record"
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

func (r *delegationSignerRecordResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data delegationSignerRecordResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53DomainsClient(ctx)

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
		response.Diagnostics.AddError("waiting for Route 53 Domains Delegation Signer Record create", err.Error())

		return
	}

	signingAttributes, diags := data.SigningAttributes.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	dnssecKey, err := findDNSSECKeyByThreePartKey(ctx, conn, data.DomainName.ValueString(), int(signingAttributes.Flags.ValueInt64()), signingAttributes.PublicKey.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 Domains Domain (%s) DNSSEC key", data.DomainName.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.DNSSECKeyID = fwflex.StringToFramework(ctx, dnssecKey.Id)
	data.setID()

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

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 Domains Domain (%s) DNSSEC key", data.DomainName.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	var signingAttributes delegationSignerRecordSigningAttributesModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, dnssecKey, &signingAttributes)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.SigningAttributes = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &signingAttributes)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *delegationSignerRecordResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data delegationSignerRecordResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53DomainsClient(ctx)

	output, err := conn.DisassociateDelegationSignerFromDomain(ctx, &route53domains.DisassociateDelegationSignerFromDomainInput{
		DomainName: aws.String(data.DomainName.ValueString()),
		Id:         aws.String(data.DNSSECKeyID.ValueString()),
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

func (data *delegationSignerRecordResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.DomainName.ValueString(), data.DNSSECKeyID.ValueString()}, delegationSignerRecordResourceIDPartCount, false)))
}
