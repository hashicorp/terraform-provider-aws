// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// @FrameworkResource("aws_cloudfront_connection_function", name="Connection Function")
// @Tags(identifierAttribute="connection_function_arn")
func newResourceConnectionFunction(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &connectionFunctionResource{}
	return r, nil
}

type connectionFunctionResource struct {
	framework.ResourceWithModel[connectionFunctionResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *connectionFunctionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"connection_function_arn": framework.ARNAttributeComputedOnly(),
			"connection_function_code": schema.StringAttribute{
				Required: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"live_stage_etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9-_]{1,64}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"publish": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"connection_function_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[functionConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrComment: schema.StringAttribute{
							Required: true,
						},
						"runtime": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.FunctionRuntime](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"key_value_store_association": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[keyValueStoreAssociationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"key_value_store_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
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

func (r *connectionFunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionFunctionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input cloudfront.CreateConnectionFunctionInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = &awstypes.Tags{
			Items: tags,
		}
	}

	outputCCF, err := conn.CreateConnectionFunction(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating CloudFront Connection Function (%s)", name), err.Error())
		return
	}

	connectionFunctionSummary := outputCCF.ConnectionFunctionSummary
	id, etag := aws.ToString(connectionFunctionSummary.Id), aws.ToString(outputCCF.ETag)

	if data.Publish.ValueBool() {
		if err := publishConnectionFunction(ctx, conn, id, etag); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("publishing CloudFront Connection Function (%s)", id), err.Error())
			return
		}

		outputDCF, err := findConnectionFunctionByTwoPartKey(ctx, conn, id, awstypes.FunctionStageLive)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Connection Function (%s) LIVE stage", id), err.Error())
			return
		}

		data.LiveStageEtag = fwflex.StringToFramework(ctx, outputDCF.ETag)
		connectionFunctionSummary = outputDCF.ConnectionFunctionSummary
	} else {
		data.LiveStageEtag = types.StringNull()
	}

	// Set values for unknowns.
	data.ConnectionFunctionARN = fwflex.StringToFramework(ctx, connectionFunctionSummary.ConnectionFunctionArn)
	data.Etag = fwflex.StringValueToFramework(ctx, etag)
	data.ID = fwflex.StringValueToFramework(ctx, id)
	data.Status = fwflex.StringToFramework(ctx, connectionFunctionSummary.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *connectionFunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionFunctionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	outputDCF, err := findConnectionFunctionByTwoPartKey(ctx, conn, id, awstypes.FunctionStageDevelopment)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Connection Function (%s) DEVELOPMENT stage", id), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, outputDCF.ConnectionFunctionSummary, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cloudfront.GetConnectionFunctionInput{
		Identifier: aws.String(id),
		Stage:      awstypes.FunctionStageDevelopment,
	}
	outputGCF, err := conn.GetConnectionFunction(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Connection Function (%s) DEVELOPMENT stage code", id), err.Error())
		return
	}

	data.ConnectionFunctionCode = fwflex.StringValueToFramework(ctx, outputGCF.ConnectionFunctionCode)
	data.Etag = fwflex.StringToFramework(ctx, outputGCF.ETag)

	outputDCF, err = findConnectionFunctionByTwoPartKey(ctx, conn, id, awstypes.FunctionStageLive)
	switch {
	case retry.NotFound(err):
		data.LiveStageEtag = types.StringNull()
	case err != nil:
		resp.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Connection Function (%s) LIVE stage", id), err.Error())
		return
	default:
		data.LiveStageEtag = fwflex.StringToFramework(ctx, outputDCF.ETag)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionFunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new connectionFunctionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		id, etag := fwflex.StringValueFromFramework(ctx, new.ID), fwflex.StringValueFromFramework(ctx, old.Etag)
		var input cloudfront.UpdateConnectionFunctionInput
		resp.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.IfMatch = aws.String(etag)

		outputUCF, err := conn.UpdateConnectionFunction(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating CloudFront Connection Function (%s)", id), err.Error())
			return
		}

		connectionFunctionSummary := outputUCF.ConnectionFunctionSummary
		etag = aws.ToString(outputUCF.ETag)

		if new.Publish.ValueBool() {
			if err := publishConnectionFunction(ctx, conn, id, etag); err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("publishing CloudFront Connection Function (%s)", id), err.Error())
				return
			}

			outputDCF, err := findConnectionFunctionByTwoPartKey(ctx, conn, id, awstypes.FunctionStageLive)
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Connection Function (%s) LIVE stage", id), err.Error())
				return
			}

			new.LiveStageEtag = fwflex.StringToFramework(ctx, outputDCF.ETag)
			connectionFunctionSummary = outputDCF.ConnectionFunctionSummary
		} else {
			new.LiveStageEtag = old.LiveStageEtag
		}

		// Set values for unknowns.
		new.Etag = fwflex.StringValueToFramework(ctx, etag)
		new.Status = fwflex.StringToFramework(ctx, connectionFunctionSummary.Status)
	} else {
		new.Etag = old.Etag
		new.LiveStageEtag = old.LiveStageEtag
		new.Status = old.Status
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *connectionFunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionFunctionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id, etag := fwflex.StringValueFromFramework(ctx, data.ID), fwflex.StringValueFromFramework(ctx, data.Etag)
	input := cloudfront.DeleteConnectionFunctionInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	_, err := conn.DeleteConnectionFunction(ctx, &input)
	if errs.IsA[*awstypes.EntityNotFound](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting CloudFront Connection Function (%s)", id), err.Error())
		return
	}
}

func publishConnectionFunction(ctx context.Context, conn *cloudfront.Client, id, etag string) error {
	input := cloudfront.PublishConnectionFunctionInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}
	_, err := conn.PublishConnectionFunction(ctx, &input)

	if err != nil {
		return err
	}

	const (
		timeout = 5 * time.Minute
	)
	if _, err := waitConnectionFunctionPublished(ctx, conn, id, awstypes.FunctionStageDevelopment, timeout); err != nil {
		return fmt.Errorf("waiting for CloudFront Connection Function (%s) publish: %w", id, err)
	}

	return err
}

func findConnectionFunctionByTwoPartKey(ctx context.Context, conn *cloudfront.Client, id string, stage awstypes.FunctionStage) (*cloudfront.DescribeConnectionFunctionOutput, error) {
	input := cloudfront.DescribeConnectionFunctionInput{
		Identifier: aws.String(id),
		Stage:      stage,
	}

	return findConnectionFunction(ctx, conn, &input)
}

func findConnectionFunction(ctx context.Context, conn *cloudfront.Client, input *cloudfront.DescribeConnectionFunctionInput) (*cloudfront.DescribeConnectionFunctionOutput, error) {
	output, err := conn.DescribeConnectionFunction(ctx, input)
	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ConnectionFunctionSummary == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusConnectionFunction(conn *cloudfront.Client, id string, stage awstypes.FunctionStage) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findConnectionFunctionByTwoPartKey(ctx, conn, id, stage)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.ConnectionFunctionSummary.Status), nil
	}
}

func waitConnectionFunctionPublished(ctx context.Context, conn *cloudfront.Client, id string, stage awstypes.FunctionStage, timeout time.Duration) (*cloudfront.DescribeConnectionFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{connectionFunctionStatusPublishing},
		Target:  []string{connectionFunctionStatusUnassociated},
		Refresh: statusConnectionFunction(conn, id, stage),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.DescribeConnectionFunctionOutput); ok {
		return output, err
	}

	return nil, err
}

type connectionFunctionResourceModel struct {
	ConnectionFunctionARN    types.String                                         `tfsdk:"connection_function_arn"`
	ConnectionFunctionCode   types.String                                         `tfsdk:"connection_function_code"`
	ConnectionFunctionConfig fwtypes.ListNestedObjectValueOf[functionConfigModel] `tfsdk:"connection_function_config"`
	Etag                     types.String                                         `tfsdk:"etag"`
	ID                       types.String                                         `tfsdk:"id"`
	LiveStageEtag            types.String                                         `tfsdk:"live_stage_etag"`
	Name                     types.String                                         `tfsdk:"name"`
	Publish                  types.Bool                                           `tfsdk:"publish"`
	Status                   types.String                                         `tfsdk:"status"`
	Tags                     tftags.Map                                           `tfsdk:"tags"`
	TagsAll                  tftags.Map                                           `tfsdk:"tags_all"`
}

type functionConfigModel struct {
	Comment                   types.String                                                   `tfsdk:"comment"`
	KeyValueStoreAssociations fwtypes.ListNestedObjectValueOf[keyValueStoreAssociationModel] `tfsdk:"key_value_store_association" autoflex:",xmlwrapper=Items"`
	Runtime                   fwtypes.StringEnum[awstypes.FunctionRuntime]                   `tfsdk:"runtime"`
}

type keyValueStoreAssociationModel struct {
	KeyValueStoreARN fwtypes.ARN `tfsdk:"key_value_store_arn"`
}
