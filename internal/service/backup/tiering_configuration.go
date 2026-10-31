// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

// @FrameworkResource("aws_backup_tiering_configuration", name="Tiering Configuration")
// @IdentityAttribute("name")
// @Tags(identifierAttribute="arn")
// @Testing(importStateIdAttribute="name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/backup/types;awstypes;awstypes.TieringConfiguration")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckTieringConfiguration")
func newTieringConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &tieringConfigurationResource{}, nil
}

const (
	tieringConfigurationMaxResources = 100
)

type tieringConfigurationResource struct {
	framework.ResourceWithModel[tieringConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *tieringConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"backup_vault_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^(\*|[0-9A-Za-z_-]{2,50})$`), "must be * or 2 to 50 alphanumeric, hyphen or underscore characters"),
				},
			},
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrLastUpdatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_]+$`), "must contain only alphanumeric characters and underscores"),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"resource_selection": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourceSelectionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(5),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrResources: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 100),
							},
						},
						names.AttrResourceType: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("S3"),
							},
						},
						"tiering_down_settings_in_days": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.Between(60, 36500),
							},
						},
					},
				},
			},
		},
	}
}

func (r *tieringConfigurationResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data tieringConfigurationResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.ResourceSelection.IsNull() || data.ResourceSelection.IsUnknown() {
		return
	}

	resourceSelections, diags := data.ResourceSelection.ToSlice(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	allVaults := !data.BackupVaultName.IsNull() && !data.BackupVaultName.IsUnknown() && data.BackupVaultName.ValueString() == "*"

	// The API limits the number of specific resource ARNs across all resource selections in a configuration.
	var n int
	for i, resourceSelection := range resourceSelections {
		if resourceSelection.Resources.IsNull() || resourceSelection.Resources.IsUnknown() {
			continue
		}

		resources := resourceSelection.Resources.Elements()
		if allVaults {
			invalid := len(resources) != 1
			if !invalid && !resources[0].IsUnknown() {
				v, ok := resources[0].(types.String)
				invalid = !ok || v.ValueString() != "*"
			}

			if invalid {
				response.Diagnostics.AddAttributeError(
					path.Root("resource_selection").AtListIndex(i).AtName(names.AttrResources),
					"Invalid Configuration",
					`resources must select all resources when backup_vault_name is "*"`,
				)
			}
		}

		for _, v := range resources {
			if v, ok := v.(types.String); ok && !v.IsUnknown() && v.ValueString() != "*" {
				n++
			}
		}
	}

	if n > tieringConfigurationMaxResources {
		response.Diagnostics.AddAttributeError(
			path.Root("resource_selection"),
			"Invalid Configuration",
			fmt.Sprintf("resource_selection blocks must select at most %d resources in total, got %d", tieringConfigurationMaxResources, n),
		)
	}
}

func (r *tieringConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BackupClient(ctx)

	var data tieringConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var tieringConfiguration awstypes.TieringConfigurationInputForCreate
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &tieringConfiguration, fwflex.WithFieldNamePrefix("TieringConfiguration")))
	if response.Diagnostics.HasError() {
		return
	}

	input := backup.CreateTieringConfigurationInput{
		CreatorRequestId:         aws.String(create.UniqueId(ctx)),
		TieringConfiguration:     &tieringConfiguration,
		TieringConfigurationTags: getTagsIn(ctx),
	}

	_, err := conn.CreateTieringConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	output, err := findTieringConfigurationByName(ctx, conn, name)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *tieringConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().BackupClient(ctx)

	var data tieringConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	output, err := findTieringConfigurationByName(ctx, conn, name)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *tieringConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().BackupClient(ctx)

	var new, old tieringConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, new.Name)

	if diff.HasChanges() {
		var tieringConfiguration awstypes.TieringConfigurationInputForUpdate
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &tieringConfiguration, fwflex.WithFieldNamePrefix("TieringConfiguration")))
		if response.Diagnostics.HasError() {
			return
		}

		input := backup.UpdateTieringConfigurationInput{
			TieringConfiguration:     &tieringConfiguration,
			TieringConfigurationName: aws.String(name),
		}

		_, err := conn.UpdateTieringConfiguration(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}
	}

	output, err := findTieringConfigurationByName(ctx, conn, name)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &new))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *tieringConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().BackupClient(ctx)

	var data tieringConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	input := backup.DeleteTieringConfigurationInput{
		TieringConfigurationName: aws.String(name),
	}
	_, err := conn.DeleteTieringConfiguration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *tieringConfigurationResource) flatten(ctx context.Context, tieringConfiguration *awstypes.TieringConfiguration, data *tieringConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, tieringConfiguration, data, fwflex.WithFieldNamePrefix("TieringConfiguration"))...)

	return diags
}

func findTieringConfigurationByName(ctx context.Context, conn *backup.Client, name string) (*awstypes.TieringConfiguration, error) {
	input := backup.GetTieringConfigurationInput{
		TieringConfigurationName: aws.String(name),
	}

	return findTieringConfiguration(ctx, conn, &input)
}

func findTieringConfiguration(ctx context.Context, conn *backup.Client, input *backup.GetTieringConfigurationInput) (*awstypes.TieringConfiguration, error) {
	output, err := conn.GetTieringConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.TieringConfiguration == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.TieringConfiguration, nil
}

type tieringConfigurationResourceModel struct {
	framework.WithRegionModel
	ARN               types.String                                            `tfsdk:"arn"`
	BackupVaultName   types.String                                            `tfsdk:"backup_vault_name"`
	CreationTime      timetypes.RFC3339                                       `tfsdk:"creation_time"`
	LastUpdatedTime   timetypes.RFC3339                                       `tfsdk:"last_updated_time"`
	Name              types.String                                            `tfsdk:"name"`
	ResourceSelection fwtypes.ListNestedObjectValueOf[resourceSelectionModel] `tfsdk:"resource_selection"`
	Tags              tftags.Map                                              `tfsdk:"tags"`
	TagsAll           tftags.Map                                              `tfsdk:"tags_all"`
}

type resourceSelectionModel struct {
	Resources                 fwtypes.SetOfString `tfsdk:"resources"`
	ResourceType              types.String        `tfsdk:"resource_type"`
	TieringDownSettingsInDays types.Int32         `tfsdk:"tiering_down_settings_in_days"`
}
