// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

// @FrameworkResource("aws_s3files_file_system", name="File System")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetFileSystemOutput")
// @Testing(existsTakesT=true, destroyTakesT=true)
// @Testing(hasNoPreExistingResource=true)
func newFileSystemResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &fileSystemResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type fileSystemResource struct {
	framework.ResourceWithModel[fileSystemResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *fileSystemResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"accept_bucket_warning": schema.BoolAttribute{
				Optional: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrBucket: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "S3 bucket ARN",
			},
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Creation time",
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKMSKeyID: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "KMS key ID for encryption",
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "File system name",
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "AWS account ID of the owner",
			},
			names.AttrPrefix: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "S3 bucket prefix",
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "IAM role ARN for S3 access",
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "File system status",
			},
			names.AttrStatusMessage: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Status message",
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

func (r *fileSystemResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data fileSystemResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.CreateFileSystemInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix("FileSystem")))
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateFileSystem(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrBucket, data.Bucket.ValueString()) // Bucket is the only required field
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.FileSystemId)
	data.ARN = fwflex.StringToFramework(ctx, output.FileSystemArn)

	fileSystem, err := waitFileSystemCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, fileSystem, &data, fwflex.WithFieldNamePrefix("FileSystem")))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *fileSystemResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data fileSystemResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	output, err := findFileSystemByID(ctx, conn, data.ID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	flattenFileSystemResource(ctx, output, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func flattenFileSystemResource(ctx context.Context, output *s3files.GetFileSystemOutput, data *fileSystemResourceModel, diags *diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	smerr.AddEnrich(ctx, diags, fwflex.Flatten(ctx, output, data, fwflex.WithFieldNamePrefix("FileSystem")))
}

func (r *fileSystemResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data fileSystemResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// Only tags can be updated; transparent tagging handles the update
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *fileSystemResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data fileSystemResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.DeleteFileSystemInput{
		FileSystemId: data.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteFileSystem(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return // Resource already deleted
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	_, err = waitFileSystemDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
	}
}

type fileSystemResourceModel struct {
	framework.WithRegionModel
	AcceptBucketWarning types.Bool        `tfsdk:"accept_bucket_warning" autoflex:",omitempty"`
	ARN                 types.String      `tfsdk:"arn"`
	Bucket              types.String      `tfsdk:"bucket"`
	CreationTime        timetypes.RFC3339 `tfsdk:"creation_time"`
	ID                  types.String      `tfsdk:"id"`
	KmsKeyId            fwtypes.ARN       `tfsdk:"kms_key_id"`
	Name                types.String      `tfsdk:"name"`
	OwnerID             types.String      `tfsdk:"owner_id"`
	Prefix              types.String      `tfsdk:"prefix" autoflex:",omitempty"`
	RoleArn             fwtypes.ARN       `tfsdk:"role_arn"`
	Status              types.String      `tfsdk:"status"`
	StatusMessage       types.String      `tfsdk:"status_message"`
	Tags                tftags.Map        `tfsdk:"tags"`
	TagsAll             tftags.Map        `tfsdk:"tags_all"`
	Timeouts            timeouts.Value    `tfsdk:"timeouts"`
}

func findFileSystemByID(ctx context.Context, conn *s3files.Client, id string) (*s3files.GetFileSystemOutput, error) {
	input := s3files.GetFileSystemInput{
		FileSystemId: &id,
	}

	output, err := conn.GetFileSystem(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func waitFileSystemCreated(ctx context.Context, conn *s3files.Client, id string, timeout time.Duration) (*s3files.GetFileSystemOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.LifeCycleStateCreating),
		Target:     enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateError),
		Refresh:    statusFileSystem(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3files.GetFileSystemOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFileSystemDeleted(ctx context.Context, conn *s3files.Client, id string, timeout time.Duration) (*s3files.GetFileSystemOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateDeleting),
		Target:  []string{},
		Refresh: statusFileSystem(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3files.GetFileSystemOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusFileSystem(_ context.Context, conn *s3files.Client, id string) retry.StateRefreshFunc {
	const iamPropagationBuffer = 3 * time.Minute
	firstErrorTime := time.Time{}

	return func(ctx context.Context) (any, string, error) {
		output, err := findFileSystemByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		if output != nil {
			status := string(output.Status)
			statusMsg := aws.ToString(output.StatusMessage)

			// s3files quirk: during creation, the resource may be in 'Creating' state with an error message
			// such as IAM permissions
			//   - Access denied: S3 Files does not have permissions to assume the provided role.
			//   - Access denied: The provided role does not have permission to call s3:HeadObject on the provided bucket.
			if output.Status == awstypes.LifeCycleStateCreating && strings.Contains(statusMsg, "Access denied") && strings.Contains(statusMsg, "does not have permission") {
				if firstErrorTime.IsZero() {
					firstErrorTime = time.Now()
				}

				// Allow time for IAM propagation
				if time.Since(firstErrorTime) > iamPropagationBuffer {
					return output, string(awstypes.LifeCycleStateError), smarterr.Errorf("in \"%s\" state with status message: %s", status, statusMsg)
				}
				// Still within propagation window, keep waiting
				return output, status, nil
			}

			// Clear error timer if no error message
			if statusMsg == "" {
				firstErrorTime = time.Time{}
			}

			// Explicit error state
			if output.Status == awstypes.LifeCycleStateError {
				return output, status, smarterr.Errorf("in \"%s\" state with status message: %s", status, statusMsg)
			}

			return output, status, nil
		}

		return output, string(output.Status), nil
	}
}
