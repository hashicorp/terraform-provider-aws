// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_backup_tiering_configuration", name="Tiering Configuration")
// @Tags(identifierAttribute="tiering_configuration_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/backup;types.TieringConfiguration")
// @Testing(generator="randomTieringConfigurationName()")
func newTieringConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &tieringConfigurationResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type tieringConfigurationResource struct {
	framework.ResourceWithModel[tieringConfigurationResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *tieringConfigurationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"backup_vault_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the backup vault where the tiering configuration applies. Use * to apply to all backup vaults.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_*-]+$`), "must contain only alphanumeric characters, hyphens, underscores, and asterisks"),
				},
			},
			names.AttrCreationTime: schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "The date and time a tiering configuration was created, in Unix format and Coordinated Universal Time (UTC). The value of CreationTime is accurate to milliseconds.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrLastUpdatedTime: schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "The date and time a tiering configuration was updated, in Unix format and Coordinated Universal Time (UTC). The value of LastUpdatedTime is accurate to milliseconds.",
			},
			names.AttrTags:              tftags.TagsAttribute(),
			names.AttrTagsAll:           tftags.TagsAttributeComputedOnly(),
			"tiering_configuration_arn": framework.ARNAttributeComputedOnly(),
			"tiering_configuration_name": schema.StringAttribute{
				Required:    true,
				Description: "The unique name of the tiering configuration. This cannot be changed after creation, and it must consist of only alphanumeric characters and underscores.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_]+$`), "must contain only alphanumeric characters and underscores"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"resource_selection": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceSelectionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(5),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrResourceType: schema.StringAttribute{
							Required:    true,
							Description: "The type of AWS resource; for example, S3 for Amazon S3. For tiering configurations, this is currently limited to S3.",
							Validators: []validator.String{
								stringvalidator.OneOf("S3"),
							},
						},
						names.AttrResources: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Required:    true,
							Description: "An array of strings that either contains ARNs of the associated resources or contains a wildcard * to specify all resources. You can specify up to 100 specific resources per tiering configuration",
							Validators: []validator.Set{
								setvalidator.SizeAtMost(100),
								setvalidator.SizeAtLeast(1),
								setvalidator.ValueStringsAre(
									stringvalidator.Any(
										stringvalidator.RegexMatches(regexache.MustCompile(`^\*$`), "must be * for all resources"),
										stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws[^:]*:s3:::.*`), "must be a valid S3 ARN"),
									),
								),
							},
						},
						"tiering_down_settings_in_days": schema.Int64Attribute{
							Required:    true,
							Description: "The number of days after creation within a backup vault that an object can transition to the low cost warm storage tier. Must be a positive integer between 60 and 36500 days.",
							Validators: []validator.Int64{
								int64validator.Between(60, 36500),
							},
						},
					},
				},
				Description: "An array of resource selection objects that specify which resources are included in the tiering configuration and their tiering settings.",
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *tieringConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tieringConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	tieringConfigurationName := plan.TieringConfigurationName.ValueString()
	input := &backup.CreateTieringConfigurationInput{
		TieringConfiguration: &awstypes.TieringConfigurationInputForCreate{
			BackupVaultName:          plan.BackupVaultName.ValueStringPointer(),
			TieringConfigurationName: aws.String(tieringConfigurationName),
		},
		TieringConfigurationTags: getTagsIn(ctx),
	}

	resp.Diagnostics.Append(fwflex.Expand(ctx, plan.ResourceSelection, &input.TieringConfiguration.ResourceSelection)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tieringConfiguration, err := conn.CreateTieringConfiguration(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating Backup Tiering Configuration (%s)", tieringConfigurationName),
			err.Error(),
		)
		return
	}

	// Set ID
	plan.ID = types.StringValue(aws.ToString(tieringConfiguration.TieringConfigurationName))

	output, err := findTieringConfigurationByName(ctx, conn, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"reading Backup Tiering Configuration after create",
			err.Error(),
		)
		return
	}

	// Set computed fields directly
	plan.TieringConfigurationArn = types.StringPointerValue(output.TieringConfigurationArn)
	plan.BackupVaultName = types.StringPointerValue(output.BackupVaultName)
	plan.TieringConfigurationName = types.StringPointerValue(output.TieringConfigurationName)

	plan.CreationTime = fwflex.TimeToFramework(ctx, output.CreationTime)
	plan.LastUpdatedTime = fwflex.TimeToFramework(ctx, output.LastUpdatedTime)

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output.ResourceSelection, &plan.ResourceSelection)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tieringConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tieringConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	output, err := findTieringConfigurationByName(ctx, conn, state.ID.ValueString())
	if errs.IsA[*sdkretry.NotFoundError](err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"reading Backup Tiering Configuration",
			err.Error(),
		)
		return
	}

	// Set computed fields directly
	state.TieringConfigurationArn = types.StringPointerValue(output.TieringConfigurationArn)
	state.BackupVaultName = types.StringPointerValue(output.BackupVaultName)
	state.TieringConfigurationName = types.StringPointerValue(output.TieringConfigurationName)

	state.CreationTime = fwflex.TimeToFramework(ctx, output.CreationTime)
	state.LastUpdatedTime = fwflex.TimeToFramework(ctx, output.LastUpdatedTime)

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output.ResourceSelection, &state.ResourceSelection)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tieringConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tieringConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	if !plan.ResourceSelection.Equal(state.ResourceSelection) || !plan.BackupVaultName.Equal(state.BackupVaultName) {
		input := &backup.UpdateTieringConfigurationInput{
			TieringConfigurationName: plan.ID.ValueStringPointer(),
			TieringConfiguration: &awstypes.TieringConfigurationInputForUpdate{
				BackupVaultName: plan.BackupVaultName.ValueStringPointer(),
			},
		}

		resp.Diagnostics.Append(fwflex.Expand(ctx, plan.ResourceSelection, &input.TieringConfiguration.ResourceSelection)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateTieringConfiguration(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				"updating Backup Tiering Configuration",
				err.Error(),
			)
			return
		}
	}

	// Always read the current state from AWS to get updated computed fields
	output, err := findTieringConfigurationByName(ctx, conn, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"reading Backup Tiering Configuration after update",
			err.Error(),
		)
		return
	}

	// Set computed fields directly after update
	plan.TieringConfigurationArn = types.StringPointerValue(output.TieringConfigurationArn)
	plan.BackupVaultName = types.StringPointerValue(output.BackupVaultName)
	plan.TieringConfigurationName = types.StringPointerValue(output.TieringConfigurationName)

	plan.CreationTime = fwflex.TimeToFramework(ctx, output.CreationTime)
	plan.LastUpdatedTime = fwflex.TimeToFramework(ctx, output.LastUpdatedTime)

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output.ResourceSelection, &plan.ResourceSelection)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tieringConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tieringConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	input := backup.DeleteTieringConfigurationInput{TieringConfigurationName: state.ID.ValueStringPointer()}
	_, err :=
		conn.DeleteTieringConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"deleting Backup Tiering Configuration",
			err.Error(),
		)
		return
	}
}

type tieringConfigurationResourceModel struct {
	framework.WithRegionModel
	BackupVaultName          types.String                                            `tfsdk:"backup_vault_name"`
	CreationTime             timetypes.RFC3339                                       `tfsdk:"creation_time"`
	ID                       types.String                                            `tfsdk:"id"`
	LastUpdatedTime          timetypes.RFC3339                                       `tfsdk:"last_updated_time"`
	ResourceSelection        fwtypes.ListNestedObjectValueOf[resourceSelectionModel] `tfsdk:"resource_selection"`
	Tags                     tftags.Map                                              `tfsdk:"tags"`
	TagsAll                  tftags.Map                                              `tfsdk:"tags_all"`
	TieringConfigurationArn  types.String                                            `tfsdk:"tiering_configuration_arn"`
	TieringConfigurationName types.String                                            `tfsdk:"tiering_configuration_name"`
	Timeouts                 timeouts.Value                                          `tfsdk:"timeouts"`
}

type resourceSelectionModel struct {
	ResourceType              types.String                     `tfsdk:"resource_type"`
	Resources                 fwtypes.SetValueOf[types.String] `tfsdk:"resources"`
	TieringDownSettingsInDays types.Int64                      `tfsdk:"tiering_down_settings_in_days"`
}

func findTieringConfigurationByName(ctx context.Context, conn *backup.Client, name string) (*awstypes.TieringConfiguration, error) {
	input := &backup.GetTieringConfigurationInput{
		TieringConfigurationName: aws.String(name),
	}

	output, err := conn.GetTieringConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TieringConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TieringConfiguration, nil
}
