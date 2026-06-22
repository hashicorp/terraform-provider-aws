// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudfront

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// @FrameworkResource("aws_cloudfront_trust_store", name="Trust Store")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newTrustStoreResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &trustStoreResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type trustStoreResource struct {
	framework.ResourceWithModel[trustStoreResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *trustStoreResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"number_of_ca_certificates": schema.Int32Attribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"ca_certificates_bundle_source": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[caCertificatesBundleSourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"ca_certificates_bundle_s3_location": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[caCertificatesBundleS3LocationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucket: schema.StringAttribute{
										Required: true,
									},
									names.AttrKey: schema.StringAttribute{
										Required: true,
									},
									names.AttrRegion: schema.StringAttribute{
										Required: true,
									},
									names.AttrVersion: schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
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

func (r *trustStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data trustStoreResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input cloudfront.CreateTrustStoreInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = &awstypes.Tags{
			Items: tags,
		}
	}

	outCTS, err := conn.CreateTrustStore(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, outCTS.TrustStore.Arn)
	id := aws.ToString(outCTS.TrustStore.Id)
	data.ID = fwflex.StringValueToFramework(ctx, id)

	outGTS, err := waitTrustStoreActive(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	// Set values for unknowns.
	data.Etag = fwflex.StringToFramework(ctx, outGTS.ETag)
	data.NumberOfCACertificates = fwflex.Int32ToFramework(ctx, outGTS.TrustStore.NumberOfCaCertificates)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, data))
}

func (r *trustStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state trustStoreResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.ID)
	out, err := findTrustStoreByID(ctx, conn, id)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.TrustStore, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.Etag = fwflex.StringToFramework(ctx, out.ETag)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *trustStoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state trustStoreResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		id, etag := fwflex.StringValueFromFramework(ctx, plan.ID), fwflex.StringValueFromFramework(ctx, state.Etag)
		var input cloudfront.UpdateTrustStoreInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.IfMatch = aws.String(etag)

		_, err := conn.UpdateTrustStore(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}

		outGTS, err := waitTrustStoreActive(ctx, conn, id, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}

		// Set values for unknowns.
		plan.Etag = fwflex.StringToFramework(ctx, outGTS.ETag)
		plan.NumberOfCACertificates = fwflex.Int32ToFramework(ctx, outGTS.TrustStore.NumberOfCaCertificates)
	} else {
		plan.Etag = state.Etag
		plan.NumberOfCACertificates = state.NumberOfCACertificates
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *trustStoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state trustStoreResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id, etag := fwflex.StringValueFromFramework(ctx, state.ID), fwflex.StringValueFromFramework(ctx, state.Etag)
	input := cloudfront.DeleteTrustStoreInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}
	_, err := conn.DeleteTrustStore(ctx, &input)
	if errs.IsA[*awstypes.EntityNotFound](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	if _, err := waitTrustStoreDeleted(ctx, conn, id, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}
}

func waitTrustStoreActive(ctx context.Context, conn *cloudfront.Client, id string, timeout time.Duration) (*cloudfront.GetTrustStoreOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.TrustStoreStatusActive),
		Refresh:                   statusTrustStore(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*cloudfront.GetTrustStoreOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.TrustStore.Reason)))
		return out, err
	}

	return nil, err
}

func waitTrustStoreDeleted(ctx context.Context, conn *cloudfront.Client, id string, timeout time.Duration) (*cloudfront.GetTrustStoreOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TrustStoreStatusActive),
		Target:  []string{},
		Refresh: statusTrustStore(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*cloudfront.GetTrustStoreOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.TrustStore.Reason)))
		return out, err
	}

	return nil, err
}

func statusTrustStore(conn *cloudfront.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findTrustStoreByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.TrustStore.Status), nil
	}
}

func findTrustStoreByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetTrustStoreOutput, error) {
	input := cloudfront.GetTrustStoreInput{
		Identifier: aws.String(id),
	}

	return findTrustStore(ctx, conn, &input)
}

func findTrustStore(ctx context.Context, conn *cloudfront.Client, input *cloudfront.GetTrustStoreInput) (*cloudfront.GetTrustStoreOutput, error) {
	out, err := conn.GetTrustStore(ctx, input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.TrustStore == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type trustStoreResourceModel struct {
	ARN                        types.String                                                     `tfsdk:"arn"`
	CACertificatesBundleSource fwtypes.ListNestedObjectValueOf[caCertificatesBundleSourceModel] `tfsdk:"ca_certificates_bundle_source"`
	Etag                       types.String                                                     `tfsdk:"etag"`
	ID                         types.String                                                     `tfsdk:"id"`
	Name                       types.String                                                     `tfsdk:"name"`
	NumberOfCACertificates     types.Int32                                                      `tfsdk:"number_of_ca_certificates"`
	Tags                       tftags.Map                                                       `tfsdk:"tags"`
	TagsAll                    tftags.Map                                                       `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                                                   `tfsdk:"timeouts"`
}

type caCertificatesBundleSourceModel struct {
	CACertificatesBundleS3Location fwtypes.ListNestedObjectValueOf[caCertificatesBundleS3LocationModel] `tfsdk:"ca_certificates_bundle_s3_location"`
}

type caCertificatesBundleS3LocationModel struct {
	Bucket  types.String `tfsdk:"bucket"`
	Key     types.String `tfsdk:"key"`
	Region  types.String `tfsdk:"region"`
	Version types.String `tfsdk:"version"`
}

var (
	_ fwflex.Expander = caCertificatesBundleSourceModel{}
)

func (m caCertificatesBundleSourceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CACertificatesBundleS3Location.IsNull():
		data, d := m.CACertificatesBundleS3Location.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.CaCertificatesBundleSourceMemberCaCertificatesBundleS3Location
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}
