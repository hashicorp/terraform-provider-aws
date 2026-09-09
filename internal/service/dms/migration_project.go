// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dms_migration_project", name="Migration Project")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(importStateIdAttribute="arn")
// @Testing(hasNoPreExistingResource=true)
func newMigrationProjectResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &migrationProjectResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)

	return r, nil
}

type migrationProjectResource struct {
	framework.ResourceWithModel[migrationProjectResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *migrationProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			},
			"instance_profile_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"instance_profile_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"transformation_rules": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"schema_conversion_application_attributes": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[migrationProjectSCApplicationAttributesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"s3_bucket_path": schema.StringAttribute{
							Optional: true,
						},
						"s3_bucket_role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
			"source_data_provider_descriptor": migrationProjectDataProviderDescriptorBlock(ctx),
			"target_data_provider_descriptor": migrationProjectDataProviderDescriptorBlock(ctx),
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func migrationProjectDataProviderDescriptorBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[migrationProjectDataProviderDescriptorModel](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"data_provider_arn": schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Required:   true,
				},
				"data_provider_name": schema.StringAttribute{
					Computed: true,
				},
				"secrets_manager_access_role_arn": schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Optional:   true,
				},
				"secrets_manager_secret_id": schema.StringAttribute{
					Optional: true,
				},
			},
		},
	}
}

func (r *migrationProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DMSClient(ctx)

	var plan migrationProjectResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input databasemigrationservice.CreateMigrationProjectInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("MigrationProject")))
	if resp.Diagnostics.HasError() {
		return
	}
	input.InstanceProfileIdentifier = plan.InstanceProfileARN.ValueStringPointer()
	input.Tags = getTagsIn(ctx)

	// DMS can't assume a just-created IAM role until it propagates, so retry the
	// AccessDeniedFault that surfaces during that window.
	out, err := tfresource.RetryWhenIsA[*databasemigrationservice.CreateMigrationProjectOutput, *awstypes.AccessDeniedFault](ctx, r.CreateTimeout(ctx, plan.Timeouts),
		func(ctx context.Context) (*databasemigrationservice.CreateMigrationProjectOutput, error) {
			return conn.CreateMigrationProject(ctx, &input)
		})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}
	if out == nil || out.MigrationProject == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.MigrationProject, &plan, flex.WithFieldNamePrefix("MigrationProject")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *migrationProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DMSClient(ctx)

	var state migrationProjectResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findMigrationProjectByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("MigrationProject")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *migrationProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DMSClient(ctx)

	var plan, state migrationProjectResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input databasemigrationservice.ModifyMigrationProjectInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("MigrationProject")))
		if resp.Diagnostics.HasError() {
			return
		}
		input.MigrationProjectIdentifier = state.ARN.ValueStringPointer()
		input.InstanceProfileIdentifier = plan.InstanceProfileARN.ValueStringPointer()
		// AutoFlex omits a null description, so explicitly send an empty string
		// to clear a previously-set description.
		if plan.Description.IsNull() && !state.Description.IsNull() {
			input.Description = aws.String("")
		}

		out, err := conn.ModifyMigrationProject(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
			return
		}
		if out == nil || out.MigrationProject == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, state.ARN.ValueString())
			return
		}

		description := plan.Description
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.MigrationProject, &plan, flex.WithFieldNamePrefix("MigrationProject")))
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Description = description
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *migrationProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DMSClient(ctx)

	var state migrationProjectResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := databasemigrationservice.DeleteMigrationProjectInput{
		MigrationProjectIdentifier: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteMigrationProject(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
}

func findMigrationProjectByARN(ctx context.Context, conn *databasemigrationservice.Client, arn string) (*awstypes.MigrationProject, error) {
	input := databasemigrationservice.DescribeMigrationProjectsInput{
		Filters: []awstypes.Filter{{
			Name:   aws.String("migration-project-identifier"),
			Values: []string{arn},
		}},
	}

	return findMigrationProject(ctx, conn, &input)
}

func findMigrationProject(ctx context.Context, conn *databasemigrationservice.Client, input *databasemigrationservice.DescribeMigrationProjectsInput) (*awstypes.MigrationProject, error) {
	var output []awstypes.MigrationProject
	for item, err := range listMigrationProjects(ctx, conn, input) {
		if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}
		output = append(output, item)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

type migrationProjectResourceModel struct {
	framework.WithRegionModel
	ARN                                   types.String                                                                  `tfsdk:"arn"`
	CreationTime                          timetypes.RFC3339                                                             `tfsdk:"creation_time"`
	Description                           types.String                                                                  `tfsdk:"description"`
	InstanceProfileARN                    fwtypes.ARN                                                                   `tfsdk:"instance_profile_arn"`
	InstanceProfileName                   types.String                                                                  `tfsdk:"instance_profile_name"`
	Name                                  types.String                                                                  `tfsdk:"name"`
	SchemaConversionApplicationAttributes fwtypes.ListNestedObjectValueOf[migrationProjectSCApplicationAttributesModel] `tfsdk:"schema_conversion_application_attributes"`
	SourceDataProviderDescriptors         fwtypes.ListNestedObjectValueOf[migrationProjectDataProviderDescriptorModel]  `tfsdk:"source_data_provider_descriptor"`
	TargetDataProviderDescriptors         fwtypes.ListNestedObjectValueOf[migrationProjectDataProviderDescriptorModel]  `tfsdk:"target_data_provider_descriptor"`
	Tags                                  tftags.Map                                                                    `tfsdk:"tags"`
	TagsAll                               tftags.Map                                                                    `tfsdk:"tags_all"`
	Timeouts                              timeouts.Value                                                                `tfsdk:"timeouts"`
	TransformationRules                   types.String                                                                  `tfsdk:"transformation_rules"`
}

type migrationProjectSCApplicationAttributesModel struct {
	S3BucketPath    types.String `tfsdk:"s3_bucket_path"`
	S3BucketRoleARN fwtypes.ARN  `tfsdk:"s3_bucket_role_arn"`
}

type migrationProjectDataProviderDescriptorModel struct {
	DataProviderARN             fwtypes.ARN  `tfsdk:"data_provider_arn"`
	DataProviderName            types.String `tfsdk:"data_provider_name"`
	SecretsManagerAccessRoleARN fwtypes.ARN  `tfsdk:"secrets_manager_access_role_arn"`
	SecretsManagerSecretID      types.String `tfsdk:"secrets_manager_secret_id"`
}

// Expand bridges a field-name mismatch AutoFlex can't resolve: the create/modify
// input names this field DataProviderIdentifier (which accepts an ARN), while the
// model uses DataProviderARN to match the DataProviderArn field on the Read
// response. AutoFlex matches by name, and those two names aren't fuzzy-equivalent.
func (m migrationProjectDataProviderDescriptorModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return &awstypes.DataProviderDescriptorDefinition{
		DataProviderIdentifier:      m.DataProviderARN.ValueStringPointer(),
		SecretsManagerAccessRoleArn: m.SecretsManagerAccessRoleARN.ValueStringPointer(),
		SecretsManagerSecretId:      m.SecretsManagerSecretID.ValueStringPointer(),
	}, nil
}
