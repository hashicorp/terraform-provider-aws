// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Generated identity tests are not used: the source savepoint is exported from a running Apache Flink application,
// whose test configuration depends on a code JAR in test-fixtures that ConfigDirectory tests cannot reference.
// @FrameworkResource("aws_kinesisanalyticsv2_snapshot_import", name="Snapshot Import")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("name")
// @Testing(identityTest=false)
// @Testing(hasNoPreExistingResource=true)
func newSnapshotImportResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &snapshotImportResource{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type snapshotImportResource struct {
	framework.ResourceWithModel[snapshotImportResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *snapshotImportResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrFormat: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SnapshotFormat](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: enhancedSnapshotNameValidators(),
			},
			"size_in_bytes": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			// Omitting this encrypts the snapshot with an AWS owned key.
			"destination_encryption_configuration": snapshotEncryptionConfigurationBlock(ctx),
			"import_source_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[importSourceConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					requiresReplaceUnlessImported(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"import_role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						names.AttrSourceType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.StorageLocationType](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"source_encryption_configuration": snapshotEncryptionConfigurationBlock(ctx),
						"source_s3_prefix_location": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3PrefixLocationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"bucket_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"key_prefix": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											// Without a trailing slash MSF reads the parent prefix.
											stringvalidator.RegexMatches(regexache.MustCompile(`/$`), "must end with a forward slash (/)"),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

// requiresReplaceUnlessImported replaces the snapshot when a block changes, but not when the prior state is null.
// DescribeSnapshot does not return the import role, source type, or source encryption, so after
// terraform import these blocks are null in state and the configured values are adopted instead.
func requiresReplaceUnlessImported() planmodifier.List {
	return listplanmodifier.RequiresReplaceIf(
		func(_ context.Context, request planmodifier.ListRequest, response *listplanmodifier.RequiresReplaceIfFuncResponse) {
			response.RequiresReplace = !request.StateValue.IsNull() && !request.PlanValue.Equal(request.StateValue)
		},
		"If the value of this attribute changes, Terraform will destroy and recreate the resource. Does not trigger replacement on import.",
		"If the value of this attribute changes, Terraform will destroy and recreate the resource. Does not trigger replacement on import.",
	)
}

func snapshotEncryptionConfigurationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[snapshotEncryptionConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		PlanModifiers: []planmodifier.List{
			requiresReplaceUnlessImported(),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrKeyID: schema.StringAttribute{
					Optional: true,
				},
				"key_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.KeyType](),
					Required:   true,
				},
			},
		},
	}
}

func (r *snapshotImportResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().KinesisAnalyticsV2Client(ctx)

	var plan snapshotImportResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	var input kinesisanalyticsv2.ImportSnapshotInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix(enhancedSnapshotFieldNamePrefix)))
	if response.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	output, err := conn.ImportSnapshot(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// Persist the identity now so that a snapshot that fails to become READY is tainted, not orphaned:
	// a FAILED snapshot keeps its name reserved until deleted. Region is needed too, or a snapshot
	// imported by region override is refreshed in the wrong Region.
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrName), name))
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrRegion), plan.Region))
	if response.Diagnostics.HasError() {
		return
	}

	snapshot, _, err := waitEnhancedSnapshotCreated(ctx, conn, name, aws.ToString(output.SnapshotOperationId), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, snapshot, &plan, fwflex.WithFieldNamePrefix(enhancedSnapshotFieldNamePrefix)))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *snapshotImportResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().KinesisAnalyticsV2Client(ctx)

	var state snapshotImportResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	snapshot, err := findEnhancedSnapshotByName(ctx, conn, state.Name.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	// DescribeSnapshot does not return the import role, source type, or source encryption,
	// so the import_source_configuration and destination_encryption_configuration arguments keep their configured values.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, snapshot, &state, fwflex.WithFieldNamePrefix(enhancedSnapshotFieldNamePrefix)))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *snapshotImportResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan snapshotImportResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	// Only tags can be updated; transparent tagging handles the update.
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *snapshotImportResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().KinesisAnalyticsV2Client(ctx)

	var state snapshotImportResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	if err := deleteEnhancedSnapshot(ctx, conn, state.Name.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.Name.String())
	}
}

type snapshotImportResourceModel struct {
	framework.WithRegionModel
	ARN                                types.String                                                          `tfsdk:"arn"`
	CreationTimestamp                  timetypes.RFC3339                                                     `tfsdk:"creation_time"`
	Description                        types.String                                                          `tfsdk:"description"`
	DestinationEncryptionConfiguration fwtypes.ListNestedObjectValueOf[snapshotEncryptionConfigurationModel] `tfsdk:"destination_encryption_configuration"`
	Format                             fwtypes.StringEnum[awstypes.SnapshotFormat]                           `tfsdk:"format"`
	ImportSourceConfiguration          fwtypes.ListNestedObjectValueOf[importSourceConfigurationModel]       `tfsdk:"import_source_configuration"`
	Name                               types.String                                                          `tfsdk:"name"`
	SizeInBytes                        types.Int64                                                           `tfsdk:"size_in_bytes"`
	Tags                               tftags.Map                                                            `tfsdk:"tags"`
	TagsAll                            tftags.Map                                                            `tfsdk:"tags_all"`
	Timeouts                           timeouts.Value                                                        `tfsdk:"timeouts"`
}

type importSourceConfigurationModel struct {
	ImportRoleARN                 fwtypes.ARN                                                           `tfsdk:"import_role_arn"`
	SourceEncryptionConfiguration fwtypes.ListNestedObjectValueOf[snapshotEncryptionConfigurationModel] `tfsdk:"source_encryption_configuration"`
	SourceS3PrefixLocation        fwtypes.ListNestedObjectValueOf[s3PrefixLocationModel]                `tfsdk:"source_s3_prefix_location"`
	SourceType                    fwtypes.StringEnum[awstypes.StorageLocationType]                      `tfsdk:"source_type"`
}

type s3PrefixLocationModel struct {
	BucketARN fwtypes.ARN  `tfsdk:"bucket_arn"`
	KeyPrefix types.String `tfsdk:"key_prefix"`
}

type snapshotEncryptionConfigurationModel struct {
	KeyID   types.String                         `tfsdk:"key_id"`
	KeyType fwtypes.StringEnum[awstypes.KeyType] `tfsdk:"key_type"`
}
