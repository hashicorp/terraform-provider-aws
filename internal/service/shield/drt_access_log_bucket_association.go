// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="DRT Log Bucket Association")
func newDRTAccessLogBucketAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &drtAccessLogBucketAssociationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type drtAccessLogBucketAssociationResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *drtAccessLogBucketAssociationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_shield_drt_access_log_bucket_association"
}

func (r *drtAccessLogBucketAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"log_bucket": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
				},
			},
			"role_arn_association_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: `Unused`,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *drtAccessLogBucketAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data drtAccessLogBucketAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	logBucket := data.LogBucket.ValueString()
	input := &shield.AssociateDRTLogBucketInput{
		LogBucket: aws.String(logBucket),
	}

	_, err := conn.AssociateDRTLogBucket(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Shield DRT Log Bucket Association (%s)", logBucket), err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	_, err = tfresource.RetryWhenNotFound(ctx, r.CreateTimeout(ctx, data.Timeouts), func() (interface{}, error) {
		return findDRTLogBucketAssociation(ctx, conn, logBucket)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Shield DRT Log Bucket Association (%s) create", logBucket), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *drtAccessLogBucketAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data drtAccessLogBucketAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().ShieldClient(ctx)

	logBucket := data.ID.ValueString()
	_, err := findDRTLogBucketAssociation(ctx, conn, logBucket)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Shield DRT Log Bucket Association (%s)", logBucket), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *drtAccessLogBucketAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data drtAccessLogBucketAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	logBucket := data.ID.ValueString()
	input := &shield.DisassociateDRTLogBucketInput{
		LogBucket: aws.String(logBucket),
	}

	_, err := conn.DisassociateDRTLogBucket(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Shield DRT Log Bucket Association (%s)", logBucket), err.Error())

		return
	}

	_, err = tfresource.RetryUntilNotFound(ctx, r.DeleteTimeout(ctx, data.Timeouts), func() (interface{}, error) {
		return findDRTLogBucketAssociation(ctx, conn, logBucket)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Shield DRT Log Bucket Association (%s) delete", logBucket), err.Error())

		return
	}
}

func findDRTLogBucketAssociation(ctx context.Context, conn *shield.Client, name string) (*string, error) {
	output, err := findDRTAccess(ctx, conn)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.LogBucketList, func(v string) bool {
		return v == name
	}))
}

func findDRTAccess(ctx context.Context, conn *shield.Client) (*shield.DescribeDRTAccessOutput, error) {
	input := &shield.DescribeDRTAccessInput{}

	output, err := conn.DescribeDRTAccess(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type drtAccessLogBucketAssociationResourceModel struct {
	ID                   types.String   `tfsdk:"id"`
	LogBucket            types.String   `tfsdk:"log_bucket"`
	RoleARNAssociationID types.String   `tfsdk:"role_arn_association_id"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}

func (model *drtAccessLogBucketAssociationResourceModel) InitFromID() error {
	model.LogBucket = model.ID

	return nil
}

func (model *drtAccessLogBucketAssociationResourceModel) setID() {
	model.ID = model.LogBucket
}
