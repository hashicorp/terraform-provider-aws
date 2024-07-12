// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Continuous Deployment Policy")
func newContinuousDeploymentPolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &continuousDeploymentPolicyResource{}, nil
}

type continuousDeploymentPolicyResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (*continuousDeploymentPolicyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cloudfront_continuous_deployment_policy"
}

func (r *continuousDeploymentPolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrEnabled: schema.BoolAttribute{
				Required: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"staging_distribution_dns_names": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[stagingDistributionDNSNamesModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"items": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							ElementType: types.StringType,
						},
						"quantity": schema.Int64Attribute{
							Required: true,
						},
					},
				},
			},
			"traffic_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[trafficConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ContinuousDeploymentPolicyType](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"single_header_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[continuousDeploymentSingleHeaderConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrHeader: schema.StringAttribute{
										Required: true,
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"single_weight_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[continuousDeploymentSingleWeightConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrWeight: schema.Float64Attribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"session_stickiness_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[sessionStickinessConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"idle_ttl": schema.Int64Attribute{
													Required: true,
												},
												"maximum_ttl": schema.Int64Attribute{
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *continuousDeploymentPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data continuousDeploymentPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	input := &cloudfront.CreateContinuousDeploymentPolicyInput{
		ContinuousDeploymentPolicyConfig: &awstypes.ContinuousDeploymentPolicyConfig{},
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input.ContinuousDeploymentPolicyConfig)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateContinuousDeploymentPolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating CloudFront Continuous Deployment Policy", err.Error())

		return
	}

	// Set values for unknowns.
	data.ETag = fwflex.StringToFramework(ctx, output.ETag)
	data.ID = fwflex.StringToFramework(ctx, output.ContinuousDeploymentPolicy.Id)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, output.ContinuousDeploymentPolicy.LastModifiedTime)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *continuousDeploymentPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data continuousDeploymentPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	output, err := findContinuousDeploymentPolicyByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Continuous Deployment Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.ContinuousDeploymentPolicy.ContinuousDeploymentPolicyConfig, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ETag = fwflex.StringToFramework(ctx, output.ETag)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, output.ContinuousDeploymentPolicy.LastModifiedTime)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *continuousDeploymentPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new continuousDeploymentPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	if !new.Enabled.Equal(old.Enabled) ||
		!new.StagingDistributionDNSNames.Equal(old.StagingDistributionDNSNames) ||
		!new.TrafficConfig.Equal(old.TrafficConfig) {
		input := &cloudfront.UpdateContinuousDeploymentPolicyInput{
			ContinuousDeploymentPolicyConfig: &awstypes.ContinuousDeploymentPolicyConfig{},
		}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input.ContinuousDeploymentPolicyConfig)...)
		if response.Diagnostics.HasError() {
			return
		}

		input.Id = aws.String(new.ID.ValueString())
		// Use state ETag value. The planned value will be unknown.
		input.IfMatch = aws.String(old.ETag.ValueString())

		output, err := conn.UpdateContinuousDeploymentPolicy(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudFront Continuous Deployment Policy (%s)", new.ID.ValueString()), err.Error())

			return
		}

		new.ETag = fwflex.StringToFramework(ctx, output.ETag)
		new.LastModifiedTime = fwflex.TimeToFramework(ctx, output.ContinuousDeploymentPolicy.LastModifiedTime)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *continuousDeploymentPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data continuousDeploymentPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := data.ID.ValueString()
	etag, err := cdpETag(ctx, conn, id)

	if tfresource.NotFound(err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError("deleting CloudFront Continuous Deployment Policy", err.Error())
	}

	input := &cloudfront.DeleteContinuousDeploymentPolicyInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	_, err = conn.DeleteContinuousDeploymentPolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchContinuousDeploymentPolicy](err) {
		return
	}

	if errs.IsA[*awstypes.PreconditionFailed](err) || errs.IsA[*awstypes.InvalidIfMatchVersion](err) {
		etag, err = cdpETag(ctx, conn, id)

		if tfresource.NotFound(err) {
			return
		}

		if err != nil {
			response.Diagnostics.AddError("deleting CloudFront Continuous Deployment Policy", err.Error())
		}

		input.IfMatch = aws.String(etag)

		_, err = conn.DeleteContinuousDeploymentPolicy(ctx, input)

		if errs.IsA[*awstypes.NoSuchContinuousDeploymentPolicy](err) {
			return
		}
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudFront Continuous Deployment Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func cdpETag(ctx context.Context, conn *cloudfront.Client, id string) (string, error) {
	output, err := findContinuousDeploymentPolicyByID(ctx, conn, id)

	if err != nil {
		return "", fmt.Errorf("reading CloudFront Continuous Deployment Policy (%s): %w", id, err)
	}

	return aws.ToString(output.ETag), nil
}

func disableContinuousDeploymentPolicy(ctx context.Context, conn *cloudfront.Client, id string) error {
	output, err := findContinuousDeploymentPolicyByID(ctx, conn, id)

	if err != nil {
		return fmt.Errorf("reading CloudFront Continuous Deployment Policy (%s): %w", id, err)
	}

	if !aws.ToBool(output.ContinuousDeploymentPolicy.ContinuousDeploymentPolicyConfig.Enabled) {
		return nil
	}

	output.ContinuousDeploymentPolicy.ContinuousDeploymentPolicyConfig.Enabled = aws.Bool(false)

	input := &cloudfront.UpdateContinuousDeploymentPolicyInput{
		ContinuousDeploymentPolicyConfig: output.ContinuousDeploymentPolicy.ContinuousDeploymentPolicyConfig,
		Id:                               output.ContinuousDeploymentPolicy.Id,
		IfMatch:                          output.ETag,
	}

	_, err = conn.UpdateContinuousDeploymentPolicy(ctx, input)

	if err != nil {
		return fmt.Errorf("updating CloudFront Continuous Deployment Policy (%s): %w", id, err)
	}

	return err
}

func findContinuousDeploymentPolicyByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetContinuousDeploymentPolicyOutput, error) {
	input := &cloudfront.GetContinuousDeploymentPolicyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetContinuousDeploymentPolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchContinuousDeploymentPolicy](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ContinuousDeploymentPolicy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type continuousDeploymentPolicyResourceModel struct {
	Enabled                     types.Bool                                                        `tfsdk:"enabled"`
	ETag                        types.String                                                      `tfsdk:"etag"`
	ID                          types.String                                                      `tfsdk:"id"`
	LastModifiedTime            timetypes.RFC3339                                                 `tfsdk:"last_modified_time"`
	StagingDistributionDNSNames fwtypes.ListNestedObjectValueOf[stagingDistributionDNSNamesModel] `tfsdk:"staging_distribution_dns_names"`
	TrafficConfig               fwtypes.ListNestedObjectValueOf[trafficConfigModel]               `tfsdk:"traffic_config"`
}

type stagingDistributionDNSNamesModel struct {
	Items    fwtypes.SetValueOf[types.String] `tfsdk:"items"`
	Quantity types.Int64                      `tfsdk:"quantity"`
}

type trafficConfigModel struct {
	SingleHeaderConfig fwtypes.ListNestedObjectValueOf[continuousDeploymentSingleHeaderConfigModel] `tfsdk:"single_header_config"`
	SingleWeightConfig fwtypes.ListNestedObjectValueOf[continuousDeploymentSingleWeightConfigModel] `tfsdk:"single_weight_config"`
	Type               fwtypes.StringEnum[awstypes.ContinuousDeploymentPolicyType]                  `tfsdk:"type"`
}

type continuousDeploymentSingleHeaderConfigModel struct {
	Header types.String `tfsdk:"header"`
	Value  types.String `tfsdk:"value"`
}

type continuousDeploymentSingleWeightConfigModel struct {
	SessionStickinessConfig fwtypes.ListNestedObjectValueOf[sessionStickinessConfigModel] `tfsdk:"session_stickiness_config"`
	Weight                  types.Float64                                                 `tfsdk:"weight"`
}

type sessionStickinessConfigModel struct {
	IdleTTL    types.Int64 `tfsdk:"idle_ttl"`
	MaximumTTL types.Int64 `tfsdk:"maximum_ttl"`
}
