// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Custom Domain Association")
func newCustomDomainAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &customDomainAssociationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type customDomainAssociationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (*customDomainAssociationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_redshiftserverless_custom_domain_association"
}

func (r *customDomainAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"custom_domain_certificate_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"custom_domain_certificate_expiry_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"custom_domain_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"workgroup_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *customDomainAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data customDomainAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftServerlessClient(ctx)

	input := &redshiftserverless.CreateCustomDomainAssociationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateCustomDomainAssociation(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Redshift Serverless Custom Domain Association", err.Error())

		return
	}

	// Set values for unknowns.
	data.CustomDomainCertificateExpiryTime = timetypes.NewRFC3339TimePointerValue(output.CustomDomainCertificateExpiryTime)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *customDomainAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data customDomainAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().RedshiftServerlessClient(ctx)

	output, err := findCustomDomainAssociationByTwoPartKey(ctx, conn, data.WorkgroupName.ValueString(), data.CustomDomainName.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Redshift Serverless Custom Domain Association (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *customDomainAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new customDomainAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftServerlessClient(ctx)

	input := &redshiftserverless.UpdateCustomDomainAssociationInput{
		CustomDomainCertificateArn: aws.String(new.CustomDomainCertificateARN.ValueString()),
		CustomDomainName:           aws.String(new.CustomDomainName.ValueString()),
		WorkgroupName:              aws.String(new.WorkgroupName.ValueString()),
	}

	output, err := conn.UpdateCustomDomainAssociation(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Redshift Serverless Custom Domain Association (%s)", new.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	new.CustomDomainCertificateExpiryTime = timetypes.NewRFC3339TimePointerValue(output.CustomDomainCertificateExpiryTime)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *customDomainAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data customDomainAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftServerlessClient(ctx)

	_, err := conn.DeleteCustomDomainAssociation(ctx, &redshiftserverless.DeleteCustomDomainAssociationInput{
		CustomDomainName: aws.String(data.CustomDomainName.ValueString()),
		WorkgroupName:    aws.String(data.WorkgroupName.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Redshift Serverless Custom Domain Association (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findCustomDomainAssociationByTwoPartKey(ctx context.Context, conn *redshiftserverless.Client, workgroupName, customDomainName string) (*redshiftserverless.GetCustomDomainAssociationOutput, error) {
	input := &redshiftserverless.GetCustomDomainAssociationInput{
		CustomDomainName: aws.String(customDomainName),
		WorkgroupName:    aws.String(workgroupName),
	}

	output, err := conn.GetCustomDomainAssociation(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CustomDomainCertificateExpiryTime == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type customDomainAssociationResourceModel struct {
	CustomDomainCertificateARN        fwtypes.ARN       `tfsdk:"custom_domain_certificate_arn"`
	CustomDomainCertificateExpiryTime timetypes.RFC3339 `tfsdk:"custom_domain_certificate_expiry_time"`
	CustomDomainName                  types.String      `tfsdk:"custom_domain_name"`
	ID                                types.String      `tfsdk:"id"`
	WorkgroupName                     types.String      `tfsdk:"workgroup_name"`
}

const (
	customDomainAssociationResourceIDPartCount = 2
)

func (data *customDomainAssociationResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, customDomainAssociationResourceIDPartCount, false)

	if err != nil {
		return err
	}

	data.WorkgroupName = types.StringValue(parts[0])
	data.CustomDomainName = types.StringValue(parts[1])

	return nil
}

func (data *customDomainAssociationResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.WorkgroupName.ValueString(), data.CustomDomainName.ValueString()}, customDomainAssociationResourceIDPartCount, false)))
}
