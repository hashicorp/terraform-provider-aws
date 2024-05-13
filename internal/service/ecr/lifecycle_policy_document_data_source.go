// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

// @FrameworkDataSource(name="Lifecycle Policy Document")
func newLifecyclePolicyDocumentDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &lifecyclePolicyDocumentDataSource{}, nil
}

type lifecyclePolicyDocumentDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *lifecyclePolicyDocumentDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_ecr_lifecycle_policy_document"
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
									names.AttrType: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.OneOf("expire"),
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
											stringvalidator.OneOf("imageCountMoreThan", "sinceImagePushed"),
										},
									},
									"count_unit": schema.StringAttribute{
										Optional: true,
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
											stringvalidator.OneOf("tagged", "untagged", "any"),
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
	Type types.String `tfsdk:"type"`
}

type lifecyclePolicyDocumentRuleSelection struct {
	CountNumber    types.Int64                       `tfsdk:"count_number"`
	CountType      types.String                      `tfsdk:"count_type"`
	CountUnit      types.String                      `tfsdk:"count_unit"`
	TagPatternList fwtypes.ListValueOf[types.String] `tfsdk:"tag_pattern_list"`
	TagPrefixList  fwtypes.ListValueOf[types.String] `tfsdk:"tag_prefix_list"`
	TagStatus      types.String                      `tfsdk:"tag_status"`
}
