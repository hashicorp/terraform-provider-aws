// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Generated identity tests are not used: a snapshot needs a running Apache Flink application,
// whose test configuration depends on a code JAR in test-fixtures that ConfigDirectory tests cannot reference.
// @FrameworkResource("aws_kinesisanalyticsv2_snapshot", name="Snapshot")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("name")
// @Testing(identityTest=false)
// @Testing(hasNoPreExistingResource=true)
func newSnapshotResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &snapshotResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	// enhancedSnapshotFieldNamePrefix maps model fields such as Name to API fields such as SnapshotName.
	enhancedSnapshotFieldNamePrefix = "Snapshot"

	snapshotErrorCodeApplicationStateCaptureFailed = "ApplicationStateCaptureFailed"
)

type snapshotResource struct {
	framework.ResourceWithModel[snapshotResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *snapshotResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: applicationNameValidators(),
			},
			"application_version_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
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
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: enhancedSnapshotNameValidators(),
			},
			"runtime_environment": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *snapshotResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().KinesisAnalyticsV2Client(ctx)

	var plan snapshotResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	var input kinesisanalyticsv2.CreateSnapshotInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix(enhancedSnapshotFieldNamePrefix)))
	if response.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	snapshot, err := createEnhancedSnapshot(ctx, conn, &input, r.CreateTimeout(ctx, plan.Timeouts), func() {
		// Persist the identity once a snapshot exists, so that one that fails to become READY is tainted, not orphaned:
		// a FAILED snapshot keeps its name reserved until deleted.
		// Region is needed too, or a snapshot created by region override is refreshed in the wrong Region.
		smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrName), name))
		smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrRegion), plan.Region))
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, snapshot, &plan, fwflex.WithFieldNamePrefix(enhancedSnapshotFieldNamePrefix)))
	if response.Diagnostics.HasError() {
		return
	}
	plan.setSourceApplication(ctx, snapshot)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *snapshotResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().KinesisAnalyticsV2Client(ctx)

	var state snapshotResourceModel
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

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, snapshot, &state, fwflex.WithFieldNamePrefix(enhancedSnapshotFieldNamePrefix)))
	if response.Diagnostics.HasError() {
		return
	}
	state.setSourceApplication(ctx, snapshot)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *snapshotResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan snapshotResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	// Only tags can be updated; transparent tagging handles the update.
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *snapshotResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().KinesisAnalyticsV2Client(ctx)

	var state snapshotResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	if err := deleteEnhancedSnapshot(ctx, conn, state.Name.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.Name.String())
	}
}

type snapshotResourceModel struct {
	framework.WithRegionModel
	ApplicationName      types.String      `tfsdk:"application_name"`
	ApplicationVersionID types.Int64       `tfsdk:"application_version_id"`
	ARN                  types.String      `tfsdk:"arn"`
	CreationTimestamp    timetypes.RFC3339 `tfsdk:"creation_time"`
	Description          types.String      `tfsdk:"description"`
	Name                 types.String      `tfsdk:"name"`
	RuntimeEnvironment   types.String      `tfsdk:"runtime_environment"`
	SizeInBytes          types.Int64       `tfsdk:"size_in_bytes"`
	Tags                 tftags.Map        `tfsdk:"tags"`
	TagsAll              tftags.Map        `tfsdk:"tags_all"`
	Timeouts             timeouts.Value    `tfsdk:"timeouts"`
}

// setSourceApplication copies the source application, which AutoFlex cannot reach, into the model.
// SnapshotSourceDescription holds either a source application or, for imported snapshots, an S3 location.
func (data *snapshotResourceModel) setSourceApplication(ctx context.Context, snapshot *awstypes.SnapshotInfoDetails) {
	if source := snapshot.SnapshotSourceDescription; source != nil && source.SourceApplicationDescription != nil {
		data.ApplicationName = fwflex.StringToFramework(ctx, source.SourceApplicationDescription.ApplicationName)
		data.ApplicationVersionID = fwflex.Int64ToFramework(ctx, source.SourceApplicationDescription.ApplicationVersionId)
	}
}

func applicationNameValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, 128),
		stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
	}
}

func enhancedSnapshotNameValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, 256),
		stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
	}
}

func findEnhancedSnapshotByName(ctx context.Context, conn *kinesisanalyticsv2.Client, name string) (*awstypes.SnapshotInfoDetails, error) {
	input := kinesisanalyticsv2.DescribeSnapshotInput{
		SnapshotName: aws.String(name),
	}

	output, err := conn.DescribeSnapshot(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.SnapshotInfoDetails == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.SnapshotInfoDetails, nil
}

func findEnhancedSnapshotOperationByTwoPartKey(ctx context.Context, conn *kinesisanalyticsv2.Client, snapshotName, operationID string) (*awstypes.SnapshotOperationInfoDetails, error) {
	input := kinesisanalyticsv2.DescribeSnapshotOperationInput{
		SnapshotName:        aws.String(snapshotName),
		SnapshotOperationId: aws.String(operationID),
	}

	output, err := conn.DescribeSnapshotOperation(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.SnapshotOperationInfoDetails == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.SnapshotOperationInfoDetails, nil
}

func statusEnhancedSnapshot(conn *kinesisanalyticsv2.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findEnhancedSnapshotByName(ctx, conn, name)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, string(output.SnapshotStatus), nil
	}
}

// createEnhancedSnapshot creates a snapshot and waits for it to become READY, calling created once a snapshot exists.
//
// For some time after an application reports RUNNING its Flink job is not yet running, and a snapshot fails with
// ApplicationStateCaptureFailed, which the API reports as retryable. Applying an application with
// start_application = true together with a snapshot of it hits this every time, so the failed snapshot is deleted,
// which frees its name, and the snapshot is retried a bounded number of times.
func createEnhancedSnapshot(ctx context.Context, conn *kinesisanalyticsv2.Client, input *kinesisanalyticsv2.CreateSnapshotInput, timeout time.Duration, created func()) (*awstypes.SnapshotInfoDetails, error) {
	const (
		retryDelay  = 30 * time.Second
		maxAttempts = 5
	)
	name := aws.ToString(input.SnapshotName)
	deadline := time.Now().Add(timeout)

	for attempt := 1; ; attempt++ {
		output, err := conn.CreateSnapshot(ctx, input)
		if err != nil {
			return nil, smarterr.NewError(err)
		}
		created()

		snapshot, failure, err := waitEnhancedSnapshotCreated(ctx, conn, name, aws.ToString(output.SnapshotOperationId), time.Until(deadline))
		if err == nil {
			return snapshot, nil
		}
		// A capture failure that persists past the startup window is not the startup race; stop retrying.
		if failure == nil || aws.ToString(failure.ErrorCode) != snapshotErrorCodeApplicationStateCaptureFailed || attempt >= maxAttempts || time.Until(deadline) < 2*retryDelay {
			if attempt > 1 {
				err = fmt.Errorf("after %d attempts: %w", attempt, err)
			}
			return snapshot, err
		}

		tflog.Info(ctx, "retrying enhanced snapshot after application state capture failure", map[string]any{"snapshot_name": name, "attempt": attempt})
		if err := deleteEnhancedSnapshot(ctx, conn, name, time.Until(deadline)); err != nil {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryDelay):
		}
	}
}

// waitEnhancedSnapshotCreated waits for a snapshot created by CreateSnapshot or ImportSnapshot to become READY.
// A failure is reported only on the snapshot operation, so its error info is returned and added to the error.
func waitEnhancedSnapshotCreated(ctx context.Context, conn *kinesisanalyticsv2.Client, name, operationID string, timeout time.Duration) (*awstypes.SnapshotInfoDetails, *awstypes.SnapshotErrorInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotStatusCreating),
		Target:  enum.Slice(awstypes.SnapshotStatusReady),
		Refresh: statusEnhancedSnapshot(conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	output, _ := outputRaw.(*awstypes.SnapshotInfoDetails)

	if output == nil || output.SnapshotStatus != awstypes.SnapshotStatusFailed || operationID == "" {
		return output, nil, smarterr.NewError(err)
	}

	operation, opErr := findEnhancedSnapshotOperationByTwoPartKey(ctx, conn, name, operationID)
	if opErr != nil {
		return output, nil, smarterr.NewError(errors.Join(err, fmt.Errorf("reading snapshot operation (%s): %w", operationID, opErr)))
	}

	failure := operation.SnapshotOperationFailureDetails
	if failure == nil || failure.ErrorInfo == nil {
		return output, nil, smarterr.NewError(err)
	}

	err = errors.Join(err, fmt.Errorf("%s: %s", aws.ToString(failure.ErrorInfo.ErrorCode), aws.ToString(failure.ErrorInfo.ErrorString)))

	return output, failure.ErrorInfo, smarterr.NewError(err)
}

func waitEnhancedSnapshotDeleted(ctx context.Context, conn *kinesisanalyticsv2.Client, name string, timeout time.Duration) (*awstypes.SnapshotInfoDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotStatusDeleting),
		Target:  []string{},
		Refresh: statusEnhancedSnapshot(conn, name),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.SnapshotInfoDetails); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// deleteEnhancedSnapshot deletes a snapshot and waits for it to be gone.
// DeleteSnapshot requires the creation timestamp, which is read fresh so that a snapshot left
// tainted by a failed create, with no timestamp in state, can still be deleted.
func deleteEnhancedSnapshot(ctx context.Context, conn *kinesisanalyticsv2.Client, name string, timeout time.Duration) error {
	snapshot, err := findEnhancedSnapshotByName(ctx, conn, name)
	if retry.NotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	input := kinesisanalyticsv2.DeleteSnapshotInput{
		SnapshotCreationTimestamp: snapshot.SnapshotCreationTimestamp,
		SnapshotName:              aws.String(name),
	}

	_, err = conn.DeleteSnapshot(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}
	if err != nil {
		return smarterr.NewError(err)
	}

	_, err = waitEnhancedSnapshotDeleted(ctx, conn, name, timeout)

	return err
}
