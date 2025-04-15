// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rds_cluster_snapshot_copy", name="Cluster Snapshot Copy")
// @Tags(identifierAttribute="db_cluster_snapshot_arn")
// @Testing(tagsTest=false)
func newResourceClusterSnapshotCopy(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceClusterSnapshotCopy{}

	r.SetDefaultCreateTimeout(20 * time.Minute)

	return r, nil
}

const (
	ResNameClusterSnapshotCopy = "Cluster Snapshot Copy"
)

type resourceClusterSnapshotCopy struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceClusterSnapshotCopy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAllocatedStorage: schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"copy_tags": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"db_cluster_snapshot_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"destination_region": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrEngine: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrEngineVersion: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"license_model": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"presigned_url": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"shared_accounts": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"snapshot_type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_db_cluster_snapshot_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStorageEncrypted: schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStorageType: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"target_db_cluster_snapshot_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z][\w-]+`), "must contain only alphanumeric, and hyphen (-) characters"),
				},
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *resourceClusterSnapshotCopy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceClusterSnapshotCopyData
	conn := r.Meta().RDSClient(ctx)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rds.CopyDBClusterSnapshotInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, data, in)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in.Tags = getTagsIn(ctx)

	if !data.DestinationRegion.IsNull() && data.PresignedURL.IsNull() {
		output, err := rds.NewPresignClient(conn, func(o *rds.PresignOptions) {
			o.ClientOptions = append(o.ClientOptions, func(o *rds.Options) {
				o.Region = data.DestinationRegion.ValueString()
			})
		}).PresignCopyDBClusterSnapshot(ctx, in)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameClusterSnapshotCopy, data.TargetDBClusterSnapshotIdentifier.String(), err),
				err.Error(),
			)
			return
		}

		in.PreSignedUrl = aws.String(output.URL)
	}

	out, err := conn.CopyDBClusterSnapshot(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameClusterSnapshotCopy, data.TargetDBClusterSnapshotIdentifier.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.DBClusterSnapshot == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameClusterSnapshotCopy, data.TargetDBClusterSnapshotIdentifier.String(), err),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.DBClusterSnapshot, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(aws.ToString(out.DBClusterSnapshot.DBClusterSnapshotIdentifier))

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	if _, err := waitDBClusterSnapshotCreated(ctx, conn, data.ID.ValueString(), createTimeout); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionWaitingForCreation, ResNameClusterSnapshotCopy, data.TargetDBClusterSnapshotIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	if !data.SharedAccounts.IsNull() {
		toAdd := []string{}
		resp.Diagnostics.Append(data.SharedAccounts.ElementsAs(ctx, &toAdd, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		input := &rds.ModifyDBClusterSnapshotAttributeInput{
			AttributeName:               aws.String(dbSnapshotAttributeNameRestore),
			DBClusterSnapshotIdentifier: data.TargetDBClusterSnapshotIdentifier.ValueStringPointer(),
			ValuesToAdd:                 toAdd,
		}

		if _, err := conn.ModifyDBClusterSnapshotAttribute(ctx, input); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameClusterSnapshotCopy, data.TargetDBClusterSnapshotIdentifier.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceClusterSnapshotCopy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceClusterSnapshotCopyData
	conn := r.Meta().RDSClient(ctx)

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDBClusterSnapshotByID(ctx, conn, data.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionReading, ResNameClusterSnapshotCopy, data.ID.String(), err),
			err.Error(),
		)
		return
	}
	// Account for variance in naming between the AWS Create and Describe APIs
	data.SourceDBClusterSnapshotIdentifier = flex.StringToFramework(ctx, out.SourceDBClusterSnapshotArn)
	data.TargetDBClusterSnapshotIdentifier = flex.StringToFramework(ctx, out.DBClusterSnapshotIdentifier)

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outAttr, err := findDBClusterSnapshotAttributeByTwoPartKey(ctx, conn, data.ID.ValueString(), dbSnapshotAttributeNameRestore)
	if err != nil && !tfresource.NotFound(err) {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionReading, ResNameClusterSnapshotCopy, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	if len(outAttr.AttributeValues) > 0 {
		resp.Diagnostics.Append(flex.Flatten(ctx, outAttr.AttributeValues, &data.SharedAccounts)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceClusterSnapshotCopy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourceClusterSnapshotCopyData
	conn := r.Meta().RDSClient(ctx)

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !old.SharedAccounts.Equal(new.SharedAccounts) {
		var have, want []string
		resp.Diagnostics.Append(old.SharedAccounts.ElementsAs(ctx, &have, false)...)
		resp.Diagnostics.Append(new.SharedAccounts.ElementsAs(ctx, &want, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		toAdd, toRemove, _ := intflex.DiffSlices(have, want, func(s1, s2 string) bool { return s1 == s2 })

		input := &rds.ModifyDBClusterSnapshotAttributeInput{
			AttributeName:               aws.String(dbSnapshotAttributeNameRestore),
			DBClusterSnapshotIdentifier: new.TargetDBClusterSnapshotIdentifier.ValueStringPointer(),
			ValuesToAdd:                 toAdd,
			ValuesToRemove:              toRemove,
		}

		if _, err := conn.ModifyDBClusterSnapshotAttribute(ctx, input); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.RDS, create.ErrActionUpdating, ResNameClusterSnapshotCopy, new.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	// StorageType can be null, and UseStateForUnknown takes no action
	// on null state values. Explicitly pass through the null value in
	// this case to prevent "invalid result object after apply" errors
	if old.StorageType.IsNull() {
		new.StorageType = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceClusterSnapshotCopy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceClusterSnapshotCopyData
	conn := r.Meta().RDSClient(ctx)

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("deleting %s", ResNameClusterSnapshotCopy), map[string]any{
		names.AttrID: data.ID.ValueString(),
	})

	_, err := conn.DeleteDBClusterSnapshot(ctx, &rds.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: data.ID.ValueStringPointer(),
	})
	if err != nil {
		if errs.IsA[*awstypes.DBClusterSnapshotNotFoundFault](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionDeleting, ResNameClusterSnapshotCopy, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceClusterSnapshotCopy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

type resourceClusterSnapshotCopyData struct {
	AllocatedStorage                  types.Int64  `tfsdk:"allocated_storage"`
	CopyTags                          types.Bool   `tfsdk:"copy_tags"`
	DBClusterSnapshotARN              types.String `tfsdk:"db_cluster_snapshot_arn"`
	DestinationRegion                 types.String `tfsdk:"destination_region"`
	Engine                            types.String `tfsdk:"engine"`
	EngineVersion                     types.String `tfsdk:"engine_version"`
	ID                                types.String `tfsdk:"id"`
	KMSKeyID                          types.String `tfsdk:"kms_key_id"`
	LicenseModel                      types.String `tfsdk:"license_model"`
	PresignedURL                      types.String `tfsdk:"presigned_url"`
	SharedAccounts                    types.Set    `tfsdk:"shared_accounts"`
	SnapshotType                      types.String `tfsdk:"snapshot_type"`
	SourceDBClusterSnapshotIdentifier types.String `tfsdk:"source_db_cluster_snapshot_identifier"`
	StorageEncrypted                  types.Bool   `tfsdk:"storage_encrypted"`
	StorageType                       types.String `tfsdk:"storage_type"`
	Tags                              tftags.Map   `tfsdk:"tags"`
	TagsAll                           tftags.Map   `tfsdk:"tags_all"`
	TargetDBClusterSnapshotIdentifier types.String `tfsdk:"target_db_cluster_snapshot_identifier"`
	VPCID                             types.String `tfsdk:"vpc_id"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
