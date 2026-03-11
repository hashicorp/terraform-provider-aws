// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package vpclattice

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpclattice_domain_verification", name="Domain Verification")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/vpclattice;vpclattice.GetDomainVerificationOutput")
func newDomainVerificationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &domainVerificationResource{}, nil
}

type domainVerificationResource struct {
	framework.ResourceWithModel[domainVerificationResourceModel]
	framework.WithImportByID
}

func (r *domainVerificationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"last_verified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VerificationStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"txt_record_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"txt_record_value": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *domainVerificationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data domainVerificationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	input := vpclattice.StartDomainVerificationInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		DomainName:  fwflex.StringFromFramework(ctx, data.DomainName),
		Tags:        getTagsIn(ctx),
	}

	output, err := conn.StartDomainVerification(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating VPCLattice Domain Verification (%s)", data.DomainName.ValueString()), err.Error())
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.Id)

	outputGet, err := findDomainVerificationByID(ctx, conn, data.ID.ValueString())

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID)
		response.Diagnostics.AddError(fmt.Sprintf("reading VPCLattice Domain Verification (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGet, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Manually set TXT record fields from TxtMethodConfig
	if outputGet.TxtMethodConfig != nil {
		data.TxtRecordName = fwflex.StringToFramework(ctx, outputGet.TxtMethodConfig.Name)
		data.TxtRecordValue = fwflex.StringToFramework(ctx, outputGet.TxtMethodConfig.Value)
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *domainVerificationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data domainVerificationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	output, err := findDomainVerificationByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading VPCLattice Domain Verification (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Manually set TXT record fields from TxtMethodConfig
	if output.TxtMethodConfig != nil {
		data.TxtRecordName = fwflex.StringToFramework(ctx, output.TxtMethodConfig.Name)
		data.TxtRecordValue = fwflex.StringToFramework(ctx, output.TxtMethodConfig.Value)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *domainVerificationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data domainVerificationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VPCLatticeClient(ctx)

	input := vpclattice.DeleteDomainVerificationInput{
		DomainVerificationIdentifier: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteDomainVerification(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting VPCLattice Domain Verification (%s)", data.ID.ValueString()), err.Error())
		return
	}
}

func findDomainVerificationByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetDomainVerificationOutput, error) {
	input := vpclattice.GetDomainVerificationInput{
		DomainVerificationIdentifier: aws.String(id),
	}

	output, err := conn.GetDomainVerification(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type domainVerificationResourceModel struct {
	framework.WithRegionModel
	ARN              types.String                                    `tfsdk:"arn"`
	CreatedAt        timetypes.RFC3339                               `tfsdk:"created_at"`
	DomainName       types.String                                    `tfsdk:"domain_name"`
	ID               types.String                                    `tfsdk:"id"`
	LastVerifiedTime timetypes.RFC3339                               `tfsdk:"last_verified_time"`
	Status           fwtypes.StringEnum[awstypes.VerificationStatus] `tfsdk:"status"`
	Tags             tftags.Map                                      `tfsdk:"tags"`
	TagsAll          tftags.Map                                      `tfsdk:"tags_all"`
	TxtRecordName    types.String                                    `tfsdk:"txt_record_name"`
	TxtRecordValue   types.String                                    `tfsdk:"txt_record_value"`
}
