// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"strings"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatchomni_organization_domain", name="Organization Domain")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("domain_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types;awstypes;awstypes.OrganizationDomain")
// @Testing(preCheck="testAccPreCheckOrganizationDomain")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(importStateIdAttribute="domain_id")
func newOrganizationDomainResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &organizationDomainResource{}, nil
}

type organizationDomainResource struct {
	framework.ResourceWithModel[organizationDomainResourceModel]
	framework.WithImportByIdentity
}

func (r *organizationDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"custom_endpoint_urls": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
			"domain_access_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`), "use lowercase letters, numbers, and hyphens. Start and end with a letter or number."),
				},
			},
			"organization_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DomainStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
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

func (r *organizationDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationDomainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input cloudwatchomni.CreateDomainForOrganizationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	// CreateDomainForOrganization takes a tag map, but the tag operations live in the cloudwatch package and use a slice.
	if tags := keyValueTags(ctx, getTagsIn(ctx)).Map(); len(tags) > 0 {
		input.Tags = tags
	}

	const (
		timeout = 2 * time.Minute
	)
	out, err := tfresource.RetryWhen(ctx, timeout, func(ctx context.Context) (*cloudwatchomni.CreateDomainForOrganizationOutput, error) {
		return conn.CreateDomainForOrganization(ctx, &input)
	}, isRetryableOrganizationDomainError(true))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.OrganizationDomain, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *organizationDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationDomainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	domainID := fwflex.StringValueFromFramework(ctx, state.DomainID)
	out, err := findOrganizationDomainByID(ctx, conn, domainID)
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

func (r *organizationDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state organizationDomainResourceModel
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
		var input cloudwatchomni.UpdateDomainForOrganizationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		const (
			timeout = 2 * time.Minute
		)
		out, err := tfresource.RetryWhen(ctx, timeout, func(ctx context.Context) (*cloudwatchomni.UpdateDomainForOrganizationOutput, error) {
			return conn.UpdateDomainForOrganization(ctx, &input)
		}, isRetryableOrganizationDomainError(false))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, domainID)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.OrganizationDomain, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		// Computed attributes without UseStateForUnknown are unknown in the plan.
		plan.CustomEndpointURLs = state.CustomEndpointURLs
		plan.IdentityCenterApplicationARN = state.IdentityCenterApplicationARN
		plan.Status = state.Status
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *organizationDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationDomainResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	domainID := fwflex.StringValueFromFramework(ctx, state.DomainID)
	input := cloudwatchomni.DeleteDomainForOrganizationInput{
		DomainId: aws.String(domainID),
	}
	_, err := conn.DeleteDomainForOrganization(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, domainID)
		return
	}

	const (
		timeout = 10 * time.Minute
	)
	if _, err := tfresource.RetryUntilNotFound(ctx, timeout, func(ctx context.Context) (any, error) {
		return findOrganizationDomainByID(ctx, conn, domainID)
	}); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, domainID)
		return
	}
}

func (r *organizationDomainResource) flatten(ctx context.Context, domain *awstypes.OrganizationDomain, data *organizationDomainResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	// IAM-only domains return an empty identity provider configuration; treat it as absent.
	if v := domain.IdentityProviderConfiguration; v != nil && v.IdentityCenterConfiguration == nil {
		domain.IdentityProviderConfiguration = nil
	}
	diags.Append(fwflex.Flatten(ctx, domain, data, fwflex.WithIgnoredFieldNamesAppend("Region"))...)
	return diags
}

// isRetryableOrganizationDomainError retries IAM Identity Center serializing mutating operations on its instance
// and, on create, IAM not yet allowing the service to assume a newly created domain access role.
func isRetryableOrganizationDomainError(isCreate bool) tfresource.Retryable {
	return func(err error) (bool, error) {
		switch {
		case errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "conflicting operation in progress"):
			return true, err
		case isCreate && errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Unable to assume the domain access role"):
			return true, err
		default:
			return false, err
		}
	}
}

func findOrganizationDomainByID(ctx context.Context, conn *cloudwatchomni.Client, id string) (*awstypes.OrganizationDomain, error) {
	input := cloudwatchomni.GetDomainForOrganizationInput{
		DomainId: aws.String(id),
	}

	return findOrganizationDomain(ctx, conn, &input)
}

func findOrganizationDomain(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.GetDomainForOrganizationInput) (*awstypes.OrganizationDomain, error) {
	out, err := conn.GetDomainForOrganization(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.OrganizationDomain == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.OrganizationDomain, nil
}

// isOrganizationDomainARN reports whether a ListDomains entry is an organization domain.
func isOrganizationDomainARN(arn string) bool {
	return strings.Contains(arn, ":organization-domain/")
}

type organizationDomainResourceModel struct {
	framework.WithRegionModel
	CustomEndpointURLs            fwtypes.ListOfString                                                `tfsdk:"custom_endpoint_urls"`
	DomainAccessRoleARN           fwtypes.ARN                                                         `tfsdk:"domain_access_role_arn"`
	DomainARN                     types.String                                                        `tfsdk:"arn"`
	DomainEndpointURL             types.String                                                        `tfsdk:"domain_endpoint_url"`
	DomainID                      types.String                                                        `tfsdk:"domain_id"`
	IdentityCenterApplicationARN  types.String                                                        `tfsdk:"identity_center_application_arn"`
	IdentityProviderConfiguration fwtypes.ListNestedObjectValueOf[identityProviderConfigurationModel] `tfsdk:"identity_provider_configuration"`
	IdentityProviders             fwtypes.SetOfStringEnum[awstypes.IdentityProvider]                  `tfsdk:"identity_providers"`
	Name                          types.String                                                        `tfsdk:"name"`
	OrganizationID                types.String                                                        `tfsdk:"organization_id"`
	OwnerAccountID                types.String                                                        `tfsdk:"owner_account_id"`
	Status                        fwtypes.StringEnum[awstypes.DomainStatus]                           `tfsdk:"status"`
	Tags                          tftags.Map                                                          `tfsdk:"tags"`
	TagsAll                       tftags.Map                                                          `tfsdk:"tags_all"`
}
