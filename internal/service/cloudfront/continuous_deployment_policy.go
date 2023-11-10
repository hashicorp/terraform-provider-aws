// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Continuous Deployment Policy")
func newResourceContinuousDeploymentPolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceContinuousDeploymentPolicy{}, nil
}

const (
	ResNameContinuousDeploymentPolicy = "Continuous Deployment Policy"
)

type resourceContinuousDeploymentPolicy struct {
	framework.ResourceWithConfigure
}

func (r *resourceContinuousDeploymentPolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_cloudfront_continuous_deployment_policy"
}

func (r *resourceContinuousDeploymentPolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Required: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			"id": framework.IDAttribute(),
			"last_modified_time": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"staging_distribution_dns_names": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"items": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
						"quantity": schema.Int64Attribute{
							Required: true,
						},
					},
				},
			},
			"traffic_config": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"single_header_config": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"header": schema.StringAttribute{
										Required: true,
									},
									"value": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"single_weight_config": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"weight": schema.Float64Attribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"session_stickiness_config": schema.ListNestedBlock{
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

func (r *resourceContinuousDeploymentPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudFrontConn(ctx)

	var plan resourceContinuousDeploymentPolicyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, d := expandContinuousDeploymentPolicyConfig(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &cloudfront.CreateContinuousDeploymentPolicyInput{
		ContinuousDeploymentPolicyConfig: cfg,
	}

	out, err := conn.CreateContinuousDeploymentPolicyWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionCreating, ResNameContinuousDeploymentPolicy, "", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ContinuousDeploymentPolicy == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionCreating, ResNameContinuousDeploymentPolicy, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ETag = flex.StringToFramework(ctx, out.ETag)
	plan.ID = flex.StringToFramework(ctx, out.ContinuousDeploymentPolicy.Id)
	plan.LastModifiedTime = flex.StringValueToFramework(ctx, out.ContinuousDeploymentPolicy.LastModifiedTime.Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceContinuousDeploymentPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudFrontConn(ctx)

	var state resourceContinuousDeploymentPolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindContinuousDeploymentPolicyByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionSetting, ResNameContinuousDeploymentPolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ETag = flex.StringToFramework(ctx, out.ETag)
	policy := out.ContinuousDeploymentPolicy
	state.ID = flex.StringToFramework(ctx, policy.Id)
	state.LastModifiedTime = flex.StringValueToFramework(ctx, policy.LastModifiedTime.Format(time.RFC3339))
	resp.Diagnostics.Append(state.refresh(ctx, policy.ContinuousDeploymentPolicyConfig)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceContinuousDeploymentPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CloudFrontConn(ctx)

	var plan, state resourceContinuousDeploymentPolicyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Enabled.Equal(state.Enabled) ||
		!plan.StagingDistributionDNSNames.Equal(state.StagingDistributionDNSNames) ||
		!plan.TrafficConfig.Equal(state.TrafficConfig) {
		cfg, d := expandContinuousDeploymentPolicyConfig(ctx, plan)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := &cloudfront.UpdateContinuousDeploymentPolicyInput{
			Id: aws.String(plan.ID.ValueString()),
			// Use state ETag value. The planned value will be unknown.
			IfMatch:                          aws.String(state.ETag.ValueString()),
			ContinuousDeploymentPolicyConfig: cfg,
		}

		out, err := conn.UpdateContinuousDeploymentPolicyWithContext(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionUpdating, ResNameContinuousDeploymentPolicy, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.ContinuousDeploymentPolicy == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionUpdating, ResNameContinuousDeploymentPolicy, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ETag = flex.StringToFramework(ctx, out.ETag)
		plan.ID = flex.StringToFramework(ctx, out.ContinuousDeploymentPolicy.Id)
		plan.LastModifiedTime = flex.StringValueToFramework(ctx, out.ContinuousDeploymentPolicy.LastModifiedTime.Format(time.RFC3339))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceContinuousDeploymentPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudFrontConn(ctx)

	var state resourceContinuousDeploymentPolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := DeleteCDP(ctx, conn, state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionDeleting, ResNameContinuousDeploymentPolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func DeleteCDP(ctx context.Context, conn *cloudfront.CloudFront, id string) error {
	etag, err := cdpETag(ctx, conn, id)
	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	in := &cloudfront.DeleteContinuousDeploymentPolicyInput{
		Id:      aws.String(id),
		IfMatch: etag,
	}

	_, err = conn.DeleteContinuousDeploymentPolicyWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchContinuousDeploymentPolicy) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodePreconditionFailed, cloudfront.ErrCodeInvalidIfMatchVersion) {
		etag, err := cdpETag(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		in.SetIfMatch(aws.StringValue(etag))

		_, err = conn.DeleteContinuousDeploymentPolicyWithContext(ctx, in)
		if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchContinuousDeploymentPolicy) {
			return nil
		}
	}

	return err
}

func disableContinuousDeploymentPolicy(ctx context.Context, conn *cloudfront.CloudFront, id string) error {
	out, err := FindContinuousDeploymentPolicyByID(ctx, conn, id)
	if tfresource.NotFound(err) || out == nil || out.ContinuousDeploymentPolicy == nil || out.ContinuousDeploymentPolicy.ContinuousDeploymentPolicyConfig == nil {
		return nil
	}

	if !aws.BoolValue(out.ContinuousDeploymentPolicy.ContinuousDeploymentPolicyConfig.Enabled) {
		return nil
	}

	out.ContinuousDeploymentPolicy.ContinuousDeploymentPolicyConfig.SetEnabled(false)

	in := &cloudfront.UpdateContinuousDeploymentPolicyInput{
		Id:                               out.ContinuousDeploymentPolicy.Id,
		IfMatch:                          out.ETag,
		ContinuousDeploymentPolicyConfig: out.ContinuousDeploymentPolicy.ContinuousDeploymentPolicyConfig,
	}

	_, err = conn.UpdateContinuousDeploymentPolicyWithContext(ctx, in)
	return err
}

func cdpETag(ctx context.Context, conn *cloudfront.CloudFront, id string) (*string, error) {
	output, err := FindContinuousDeploymentPolicyByID(ctx, conn, id)
	if err != nil {
		return nil, err
	}

	return output.ETag, nil
}

func (r *resourceContinuousDeploymentPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func FindContinuousDeploymentPolicyByID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.GetContinuousDeploymentPolicyOutput, error) {
	in := &cloudfront.GetContinuousDeploymentPolicyInput{
		Id: aws.String(id),
	}

	out, err := conn.GetContinuousDeploymentPolicyWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchContinuousDeploymentPolicy) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.ContinuousDeploymentPolicy == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenStagingDistributionDNSNames(ctx context.Context, apiObject *cloudfront.StagingDistributionDnsNames) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: stagingDistributionDNSNamesAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"items":    flex.FlattenFrameworkStringSet(ctx, apiObject.Items),
		"quantity": flex.Int64ToFramework(ctx, apiObject.Quantity),
	}
	objVal, d := types.ObjectValue(stagingDistributionDNSNamesAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenTrafficConfig(ctx context.Context, apiObject *cloudfront.TrafficConfig) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: trafficConfigAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	singleHeaderConfig, d := flattenSingleHeaderConfig(ctx, apiObject.SingleHeaderConfig)
	diags.Append(d...)

	singleWeightConfig, d := flattenSingleWeightConfig(ctx, apiObject.SingleWeightConfig)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"type":                 flex.StringToFramework(ctx, apiObject.Type),
		"single_header_config": singleHeaderConfig,
		"single_weight_config": singleWeightConfig,
	}
	objVal, d := types.ObjectValue(trafficConfigAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenSingleHeaderConfig(ctx context.Context, apiObject *cloudfront.ContinuousDeploymentSingleHeaderConfig) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: singleHeaderConfigAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"header": flex.StringToFramework(ctx, apiObject.Header),
		"value":  flex.StringToFramework(ctx, apiObject.Value),
	}
	objVal, d := types.ObjectValue(singleHeaderConfigAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenSingleWeightConfig(ctx context.Context, apiObject *cloudfront.ContinuousDeploymentSingleWeightConfig) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: singleWeightConfigAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	sessionStickinessConfig, d := flattenSessionStickinessConfig(ctx, apiObject.SessionStickinessConfig)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"session_stickiness_config": sessionStickinessConfig,
		"weight":                    flex.Float64ToFramework(ctx, apiObject.Weight),
	}
	objVal, d := types.ObjectValue(singleWeightConfigAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenSessionStickinessConfig(ctx context.Context, apiObject *cloudfront.SessionStickinessConfig) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: sessionStickinessConfigAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"idle_ttl":    flex.Int64ToFramework(ctx, apiObject.IdleTTL),
		"maximum_ttl": flex.Int64ToFramework(ctx, apiObject.MaximumTTL),
	}
	objVal, d := types.ObjectValue(sessionStickinessConfigAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

// expandContinuousDeploymentPolicyConfig translates a resource plan into a
// continuous deployment policy config
func expandContinuousDeploymentPolicyConfig(ctx context.Context, data resourceContinuousDeploymentPolicyData) (*cloudfront.ContinuousDeploymentPolicyConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var stagingDistributionDNSNames []stagingDistributionDNSNamesData
	diags.Append(data.StagingDistributionDNSNames.ElementsAs(ctx, &stagingDistributionDNSNames, false)...)

	apiObject := &cloudfront.ContinuousDeploymentPolicyConfig{
		Enabled:                     aws.Bool(data.Enabled.ValueBool()),
		StagingDistributionDnsNames: expandStagingDistributionDNSNames(ctx, stagingDistributionDNSNames),
	}
	if !data.TrafficConfig.IsNull() {
		var tcData []trafficConfigData
		diags.Append(data.TrafficConfig.ElementsAs(ctx, &tcData, false)...)

		trafficConfig, d := expandTrafficConfig(ctx, tcData)
		diags.Append(d...)

		apiObject.TrafficConfig = trafficConfig
	}

	return apiObject, diags
}

func expandStagingDistributionDNSNames(ctx context.Context, tfList []stagingDistributionDNSNamesData) *cloudfront.StagingDistributionDnsNames {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &cloudfront.StagingDistributionDnsNames{
		Quantity: aws.Int64(tfObj.Quantity.ValueInt64()),
		Items:    flex.ExpandFrameworkStringSet(ctx, tfObj.Items),
	}

	return apiObject
}

func expandTrafficConfig(ctx context.Context, tfList []trafficConfigData) (*cloudfront.TrafficConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(tfList) == 0 {
		return nil, diags
	}

	tfObj := tfList[0]
	apiObject := &cloudfront.TrafficConfig{
		Type: aws.String(tfObj.Type.ValueString()),
	}
	if !tfObj.SingleHeaderConfig.IsNull() {
		var data []singleHeaderConfigData
		diags.Append(tfObj.SingleHeaderConfig.ElementsAs(ctx, &data, false)...)

		apiObject.SingleHeaderConfig = expandSingleHeaderConfig(data)
	}
	if !tfObj.SingleWeightConfig.IsNull() {
		var data []singleWeightConfigData
		diags.Append(tfObj.SingleWeightConfig.ElementsAs(ctx, &data, false)...)

		singleWeightConfig, d := expandSingleWeightConfig(ctx, data)
		diags.Append(d...)

		apiObject.SingleWeightConfig = singleWeightConfig
	}

	return apiObject, diags
}

func expandSingleHeaderConfig(tfList []singleHeaderConfigData) *cloudfront.ContinuousDeploymentSingleHeaderConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &cloudfront.ContinuousDeploymentSingleHeaderConfig{
		Header: aws.String(tfObj.Header.ValueString()),
		Value:  aws.String(tfObj.Value.ValueString()),
	}

	return apiObject
}

func expandSingleWeightConfig(ctx context.Context, tfList []singleWeightConfigData) (*cloudfront.ContinuousDeploymentSingleWeightConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(tfList) == 0 {
		return nil, diags
	}

	tfObj := tfList[0]
	apiObject := &cloudfront.ContinuousDeploymentSingleWeightConfig{
		Weight: aws.Float64(tfObj.Weight.ValueFloat64()),
	}
	if !tfObj.SessionStickinessConfig.IsNull() {
		var data []sessionStickinessConfigData
		diags.Append(tfObj.SessionStickinessConfig.ElementsAs(ctx, &data, false)...)

		apiObject.SessionStickinessConfig = expandSessionStickinessConfig(data)
	}

	return apiObject, diags
}

func expandSessionStickinessConfig(tfList []sessionStickinessConfigData) *cloudfront.SessionStickinessConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &cloudfront.SessionStickinessConfig{
		IdleTTL:    aws.Int64(tfObj.IdleTTL.ValueInt64()),
		MaximumTTL: aws.Int64(tfObj.MaximumTTL.ValueInt64()),
	}

	return apiObject
}

type resourceContinuousDeploymentPolicyData struct {
	Enabled                     types.Bool   `tfsdk:"enabled"`
	ETag                        types.String `tfsdk:"etag"`
	ID                          types.String `tfsdk:"id"`
	LastModifiedTime            types.String `tfsdk:"last_modified_time"`
	StagingDistributionDNSNames types.List   `tfsdk:"staging_distribution_dns_names"`
	TrafficConfig               types.List   `tfsdk:"traffic_config"`
}

type stagingDistributionDNSNamesData struct {
	Items    types.Set   `tfsdk:"items"`
	Quantity types.Int64 `tfsdk:"quantity"`
}

type trafficConfigData struct {
	SingleHeaderConfig types.List   `tfsdk:"single_header_config"`
	SingleWeightConfig types.List   `tfsdk:"single_weight_config"`
	Type               types.String `tfsdk:"type"`
}

type singleHeaderConfigData struct {
	Header types.String `tfsdk:"header"`
	Value  types.String `tfsdk:"value"`
}

type singleWeightConfigData struct {
	SessionStickinessConfig types.List    `tfsdk:"session_stickiness_config"`
	Weight                  types.Float64 `tfsdk:"weight"`
}

type sessionStickinessConfigData struct {
	IdleTTL    types.Int64 `tfsdk:"idle_ttl"`
	MaximumTTL types.Int64 `tfsdk:"maximum_ttl"`
}

var stagingDistributionDNSNamesAttrTypes = map[string]attr.Type{
	"items":    types.SetType{ElemType: types.StringType},
	"quantity": types.Int64Type,
}

var trafficConfigAttrTypes = map[string]attr.Type{
	"single_header_config": types.ListType{ElemType: types.ObjectType{AttrTypes: singleHeaderConfigAttrTypes}},
	"single_weight_config": types.ListType{ElemType: types.ObjectType{AttrTypes: singleWeightConfigAttrTypes}},
	"type":                 types.StringType,
}

var singleHeaderConfigAttrTypes = map[string]attr.Type{
	"header": types.StringType,
	"value":  types.StringType,
}

var singleWeightConfigAttrTypes = map[string]attr.Type{
	"session_stickiness_config": types.ListType{ElemType: types.ObjectType{AttrTypes: sessionStickinessConfigAttrTypes}},
	"weight":                    types.Float64Type,
}

var sessionStickinessConfigAttrTypes = map[string]attr.Type{
	"idle_ttl":    types.Int64Type,
	"maximum_ttl": types.Int64Type,
}

// refresh updates state data from the returned API response
func (rd *resourceContinuousDeploymentPolicyData) refresh(ctx context.Context, apiObject *cloudfront.ContinuousDeploymentPolicyConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if apiObject == nil {
		return diags
	}

	rd.Enabled = flex.BoolToFramework(ctx, apiObject.Enabled)

	stagingDistributionDNSNames, d := flattenStagingDistributionDNSNames(ctx, apiObject.StagingDistributionDnsNames)
	diags.Append(d...)
	rd.StagingDistributionDNSNames = stagingDistributionDNSNames

	trafficConfig, d := flattenTrafficConfig(ctx, apiObject.TrafficConfig)
	diags.Append(d...)
	rd.TrafficConfig = trafficConfig

	return diags
}
