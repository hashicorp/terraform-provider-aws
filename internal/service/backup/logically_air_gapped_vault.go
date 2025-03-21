// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_backup_logically_air_gapped_vault", name="Logically Air Gapped Vault")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newLogicallyAirGappedVaultResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &logicallyAirGappedVaultResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)

	return r, nil
}

type logicallyAirGappedVaultResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[logicallyAirGappedVaultResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *logicallyAirGappedVaultResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"max_retention_days": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"min_retention_days": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(7),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-\_]{2,50}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *logicallyAirGappedVaultResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data logicallyAirGappedVaultResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	name := data.BackupVaultName.ValueString()
	input := &backup.CreateLogicallyAirGappedBackupVaultInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.BackupVaultTags = getTagsIn(ctx)
	input.CreatorRequestId = aws.String(sdkid.UniqueId())

	output, err := conn.CreateLogicallyAirGappedBackupVault(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Backup Logically Air Gapped Vault (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.BackupVaultARN = fwflex.StringToFramework(ctx, output.BackupVaultArn)
	data.ID = fwflex.StringToFramework(ctx, output.BackupVaultName)

	if _, err := waitLogicallyAirGappedVaultCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Backup Logically Air Gapped Vault (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *logicallyAirGappedVaultResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data logicallyAirGappedVaultResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	output, err := findLogicallyAirGappedBackupVaultByName(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Backup Logically Air Gapped Vault (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *logicallyAirGappedVaultResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data logicallyAirGappedVaultResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BackupClient(ctx)

	input := backup.DeleteBackupVaultInput{
		BackupVaultName: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteBackupVault(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || tfawserr.ErrCodeEquals(err, errCodeAccessDeniedException) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Backup Logically Air Gapped Vault (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type logicallyAirGappedVaultResourceModel struct {
	BackupVaultARN   types.String   `tfsdk:"arn"`
	BackupVaultName  types.String   `tfsdk:"name"`
	ID               types.String   `tfsdk:"id"`
	MaxRetentionDays types.Int64    `tfsdk:"max_retention_days"`
	MinRetentionDays types.Int64    `tfsdk:"min_retention_days"`
	Tags             tftags.Map     `tfsdk:"tags"`
	TagsAll          tftags.Map     `tfsdk:"tags_all"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

func findLogicallyAirGappedBackupVaultByName(ctx context.Context, conn *backup.Client, name string) (*backup.DescribeBackupVaultOutput, error) { // nosemgrep:ci.backup-in-func-name
	output, err := findVaultByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.VaultType != awstypes.VaultTypeLogicallyAirGappedBackupVault {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output, nil
}

func statusLogicallyAirGappedVault(ctx context.Context, conn *backup.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findLogicallyAirGappedBackupVaultByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.VaultState), nil
	}
}

func waitLogicallyAirGappedVaultCreated(ctx context.Context, conn *backup.Client, name string, timeout time.Duration) (*backup.DescribeBackupVaultOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VaultStateCreating),
		Target:                    enum.Slice(awstypes.VaultStateAvailable),
		Refresh:                   statusLogicallyAirGappedVault(ctx, conn, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*backup.DescribeBackupVaultOutput); ok {
		return output, err
	}

	return nil, err
}
