// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudfront_vpc_origin", name="VPC Origin")
// @Tags(identifierAttribute="arn")
func newVPCOriginResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &vpcOriginResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultUpdateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

type vpcOriginResource struct {
	framework.ResourceWithModel[vpcOriginResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *vpcOriginResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"vpc_origin_endpoint_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vpcOriginEndpointConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.ARNType,
						},
						"http_port": schema.Int64Attribute{
							Required: true,
						},
						"https_port": schema.Int64Attribute{
							Required: true,
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						"origin_protocol_policy": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.OriginProtocolPolicy](),
						},
					},
					Blocks: map[string]schema.Block{
						"origin_ssl_protocols": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[originSSLProtocolsModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"items": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringEnumType[awstypes.SslProtocol](),
										Required:   true,
									},
									"quantity": schema.Int64Attribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *vpcOriginResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data vpcOriginResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	input := &cloudfront.CreateVpcOriginInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = &awstypes.Tags{
			Items: tags,
		}
	}

	output, err := conn.CreateVpcOrigin(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating CloudFront VPC Origin", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.VpcOrigin.Id)
	data.ARN = fwflex.StringToFramework(ctx, output.VpcOrigin.Arn)
	data.ETag = fwflex.StringToFramework(ctx, output.ETag)

	if _, err := waitVPCOriginDeployed(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudFront VPC Origin (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcOriginResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data vpcOriginResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	output, err := findVPCOriginByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront VPC Origin (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.VpcOrigin, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ETag = fwflex.StringToFramework(ctx, output.ETag)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *vpcOriginResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new vpcOriginResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	if !new.VPCOriginEndpointConfig.Equal(old.VPCOriginEndpointConfig) {
		input := &cloudfront.UpdateVpcOriginInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		input.Id = new.ID.ValueStringPointer()
		// Use state ETag value. The planned value will be unknown.
		input.IfMatch = old.ETag.ValueStringPointer()

		_, err := conn.UpdateVpcOrigin(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudFront VPC Origin (%s)", new.ID.ValueString()), err.Error())

			return
		}

		output, err := waitVPCOriginDeployed(ctx, conn, old.ID.ValueString(), r.UpdateTimeout(ctx, old.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudFront VPC Origin (%s) update", new.ID.ValueString()), err.Error())

			return
		}

		new.ETag = fwflex.StringToFramework(ctx, output.ETag)
	} else {
		new.ETag = old.ETag
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *vpcOriginResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data vpcOriginResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := data.ID.ValueString()
	etag, err := vpcOriginETag(ctx, conn, id)

	if tfresource.NotFound(err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront VPC Origin (%s)", data.ID.ValueString()), err.Error())

		return
	}

	input := &cloudfront.DeleteVpcOriginInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	_, err = conn.DeleteVpcOrigin(ctx, input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return
	}

	if errs.IsA[*awstypes.PreconditionFailed](err) || errs.IsA[*awstypes.InvalidIfMatchVersion](err) {
		etag, err = vpcOriginETag(ctx, conn, id)

		if tfresource.NotFound(err) {
			return
		}

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront VPC Origin (%s)", data.ID.ValueString()), err.Error())

			return
		}

		input.IfMatch = aws.String(etag)

		_, err = conn.DeleteVpcOrigin(ctx, input)

		if errs.IsA[*awstypes.EntityNotFound](err) {
			return
		}
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudFront VPC Origin (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitVPCOriginDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudFront VPC Origin (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func vpcOriginETag(ctx context.Context, conn *cloudfront.Client, id string) (string, error) {
	output, err := findVPCOriginByID(ctx, conn, id)

	if err != nil {
		return "", err
	}

	return aws.ToString(output.ETag), nil
}

func findVPCOriginByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetVpcOriginOutput, error) {
	input := &cloudfront.GetVpcOriginInput{
		Id: aws.String(id),
	}

	return findVPCOrigin(ctx, conn, input)
}

func findVPCOrigin(ctx context.Context, conn *cloudfront.Client, input *cloudfront.GetVpcOriginInput) (*cloudfront.GetVpcOriginOutput, error) {
	output, err := conn.GetVpcOrigin(ctx, input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VpcOrigin == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func vpcOriginStatus(ctx context.Context, conn *cloudfront.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findVPCOriginByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.VpcOrigin.Status), nil
	}
}

func waitVPCOriginDeployed(ctx context.Context, conn *cloudfront.Client, id string, timeout time.Duration) (*cloudfront.GetVpcOriginOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{vpcOriginStatusDeploying},
		Target:  []string{vpcOriginStatusDeployed},
		Refresh: vpcOriginStatus(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetVpcOriginOutput); ok {
		return output, err
	}

	return nil, err
}

func waitVPCOriginDeleted(ctx context.Context, conn *cloudfront.Client, id string, timeout time.Duration) (*cloudfront.GetVpcOriginOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{vpcOriginStatusDeployed, vpcOriginStatusDeploying},
		Target:  []string{},
		Refresh: vpcOriginStatus(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetVpcOriginOutput); ok {
		return output, err
	}

	return nil, err
}

type vpcOriginResourceModel struct {
	ARN                     types.String                                                  `tfsdk:"arn"`
	ETag                    types.String                                                  `tfsdk:"etag"`
	ID                      types.String                                                  `tfsdk:"id"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
	VPCOriginEndpointConfig fwtypes.ListNestedObjectValueOf[vpcOriginEndpointConfigModel] `tfsdk:"vpc_origin_endpoint_config"`
}

type vpcOriginEndpointConfigModel struct {
	ARN                  types.String                                             `tfsdk:"arn"`
	HTTPPort             types.Int64                                              `tfsdk:"http_port"`
	HTTPSPort            types.Int64                                              `tfsdk:"https_port"`
	Name                 types.String                                             `tfsdk:"name"`
	OriginProtocolPolicy fwtypes.StringEnum[awstypes.OriginProtocolPolicy]        `tfsdk:"origin_protocol_policy"`
	OriginSSLProtocols   fwtypes.ListNestedObjectValueOf[originSSLProtocolsModel] `tfsdk:"origin_ssl_protocols"`
}

type originSSLProtocolsModel struct {
	Items    fwtypes.SetOfStringEnum[awstypes.SslProtocol] `tfsdk:"items"`
	Quantity types.Int64                                   `tfsdk:"quantity"`
}

var (
	_ fwflex.Expander  = vpcOriginEndpointConfigModel{}
	_ fwflex.Flattener = &vpcOriginEndpointConfigModel{}
)

func (m vpcOriginEndpointConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	config := &awstypes.VpcOriginEndpointConfig{
		Arn:                  fwflex.StringFromFramework(ctx, m.ARN),
		HTTPPort:             aws.Int32(int32(m.HTTPPort.ValueInt64())),
		HTTPSPort:            aws.Int32(int32(m.HTTPSPort.ValueInt64())),
		Name:                 fwflex.StringFromFramework(ctx, m.Name),
		OriginProtocolPolicy: m.OriginProtocolPolicy.ValueEnum(),
	}

	if !m.OriginSSLProtocols.IsNull() {
		sslList, d := m.OriginSSLProtocols.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		if len(sslList) > 0 {
			expanded, d := sslList[0].Expand(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			if sslProto, ok := expanded.(*awstypes.OriginSslProtocols); ok {
				config.OriginSslProtocols = sslProto
			} else {
				diags.AddError("error expanding Origin SSL Protocols", fmt.Sprintf("expected OriginSslProtocols, got: %T", expanded))
			}
		}
	}

	return config, diags
}

func (m *vpcOriginEndpointConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil {
		return diags
	}
	if t, ok := v.(awstypes.VpcOriginEndpointConfig); ok {
		m.ARN = fwflex.StringToFramework(ctx, t.Arn)
		m.HTTPPort = types.Int64Value(int64(aws.ToInt32(t.HTTPPort)))
		m.HTTPSPort = types.Int64Value(int64(aws.ToInt32(t.HTTPSPort)))
		m.Name = fwflex.StringToFramework(ctx, t.Name)
		m.OriginProtocolPolicy = fwtypes.StringEnumValue(t.OriginProtocolPolicy)

		if t.OriginSslProtocols != nil {
			var sslModel originSSLProtocolsModel
			diags.Append(fwflex.Flatten(ctx, t.OriginSslProtocols, &sslModel)...)
			if diags.HasError() {
				return diags
			}
			sslList, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &sslModel)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			m.OriginSSLProtocols = sslList
		}
	} else {
		diags.AddError("error flattening VPC Origin Endpoint Config", fmt.Sprintf("expected VpcOriginEndpointConfig, got: %T", v))
	}
	return diags
}

var (
	_ fwflex.Expander  = originSSLProtocolsModel{}
	_ fwflex.Flattener = &originSSLProtocolsModel{}
)

func (m originSSLProtocolsModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	var protocols []awstypes.SslProtocol
	diags.Append(fwflex.Expand(ctx, m.Items, &protocols)...)
	if diags.HasError() {
		return nil, diags
	}

	result := awstypes.OriginSslProtocols{
		Items:    protocols,
		Quantity: aws.Int32(int32(m.Quantity.ValueInt64())),
	}

	return &result, diags
}

func (m *originSSLProtocolsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil {
		return diags
	}

	if t, ok := v.(awstypes.OriginSslProtocols); ok {
		diags.Append(fwflex.Flatten(ctx, t.Items, &m.Items)...)
		if diags.HasError() {
			return diags
		}

		m.Quantity = types.Int64Value(int64(aws.ToInt32(t.Quantity)))
	} else {
		diags.AddError("error flattening Origin SSL Protocols", fmt.Sprintf("expected OriginSslProtocols, got: %T", v))
	}
	return diags
}
