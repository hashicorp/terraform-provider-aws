// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockdataautomation

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

// @FrameworkResource("aws_bedrockdataautomation_library", name="Library")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockdataautomation/types;awstypes;awstypes.DataAutomationLibrary")
// @Testing(hasNoPreExistingResource=true)
func newLibraryResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &libraryResource{}

	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type libraryResource struct {
	framework.ResourceWithModel[libraryResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *libraryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 300),
				},
				PlanModifiers: []planmodifier.String{
					// The API can't remove a description.
					stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						resp.RequiresReplace = req.PlanValue.IsNull() && !req.StateValue.IsNull()
					}, "", ""),
				},
			},
			"entity_types": framework.ResourceComputedListOfObjectsAttribute[entityTypeInfoModel](ctx, listplanmodifier.UseStateForUnknown()),
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
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataAutomationLibraryStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrEncryptionConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"kms_encryption_context": schema.MapAttribute{
							CustomType: fwtypes.MapOfStringType,
							Optional:   true,
							PlanModifiers: []planmodifier.Map{
								mapplanmodifier.RequiresReplace(),
							},
						},
						names.AttrKMSKeyID: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Delete: true,
			}),
		},
	}
}

func (r *libraryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan libraryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input bedrockdataautomation.CreateDataAutomationLibraryInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Library")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateDataAutomationLibrary(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	arn := aws.ToString(out.LibraryArn)
	// Save the ARN and Region so that a failed read leaves a tainted resource rather than an orphan.
	plan.ARN = fwflex.StringValueToFramework(ctx, arn)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrARN), plan.ARN))
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrRegion), plan.Region))
	if resp.Diagnostics.HasError() {
		return
	}

	// The create response doesn't include the library.
	library, err := findLibraryByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, library, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *libraryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state libraryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findLibraryByARN(ctx, conn, arn)
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

func (r *libraryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state libraryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, plan.ARN)

	if !plan.Description.Equal(state.Description) {
		input := bedrockdataautomation.UpdateDataAutomationLibraryInput{
			ClientToken:        aws.String(create.UniqueId(ctx)),
			LibraryArn:         aws.String(arn),
			LibraryDescription: fwflex.StringFromFramework(ctx, plan.Description),
		}
		if _, err := conn.UpdateDataAutomationLibrary(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *libraryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state libraryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockDataAutomationClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := bedrockdataautomation.DeleteDataAutomationLibraryInput{
		LibraryArn: aws.String(arn),
	}
	_, err := conn.DeleteDataAutomationLibrary(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitLibraryDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

// flatten flattens the API object into the model. The API returns the
// encryption settings as top-level fields rather than as a nested object.
func (r *libraryResource) flatten(ctx context.Context, library *awstypes.DataAutomationLibrary, data *libraryResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, library, data, fwflex.WithFieldNamePrefix("Library")))
	if diags.HasError() {
		return diags
	}

	if library.KmsKeyId == nil {
		data.EncryptionConfiguration = fwtypes.NewListNestedObjectValueOfNull[encryptionConfigurationModel](ctx)
		return diags
	}

	encryptionConfiguration := awstypes.EncryptionConfiguration{
		KmsEncryptionContext: library.KmsEncryptionContext,
		KmsKeyId:             library.KmsKeyId,
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

func findLibraryByARN(ctx context.Context, conn *bedrockdataautomation.Client, arn string) (*awstypes.DataAutomationLibrary, error) {
	input := bedrockdataautomation.GetDataAutomationLibraryInput{
		LibraryArn: aws.String(arn),
	}

	out, err := conn.GetDataAutomationLibrary(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Library == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Library, nil
}

func statusLibrary(conn *bedrockdataautomation.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findLibraryByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func waitLibraryDeleted(ctx context.Context, conn *bedrockdataautomation.Client, arn string, timeout time.Duration) (*awstypes.DataAutomationLibrary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataAutomationLibraryStatusActive, awstypes.DataAutomationLibraryStatusDeleting),
		Target:  []string{},
		Refresh: statusLibrary(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DataAutomationLibrary); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type libraryResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String                                                  `tfsdk:"arn"`
	CreationTime            timetypes.RFC3339                                             `tfsdk:"creation_time"`
	Description             types.String                                                  `tfsdk:"description"`
	EncryptionConfiguration fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	EntityTypes             fwtypes.ListNestedObjectValueOf[entityTypeInfoModel]          `tfsdk:"entity_types"`
	Name                    types.String                                                  `tfsdk:"name"`
	Status                  fwtypes.StringEnum[awstypes.DataAutomationLibraryStatus]      `tfsdk:"status"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
}

type entityTypeInfoModel struct {
	EntityMetadata types.String                            `tfsdk:"entity_metadata"`
	EntityType     fwtypes.StringEnum[awstypes.EntityType] `tfsdk:"entity_type"`
}
