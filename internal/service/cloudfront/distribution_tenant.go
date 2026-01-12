// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudfront_distribution_tenant", name="Distribution Tenant")
// @Tags(identifierAttribute="arn")
func newDistributionTenantResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &distributionTenantResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultUpdateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

const (
	distributionTenantPollInterval = 30 * time.Second
)

type distributionTenantResource struct {
	framework.ResourceWithModel[distributionTenantResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *distributionTenantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"connection_group_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"distribution_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrEnabled: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"wait_for_deployment": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
		},
		Blocks: map[string]schema.Block{
			"customizations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customizationsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrCertificate: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[certificateModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
						"geo_restriction": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[geoRestrictionCustomizationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"locations": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Optional:   true,
										Computed:   true,
									},
									"restriction_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.GeoRestrictionType](),
										Optional:   true,
									},
								},
							},
						},
						"web_acl": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[webAclCustomizationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrAction: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.CustomizationActionType](),
										Optional:   true,
									},
									names.AttrARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrDomain: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[domainResultModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDomain: schema.StringAttribute{
							Required: true,
						},
						names.AttrStatus: schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"managed_certificate_request": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[managedCertificateRequestModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"certificate_transparency_logging_preference": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.CertificateTransparencyLoggingPreference](),
							Optional:   true,
						},
						"primary_domain_name": schema.StringAttribute{
							Optional: true,
						},
						"validation_token_host": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ValidationTokenHost](),
							Optional:   true,
						},
					},
				},
			},
			names.AttrParameter: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[parameterModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						names.AttrValue: schema.StringAttribute{
							Required: true,
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

type distributionTenantResourceModel struct {
	ARN                       types.String                                                    `tfsdk:"arn"`
	ConnectionGroupID         types.String                                                    `tfsdk:"connection_group_id"`
	Customizations            fwtypes.ListNestedObjectValueOf[customizationsModel]            `tfsdk:"customizations"`
	DistributionID            types.String                                                    `tfsdk:"distribution_id"`
	Domains                   fwtypes.SetNestedObjectValueOf[domainResultModel]               `tfsdk:"domain" autoflex:",xmlwrapper=Items"`
	Enabled                   types.Bool                                                      `tfsdk:"enabled"`
	ETag                      types.String                                                    `tfsdk:"etag"`
	ID                        types.String                                                    `tfsdk:"id"`
	ManagedCertificateRequest fwtypes.ListNestedObjectValueOf[managedCertificateRequestModel] `tfsdk:"managed_certificate_request"`
	Name                      types.String                                                    `tfsdk:"name"`
	Parameters                fwtypes.SetNestedObjectValueOf[parameterModel]                  `tfsdk:"parameter" autoflex:",xmlwrapper=Items"`
	Status                    types.String                                                    `tfsdk:"status"`
	Tags                      tftags.Map                                                      `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                      `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                  `tfsdk:"timeouts"`
	WaitForDeployment         types.Bool                                                      `tfsdk:"wait_for_deployment"`
}

type customizationsModel struct {
	Certificate    fwtypes.ListNestedObjectValueOf[certificateModel]                 `tfsdk:"certificate"`
	GeoRestriction fwtypes.ListNestedObjectValueOf[geoRestrictionCustomizationModel] `tfsdk:"geo_restriction"`
	WebAcl         fwtypes.ListNestedObjectValueOf[webAclCustomizationModel]         `tfsdk:"web_acl"`
}

// Remove manual flattener interfaces - let AutoFlex handle everything

type domainResultModel struct {
	Domain types.String `tfsdk:"domain"`
	Status types.String `tfsdk:"status"`
}

type geoRestrictionCustomizationModel struct {
	Locations       fwtypes.SetOfString                             `tfsdk:"locations"`
	RestrictionType fwtypes.StringEnum[awstypes.GeoRestrictionType] `tfsdk:"restriction_type"`
}

type certificateModel struct {
	ARN fwtypes.ARN `tfsdk:"arn"`
}

type webAclCustomizationModel struct {
	Action fwtypes.StringEnum[awstypes.CustomizationActionType] `tfsdk:"action"`
	ARN    fwtypes.ARN                                          `tfsdk:"arn"`
}

type managedCertificateRequestModel struct {
	CertificateTransparencyLoggingPreference fwtypes.StringEnum[awstypes.CertificateTransparencyLoggingPreference] `tfsdk:"certificate_transparency_logging_preference"`
	PrimaryDomainName                        types.String                                                          `tfsdk:"primary_domain_name"`
	ValidationTokenHost                      fwtypes.StringEnum[awstypes.ValidationTokenHost]                      `tfsdk:"validation_token_host"`
}

type parameterModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (r *distributionTenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data distributionTenantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input cloudfront.CreateDistributionTenantInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = &awstypes.Tags{
			Items: tags,
		}
	}

	output, err := conn.CreateDistributionTenant(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating CloudFront Distribution Tenant (%s)", name), err.Error())
		return
	}

	// Use create response directly - no extra read needed
	resp.Diagnostics.Append(fwflex.Flatten(ctx, output.DistributionTenant, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use AutoFlex to flatten the response
	resp.Diagnostics.Append(fwflex.Flatten(ctx, output.DistributionTenant, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set fields that AutoFlex might not handle correctly
	id := aws.ToString(output.DistributionTenant.Id)
	data.ID = fwflex.StringValueToFramework(ctx, id)
	data.ARN = fwflex.StringToFramework(ctx, output.DistributionTenant.Arn)
	data.ETag = fwflex.StringToFramework(ctx, output.ETag)

	if data.WaitForDeployment.ValueBool() {
		if _, err := waitDistributionTenantDeployed(ctx, conn, id); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("waiting CloudFront Distribution Tenant (%s) deploy", id), err.Error())
			return
		}

		// Wait for managed certificate if specified
		if !data.ManagedCertificateRequest.IsNull() && !data.ManagedCertificateRequest.IsUnknown() {
			var managedCertRequest *awstypes.ManagedCertificateRequest
			resp.Diagnostics.Append(fwflex.Expand(ctx, data.ManagedCertificateRequest, &managedCertRequest)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if err := waitManagedCertificateReady(ctx, conn, id, managedCertRequest); err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("waiting CloudFront Distribution Tenant (%s) managed certificate", id), err.Error())
				return
			}

			// Refresh the distribution tenant data after managed certificate processing
			refreshedOutput, err := findDistributionTenantByIdentifier(ctx, conn, id)
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Distribution Tenant (%s)", id), err.Error())
				return
			}

			// Update the data model with refreshed information
			resp.Diagnostics.Append(fwflex.Flatten(ctx, refreshedOutput.DistributionTenant, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			// Use AutoFlex to flatten the refreshed response
			resp.Diagnostics.Append(fwflex.Flatten(ctx, refreshedOutput.DistributionTenant, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			data.ETag = fwflex.StringToFramework(ctx, refreshedOutput.ETag)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *distributionTenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data distributionTenantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	output, err := findDistributionTenantByIdentifier(ctx, conn, id)
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Distribution Tenant (%s)", id), err.Error())
		return
	}

	tenant := output.DistributionTenant

	// Flatten the distribution tenant data into the model
	resp.Diagnostics.Append(fwflex.Flatten(ctx, tenant, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use AutoFlex to flatten the response
	resp.Diagnostics.Append(fwflex.Flatten(ctx, tenant, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set computed fields that need special handling
	data.ETag = fwflex.StringToFramework(ctx, output.ETag)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *distributionTenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new distributionTenantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, new.ID)
	var output *cloudfront.UpdateDistributionTenantOutput

	// Check if configuration changed (excluding tags)
	if !new.ConnectionGroupID.Equal(old.ConnectionGroupID) ||
		!new.Customizations.Equal(old.Customizations) ||
		!new.DistributionID.Equal(old.DistributionID) ||
		!new.Domains.Equal(old.Domains) ||
		!new.Enabled.Equal(old.Enabled) ||
		!new.ManagedCertificateRequest.Equal(old.ManagedCertificateRequest) ||
		!new.Parameters.Equal(old.Parameters) {
		input := &cloudfront.UpdateDistributionTenantInput{}
		resp.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Handle special fields manually
		input.Id = aws.String(id)
		input.IfMatch = fwflex.StringFromFramework(ctx, old.ETag)

		_, err := conn.UpdateDistributionTenant(ctx, input)

		// Refresh our ETag if it is out of date and attempt update again.
		if errs.IsA[*awstypes.PreconditionFailed](err) {
			var etag string
			etag, err = distributionTenantETag(ctx, conn, id)

			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Distribution Tenant (%s) Etag", id), err.Error())
				return
			}

			input.IfMatch = aws.String(etag)
			output, err = conn.UpdateDistributionTenant(ctx, input)
		}

		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating CloudFront Distribution Tenant (%s)", id), err.Error())
			return
		}

		if new.WaitForDeployment.ValueBool() {
			if _, err := waitDistributionTenantDeployed(ctx, conn, id); err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("waiting CloudFront Distribution Tenant (%s) deploy", id), err.Error())
				return
			}

			// Wait for managed certificate if specified
			if !new.ManagedCertificateRequest.IsNull() && !new.ManagedCertificateRequest.IsUnknown() {
				var managedCertRequest *awstypes.ManagedCertificateRequest
				resp.Diagnostics.Append(fwflex.Expand(ctx, new.ManagedCertificateRequest, &managedCertRequest)...)
				if resp.Diagnostics.HasError() {
					return
				}

				if err := waitManagedCertificateReady(ctx, conn, id, managedCertRequest); err != nil {
					resp.Diagnostics.AddError(fmt.Sprintf("waiting CloudFront Distribution Tenant (%s) managed certificate", id), err.Error())
					return
				}

				// Refresh the distribution tenant data after managed certificate processing
				refreshedOutput, err := findDistributionTenantByIdentifier(ctx, conn, id)
				if err != nil {
					resp.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Distribution Tenant (%s)", id), err.Error())
					return
				}

				// Update the model with refreshed information
				resp.Diagnostics.Append(fwflex.Flatten(ctx, refreshedOutput.DistributionTenant, &new)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Manually flatten domains and parameters
				// Use AutoFlex to flatten the refreshed response
				resp.Diagnostics.Append(fwflex.Flatten(ctx, refreshedOutput.DistributionTenant, &new)...)
				if resp.Diagnostics.HasError() {
					return
				}

				new.ETag = fwflex.StringToFramework(ctx, refreshedOutput.ETag)
			}
		}
	}

	// Flatten the distribution tenant data into the model
	if output != nil {
		resp.Diagnostics.Append(fwflex.Flatten(ctx, output.DistributionTenant, &new)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Use AutoFlex to flatten the response
		resp.Diagnostics.Append(fwflex.Flatten(ctx, output.DistributionTenant, &new)...)
		if resp.Diagnostics.HasError() {
			return
		}

		new.ETag = fwflex.StringToFramework(ctx, output.ETag)
	} else {
		// If no update was performed (e.g., tag-only changes), we still need to refresh the distribution tenant data
		// to ensure all computed fields are properly set
		getOutput, err := findDistributionTenantByIdentifier(ctx, conn, id)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Distribution Tenant (%s)", id), err.Error())
			return
		}

		resp.Diagnostics.Append(fwflex.Flatten(ctx, getOutput.DistributionTenant, &new)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Use AutoFlex to flatten the response
		resp.Diagnostics.Append(fwflex.Flatten(ctx, getOutput.DistributionTenant, &new)...)
		if resp.Diagnostics.HasError() {
			return
		}

		new.ETag = fwflex.StringToFramework(ctx, getOutput.ETag)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *distributionTenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data distributionTenantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)
	id := fwflex.StringValueFromFramework(ctx, data.ID)

	if err := disableDistributionTenant(ctx, conn, id); err != nil {
		if retry.NotFound(err) || errs.IsA[*awstypes.EntityNotFound](err) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("disabling CloudFront Distribution Tenant (%s)", id), err.Error())
		return
	}

	err := deleteDistributionTenant(ctx, conn, id)

	if err == nil || retry.NotFound(err) || errs.IsA[*awstypes.EntityNotFound](err) {
		return
	}

	// Disable distribution tenant if it is not yet disabled and attempt deletion again.
	if errs.IsA[*awstypes.ResourceNotDisabled](err) {
		if err := disableDistributionTenant(ctx, conn, id); err != nil {
			if retry.NotFound(err) || errs.IsA[*awstypes.EntityNotFound](err) {
				return
			}
			resp.Diagnostics.AddError(fmt.Sprintf("disabling CloudFront Distribution Tenant (%s)", id), err.Error())
			return
		}

		_, err = tfresource.RetryWhenIsA[any, *awstypes.ResourceNotDisabled](ctx, distributionTenantPollInterval, func(ctx context.Context) (any, error) {
			return nil, deleteDistributionTenant(ctx, conn, id)
		})
	}

	if errs.IsA[*awstypes.PreconditionFailed](err) || errs.IsA[*awstypes.InvalidIfMatchVersion](err) {
		_, err = tfresource.RetryWhenIsOneOf2[any, *awstypes.PreconditionFailed, *awstypes.InvalidIfMatchVersion](ctx, distributionTenantPollInterval, func(ctx context.Context) (any, error) {
			return nil, deleteDistributionTenant(ctx, conn, id)
		})
	}

	if errs.IsA[*awstypes.ResourceNotDisabled](err) {
		if err := disableDistributionTenant(ctx, conn, id); err != nil {
			if retry.NotFound(err) || errs.IsA[*awstypes.EntityNotFound](err) {
				return
			}
			resp.Diagnostics.AddError(fmt.Sprintf("disabling CloudFront Distribution Tenant (%s)", id), err.Error())
			return
		}

		err = deleteDistributionTenant(ctx, conn, id)
	}

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting CloudFront Distribution Tenant (%s)", id), err.Error())
		return
	}
}

func findDistributionTenantByIdentifier(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetDistributionTenantOutput, error) {
	input := cloudfront.GetDistributionTenantInput{
		Identifier: aws.String(id),
	}

	return findDistributionTenant(ctx, conn, &input)
}

func findDistributionTenant(ctx context.Context, conn *cloudfront.Client, input *cloudfront.GetDistributionTenantInput) (*cloudfront.GetDistributionTenantOutput, error) {
	output, err := conn.GetDistributionTenant(ctx, input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DistributionTenant == nil || output.DistributionTenant.Domains == nil || output.DistributionTenant.DistributionId == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func disableDistributionTenant(ctx context.Context, conn *cloudfront.Client, id string) error {
	output, err := findDistributionTenantByIdentifier(ctx, conn, id)

	if err != nil {
		return fmt.Errorf("reading CloudFront Distribution Tenant (%s): %w", id, err)
	}

	if aws.ToString(output.DistributionTenant.Status) == distributionTenantStatusInProgress {
		output, err = waitDistributionTenantDeployed(ctx, conn, id)

		if err != nil {
			return fmt.Errorf("waiting for CloudFront Distribution Tenant (%s) deploy: %w", id, err)
		}
	}

	if !aws.ToBool(output.DistributionTenant.Enabled) {
		return nil
	}

	input := cloudfront.UpdateDistributionTenantInput{
		ConnectionGroupId: output.DistributionTenant.ConnectionGroupId,
		Customizations:    output.DistributionTenant.Customizations,
		DistributionId:    output.DistributionTenant.DistributionId,
		Domains:           convertDomainResultsToDomainItems(output.DistributionTenant.Domains),
		Enabled:           aws.Bool(false),
		Id:                aws.String(id),
		IfMatch:           output.ETag,
		Parameters:        output.DistributionTenant.Parameters,
	}
	_, err = conn.UpdateDistributionTenant(ctx, &input)

	if err != nil {
		return fmt.Errorf("updating CloudFront Distribution Tenant (%s): %w", id, err)
	}

	if _, err := waitDistributionTenantDeployed(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Distribution Tenant (%s) deploy: %w", id, err)
	}

	return nil
}

func deleteDistributionTenant(ctx context.Context, conn *cloudfront.Client, id string) error {
	etag, err := distributionTenantETag(ctx, conn, id)

	if err != nil {
		return err
	}

	input := cloudfront.DeleteDistributionTenantInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}
	_, err = conn.DeleteDistributionTenant(ctx, &input)

	if err != nil {
		return fmt.Errorf("deleting CloudFront Distribution Tenant (%s): %w", id, err)
	}

	if _, err := waitDistributionTenantDeleted(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Distribution Tenant (%s) delete: %w", id, err)
	}

	return nil
}

func waitDistributionTenantDeployed(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetDistributionTenantOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{distributionTenantStatusInProgress},
		Target:     []string{distributionTenantStatusDeployed},
		Refresh:    statusDistributionTenant(conn, id),
		Timeout:    30 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetDistributionTenantOutput); ok {
		return output, err
	}

	return nil, err
}

func waitDistributionTenantDeleted(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetDistributionTenantOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{distributionTenantStatusInProgress, distributionTenantStatusDeployed},
		Target:     []string{},
		Refresh:    statusDistributionTenant(conn, id),
		Timeout:    30 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetDistributionTenantOutput); ok {
		return output, err
	}

	return nil, err
}

func statusDistributionTenant(conn *cloudfront.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findDistributionTenantByIdentifier(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.ToString(output.DistributionTenant.Status), nil
	}
}

func distributionTenantETag(ctx context.Context, conn *cloudfront.Client, id string) (string, error) {
	output, err := findDistributionTenantByIdentifier(ctx, conn, id)

	if err != nil {
		return "", fmt.Errorf("reading CloudFront Distribution Tenant (%s): %w", id, err)
	}

	return aws.ToString(output.ETag), nil
}

func waitManagedCertificateReady(ctx context.Context, conn *cloudfront.Client, id string, managedCertRequest *awstypes.ManagedCertificateRequest) error {
	if managedCertRequest == nil {
		// No managed certificate request, nothing to wait for
		return nil
	}

	// Wait for distribution tenant to be deployed first
	dtOutput, err := waitForDistributionTenantDeployed(ctx, conn, id)
	if err != nil {
		return fmt.Errorf("waiting for CloudFront Distribution Tenant (%s) deploy: %w", id, err)
	}

	// Step 1: Wait for managed certificate to be issued (3 hours max)
	mcOutput, err := waitForManagedCertificateIssued(ctx, conn, id)
	if err != nil {
		return fmt.Errorf("CloudFront Distribution Tenant (%s) managed certificate issuance failed: %w", id, err)
	}

	// Step 2: Update distribution tenant with the issued certificate
	return updateDistributionTenantWithManagedCertificate(ctx, conn, dtOutput, mcOutput)
}

func waitForDistributionTenantDeployed(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetDistributionTenantOutput, error) {
	// Simple loop to wait for deployment - reuse existing logic if needed
	for {
		dtOutput, err := findDistributionTenantByIdentifier(ctx, conn, id)
		if err != nil {
			return nil, fmt.Errorf("failed reading CloudFront Distribution Tenant (%s): %w", id, err)
		}

		if aws.ToString(dtOutput.DistributionTenant.Status) == distributionTenantStatusDeployed {
			return dtOutput, nil
		}

		time.Sleep(distributionTenantPollInterval)
	}
}

func waitForManagedCertificateIssued(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetManagedCertificateDetailsOutput, error) {
	timeout := 3 * time.Hour
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		mcInput := &cloudfront.GetManagedCertificateDetailsInput{
			Identifier: aws.String(id),
		}

		mcOutput, err := conn.GetManagedCertificateDetails(ctx, mcInput)
		if errs.IsA[*awstypes.EntityNotFound](err) {
			// No managed certificate found - domains are covered by existing certs
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("failed getting CloudFront Distribution Tenant (%s) managed certificate details: %w", id, err)
		}

		// Check certificate status
		switch mcOutput.ManagedCertificateDetails.CertificateStatus {
		case awstypes.ManagedCertificateStatusIssued:
			return mcOutput, nil

		case awstypes.ManagedCertificateStatusPendingValidation:
			// Certificate still being validated, continue waiting
			time.Sleep(1 * time.Minute) // Longer sleep for certificate issuance
			continue

		default:
			return nil, fmt.Errorf("CloudFront Distribution Tenant (%s) managed certificate failed with status: %s", id, mcOutput.ManagedCertificateDetails.CertificateStatus)
		}
	}

	return nil, fmt.Errorf("CloudFront Distribution Tenant (%s) timeout after 3 hours waiting for managed certificate to be issued", id)
}

func updateDistributionTenantWithManagedCertificate(ctx context.Context, conn *cloudfront.Client, dtOutput *cloudfront.GetDistributionTenantOutput, mcOutput *cloudfront.GetManagedCertificateDetailsOutput) error {
	// Check if we need to update the certificate ARN
	if !needToUpdateCertificateARN(dtOutput.DistributionTenant, aws.ToString(mcOutput.ManagedCertificateDetails.CertificateArn)) {
		// Certificate ARN already matches, nothing to do
		return nil
	}

	// Get fresh ETag before update
	freshOutput, err := findDistributionTenantByIdentifier(ctx, conn, aws.ToString(dtOutput.DistributionTenant.Id))
	if err != nil {
		return fmt.Errorf("failed reading CloudFront Distribution Tenant (%s): %w", aws.ToString(dtOutput.DistributionTenant.Id), err)
	}

	// Update distribution tenant with managed certificate ARN
	updateInput := &cloudfront.UpdateDistributionTenantInput{
		Id:      dtOutput.DistributionTenant.Id,
		IfMatch: freshOutput.ETag,
		Customizations: &awstypes.Customizations{
			Certificate: &awstypes.Certificate{
				Arn: mcOutput.ManagedCertificateDetails.CertificateArn,
			},
		},
	}

	// Copy other required fields from current distribution tenant
	updateInput.ConnectionGroupId = dtOutput.DistributionTenant.ConnectionGroupId
	updateInput.DistributionId = dtOutput.DistributionTenant.DistributionId
	updateInput.Domains = convertDomainResultsToDomainItems(dtOutput.DistributionTenant.Domains)
	updateInput.Enabled = dtOutput.DistributionTenant.Enabled
	updateInput.Parameters = dtOutput.DistributionTenant.Parameters

	_, err = conn.UpdateDistributionTenant(ctx, updateInput)
	if err != nil {
		return fmt.Errorf("updating CloudFront Distribution Tenant (%s) with managed certificate: %w", aws.ToString(dtOutput.DistributionTenant.Id), err)
	}

	// Wait for the distribution tenant update to be deployed
	_, err = waitForDistributionTenantDeployed(ctx, conn, aws.ToString(dtOutput.DistributionTenant.Id))
	if err != nil {
		return fmt.Errorf("failed waiting for CloudFront Distribution Tenant (%s) deploy: %w", aws.ToString(dtOutput.DistributionTenant.Id), err)
	}

	return nil
}

func needToUpdateCertificateARN(dt *awstypes.DistributionTenant, certArn string) bool {
	if dt.Customizations == nil || dt.Customizations.Certificate == nil {
		return true
	}
	return certArn != aws.ToString(dt.Customizations.Certificate.Arn)
}

func convertDomainResultsToDomainItems(domainResults []awstypes.DomainResult) []awstypes.DomainItem {
	if len(domainResults) == 0 {
		return nil
	}

	domainItems := make([]awstypes.DomainItem, len(domainResults))
	for i, domainResult := range domainResults {
		domainItems[i] = awstypes.DomainItem{
			Domain: domainResult.Domain,
		}
	}

	return domainItems
}
