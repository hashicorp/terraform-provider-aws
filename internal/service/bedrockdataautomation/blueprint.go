// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockdataautomation

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockdataautomation_blueprint", name="Blueprint")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation/types;awstypes;awstypes.Blueprint")
// @Testing(hasNoPreExistingResource=true)
func newBlueprintResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &blueprintResource{}, nil
}

type blueprintResource struct {
	framework.ResourceWithModel[blueprintResourceModel]
	framework.WithImportByIdentity
}

func (r *blueprintResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"blueprint_stage": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.BlueprintStage](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric characters, hyphens, and underscores"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrSchema: schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100000),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Type](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrEncryptionConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"kms_encryption_context": schema.MapAttribute{
							CustomType: fwtypes.MapOfStringType,
							Optional:   true,
						},
						names.AttrKMSKeyID: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
							},
						},
					},
				},
			},
		},
	}
}

func (r *blueprintResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan blueprintResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input bedrockdataautomation.CreateBlueprintInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Blueprint")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateBlueprint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.Blueprint, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *blueprintResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state blueprintResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findBlueprintByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *blueprintResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state blueprintResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, plan.ARN)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	var blueprint *awstypes.Blueprint
	if diff.HasChanges() {
		var input bedrockdataautomation.UpdateBlueprintInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Blueprint")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateBlueprint(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		blueprint = out.Blueprint
	} else {
		// Only tags changed. Read back the last modified time, which is unknown in the plan.
		var err error
		blueprint, err = findBlueprintByARN(ctx, conn, arn)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, blueprint, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *blueprintResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state blueprintResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := bedrockdataautomation.DeleteBlueprintInput{
		BlueprintArn: aws.String(arn),
	}
	_, err := conn.DeleteBlueprint(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

// flatten flattens the API object into the model. The API returns the
// encryption settings as top-level fields rather than as a nested object.
func (r *blueprintResource) flatten(ctx context.Context, blueprint *awstypes.Blueprint, data *blueprintResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, blueprint, data, fwflex.WithFieldNamePrefix("Blueprint")))
	if diags.HasError() {
		return diags
	}

	if blueprint.KmsKeyId == nil {
		data.EncryptionConfiguration = fwtypes.NewListNestedObjectValueOfNull[encryptionConfigurationModel](ctx)
		return diags
	}

	encryptionConfiguration := awstypes.EncryptionConfiguration{
		KmsEncryptionContext: blueprint.KmsEncryptionContext,
		KmsKeyId:             blueprint.KmsKeyId,
	}
	var model encryptionConfigurationModel
	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, &encryptionConfiguration, &model))
	if diags.HasError() {
		return diags
	}

	var d diag.Diagnostics
	data.EncryptionConfiguration, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
	smerr.AddEnrich(ctx, &diags, d)

	return diags
}

func findBlueprintByARN(ctx context.Context, conn *bedrockdataautomation.Client, arn string) (*awstypes.Blueprint, error) {
	input := bedrockdataautomation.GetBlueprintInput{
		BlueprintArn: aws.String(arn),
	}

	out, err := conn.GetBlueprint(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Blueprint == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Blueprint, nil
}

type blueprintResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String                                                  `tfsdk:"arn"`
	BlueprintStage          fwtypes.StringEnum[awstypes.BlueprintStage]                   `tfsdk:"blueprint_stage"`
	CreationTime            timetypes.RFC3339                                             `tfsdk:"creation_time"`
	EncryptionConfiguration fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	LastModifiedTime        timetypes.RFC3339                                             `tfsdk:"last_modified_time"`
	Name                    types.String                                                  `tfsdk:"name"`
	Schema                  jsontypes.Normalized                                          `tfsdk:"schema"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Type                    fwtypes.StringEnum[awstypes.Type]                             `tfsdk:"type"`
}

type encryptionConfigurationModel struct {
	KMSEncryptionContext fwtypes.MapOfString `tfsdk:"kms_encryption_context"`
	KMSKeyID             types.String        `tfsdk:"kms_key_id"`
}
