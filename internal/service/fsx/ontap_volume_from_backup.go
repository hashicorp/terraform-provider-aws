// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Ontap Volume From Backup")
func newResourceOntapVolumeFromBackup(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceOntapVolumeFromBackup{}

	/*	r.SetDefaultCreateTimeout(30 * time.Minute)
		r.SetDefaultUpdateTimeout(30 * time.Minute)
		r.SetDefaultDeleteTimeout(30 * time.Minute)*/

	return r, nil
}

const (
	ResNameOntapVolumeFromBackup = "Ontap Volume From Backup"
)

type resourceOntapVolumeFromBackup struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceOntapVolumeFromBackup) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_fsx_ontap_volume_from_backup"
}

func (r *resourceOntapVolumeFromBackup) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"backup_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"ontap_configuration": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"copy_tags_to_backups": schema.BoolAttribute{
							Optional: true,
						},
						"junction_path": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^\/[^\x00\x85\x{2028}\x{2029}\r\n\/]{1,255}$`),
									"The JunctionPath must have a leading forward slash, such as /vol3",
								),
							},
						},
						"ontap_volume_type": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
								stringvalidator.RegexMatches(
									regexache.MustCompile(`^RW$|^DP$`),
									"Valid values one of RW, DP",
								),
							},
						},
						"security_style": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
								stringvalidator.RegexMatches(
									regexache.MustCompile(`^UNIX$|^NTFS$|^MIXED$`),
									"Valid values one of UNIX | NTFS | MIXED",
								),
							},
						},
						"size_in_megabytes": schema.Int64Attribute{
							Required: true,
						},
						"storage_efficiency_enabled": schema.BoolAttribute{
							Optional: true,
						},
						"storage_virtualmachine_id": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(21),
								stringvalidator.RegexMatches(
									regexache.MustCompile(`^(svm-[0-9a-f]{17,})$`),
									"value must contain 21 lowercase letters or numbers, prefixed with svm-",
								),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"snaplock_configuration": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"audit_log_volume": schema.BoolAttribute{
										Optional: true,
									},
									"privileged_delete": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 255),
											stringvalidator.RegexMatches(
												regexache.MustCompile(`^DISABLED$|^ENABLED$|^PERMANENTLY_DISABLED$`),
												"Valid values one of DISABLED | ENABLED | PERMANENTLY_DISABLED",
											),
										},
									},
									"snaplock_type": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 255),
											stringvalidator.RegexMatches(
												regexache.MustCompile(`^COMPLIANCE$|^ENTERPRISE$`),
												"Valid values one of COMPLIANCE | ENTERPRISE",
											),
										},
									},
									"volume_append_mode_enabled": schema.BoolAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
									"autocommit_period": schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Required: true,
												},
												"value": schema.Int64Attribute{
													Optional: true,
												},
											},
										},
									},
									"retention_period": schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"default_retention": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"type": schema.StringAttribute{
																Required: true,
															},
															"value": schema.Int64Attribute{
																Optional: true,
															},
														},
													},
												},
												"maximum_retention": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"type": schema.StringAttribute{
																Required: true,
															},
															"value": schema.Int64Attribute{
																Optional: true,
															},
														},
													},
												},
												"minimum_retention": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"type": schema.StringAttribute{
																Required: true,
															},
															"value": schema.Int64Attribute{
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						// ... other blocks inside ontap_configuration ...
					},
				},
			},
			// ... other top-level blocks ...
		},
	}
}

func (r *resourceOntapVolumeFromBackup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().FSxClient(ctx)
	var plan resourceOntapVolumeFromBackupData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := plan.ONTAPConfiguration

	var ontapConfig = awstypes.CreateOntapVolumeConfiguration{}

	if &plan.ONTAPConfiguration != nil {
		ontapConfig = awstypes.CreateOntapVolumeConfiguration{
			// Map the fields from config to ontapConfig
			CopyTagsToBackups:        aws.Bool(config.CopyTagsToBackup.ValueBool()), //config.CopyTagsToBackups,
			JunctionPath:             aws.String(config.JunctionPath.ValueString()), //config.JunctionPath,
			OntapVolumeType:          config.ONTAPVolumeType,
			SecurityStyle:            config.SecurityStyle,
			SizeInMegabytes:          aws.Int32(config.SizeInMegaBytes),                        //config.SizeInMegabytes,
			SnapshotPolicy:           aws.String(config.SnapshotPolicy.ValueString()),          //config.SnapshotPolicy,
			StorageEfficiencyEnabled: aws.Bool(config.StorageEfficiencyEnabled.ValueBool()),    //config.StorageEfficiencyEnabled,
			StorageVirtualMachineId:  aws.String(config.StorageVirtualMachineID.ValueString()), //config.StorageVirtualMachineId,
			TieringPolicy:            &config.TieringPolicy,
		}
		// Check and map nested configurations for Snaplock if present
		if &config.SnaplockConfiguration != nil {
			snapConfig := config.SnaplockConfiguration
			ontapConfig.SnaplockConfiguration = &awstypes.CreateSnaplockConfiguration{
				AuditLogVolume:          snapConfig.AuditLogVolume,
				AutocommitPeriod:        snapConfig.AutocommitPeriod,
				PrivilegedDelete:        snapConfig.PrivilegedDelete,
				RetentionPeriod:         snapConfig.RetentionPeriod,
				SnaplockType:            snapConfig.SnaplockType,
				VolumeAppendModeEnabled: snapConfig.VolumeAppendModeEnabled,
			}
		}
	}

	in := &fsx.CreateVolumeFromBackupInput{
		BackupId:           aws.String(plan.BackupID.ValueString()),
		Name:               aws.String(plan.Name.ValueString()),
		OntapConfiguration: &ontapConfig,
		//Tags: ,
	}

	out, err := conn.CreateVolumeFromBackup(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FSx, create.ErrActionCreating, ResNameOntapVolumeFromBackup, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Volume.VolumeId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FSx, create.ErrActionCreating, ResNameOntapVolumeFromBackup, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.Volume.ResourceARN)
	plan.ID = flex.StringToFramework(ctx, out.Volume.VolumeId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitOntapVolumeFromBackupCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FSx, create.ErrActionWaitingForCreation, ResNameOntapVolumeFromBackup, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceOntapVolumeFromBackup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().FSxClient(ctx)

	var state resourceOntapVolumeFromBackupData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindOntapVolumeByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FSx, create.ErrActionSetting, ResNameOntapVolumeFromBackup, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.ResourceARN)
	state.ID = flex.StringToFramework(ctx, out.VolumeId)
	state.Name = flex.StringToFramework(ctx, out.Name)
	/*	if out.OntapConfiguration != nil {
		state.ONTAPConfiguration.ONTAPVolumeType = out.OntapConfiguration.OntapVolumeType

	}*/

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceOntapVolumeFromBackup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().FSxClient(ctx)

	var plan, state resourceOntapVolumeFromBackupData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.Name.Equal(state.Name) ||
		!plan.ID.Equal(state.ID) ||
		!plan.BackupID.Equal(state.BackupID) ||
		!plan.Type.Equal(state.Type) {

		updateOntapConfig := &fsx.UpdateVolumeInput{}

		if !plan.ONTAPConfiguration.CopyTagsToBackup.Equal(state.ONTAPConfiguration.CopyTagsToBackup) {
			updateOntapConfig.OntapConfiguration.CopyTagsToBackups = aws.Bool(plan.ONTAPConfiguration.CopyTagsToBackup.ValueBool())
		}

		if !plan.ONTAPConfiguration.JunctionPath.Equal(state.ONTAPConfiguration.JunctionPath) {
			updateOntapConfig.OntapConfiguration.JunctionPath = aws.String(plan.ONTAPConfiguration.JunctionPath.ValueString())
		}

		in := &fsx.UpdateVolumeInput{
			VolumeId:           aws.String(plan.ID.ValueString()),
			OntapConfiguration: updateOntapConfig.OntapConfiguration,
		}

		if &plan.ONTAPConfiguration.SnaplockConfiguration != nil {
			in.OntapConfiguration.SnaplockConfiguration = updateOntapConfig.OntapConfiguration.SnaplockConfiguration
		}

		out, err := conn.UpdateVolume(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.FSx, create.ErrActionUpdating, ResNameOntapVolumeFromBackup, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Volume == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.FSx, create.ErrActionUpdating, ResNameOntapVolumeFromBackup, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.Volume.ResourceARN)
		plan.ID = flex.StringToFramework(ctx, out.Volume.VolumeId)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitOntapVolumeFromBackupAvailable(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FSx, create.ErrActionWaitingForUpdate, ResNameOntapVolumeFromBackup, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceOntapVolumeFromBackup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().FSxClient(ctx)

	var state resourceOntapVolumeFromBackupData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &fsx.DeleteVolumeInput{
		VolumeId: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteVolume(ctx, in)
	if err != nil {
		var nfe *awstypes.VolumeNotFound
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FSx, create.ErrActionDeleting, ResNameOntapVolumeFromBackup, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitOntapVolumeFromBackupDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FSx, create.ErrActionWaitingForDeletion, ResNameOntapVolumeFromBackup, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceOntapVolumeFromBackup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Available"
	statusCreated       = "Created"
	statusCreating      = "Creating"
	statusFailed        = "Failed"
	statusMisconfigured = "Misconfigured"
)

func waitOntapVolumeFromBackupCreated(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*fsx.DescribeVolumesOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusCreated},
		Refresh:                   statusOntapVolumeFromBackup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*fsx.DescribeVolumesOutput); ok {
		return out, err
	}

	return nil, err
}

func waitOntapVolumeFromBackupAvailable(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*fsx.DescribeVolumesOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusNormal},
		Refresh:                   statusOntapVolumeFromBackup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*fsx.DescribeVolumesOutput); ok {
		return out, err
	}

	return nil, err
}

func waitOntapVolumeFromBackupDeleted(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*fsx.DescribeVolumesOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusOntapVolumeFromBackup(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*fsx.DescribeVolumesOutput); ok {
		return out, err
	}

	return nil, err
}

func statusOntapVolumeFromBackup(ctx context.Context, conn *fsx.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindOntapVolumeByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		volumeStatus := getVolumeLifecycleString(out.Lifecycle)

		return out, volumeStatus, nil
	}
}

func getVolumeLifecycleString(lifecycle awstypes.VolumeLifecycle) string {
	switch lifecycle {
	case awstypes.VolumeLifecycleAvailable:
		return "AVAILABLE"
	case awstypes.VolumeLifecycleCreated:
		return "CREATED"
	case awstypes.VolumeLifecycleCreating:
		return "CREATING"
	case awstypes.VolumeLifecycleDeleting:
		return "DELETING"
	case awstypes.VolumeLifecycleFailed:
		return "FAILED"
	case awstypes.VolumeLifecycleMisconfigured:
		return "MISCONFIGURED"
	case awstypes.VolumeLifecyclePending:
		return "PENDING"
	default:
		return "UNKNOWN"
	}
}

func FindOntapVolumeByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.Volume, error) {
	in := &fsx.DescribeVolumesInput{
		VolumeIds: []string{id},
	}

	out, err := conn.DescribeVolumes(ctx, in)
	if err != nil {
		var nfe *awstypes.VolumeNotFound
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}
	volume := out.Volumes[0]
	awsVolume := &awstypes.Volume{
		VolumeId:           volume.VolumeId,
		Name:               volume.Name,
		OntapConfiguration: volume.OntapConfiguration,
		ResourceARN:        volume.ResourceARN,
		Tags:               volume.Tags,
		VolumeType:         volume.VolumeType,
		Lifecycle:          volume.Lifecycle,
	}

	return awsVolume, nil
}

type resourceOntapVolumeFromBackupData struct {
	ID  types.String `tfsdk:"id"`
	ARN types.String `tfsdk:"arn"`
	//ComplexArgument types.List     `tfsdk:"complex_argument"`
	BackupID           types.String       `tfsdk:"backup_id"`
	Name               types.String       `tfsdk:"name"`
	ONTAPConfiguration ontapConfiguration `tfsdk:"ontap_configuration"`
	Timeouts           timeouts.Value     `tfsdk:"timeouts"`
	Type               types.String       `tfsdk:"type"`
}

type ontapConfiguration struct {
	CopyTagsToBackup         types.Bool                           `tfsdk:"copy_tags_to_backups"`
	JunctionPath             types.String                         `tfsdk:"junction_path"`
	ONTAPVolumeType          awstypes.InputOntapVolumeType        `tfsdk:"ontap_volume_type"`
	SecurityStyle            awstypes.SecurityStyle               `tfsdk:"security_style"`
	SizeInMegaBytes          int32                                `tfsdk:"size_in_megabytes"`
	SnaplockConfiguration    awstypes.CreateSnaplockConfiguration `tfsdk:"snaplock_configuraiton"`
	SnapshotPolicy           types.String                         `tfsdk:"snapshot_policy"`
	StorageEfficiencyEnabled types.Bool                           `tfsdk:"storage_efficiency_enabled"`
	StorageVirtualMachineID  types.String                         `tfsdk:"storage_virtualmachine_id"`
	TieringPolicy            awstypes.TieringPolicy               `tfsdk:"tiering_policy"`
}
type snaplockConfiguration struct {
	AuditLogVolume          types.Bool                       `tfsdk:"audit_log_volume"`
	AutoCommitPeriod        awstypes.AutocommitPeriod        `tfsdk:"auto_commit_period"`
	PrivilegedDelete        awstypes.PrivilegedDelete        `tfsdk:"privileged_delete"`
	RetentionPeriod         awstypes.SnaplockRetentionPeriod `tfsdk:"retention_period"`
	SnaplockType            awstypes.SnaplockType            `tfsdk:"snaplock_type"`
	VolumeAppendModeEnabled types.Bool                       `tfsdk:"volume_append_mode_enabled"`
}

// Exports for use in tests only.
var (
	ResourceOntapVolumeFromBackup = newResourceOntapVolumeFromBackup
)
