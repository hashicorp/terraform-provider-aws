// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatchomni_domain", name="Domain")
// @IdentityAttribute("domain_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types;awstypes;awstypes.Domain")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(importStateIdAttribute="domain_id")
// @Testing(tagsTest=false)
func newDomainResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &domainResource{}, nil
}

const (
	domainDeletedTimeout = 10 * time.Minute
)

type domainResource struct {
	framework.ResourceWithModel[domainResourceModel]
	framework.WithImportByIdentity
}

func (r *domainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"custom_endpoint_urls": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
			"domain_endpoint_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_id": framework.IDAttribute(),
			"identity_center_application_arn": schema.StringAttribute{
				Computed: true,
			},
			"identity_providers": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringEnumType[awstypes.IdentityProvider](),
				Required:   true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z]+(-[0-9a-z]+)*$`), "must contain only lowercase letters, numbers, and single hyphens, and must begin and end with a letter or number"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"identity_provider_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[identityProviderConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"identity_center_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[identityCenterConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"identity_center_instance_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input cloudwatchomni.CreateDomainInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	out, err := conn.CreateDomain(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.Domain, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	domainID := fwflex.StringValueFromFramework(ctx, state.DomainID)
	out, err := findDomainByID(ctx, conn, domainID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, domainID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state domainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		domainID := fwflex.StringValueFromFramework(ctx, plan.DomainID)
		var input cloudwatchomni.UpdateDomainInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateDomain(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, domainID)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.Domain, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// Computed attributes without UseStateForUnknown are unknown in the plan.
		plan.CustomEndpointURLs = state.CustomEndpointURLs
		plan.IdentityCenterApplicationARN = state.IdentityCenterApplicationARN
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	domainID := fwflex.StringValueFromFramework(ctx, state.DomainID)
	input := cloudwatchomni.DeleteDomainInput{
		DomainId: aws.String(domainID),
	}
	_, err := conn.DeleteDomain(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, domainID)
		return
	}

	if _, err := tfresource.RetryUntilNotFound(ctx, domainDeletedTimeout, func(ctx context.Context) (any, error) {
		return findDomainByID(ctx, conn, domainID)
	}); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, domainID)
		return
	}
}

func (r *domainResource) flatten(ctx context.Context, domain *awstypes.Domain, data *domainResourceModel) diag.Diagnostics {
	// IAM-only domains return an empty identity provider configuration; treat it as absent.
	if v := domain.IdentityProviderConfiguration; v != nil && v.IdentityCenterConfiguration == nil {
		domain.IdentityProviderConfiguration = nil
	}

	return fwflex.Flatten(ctx, domain, data, fwflex.WithIgnoredFieldNamesAppend("Region"))
}

func findDomainByID(ctx context.Context, conn *cloudwatchomni.Client, id string) (*awstypes.Domain, error) {
	input := cloudwatchomni.GetDomainInput{
		DomainId: aws.String(id),
	}

	return findDomain(ctx, conn, &input)
}

func findDomain(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.GetDomainInput) (*awstypes.Domain, error) {
	out, err := conn.GetDomain(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Domain == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Domain, nil
}

type domainResourceModel struct {
	framework.WithRegionModel
	CustomEndpointURLs            fwtypes.ListOfString                                                `tfsdk:"custom_endpoint_urls"`
	DomainARN                     types.String                                                        `tfsdk:"arn"`
	DomainEndpointURL             types.String                                                        `tfsdk:"domain_endpoint_url"`
	DomainID                      types.String                                                        `tfsdk:"domain_id"`
	IdentityCenterApplicationARN  types.String                                                        `tfsdk:"identity_center_application_arn"`
	IdentityProviderConfiguration fwtypes.ListNestedObjectValueOf[identityProviderConfigurationModel] `tfsdk:"identity_provider_configuration"`
	IdentityProviders             fwtypes.SetOfStringEnum[awstypes.IdentityProvider]                  `tfsdk:"identity_providers"`
	Name                          types.String                                                        `tfsdk:"name"`
}

type identityProviderConfigurationModel struct {
	IdentityCenterConfiguration fwtypes.ListNestedObjectValueOf[identityCenterConfigurationModel] `tfsdk:"identity_center_configuration"`
}

type identityCenterConfigurationModel struct {
	IdentityCenterInstanceARN fwtypes.ARN `tfsdk:"identity_center_instance_arn"`
}
