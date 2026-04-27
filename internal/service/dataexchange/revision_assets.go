// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dataexchange

import (
	"context"
	"crypto/md5" // nosemgrep: go/sast/internal/crypto/md5 -- AWS DataExchange API requires MD5 for asset upload integrity checking
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/hashicorp/terraform-provider-aws/version"
)

// @FrameworkResource("aws_dataexchange_revision_assets", name="Revision Assets")
// @Tags(identifierAttribute="arn")
// @NoImport
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/dataexchange;dataexchange.GetRevisionOutput")
func newRevisionAssetsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &revisionAssetsResource{}
	r.SetDefaultCreateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameRevisionAssets = "Revision Assets"
)

type revisionAssetsResource struct {
	framework.ResourceWithModel[revisionAssetsResourceModel]
	framework.WithTimeouts
}

func (r *revisionAssetsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrComment: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 16_348),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_set_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"finalized": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrForceDestroy: schema.BoolAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},

			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"asset": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[assetModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN: framework.ARNAttributeComputedOnly(),
						names.AttrCreatedAt: schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrID: framework.IDAttribute(),
						names.AttrName: schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"updated_at": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"create_s3_data_access_from_s3_bucket": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[createS3DataAccessFromS3BucketModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"access_point_alias": schema.StringAttribute{
										Computed: true,
									},
									"access_point_arn": schema.StringAttribute{
										Computed: true,
									},
								},
								Blocks: map[string]schema.Block{
									"asset_source": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[s3DataAccessAssetSourceModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucket: schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"key_prefixes": schema.SetAttribute{
													CustomType: fwtypes.SetOfStringType,
													Optional:   true,
													PlanModifiers: []planmodifier.Set{
														setplanmodifier.RequiresReplace(),
													},
												},
												"keys": schema.SetAttribute{
													CustomType: fwtypes.SetOfStringType,
													Optional:   true,
													PlanModifiers: []planmodifier.Set{
														setplanmodifier.RequiresReplace(),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"kms_keys_to_grant": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[kmsKeyToGrantModel](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrKMSKeyARN: schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Required:   true,
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
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
						},
						"import_assets_from_s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[importAssetsFromS3Model](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"asset_source": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[assetSourceModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucket: schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												names.AttrKey: schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
										},
									},
								},
							},
						},
						"import_assets_from_signed_url": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[importAssetsFromSignedURLModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"filename": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *revisionAssetsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var plan revisionAssetsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalized := plan.Finalized.ValueBool()
	var input dataexchange.CreateRevisionInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateRevision(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, plan.DataSetID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, plan.DataSetID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// NOTE: This is currently sequential, because there is no direct linkage between the Job and the Asset that it creates.
	// If unique names were enforced for each asset, they could be done in parallel, but alas.
	// Assets can be renamed *after* the Job that creates them is complete.
	// The `ImportAssetsFromSignedURL` Job technically requires a `Name` parameter, but I've defaulted to the name of the file.
	// This should probably be changed to explicitly require the `name`
	revisionID := aws.ToString(out.Id)
	assets := make([]assetModel, len(plan.Assets.Elements()))
	existingAssetIDs := make([]string, 0, len(plan.Assets.Elements()))
	for i, asset := range nestedObjectCollectionAllMust[assetModel](ctx, plan.Assets) {
		switch {
		case !asset.ImportAssetsFromS3.IsNull():
			importAssetsFromS3, d := asset.ImportAssetsFromS3.ToPtr(ctx)
			resp.Diagnostics.Append(d...)
			if d.HasError() {
				return
			}

			var importAssetsFromS3RequestDetails awstypes.ImportAssetsFromS3RequestDetails
			resp.Diagnostics.Append(flex.Expand(ctx, importAssetsFromS3, &importAssetsFromS3RequestDetails)...)
			if resp.Diagnostics.HasError() {
				return
			}
			importAssetsFromS3RequestDetails.DataSetId = plan.DataSetID.ValueStringPointer()
			importAssetsFromS3RequestDetails.RevisionId = plan.ID.ValueStringPointer()

			requestDetails := awstypes.RequestDetails{
				ImportAssetsFromS3: &importAssetsFromS3RequestDetails,
			}
			createJobInput := dataexchange.CreateJobInput{
				Type:    awstypes.TypeImportAssetsFromS3,
				Details: &requestDetails,
			}
			createJobOutput, err := conn.CreateJob(ctx, &createJobInput)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
					err.Error(),
				)
				return
			}
			if createJobOutput == nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, nil),
					errors.New("empty output").Error(),
				)
				return
			}

			err = startJob(ctx, createJobOutput.Id, conn)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
					err.Error(),
				)
				return
			}

			createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
			_, err = waitJobCompleted(ctx, conn, aws.ToString(createJobOutput.Id), createTimeout)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
					err.Error(),
				)
				return
			}

			listAssetsInput := dataexchange.ListRevisionAssetsInput{
				DataSetId:  plan.DataSetID.ValueStringPointer(),
				RevisionId: plan.ID.ValueStringPointer(),
			}
			newAsset, err := getRevisionAsset(ctx, conn, listAssetsInput, existingAssetIDs)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
					err.Error(),
				)
			}

			resp.Diagnostics.Append(flex.Flatten(ctx, newAsset, asset)...)
			if resp.Diagnostics.HasError() {
				return
			}
			assets[i] = *asset // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment
			existingAssetIDs = append(existingAssetIDs, aws.ToString(newAsset.Id))

		case !asset.ImportAssetsFromSignedURL.IsNull():
			/* calling defer functions directly in a loop is bad practice and can cause resource leaks
			by deferring all deferred actions until the end of the function.
			Wrapping execution in anonymous function to ensure deferred actions are executed at the end of each iteration.
			*/
			func() {
				importAssetsFromSignedURL, d := asset.ImportAssetsFromSignedURL.ToPtr(ctx)
				resp.Diagnostics.Append(d...)
				if d.HasError() {
					return
				}

				var importAssetFromSignedUrlRequestDetails awstypes.ImportAssetFromSignedUrlRequestDetails
				importAssetFromSignedUrlRequestDetails.DataSetId = plan.DataSetID.ValueStringPointer()
				importAssetFromSignedUrlRequestDetails.RevisionId = plan.ID.ValueStringPointer()
				// Default `AssetName` to last path component?
				importAssetFromSignedUrlRequestDetails.AssetName = importAssetsFromSignedURL.Filename.ValueStringPointer()
				// Stream MD5
				f, err := os.Open(importAssetsFromSignedURL.Filename.ValueString())
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
					return
				}

				defer func() {
					err := f.Close()
					if err != nil {
						tflog.Warn(ctx, "error closing file", map[string]any{
							"file":  f.Name(),
							"error": err.Error(),
						})
					}
				}()

				hash, err := md5Reader(f)
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
					return
				}

				importAssetFromSignedUrlRequestDetails.Md5Hash = aws.String(hash)

				_, err = f.Seek(0, 0)
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
					return
				}

				requestDetails := awstypes.RequestDetails{
					ImportAssetFromSignedUrl: &importAssetFromSignedUrlRequestDetails,
				}
				createJobInput := dataexchange.CreateJobInput{
					Type:    awstypes.TypeImportAssetFromSignedUrl,
					Details: &requestDetails,
				}
				createJobOutput, err := conn.CreateJob(ctx, &createJobInput)
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
					return
				}
				if createJobOutput == nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, nil),
						errors.New("empty output").Error(),
					)
					return
				}

				// Upload file to URL with PUT operation
				info, err := f.Stat()
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
					return
				}

				const (
					uploadTimeout = 1 * time.Minute
				)
				ctxUpload, cancel := context.WithTimeout(ctx, uploadTimeout)
				defer cancel()

				request, err := http.NewRequestWithContext(ctxUpload, http.MethodPut, *createJobOutput.Details.ImportAssetFromSignedUrl.SignedUrl, f)
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
					return
				}
				request.ContentLength = info.Size()
				request.Header.Set("Content-MD5", hash)
				request.Header.Set("Provider-Version", version.ProviderVersion)
				request.Header.Set("Terraform-Version", r.Meta().TerraformVersion(ctx))

				httpClient := r.Meta().AwsConfig(ctx).HTTPClient
				response, err := httpClient.Do(request)
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
					return
				}

				defer func() {
					err := response.Body.Close()
					if err != nil {
						tflog.Warn(ctx, "error closing file", map[string]any{
							"file":  f.Name(),
							"error": err.Error(),
						})
					}
				}()

				if !(response.StatusCode >= http.StatusOK && response.StatusCode <= http.StatusIMUsed) {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, nil),
						fmt.Sprintf("Uploading to %q\n\nUnexpected HTTP response: %s", *createJobOutput.Details.ImportAssetFromSignedUrl.SignedUrl, response.Status),
					)
					return
				}
				_, err = io.Copy(io.Discard, response.Body)
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
					return
				}

				// Start Job
				err = startJob(ctx, createJobOutput.Id, conn)
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
					return
				}

				createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
				_, err = waitJobCompleted(ctx, conn, aws.ToString(createJobOutput.Id), createTimeout)
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
					return
				}

				// List assets
				listAssetsInput := dataexchange.ListRevisionAssetsInput{
					DataSetId:  plan.DataSetID.ValueStringPointer(),
					RevisionId: plan.ID.ValueStringPointer(),
				}
				newAsset, err := getRevisionAsset(ctx, conn, listAssetsInput, existingAssetIDs)
				if err != nil {
					resp.Diagnostics.AddError(
						create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
						err.Error(),
					)
				}

				resp.Diagnostics.Append(flex.Flatten(ctx, newAsset, asset)...)
				if resp.Diagnostics.HasError() {
					return
				}
				assets[i] = *asset // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment
				existingAssetIDs = append(existingAssetIDs, aws.ToString(newAsset.Id))
			}()
		case !asset.CreateS3DataAccessFromS3Bucket.IsNull():
			createS3DataAccessFromS3Bucket, d := asset.CreateS3DataAccessFromS3Bucket.ToPtr(ctx)
			resp.Diagnostics.Append(d...)
			if d.HasError() {
				return
			}

			var createS3DataAccessFromS3BucketRequestDetails awstypes.CreateS3DataAccessFromS3BucketRequestDetails
			resp.Diagnostics.Append(flex.Expand(ctx, createS3DataAccessFromS3Bucket, &createS3DataAccessFromS3BucketRequestDetails)...)
			if resp.Diagnostics.HasError() {
				return
			}
			createS3DataAccessFromS3BucketRequestDetails.DataSetId = plan.DataSetID.ValueStringPointer()
			createS3DataAccessFromS3BucketRequestDetails.RevisionId = plan.ID.ValueStringPointer()

			requestDetails := awstypes.RequestDetails{
				CreateS3DataAccessFromS3Bucket: &createS3DataAccessFromS3BucketRequestDetails,
			}
			createJobInput := dataexchange.CreateJobInput{
				Type:    awstypes.TypeCreateS3DataAccessFromS3Bucket,
				Details: &requestDetails,
			}
			createJobOutput, err := conn.CreateJob(ctx, &createJobInput)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
					err.Error(),
				)
				return
			}
			if createJobOutput == nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, nil),
					errors.New("empty output").Error(),
				)
				return
			}

			err = startJob(ctx, createJobOutput.Id, conn)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
					err.Error(),
				)
				return
			}

			createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
			_, err = waitJobCompleted(ctx, conn, aws.ToString(createJobOutput.Id), createTimeout)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
					err.Error(),
				)
				return
			}

			in := dataexchange.ListRevisionAssetsInput{
				DataSetId:  plan.DataSetID.ValueStringPointer(),
				RevisionId: plan.ID.ValueStringPointer(),
			}
			newAsset, err := getRevisionAsset(ctx, conn, in, existingAssetIDs)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
					err.Error(),
				)
			}

			resp.Diagnostics.Append(flex.Flatten(ctx, newAsset, asset)...)
			if resp.Diagnostics.HasError() {
				return
			}

			createS3DataAccessFromS3Bucket.AccessPointARN = flex.StringToFramework(ctx, newAsset.AssetDetails.S3DataAccessAsset.S3AccessPointArn)
			createS3DataAccessFromS3Bucket.AccessPointAlias = flex.StringToFramework(ctx, newAsset.AssetDetails.S3DataAccessAsset.S3AccessPointAlias)
			asset.CreateS3DataAccessFromS3Bucket = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, createS3DataAccessFromS3Bucket)
			assets[i] = *asset // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-assignment
			existingAssetIDs = append(existingAssetIDs, aws.ToString(newAsset.Id))
		}
	}

	assetsVal, d := fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, assets)
	resp.Diagnostics.Append(d...)
	if d.HasError() {
		return
	}
	plan.Assets = assetsVal

	// finalize asset if requested
	if finalized {
		err = finalizeAsset(ctx, conn, plan.DataSetID.ValueString(), plan.ID.ValueString(), finalized)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionCreating, ResNameRevisionAssets, revisionID, err),
				err.Error(),
			)
			return
		}
		plan.Finalized = types.BoolValue(finalized)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *revisionAssetsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var state revisionAssetsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRevisionByID(ctx, conn, state.DataSetID.ValueString(), state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionReading, ResNameRevisionAssets, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *revisionAssetsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var plan, state revisionAssetsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state, flex.WithIgnoredField("ForceDestroy"))
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := dataexchange.UpdateRevisionInput{
			DataSetId:  plan.DataSetID.ValueStringPointer(),
			RevisionId: plan.ID.ValueStringPointer(),
		}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.Comment.Equal(state.Comment) && state.Finalized.ValueBool() {
			err := finalizeAsset(ctx, conn, plan.DataSetID.ValueString(), plan.ID.ValueString(), false)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.DataExchange, create.ErrActionUpdating, ResNameRevisionAssets, plan.ID.String(), err),
					err.Error(),
				)
				return
			}
		}

		out, err := conn.UpdateRevision(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionUpdating, ResNameRevisionAssets, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *revisionAssetsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataExchangeClient(ctx)

	var state revisionAssetsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ForceDestroy.ValueBool() && state.Finalized.ValueBool() {
		err := finalizeAsset(ctx, conn, state.DataSetID.ValueString(), state.ID.ValueString(), false)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataExchange, create.ErrActionDeleting, ResNameRevisionAssets, state.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	input := dataexchange.DeleteRevisionInput{
		DataSetId:  state.DataSetID.ValueStringPointer(),
		RevisionId: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteRevision(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataExchange, create.ErrActionDeleting, ResNameRevisionAssets, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *revisionAssetsResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// check if resource is being destroyed
	if req.Plan.Raw.IsNull() {
		var forceDestroy, finalized types.Bool
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(names.AttrForceDestroy), &forceDestroy)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("finalized"), &finalized)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !forceDestroy.ValueBool() && finalized.ValueBool() {
			resp.Diagnostics.AddError(
				"Unable to destroy a finalized revision",
				"Cannot destroy a finalized revision without setting `force_destroy` to `true`",
			)
			return
		}
	}

	if !req.Plan.Raw.IsNull() && !req.State.Raw.IsNull() {
		var plan, state revisionAssetsResourceModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.Comment.Equal(state.Comment) || !plan.Finalized.Equal(state.Finalized) {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("updated_at"), timetypes.NewRFC3339Unknown())...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("updated_at"), state.UpdatedAt)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}
}

func findRevisionByID(ctx context.Context, conn *dataexchange.Client, dataSetId, revisionId string) (*dataexchange.GetRevisionOutput, error) {
	input := dataexchange.GetRevisionInput{
		DataSetId:  aws.String(dataSetId),
		RevisionId: aws.String(revisionId),
	}
	output, err := conn.GetRevision(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func getRevisionAsset(ctx context.Context, conn *dataexchange.Client, input dataexchange.ListRevisionAssetsInput, existingAssetIDs []string) (awstypes.AssetEntry, error) {
	var newAsset awstypes.AssetEntry
	paginator := dataexchange.NewListRevisionAssetsPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return awstypes.AssetEntry{}, err
		}

		for _, v := range page.Assets {
			if !slices.Contains(existingAssetIDs, aws.ToString(v.Id)) {
				newAsset = v
			}
		}
	}
	if newAsset.Id == nil {
		return awstypes.AssetEntry{}, errors.New("missing new asset")
	}

	return newAsset, nil
}

// finalizeAsset will either finalize or de-finalize the asset
func finalizeAsset(ctx context.Context, conn *dataexchange.Client, datasetId, revisionId string, finalized bool) error {
	input := dataexchange.UpdateRevisionInput{
		DataSetId:  aws.String(datasetId),
		RevisionId: aws.String(revisionId),
		Finalized:  aws.Bool(finalized),
	}
	_, err := conn.UpdateRevision(ctx, &input)

	return err
}

type revisionAssetsResourceModel struct {
	framework.WithRegionModel
	ARN          types.String                               `tfsdk:"arn"`
	Assets       fwtypes.SetNestedObjectValueOf[assetModel] `tfsdk:"asset"`
	Comment      types.String                               `tfsdk:"comment"`
	CreatedAt    timetypes.RFC3339                          `tfsdk:"created_at"`
	DataSetID    types.String                               `tfsdk:"data_set_id"`
	Finalized    types.Bool                                 `tfsdk:"finalized"`
	ForceDestroy types.Bool                                 `tfsdk:"force_destroy"`
	ID           types.String                               `tfsdk:"id"`
	UpdatedAt    timetypes.RFC3339                          `tfsdk:"updated_at"`
	Tags         tftags.Map                                 `tfsdk:"tags"`
	TagsAll      tftags.Map                                 `tfsdk:"tags_all"`
	Timeouts     timeouts.Value                             `tfsdk:"timeouts"`
}

type assetModel struct {
	ARN                            types.String                                                         `tfsdk:"arn"`
	CreatedAt                      timetypes.RFC3339                                                    `tfsdk:"created_at"`
	CreateS3DataAccessFromS3Bucket fwtypes.ListNestedObjectValueOf[createS3DataAccessFromS3BucketModel] `tfsdk:"create_s3_data_access_from_s3_bucket"`
	ID                             types.String                                                         `tfsdk:"id"`
	ImportAssetsFromS3             fwtypes.ListNestedObjectValueOf[importAssetsFromS3Model]             `tfsdk:"import_assets_from_s3"`
	ImportAssetsFromSignedURL      fwtypes.ListNestedObjectValueOf[importAssetsFromSignedURLModel]      `tfsdk:"import_assets_from_signed_url"`
	Name                           types.String                                                         `tfsdk:"name"`
	UpdatedAt                      timetypes.RFC3339                                                    `tfsdk:"updated_at"`
}

type importAssetsFromS3Model struct {
	AssetSources fwtypes.ListNestedObjectValueOf[assetSourceModel] `tfsdk:"asset_source"`
}

type assetSourceModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Key    types.String `tfsdk:"key"`
}

type importAssetsFromSignedURLModel struct {
	Filename types.String `tfsdk:"filename"`
}

type createS3DataAccessFromS3BucketModel struct {
	AccessPointAlias types.String                                                  `tfsdk:"access_point_alias"`
	AccessPointARN   types.String                                                  `tfsdk:"access_point_arn"`
	AssetSource      fwtypes.ListNestedObjectValueOf[s3DataAccessAssetSourceModel] `tfsdk:"asset_source"`
}

type s3DataAccessAssetSourceModel struct {
	Bucket         types.String                                        `tfsdk:"bucket"`
	KeyPrefixes    fwtypes.SetOfString                                 `tfsdk:"key_prefixes"`
	Keys           fwtypes.SetOfString                                 `tfsdk:"keys"`
	KmsKeysToGrant fwtypes.ListNestedObjectValueOf[kmsKeyToGrantModel] `tfsdk:"kms_keys_to_grant"`
}

type kmsKeyToGrantModel struct {
	KmsKeyArn fwtypes.ARN `tfsdk:"kms_key_arn"`
}

func waitJobCompleted(ctx context.Context, conn *dataexchange.Client, jobID string, timeout time.Duration) (*dataexchange.GetJobOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.StateWaiting, awstypes.StateInProgress),
		Target:       enum.Slice(awstypes.StateCompleted),
		Refresh:      statusJob(conn, jobID),
		Timeout:      timeout,
		PollInterval: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dataexchange.GetJobOutput); ok {
		if output.State == awstypes.StateError {
			return output, jobError(output.Errors)
		}

		return output, err
	}

	return nil, err
}

func jobError(errs []awstypes.JobError) error {
	return errors.Join(tfslices.ApplyToAll(errs, func(e awstypes.JobError) error {
		return fmt.Errorf("%s: %s", e.Code, *e.Message)
	})...)
}

func startJob(ctx context.Context, id *string, conn *dataexchange.Client) error {
	startJobInput := dataexchange.StartJobInput{
		JobId: id,
	}
	_, err := conn.StartJob(ctx, &startJobInput)

	return err
}

func statusJob(conn *dataexchange.Client, jobID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findJobByID(ctx, conn, jobID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func findJobByID(ctx context.Context, conn *dataexchange.Client, jobID string) (*dataexchange.GetJobOutput, error) {
	input := dataexchange.GetJobInput{
		JobId: aws.String(jobID),
	}

	out, err := conn.GetJob(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func nestedObjectCollectionAll[T any](ctx context.Context, v nestedObjectCollectionValue[T]) (iter.Seq2[int, *T], diag.Diagnostics) {
	s, diags := v.ToSlice(ctx)
	if diags.HasError() {
		return nil, diags
	}

	return func(yield func(int, *T) bool) {
		for i, e := range slices.All(s) {
			if !yield(i, e) {
				return
			}
		}
	}, diags
}

func nestedObjectCollectionAllMust[T any](ctx context.Context, v nestedObjectCollectionValue[T]) iter.Seq2[int, *T] {
	return fwdiag.Must(nestedObjectCollectionAll(ctx, v))
}

type nestedObjectCollectionValue[T any] interface {
	ToSlice(context.Context) ([]*T, diag.Diagnostics)
}

func md5Reader(src io.Reader) (string, error) {
	// MD5 is required by AWS DataExchange API for asset upload integrity checking.
	// This is not used for cryptographic security purposes.
	h := md5.New()
	if _, err := io.Copy(h, src); err != nil {
		return "", err
	}

	return inttypes.Base64Encode(h.Sum(nil)), nil
}
