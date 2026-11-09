// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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

// @FrameworkResource("aws_observabilityadmin_s3_table_integration", name="S3 Table Integration")
// @ArnIdentity
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/observabilityadmin;observabilityadmin;observabilityadmin.GetS3TableIntegrationOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccS3TableIntegrationPreCheck")
// @Testing(tagsTest=false)
// @Testing(serialize=true)
func newS3TableIntegrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &s3TableIntegrationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type s3TableIntegrationResource struct {
	framework.ResourceWithModel[s3TableIntegrationResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *s3TableIntegrationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"destination_table_bucket_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"encryption": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKMSKeyARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"sse_algorithm": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.SSEAlgorithm](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
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

func (r *s3TableIntegrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data s3TableIntegrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	var input observabilityadmin.CreateS3TableIntegrationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateS3TableIntegration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	arn := aws.ToString(output.Arn)
	out, err := waitS3TableIntegrationActive(ctx, conn, arn, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringValueToFramework(ctx, arn)
	data.DestinationTableBucketARN = fwflex.StringToFramework(ctx, out.DestinationTableBucketArn)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *s3TableIntegrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data s3TableIntegrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	out, err := findS3TableIntegrationByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *s3TableIntegrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data s3TableIntegrationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	input := observabilityadmin.DeleteS3TableIntegrationInput{
		Arn: aws.String(arn),
	}
	_, err := conn.DeleteS3TableIntegration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) || tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitS3TableIntegrationDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}
}

type s3TableIntegrationResourceModel struct {
	framework.WithRegionModel
	ARN                       types.String                                     `tfsdk:"arn"`
	DestinationTableBucketARN types.String                                     `tfsdk:"destination_table_bucket_arn"`
	Encryption                fwtypes.ListNestedObjectValueOf[encryptionModel] `tfsdk:"encryption"`
	RoleARN                   fwtypes.ARN                                      `tfsdk:"role_arn"`
	Tags                      tftags.Map                                       `tfsdk:"tags"`
	TagsAll                   tftags.Map                                       `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                   `tfsdk:"timeouts"`
}

type encryptionModel struct {
	KMSKeyARN    fwtypes.ARN                               `tfsdk:"kms_key_arn"`
	SSEAlgorithm fwtypes.StringEnum[awstypes.SSEAlgorithm] `tfsdk:"sse_algorithm"`
}

func findS3TableIntegrationByARN(ctx context.Context, conn *observabilityadmin.Client, arn string) (*observabilityadmin.GetS3TableIntegrationOutput, error) {
	input := observabilityadmin.GetS3TableIntegrationInput{
		Arn: aws.String(arn),
	}
	return findS3TableIntegration(ctx, conn, &input)
}

func findS3TableIntegration(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.GetS3TableIntegrationInput) (*observabilityadmin.GetS3TableIntegrationOutput, error) {
	output, err := conn.GetS3TableIntegration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusS3TableIntegration(conn *observabilityadmin.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findS3TableIntegrationByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitS3TableIntegrationActive(ctx context.Context, conn *observabilityadmin.Client, arn string, timeout time.Duration) (*observabilityadmin.GetS3TableIntegrationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{""},
		Target:                    enum.Slice(awstypes.IntegrationStatusActive),
		Refresh:                   statusS3TableIntegration(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*observabilityadmin.GetS3TableIntegrationOutput); ok {
		return output, err
	}

	return nil, err
}

func waitS3TableIntegrationDeleted(ctx context.Context, conn *observabilityadmin.Client, arn string, timeout time.Duration) (*observabilityadmin.GetS3TableIntegrationOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IntegrationStatusDeleting),
		Target:  []string{},
		Refresh: statusS3TableIntegration(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*observabilityadmin.GetS3TableIntegrationOutput); ok {
		return output, err
	}

	return nil, err
}
