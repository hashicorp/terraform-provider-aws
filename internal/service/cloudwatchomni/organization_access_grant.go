// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatchomni_organization_access_grant", name="Organization Access Grant")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("grant_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types;awstypes;awstypes.OrganizationAccessGrant")
// @Testing(preCheck="testAccPreCheckOrganizationAccessGrant")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(importStateIdAttribute="grant_id")
func newOrganizationAccessGrantResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &organizationAccessGrantResource{}, nil
}

// There is no UpdateDomainAccessGrantForOrganization operation, so every argument except tags forces replacement.
type organizationAccessGrantResource struct {
	framework.ResourceWithModel[organizationAccessGrantResourceModel]
	framework.WithImportByIdentity
}

func (r *organizationAccessGrantResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"domain_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^d-[0-9a-z]{1,25}$`), "must be a CloudWatch Omni domain ID"),
				},
			},
			"grant_id": framework.IDAttribute(),
			"grant_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AccessGrantType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_-]+$`), "use letters, numbers, underscores, and hyphens."),
				},
			},
			"permission": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OrganizationGrantPermission](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrPrincipal: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[organizationAccessGrantPrincipalModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.AtLeastOneOfChildren(
							path.MatchRelative().AtName("principal_attribute"),
							path.MatchRelative().AtName("principal_id"),
						),
					},
					Attributes: map[string]schema.Attribute{
						"principal_id": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
								stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9:/_+=,.@-]*$`), "start with a letter or number and use letters, numbers, and the characters :/_+=,.@-"),
							},
						},
						"principal_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.OrganizationGrantPrincipalType](),
							Required:   true,
							Validators: []validator.String{
								// Principal attributes are only valid for Identity Center users.
								tfstringvalidator.ConflictsWithWhenNotEquals(
									awstypes.OrganizationGrantPrincipalTypeIdcUser,
									path.MatchRelative().AtParent().AtName("principal_attribute"),
								),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"principal_attribute": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[organizationAccessGrantPrincipalAttributeModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(10),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
										},
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

func (r *organizationAccessGrantResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data organizationAccessGrantResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input cloudwatchomni.CreateDomainAccessGrantForOrganizationInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	// CreateDomainAccessGrantForOrganization takes a tag map, but the tag operations live in the cloudwatch package and use a slice.
	if tags := keyValueTags(ctx, getTagsIn(ctx)).Map(); len(tags) > 0 {
		input.Tags = tags
	}

	const (
		timeout = 2 * time.Minute
	)
	output, err := tfresource.RetryWhen(ctx, timeout, func(ctx context.Context) (*cloudwatchomni.CreateDomainAccessGrantForOrganizationOutput, error) {
		return conn.CreateDomainAccessGrantForOrganization(ctx, &input)
	}, isRetryableOrganizationAccessGrantConflict)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output.AccessGrant, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *organizationAccessGrantResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data organizationAccessGrantResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	grantID := fwflex.StringValueFromFramework(ctx, data.GrantID)
	output, err := findOrganizationAccessGrantByID(ctx, conn, grantID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, grantID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *organizationAccessGrantResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data organizationAccessGrantResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	grantID := fwflex.StringValueFromFramework(ctx, data.GrantID)
	input := cloudwatchomni.DeleteDomainAccessGrantForOrganizationInput{
		GrantId: aws.String(grantID),
	}
	// CloudWatch Omni serializes grant mutations per domain, returning a ConflictException
	// while another grant on the domain is being created or deleted.
	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhen(ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.DeleteDomainAccessGrantForOrganization(ctx, &input)
	}, isRetryableOrganizationAccessGrantConflict)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, grantID)
		return
	}
}

func (r *organizationAccessGrantResource) flatten(ctx context.Context, grant *awstypes.OrganizationAccessGrant, data *organizationAccessGrantResourceModel) diag.Diagnostics {
	// Unset collections are returned empty rather than absent; treat them as absent.
	if grant.Principal != nil && len(grant.Principal.PrincipalAttributes) == 0 {
		grant.Principal.PrincipalAttributes = nil
	}

	return fwflex.Flatten(ctx, grant, data, fwflex.WithIgnoredFieldNamesAppend("Region"))
}

// isRetryableOrganizationAccessGrantConflict retries while CloudWatch Omni serializes a
// concurrent grant mutation on the domain.
func isRetryableOrganizationAccessGrantConflict(err error) (bool, error) {
	if errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "concurrent") {
		return true, err
	}
	return false, err
}

func findOrganizationAccessGrantByID(ctx context.Context, conn *cloudwatchomni.Client, id string) (*awstypes.OrganizationAccessGrant, error) {
	input := cloudwatchomni.GetDomainAccessGrantForOrganizationInput{
		GrantId: aws.String(id),
	}

	return findOrganizationAccessGrant(ctx, conn, &input)
}

func findOrganizationAccessGrant(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.GetDomainAccessGrantForOrganizationInput) (*awstypes.OrganizationAccessGrant, error) {
	output, err := conn.GetDomainAccessGrantForOrganization(ctx, input)
	switch {
	case errs.IsA[*awstypes.ResourceNotFoundException](err):
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	case err != nil:
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.AccessGrant == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.AccessGrant, nil
}

func listOrganizationAccessGrantPages(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListDomainAccessGrantsForOrganizationInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[[]awstypes.OrganizationAccessGrantSummary, error] {
	return func(yield func([]awstypes.OrganizationAccessGrantSummary, error) bool) {
		pages := cloudwatchomni.NewListDomainAccessGrantsForOrganizationPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Omni Organization Access Grants: %w", err))
				return
			}

			if !yield(page.Items, nil) {
				return
			}
		}
	}
}

type organizationAccessGrantResourceModel struct {
	framework.WithRegionModel
	DomainID   types.String                                                           `tfsdk:"domain_id"`
	GrantARN   types.String                                                           `tfsdk:"arn"`
	GrantID    types.String                                                           `tfsdk:"grant_id"`
	GrantType  fwtypes.StringEnum[awstypes.AccessGrantType]                           `tfsdk:"grant_type"`
	Name       types.String                                                           `tfsdk:"name"`
	Permission fwtypes.StringEnum[awstypes.OrganizationGrantPermission]               `tfsdk:"permission"`
	Principal  fwtypes.ListNestedObjectValueOf[organizationAccessGrantPrincipalModel] `tfsdk:"principal"`
	Tags       tftags.Map                                                             `tfsdk:"tags"`
	TagsAll    tftags.Map                                                             `tfsdk:"tags_all"`
}

type organizationAccessGrantPrincipalModel struct {
	PrincipalAttributes fwtypes.ListNestedObjectValueOf[organizationAccessGrantPrincipalAttributeModel] `tfsdk:"principal_attribute"`
	PrincipalID         types.String                                                                    `tfsdk:"principal_id"`
	PrincipalType       fwtypes.StringEnum[awstypes.OrganizationGrantPrincipalType]                     `tfsdk:"principal_type"`
}

type organizationAccessGrantPrincipalAttributeModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}
