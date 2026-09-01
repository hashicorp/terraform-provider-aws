// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ecr

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ecr_lifecycle_policy_document", name="Lifecycle Policy Document")
// @Region(overrideEnabled=false)
func newLifecyclePolicyDocumentDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &lifecyclePolicyDocumentDataSource{}, nil
}

type lifecyclePolicyDocumentDataSource struct {
	framework.DataSourceWithModel[lifecyclePolicyDocumentDataSourceModel]
}

func (d *lifecyclePolicyDocumentDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrJSON: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[lifecyclePolicyDocumentRule](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						names.AttrPriority: schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrAction: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lifecyclePolicyDocumentRuleAction](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"target_storage_class": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.OneOf("archive"),
										},
									},
									names.AttrType: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.OneOf("expire", "transition"),
										},
									},
								},
							},
						},
						"selection": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lifecyclePolicyDocumentRuleSelection](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"count_number": schema.Int64Attribute{
										Required: true,
										Validators: []validator.Int64{
											int64validator.AtLeast(1),
										},
									},
									"count_type": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.OneOf("imageCountMoreThan", "sinceImagePulled", "sinceImagePushed", "sinceImageTransitioned"),
										},
									},
									"count_unit": schema.StringAttribute{
										Optional: true,
									},
									names.AttrStorageClass: schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.OneOf("archive", "standard"),
										},
									},
									"tag_pattern_list": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									"tag_prefix_list": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									"tag_status": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.OneOf("any", "tagged", "untagged"),
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

func (d *lifecyclePolicyDocumentDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data lifecyclePolicyDocumentDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &lifecyclePolicy{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Default values.
	for _, v := range input.Rules {
		if v.Action == nil {
			v.Action = &lifecyclePolicyRuleAction{
				Type: aws.String("expire"),
			}
		}
	}

	bytes, err := json.MarshalIndent(input, "", "  ")

	if err != nil {
		response.Diagnostics.AddError("Marshalling lifecycle policy to JSON", err.Error())
	}

	data.JSON = types.StringValue(string(bytes))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type lifecyclePolicyDocumentDataSourceModel struct {
	JSON  types.String                                                 `tfsdk:"json"`
	Rules fwtypes.ListNestedObjectValueOf[lifecyclePolicyDocumentRule] `tfsdk:"rule"`
}

type lifecyclePolicyDocumentRule struct {
	Action       fwtypes.ListNestedObjectValueOf[lifecyclePolicyDocumentRuleAction]    `tfsdk:"action"`
	Description  types.String                                                          `tfsdk:"description"`
	RulePriority types.Int64                                                           `tfsdk:"priority"`
	Selection    fwtypes.ListNestedObjectValueOf[lifecyclePolicyDocumentRuleSelection] `tfsdk:"selection"`
}

type lifecyclePolicyDocumentRuleAction struct {
	Type               types.String `tfsdk:"type"`
	TargetStorageClass types.String `tfsdk:"target_storage_class"`
}

type lifecyclePolicyDocumentRuleSelection struct {
	CountNumber    types.Int64                       `tfsdk:"count_number"`
	CountType      types.String                      `tfsdk:"count_type"`
	CountUnit      types.String                      `tfsdk:"count_unit"`
	StorageClass   types.String                      `tfsdk:"storage_class"`
	TagPatternList fwtypes.ListValueOf[types.String] `tfsdk:"tag_pattern_list"`
	TagPrefixList  fwtypes.ListValueOf[types.String] `tfsdk:"tag_prefix_list"`
	TagStatus      types.String                      `tfsdk:"tag_status"`
}
