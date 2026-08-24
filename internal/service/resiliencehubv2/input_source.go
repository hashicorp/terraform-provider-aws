// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehubv2_input_source", name="Input Source")
// @IdentityAttribute("service_arn")
// @IdentityAttribute("input_source_id")
// @ImportIDHandler("inputSourceImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.InputSourceSummary")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccCheckInputSourceImportStateIDFunc")
// @Testing(importStateIdAttribute="input_source_id")
func newInputSourceResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &inputSourceResource{}, nil
}

type inputSourceResource struct {
	framework.ResourceWithModel[inputSourceResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
}

func (r *inputSourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"input_source_id": framework.IDAttribute(),
			"service_arn": fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]fwschema.Block{
			"resource_configuration": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: fwschema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("cfn_stack_arn"),
							path.MatchRelative().AtName("design_file_s3_url"),
							path.MatchRelative().AtName("eks"),
							path.MatchRelative().AtName("resource_tag"),
							path.MatchRelative().AtName("tf_state_file_url"),
						),
					},
					Attributes: map[string]fwschema.Attribute{
						"cfn_stack_arn": fwschema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"design_file_s3_url": fwschema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"tf_state_file_url": fwschema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]fwschema.Block{
						"eks": fwschema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[eksSourceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: fwschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"cluster_arn": fwschema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"namespaces": fwschema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
										Validators: []validator.Set{
											setvalidator.SizeBetween(1, 10),
											setvalidator.ValueStringsAre(
												stringvalidator.LengthBetween(1, 63),
											),
										},
									},
								},
							},
						},
						"resource_tag": fwschema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[resourceTagModel](ctx),
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 10),
							},
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							NestedObject: fwschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									names.AttrKey: fwschema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrValues: fwschema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
										Validators: []validator.Set{
											setvalidator.SizeBetween(0, 10),
											setvalidator.ValueStringsAre(
												stringvalidator.LengthBetween(0, 256),
											),
										},
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
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

func (r *inputSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan inputSourceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	var input resiliencehubv2.CreateInputSourceInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	output, err := tfresource.RetryWhenIsAErrorMessageContains[*resiliencehubv2.CreateInputSourceOutput, *awstypes.AccessDeniedException](ctx, propagationTimeout, func(ctx context.Context) (*resiliencehubv2.CreateInputSourceOutput, error) {
		return conn.CreateInputSource(ctx, &input)
	}, "The invoker role does not have access to")
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	// Set values for unknowns.
	plan.InputSourceID = fwflex.StringToFramework(ctx, output.InputSourceId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *inputSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state inputSourceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	serviceARN, inputSourceID := fwflex.StringValueFromFramework(ctx, state.ServiceARN), fwflex.StringValueFromFramework(ctx, state.InputSourceID)
	is, err := findInputSourceByTwoPartKey(ctx, conn, serviceARN, inputSourceID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, inputSourceID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, is, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *inputSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state inputSourceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	serviceARN, inputSourceID := fwflex.StringValueFromFramework(ctx, state.ServiceARN), fwflex.StringValueFromFramework(ctx, state.InputSourceID)
	input := resiliencehubv2.DeleteInputSourceInput{
		InputSourceId: aws.String(inputSourceID),
		ServiceArn:    aws.String(serviceARN),
	}
	_, err := conn.DeleteInputSource(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, inputSourceID)
	}
}

func (r *inputSourceResource) flatten(ctx context.Context, is *awstypes.InputSourceSummary, data *inputSourceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, is, data)...)
	if diags.HasError() {
		return diags
	}

	resourceConfiguration, d := fwtypes.Nullified[resourceConfigurationModel](ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch is.Type {
	case awstypes.InputSourceTypeCfnStack:
		resourceConfiguration.CFNStackARN = fwflex.StringToFrameworkARN(ctx, is.CfnStackArn)

	case awstypes.InputSourceTypeDesignFile:
		resourceConfiguration.DesignFileS3URL = fwflex.StringToFramework(ctx, is.DesignFileS3Url)

	case awstypes.InputSourceTypeEks:
		diags.Append(fwflex.Flatten(ctx, is.Eks, &resourceConfiguration.EKS)...)
		if diags.HasError() {
			return diags
		}

	case awstypes.InputSourceTypeTags:
		diags.Append(fwflex.Flatten(ctx, is.ResourceTags, &resourceConfiguration.ResourceTags)...)
		if diags.HasError() {
			return diags
		}

	case awstypes.InputSourceTypeTerraform:
		resourceConfiguration.TFStateFileURL = fwflex.StringToFramework(ctx, is.TfStateFileUrl)
	}

	data.ResourceConfiguration, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &resourceConfiguration)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	return diags
}

var (
	_ inttypes.ImportIDParser = inputSourceImportID{}
)

type inputSourceImportID struct{}

func (inputSourceImportID) Parse(id string) (string, map[string]any, error) {
	const (
		inputSourceImportIDPartCount = 2
	)
	parts, err := flex.ExpandResourceId(id, inputSourceImportIDPartCount, false)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"input_source_id": parts[1],
		"service_arn":     parts[0],
	}

	return id, result, nil
}

func findInputSourceByTwoPartKey(ctx context.Context, conn *resiliencehubv2.Client, serviceARN, inputSourceID string) (*awstypes.InputSourceSummary, error) {
	input := resiliencehubv2.ListInputSourcesInput{
		ServiceArn: aws.String(serviceARN),
	}

	return findInputSource(ctx, conn, &input, tfslices.WithFilter(func(v awstypes.InputSourceSummary) bool {
		return aws.ToString(v.InputSourceId) == inputSourceID
	}))
}

func findInputSource(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListInputSourcesInput, optFns ...tfslices.FinderOptionsFunc[awstypes.InputSourceSummary]) (*awstypes.InputSourceSummary, error) {
	output, err := findInputSources(ctx, conn, input, optFns...)

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findInputSources(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListInputSourcesInput, optFns ...tfslices.FinderOptionsFunc[awstypes.InputSourceSummary]) ([]awstypes.InputSourceSummary, error) {
	output, err := tfslices.CollectAndConcatWithError(listInputSourcePages(ctx, conn, input), optFns...)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	return output, nil
}

func listInputSourcePages(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListInputSourcesInput, optFns ...func(*resiliencehubv2.Options)) iter.Seq2[[]awstypes.InputSourceSummary, error] {
	return func(yield func([]awstypes.InputSourceSummary, error) bool) {
		pages := resiliencehubv2.NewListInputSourcesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Resilience Hub V2 Input Sources: %w", err))
				return
			}

			if !yield(page.InputSourceSummaries, nil) {
				return
			}
		}
	}
}

type inputSourceResourceModel struct {
	framework.WithRegionModel
	InputSourceID         types.String                                                `tfsdk:"input_source_id"`
	ResourceConfiguration fwtypes.ListNestedObjectValueOf[resourceConfigurationModel] `tfsdk:"resource_configuration" autoflex:",noflatten"`
	ServiceARN            fwtypes.ARN                                                 `tfsdk:"service_arn"`
}

type resourceConfigurationModel struct {
	CFNStackARN     fwtypes.ARN                                      `tfsdk:"cfn_stack_arn"`
	DesignFileS3URL types.String                                     `tfsdk:"design_file_s3_url"`
	EKS             fwtypes.ListNestedObjectValueOf[eksSourceModel]  `tfsdk:"eks"`
	ResourceTags    fwtypes.SetNestedObjectValueOf[resourceTagModel] `tfsdk:"resource_tag"`
	TFStateFileURL  types.String                                     `tfsdk:"tf_state_file_url"`
}

var (
	_ fwflex.Expander = resourceConfigurationModel{}
)

func (m *resourceConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.ResourceConfigurationMemberCfnStackArn:
		m.CFNStackARN = fwtypes.ARNValue(t.Value)

	case awstypes.ResourceConfigurationMemberDesignFileS3Url:
		m.DesignFileS3URL = fwflex.StringValueToFramework(ctx, t.Value)

	case awstypes.ResourceConfigurationMemberEks:
		var data eksSourceModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.EKS = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	case awstypes.ResourceConfigurationMemberResourceTags:
		var data resourceTagModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.ResourceTags = fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &data)

	case awstypes.ResourceConfigurationMemberTfStateFileUrl:
		m.TFStateFileURL = fwflex.StringValueToFramework(ctx, t.Value)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("resourceConfigurationModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m resourceConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CFNStackARN.IsNull():
		var r awstypes.ResourceConfigurationMemberCfnStackArn
		r.Value = fwflex.StringValueFromFramework(ctx, m.CFNStackARN)
		return &r, diags

	case !m.DesignFileS3URL.IsNull():
		var r awstypes.ResourceConfigurationMemberDesignFileS3Url
		r.Value = fwflex.StringValueFromFramework(ctx, m.DesignFileS3URL)
		return &r, diags

	case !m.EKS.IsNull():
		data, d := m.EKS.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ResourceConfigurationMemberEks
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags

	case !m.ResourceTags.IsNull():
		var r awstypes.ResourceConfigurationMemberResourceTags
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.ResourceTags, &r.Value))
		return &r, diags

	case !m.TFStateFileURL.IsNull():
		var r awstypes.ResourceConfigurationMemberTfStateFileUrl
		r.Value = fwflex.StringValueFromFramework(ctx, m.TFStateFileURL)
		return &r, diags
	}

	return nil, diags
}

type eksSourceModel struct {
	ClusterARN fwtypes.ARN         `tfsdk:"cluster_arn"`
	Namespaces fwtypes.SetOfString `tfsdk:"namespaces"`
}

type resourceTagModel struct {
	Key    types.String        `tfsdk:"key"`
	Values fwtypes.SetOfString `tfsdk:"values"`
}
