// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3files_file_system_policy", name="File System Policy")
// @IdentityAttribute("file_system_id")
// @Testing(existsTakesT=true, destroyTakesT=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="file_system_id")
func newFileSystemPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &fileSystemPolicyResource{}, nil
}

type fileSystemPolicyResource struct {
	framework.ResourceWithModel[fileSystemPolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *fileSystemPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrFileSystemID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "File system ID",
			},
			names.AttrPolicy: schema.StringAttribute{
				CustomType:  fwtypes.IAMPolicyType,
				Required:    true,
				Description: "File system policy JSON",
			},
		},
	}
}

func (r *fileSystemPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data fileSystemPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.PutFileSystemPolicyInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix("FileSystem")))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutFileSystemPolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *fileSystemPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data fileSystemPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	output, err := findFileSystemPolicyByID(ctx, conn, data.FileSystemID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.FileSystemID.ValueString())
		return
	}

	flattenFileSystemPolicyResource(ctx, output, &data, &response.Diagnostics)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func flattenFileSystemPolicyResource(ctx context.Context, output *s3files.GetFileSystemPolicyOutput, data *fileSystemPolicyResourceModel, diags *diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	policyToSet, err := verify.PolicyToSet(data.Policy.ValueString(), *output.Policy)
	if err != nil {
		smerr.AddError(ctx, diags, err, smerr.ID, data.FileSystemID.ValueString())
		return
	}
	data.Policy = fwtypes.IAMPolicyValue(policyToSet)
}

func (r *fileSystemPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data fileSystemPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.PutFileSystemPolicyInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix("FileSystem")))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutFileSystemPolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.FileSystemID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *fileSystemPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data fileSystemPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.DeleteFileSystemPolicyInput{
		FileSystemId: data.FileSystemID.ValueStringPointer(),
	}

	_, err := conn.DeleteFileSystemPolicy(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return // Resource already deleted
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.FileSystemID.ValueString())
	}
}

type fileSystemPolicyResourceModel struct {
	framework.WithRegionModel
	FileSystemID types.String      `tfsdk:"file_system_id"`
	Policy       fwtypes.IAMPolicy `tfsdk:"policy"`
}

func findFileSystemPolicyByID(ctx context.Context, conn *s3files.Client, id string) (*s3files.GetFileSystemPolicyOutput, error) {
	input := s3files.GetFileSystemPolicyInput{
		FileSystemId: &id,
	}

	output, err := conn.GetFileSystemPolicy(ctx, &input)
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
