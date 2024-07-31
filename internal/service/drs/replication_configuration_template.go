// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package drs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/drs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/drs/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Replication Configuration Template")
// @Tags(identifierAttribute="arn")
func newReplicationConfigurationTemplateResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &replicationConfigurationTemplateResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultUpdateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type replicationConfigurationTemplateResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *replicationConfigurationTemplateResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_drs_replication_configuration_template"
}

func (r *replicationConfigurationTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"associate_default_security_group": schema.BoolAttribute{
				Required: true,
			},
			"auto_replicate_new_disks": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"bandwidth_throttling": schema.Int64Attribute{
				Required: true,
			},
			"create_public_ip": schema.BoolAttribute{
				Required: true,
			},
			"data_plane_routing": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ReplicationConfigurationDataPlaneRouting](),
			},
			"default_large_staging_disk_type": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ReplicationConfigurationDefaultLargeStagingDiskType](),
			},
			"ebs_encryption": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ReplicationConfigurationEbsEncryption](),
			},
			"ebs_encryption_key_arn": schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
			"replication_server_instance_type": schema.StringAttribute{
				Required: true,
			},
			"replication_servers_security_groups_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"staging_area_subnet_id": schema.StringAttribute{
				Required: true,
			},

			"staging_area_tags": tftags.TagsAttribute(),
			names.AttrTags:      tftags.TagsAttribute(),
			names.AttrTagsAll:   tftags.TagsAttributeComputedOnly(),

			"use_dedicated_replication_server": schema.BoolAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"pit_policy": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[pitPolicy](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrEnabled: schema.BoolAttribute{
							Optional: true,
						},
						names.AttrInterval: schema.Int64Attribute{
							Required: true,
						},
						"retention_duration": schema.Int64Attribute{
							Required: true,
						},
						"rule_id": schema.Int64Attribute{
							Optional: true,
						},
						"units": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.PITPolicyRuleUnits](),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *replicationConfigurationTemplateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data replicationConfigurationTemplateResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DRSClient(ctx)

	input := &drs.CreateReplicationConfigurationTemplateInput{}
	response.Diagnostics.Append(flex.Expand(context.WithValue(ctx, flex.ResourcePrefix, ResPrefixReplicationConfigurationTemplate), data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	_, err := conn.CreateReplicationConfigurationTemplate(ctx, input)
	if err != nil {
		create.AddError(&response.Diagnostics, names.DRS, create.ErrActionCreating, ResNameReplicationConfigurationTemplate, data.ID.ValueString(), err)

		return
	}

	output, err := waitReplicationConfigurationTemplateAvailable(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		create.AddError(&response.Diagnostics, names.DRS, create.ErrActionWaitingForCreation, ResNameReplicationConfigurationTemplate, data.ID.ValueString(), err)

		return
	}

	response.Diagnostics.Append(flex.Flatten(context.WithValue(ctx, flex.ResourcePrefix, ResPrefixReplicationConfigurationTemplate), output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *replicationConfigurationTemplateResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data replicationConfigurationTemplateResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DRSClient(ctx)

	output, err := findReplicationConfigurationTemplateByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		create.AddError(&response.Diagnostics, names.DRS, create.ErrActionReading, ResNameReplicationConfigurationTemplate, data.ID.ValueString(), err)

		return
	}

	response.Diagnostics.Append(flex.Flatten(context.WithValue(ctx, flex.ResourcePrefix, ResPrefixReplicationConfigurationTemplate), output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *replicationConfigurationTemplateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new replicationConfigurationTemplateResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DRSClient(ctx)

	if replicationConfigurationTemplateHasChanges(ctx, new, old) {
		input := &drs.UpdateReplicationConfigurationTemplateInput{}
		response.Diagnostics.Append(flex.Expand(context.WithValue(ctx, flex.ResourcePrefix, ResPrefixReplicationConfigurationTemplate), new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateReplicationConfigurationTemplate(ctx, input)
		if err != nil {
			create.AddError(&response.Diagnostics, names.DRS, create.ErrActionUpdating, ResNameReplicationConfigurationTemplate, new.ID.ValueString(), err)

			return
		}

		if _, err := waitReplicationConfigurationTemplateAvailable(ctx, conn, old.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			create.AddError(&response.Diagnostics, names.DRS, create.ErrActionWaitingForUpdate, ResNameReplicationConfigurationTemplate, new.ID.ValueString(), err)

			return
		}
	}

	output, err := findReplicationConfigurationTemplateByID(ctx, conn, old.ID.ValueString())
	if err != nil {
		create.AddError(&response.Diagnostics, names.DRS, create.ErrActionUpdating, ResNameReplicationConfigurationTemplate, old.ID.ValueString(), err)

		return
	}

	response.Diagnostics.Append(flex.Flatten(context.WithValue(ctx, flex.ResourcePrefix, ResPrefixReplicationConfigurationTemplate), output, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *replicationConfigurationTemplateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data replicationConfigurationTemplateResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DRSClient(ctx)

	tflog.Debug(ctx, "deleting DRS Replication Configuration Template", map[string]interface{}{
		names.AttrID: data.ID.ValueString(),
	})

	input := &drs.DeleteReplicationConfigurationTemplateInput{
		ReplicationConfigurationTemplateID: aws.String(data.ID.ValueString()),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute, func() (interface{}, error) {
		return conn.DeleteReplicationConfigurationTemplate(ctx, input)
	}, "DependencyViolation")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		create.AddError(&response.Diagnostics, names.DRS, create.ErrActionDeleting, ResNameReplicationConfigurationTemplate, data.ID.ValueString(), err)

		return
	}

	if _, err := waitReplicationConfigurationTemplateDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		create.AddError(&response.Diagnostics, names.DRS, create.ErrActionWaitingForDeletion, ResNameReplicationConfigurationTemplate, data.ID.ValueString(), err)

		return
	}
}

func (r *replicationConfigurationTemplateResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findReplicationConfigurationTemplate(ctx context.Context, conn *drs.Client, input *drs.DescribeReplicationConfigurationTemplatesInput) (*awstypes.ReplicationConfigurationTemplate, error) {
	output, err := findReplicationConfigurationTemplates(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReplicationConfigurationTemplates(ctx context.Context, conn *drs.Client, input *drs.DescribeReplicationConfigurationTemplatesInput) ([]awstypes.ReplicationConfigurationTemplate, error) {
	var output []awstypes.ReplicationConfigurationTemplate

	pages := drs.NewDescribeReplicationConfigurationTemplatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Items...)
	}

	return output, nil
}

func findReplicationConfigurationTemplateByID(ctx context.Context, conn *drs.Client, id string) (*awstypes.ReplicationConfigurationTemplate, error) {
	input := &drs.DescribeReplicationConfigurationTemplatesInput{
		//ReplicationConfigurationTemplateIDs: []string{id}, // Uncomment when SDK supports this, currently MAX of 1 so you find it anyway
	}

	return findReplicationConfigurationTemplate(ctx, conn, input)
}

const (
	replicationConfigurationTemplateAvailable = "AVAILABLE"
)

func statusReplicationConfigurationTemplate(ctx context.Context, conn *drs.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findReplicationConfigurationTemplateByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, replicationConfigurationTemplateAvailable, nil
	}
}

func waitReplicationConfigurationTemplateAvailable(ctx context.Context, conn *drs.Client, id string, timeout time.Duration) (*awstypes.ReplicationConfigurationTemplate, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{},
		Target:     []string{replicationConfigurationTemplateAvailable},
		Refresh:    statusReplicationConfigurationTemplate(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationConfigurationTemplate); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationConfigurationTemplateDeleted(ctx context.Context, conn *drs.Client, id string, timeout time.Duration) (*awstypes.ReplicationConfigurationTemplate, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationConfigurationTemplateAvailable},
		Target:     []string{},
		Refresh:    statusReplicationConfigurationTemplate(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationConfigurationTemplate); ok {
		return output, err
	}

	return nil, err
}

type replicationConfigurationTemplateResourceModel struct {
	ARN                                 types.String                                                                     `tfsdk:"arn"`
	AssociateDefaultSecurityGroup       types.Bool                                                                       `tfsdk:"associate_default_security_group"`
	AutoReplicateNewDisks               types.Bool                                                                       `tfsdk:"auto_replicate_new_disks"`
	BandwidthThrottling                 types.Int64                                                                      `tfsdk:"bandwidth_throttling"`
	CreatePublicIP                      types.Bool                                                                       `tfsdk:"create_public_ip"`
	DataPlaneRouting                    fwtypes.StringEnum[awstypes.ReplicationConfigurationDataPlaneRouting]            `tfsdk:"data_plane_routing"`
	DefaultLargeStagingDiskType         fwtypes.StringEnum[awstypes.ReplicationConfigurationDefaultLargeStagingDiskType] `tfsdk:"default_large_staging_disk_type"`
	EBSEncryption                       fwtypes.StringEnum[awstypes.ReplicationConfigurationEbsEncryption]               `tfsdk:"ebs_encryption"`
	EBSEncryptionKeyARN                 types.String                                                                     `tfsdk:"ebs_encryption_key_arn"`
	ID                                  types.String                                                                     `tfsdk:"id"`
	PitPolicy                           fwtypes.ListNestedObjectValueOf[pitPolicy]                                       `tfsdk:"pit_policy"`
	ReplicationServerInstanceType       types.String                                                                     `tfsdk:"replication_server_instance_type"`
	ReplicationServersSecurityGroupsIDs types.List                                                                       `tfsdk:"replication_servers_security_groups_ids"`
	StagingAreaSubnetID                 types.String                                                                     `tfsdk:"staging_area_subnet_id"`
	UseDedicatedReplicationServer       types.Bool                                                                       `tfsdk:"use_dedicated_replication_server"`
	StagingAreaTags                     types.Map                                                                        `tfsdk:"staging_area_tags"`
	Tags                                types.Map                                                                        `tfsdk:"tags"`
	TagsAll                             types.Map                                                                        `tfsdk:"tags_all"`
	Timeouts                            timeouts.Value                                                                   `tfsdk:"timeouts"`
}

type pitPolicy struct {
	Enabled           types.Bool                                      `tfsdk:"enabled"`
	Interval          types.Int64                                     `tfsdk:"interval"`
	RetentionDuration types.Int64                                     `tfsdk:"retention_duration"`
	RuleID            types.Int64                                     `tfsdk:"rule_id"`
	Units             fwtypes.StringEnum[awstypes.PITPolicyRuleUnits] `tfsdk:"units"`
}

func replicationConfigurationTemplateHasChanges(_ context.Context, plan, state replicationConfigurationTemplateResourceModel) bool {
	return !plan.AssociateDefaultSecurityGroup.Equal(state.AssociateDefaultSecurityGroup) ||
		!plan.AutoReplicateNewDisks.Equal(state.AutoReplicateNewDisks) ||
		!plan.BandwidthThrottling.Equal(state.BandwidthThrottling) ||
		!plan.CreatePublicIP.Equal(state.CreatePublicIP) ||
		!plan.DataPlaneRouting.Equal(state.DataPlaneRouting) ||
		!plan.DefaultLargeStagingDiskType.Equal(state.DefaultLargeStagingDiskType) ||
		!plan.EBSEncryption.Equal(state.EBSEncryption) ||
		!plan.EBSEncryptionKeyARN.Equal(state.EBSEncryptionKeyARN) ||
		!plan.ID.Equal(state.ID) ||
		!plan.PitPolicy.Equal(state.PitPolicy) ||
		!plan.ReplicationServerInstanceType.Equal(state.ReplicationServerInstanceType) ||
		!plan.ReplicationServersSecurityGroupsIDs.Equal(state.ReplicationServersSecurityGroupsIDs) ||
		!plan.StagingAreaSubnetID.Equal(state.StagingAreaSubnetID) ||
		!plan.UseDedicatedReplicationServer.Equal(state.UseDedicatedReplicationServer)
}
