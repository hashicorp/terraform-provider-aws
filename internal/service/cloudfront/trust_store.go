// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudfront_trust_store", name="Trust Store")
func newResourceTrustStore(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTrustStore{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameTrustStore = "Trust Store"
)

type resourceTrustStore struct {
	framework.ResourceWithModel[trustStoreResourceModel]
	framework.WithTimeouts
}

func (r *resourceTrustStore) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"number_of_ca_certificates": schema.Int32Attribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3LocationModel](ctx),
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

func (r *resourceTrustStore) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data trustStoreResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	input := cloudfront.CreateTrustStoreInput{
		Name: fwflex.StringFromFramework(ctx, data.Name),
	}

	// Add CA certificates bundle source if provided
	if !data.CACertificatesBundleSource.IsNull() {
		bundleSourceList, d := data.CACertificatesBundleSource.ToSlice(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, d)
		if response.Diagnostics.HasError() {
			return
		}

		if len(bundleSourceList) > 0 {
			bundleSource := bundleSourceList[0]
			if !bundleSource.CACertificatesBundleS3Location.IsNull() {
				s3LocationList, d := bundleSource.CACertificatesBundleS3Location.ToSlice(ctx)
				smerr.AddEnrich(ctx, &response.Diagnostics, d)
				if response.Diagnostics.HasError() {
					return
				}

				if len(s3LocationList) > 0 {
					s3Location := s3LocationList[0]
					input.CaCertificatesBundleSource = &awstypes.CaCertificatesBundleSourceMemberCaCertificatesBundleS3Location{
						Value: awstypes.CaCertificatesBundleS3Location{
							Bucket:  s3Location.Bucket.ValueStringPointer(),
							Key:     s3Location.Key.ValueStringPointer(),
							Region:  s3Location.Region.ValueStringPointer(),
							Version: s3Location.Version.ValueStringPointer(),
						},
					}
				}
			}
		}
	}

	out, err := conn.CreateTrustStore(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	data.ID = flex.StringToFramework(ctx, out.TrustStore.Id)
	data.ARN = flex.StringToFramework(ctx, out.TrustStore.Arn)

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	_, err = waitTrustStoreActive(ctx, conn, data.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	// Fetch the created trust store and populate the resource model
	output, err := findTrustStoreByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	// Use custom flattener to populate the resource model
	flattenTrustStoreIntoResourceModel(ctx, output, &data)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *resourceTrustStore) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var state trustStoreResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTrustStoreByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// Use custom flattener to populate the resource model
	flattenTrustStoreIntoResourceModel(ctx, out, &state)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceTrustStore) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var plan, state trustStoreResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.CACertificatesBundleSource.Equal(state.CACertificatesBundleSource) {
		input := cloudfront.UpdateTrustStoreInput{
			Id:      fwflex.StringFromFramework(ctx, state.ID),
			IfMatch: fwflex.StringFromFramework(ctx, state.Etag),
		}

		// Add CA certificates bundle source from plan
		if !plan.CACertificatesBundleSource.IsNull() {
			bundleSourceList, d := plan.CACertificatesBundleSource.ToSlice(ctx)
			smerr.AddEnrich(ctx, &resp.Diagnostics, d)
			if resp.Diagnostics.HasError() {
				return
			}

			if len(bundleSourceList) > 0 {
				bundleSource := bundleSourceList[0]
				if !bundleSource.CACertificatesBundleS3Location.IsNull() {
					s3LocationList, d := bundleSource.CACertificatesBundleS3Location.ToSlice(ctx)
					smerr.AddEnrich(ctx, &resp.Diagnostics, d)
					if resp.Diagnostics.HasError() {
						return
					}

					if len(s3LocationList) > 0 {
						s3Location := s3LocationList[0]
						input.CaCertificatesBundleSource = &awstypes.CaCertificatesBundleSourceMemberCaCertificatesBundleS3Location{
							Value: awstypes.CaCertificatesBundleS3Location{
								Bucket:  s3Location.Bucket.ValueStringPointer(),
								Key:     s3Location.Key.ValueStringPointer(),
								Region:  s3Location.Region.ValueStringPointer(),
								Version: s3Location.Version.ValueStringPointer(),
							},
						}
					}
				}
			}
		}

		out, err := conn.UpdateTrustStore(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil || out.TrustStore == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
			return
		}

		plan.Etag = flex.StringToFramework(ctx, out.ETag)
		plan.Status = flex.StringValueToFramework(ctx, out.TrustStore.Status)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitTrustStoreActive(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	// Fetch the updated trust store and populate the resource model
	output, err := findTrustStoreByID(ctx, conn, plan.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	// Use custom flattener to populate the resource model
	flattenTrustStoreIntoResourceModel(ctx, output, &plan)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceTrustStore) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var state trustStoreResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := cloudfront.DeleteTrustStoreInput{
		Id:      state.ID.ValueStringPointer(),
		IfMatch: state.Etag.ValueStringPointer(),
	}

	_, err := conn.DeleteTrustStore(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFound](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitTrustStoreDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func (r *resourceTrustStore) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitTrustStoreActive(ctx context.Context, conn *cloudfront.Client, id string, timeout time.Duration) (*awstypes.TrustStore, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{string(awstypes.TrustStoreStatusActive)},
		Refresh:                   statusTrustStore(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.TrustStore); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitTrustStoreDeleted(ctx context.Context, conn *cloudfront.Client, id string, timeout time.Duration) (*awstypes.TrustStore, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.TrustStoreStatusActive)},
		Target:  []string{},
		Refresh: statusTrustStore(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.TrustStore); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusTrustStore(ctx context.Context, conn *cloudfront.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findTrustStoreByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out.TrustStore, string(out.TrustStore.Status), nil
	}
}

func findTrustStoreByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetTrustStoreOutput, error) {
	input := cloudfront.GetTrustStoreInput{
		Identifier: aws.String(id),
	}

	out, err := conn.GetTrustStore(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFound](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.TrustStore == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type trustStoreResourceModel struct {
	ARN                        types.String                                                     `tfsdk:"arn"`
	CACertificatesBundleSource fwtypes.ListNestedObjectValueOf[caCertificatesBundleSourceModel] `tfsdk:"ca_certificates_bundle_source"`
	Etag                       types.String                                                     `tfsdk:"etag"`
	ID                         types.String                                                     `tfsdk:"id"`
	LastModifiedTime           timetypes.RFC3339                                                `tfsdk:"last_modified_time"`
	Name                       types.String                                                     `tfsdk:"name"`
	NumberOfCaCertificates     types.Int32                                                      `tfsdk:"number_of_ca_certificates"`
	Status                     types.String                                                     `tfsdk:"status"`
	Timeouts                   timeouts.Value                                                   `tfsdk:"timeouts"`
}

type caCertificatesBundleSourceModel struct {
	CACertificatesBundleS3Location fwtypes.ListNestedObjectValueOf[s3LocationModel] `tfsdk:"ca_certificates_bundle_s3_location"`
}

type s3LocationModel struct {
	Bucket  types.String `tfsdk:"bucket"`
	Key     types.String `tfsdk:"key"`
	Region  types.String `tfsdk:"region"`
	Version types.String `tfsdk:"version"`
}

// flattenTrustStoreIntoResourceModel populates the Terraform resource model with data from the AWS API response
func flattenTrustStoreIntoResourceModel(ctx context.Context, output *cloudfront.GetTrustStoreOutput, data *trustStoreResourceModel) {
	if output == nil || output.TrustStore == nil {
		return
	}

	// Populate computed attributes from the API response
	data.ARN = flex.StringToFramework(ctx, output.TrustStore.Arn)
	data.Name = flex.StringToFramework(ctx, output.TrustStore.Name)
	data.Status = flex.StringValueToFramework(ctx, output.TrustStore.Status)
	data.Etag = flex.StringToFramework(ctx, output.ETag)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, output.TrustStore.LastModifiedTime)
	data.NumberOfCaCertificates = flex.Int32ToFramework(ctx, output.TrustStore.NumberOfCaCertificates)

	// Note: CACertificatesBundleSource is not returned by the GetTrustStore API
	// The value from the plan/state is preserved as it was set during creation
}

func sweepTrustStores(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := cloudfront.ListTrustStoresInput{}
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := cloudfront.NewListTrustStoresPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.TrustStoreList {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceTrustStore, client,
				sweepfw.NewAttribute(names.AttrID, aws.ToString(v.Id)),
			))
		}
	}

	return sweepResources, nil
}
