// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_apigatewayv2_portal", name="Portal")
// @Tags(identifierAttribute="portal_arn")
// @IdentityAttribute("portal_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigatewayv2;apigatewayv2.GetPortalOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="portal_id")
// @Testing(importIgnore="logo_uri;publish_description")
// @Testing(plannableImportAction="NoOp")
// @Testing(skipEmptyTags=true)
// @Testing(skipNullTags=true)
func newPortalResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &portalResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type portalResource struct {
	framework.ResourceWithModel[portalResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *portalResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	colorAttribute := func() schema.StringAttribute {
		return schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 16),
			},
		}
	}

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"included_portal_product_arns": schema.SetAttribute{
				CustomType:  fwtypes.SetOfARNType,
				ElementType: fwtypes.ARNType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"last_modified": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"last_published": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"last_published_description": schema.StringAttribute{
				Computed: true,
			},
			"logo_uri": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1092),
				},
			},
			"portal_arn": framework.ARNAttributeComputedOnly(),
			"portal_default_domain_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_domain_hosted_zone_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_id": framework.IDAttribute(),
			"publish": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"publish_description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
				},
			},
			"publish_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PublishStatus](),
				Computed:   true,
			},
			"rum_app_monitor_name": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"authorization": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[portalAuthorizationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("cognito_config"),
							path.MatchRelative().AtName("none"),
						),
					},
					Blocks: map[string]schema.Block{
						"cognito_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[portalCognitoConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"app_client_id": schema.StringAttribute{
										Required: true,
									},
									"user_pool_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"user_pool_domain": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"none": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[portalNoneModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
					},
				},
			},
			"endpoint_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[portalEndpointConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("acm_managed"),
							path.MatchRelative().AtName("none"),
						),
					},
					Blocks: map[string]schema.Block{
						"acm_managed": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[portalACMManagedModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrCertificateARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									names.AttrDomainName: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(3, 256),
										},
									},
								},
							},
						},
						"none": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[portalNoneModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
					},
				},
			},
			"portal_content": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[portalContentModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 1024),
							},
						},
						names.AttrDisplayName: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(3, 255),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"theme": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[portalThemeModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"custom_colors": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[portalCustomColorsModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"accent_color":           colorAttribute(),
												"background_color":       colorAttribute(),
												"error_validation_color": colorAttribute(),
												"header_color":           colorAttribute(),
												"navigation_color":       colorAttribute(),
												"text_color":             colorAttribute(),
											},
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
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *portalResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data portalResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	var input apigatewayv2.CreatePortalInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	tags := getTagsIn(ctx)
	// Tags on create are also authorized against every included portal product, which
	// fails for a product shared from another account. Tag such portals after creation.
	tagOnCreate := !slices.ContainsFunc(input.IncludedPortalProductArns, func(v string) bool {
		return !portalProductARNInAccount(v, r.Meta().AccountID(ctx))
	})
	if tagOnCreate {
		input.Tags = tags
	}

	out, err := conn.CreatePortal(ctx, &input)
	if err != nil {
		// There is no ID yet; identify the portal by its display name.
		var displayName string
		if v := input.PortalContent; v != nil {
			displayName = aws.ToString(v.DisplayName)
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, displayName)
		return
	}

	id := aws.ToString(out.PortalId)

	if !tagOnCreate && len(tags) > 0 {
		if err := updateTags(ctx, conn, aws.ToString(out.PortalArn), nil, tags); err != nil {
			// Taint the resource.
			response.State.SetAttribute(ctx, path.Root("portal_id"), id)
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	if data.Publish.ValueBool() {
		if err := publishPortal(ctx, conn, id, data.PublishDescription.ValueStringPointer(), out.LastPublished, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
			// Taint the resource.
			response.State.SetAttribute(ctx, path.Root("portal_id"), id)
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	portal, err := findPortalByID(ctx, conn, id)
	if err != nil {
		response.State.SetAttribute(ctx, path.Root("portal_id"), id)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, portal, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *portalResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data portalResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.PortalID)
	out, err := findPortalByID(ctx, conn, id)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *portalResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old portalResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	// Publishing is handled separately from the portal's configuration.
	diff, d := fwflex.Diff(ctx, new, old, fwflex.WithIgnoredField("Publish"), fwflex.WithIgnoredField("PublishDescription"))
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	id := fwflex.StringValueFromFramework(ctx, new.PortalID)
	var lastPublished *time.Time
	if v, d := old.LastPublished.ValueRFC3339Time(); !d.HasError() && !old.LastPublished.IsNull() {
		lastPublished = &v
	}

	if diff.HasChanges() {
		var input apigatewayv2.UpdatePortalInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// UpdatePortal is a PATCH: omitted fields are left unchanged, so only send the
		// blocks that changed and clear removed optional arguments explicitly.
		changed := diff.ChangedFieldNames()
		if !slices.Contains(changed, "Authorization") {
			input.Authorization = nil
		}
		if !slices.Contains(changed, "EndpointConfiguration") {
			input.EndpointConfiguration = nil
		}
		if !slices.Contains(changed, "PortalContent") {
			input.PortalContent = nil
		}
		if new.IncludedPortalProductARNs.IsNull() && !old.IncludedPortalProductARNs.IsNull() {
			input.IncludedPortalProductArns = []string{}
		}
		if new.LogoURI.IsNull() && !old.LogoURI.IsNull() {
			input.LogoUri = aws.String("")
		}
		if new.RumAppMonitorName.IsNull() && !old.RumAppMonitorName.IsNull() {
			input.RumAppMonitorName = aws.String("")
		}

		_, err := conn.UpdatePortal(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	publish, wasPublished := new.Publish.ValueBool(), old.Publish.ValueBool()
	switch {
	case publish && (!wasPublished || diff.HasChanges() || !new.PublishDescription.Equal(old.PublishDescription)):
		// Changes to a published portal are not live until the portal is published again.
		if err := publishPortal(ctx, conn, id, new.PublishDescription.ValueStringPointer(), lastPublished, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}
	case !publish && wasPublished:
		if err := disablePortal(ctx, conn, id, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}
	case !diff.HasChanges():
		// Nothing changed in AWS, e.g. a tags-only update.
		new.LastModified = old.LastModified
		new.LastPublished = old.LastPublished
		new.LastPublishedDescription = old.LastPublishedDescription
		new.PublishStatus = old.PublishStatus
		smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
		return
	}

	out, err := findPortalByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &new))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *portalResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data portalResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayV2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.PortalID)

	// A portal must be disabled before it can be deleted.
	out, err := findPortalByID(ctx, conn, id)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if out.PublishStatus != awstypes.PublishStatusDisabled {
		err := disablePortal(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts))
		if retry.NotFound(err) {
			return
		}
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	input := apigatewayv2.DeletePortalInput{
		PortalId: aws.String(id),
	}
	_, err = conn.DeletePortal(ctx, &input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if _, err := waitPortalDeleted(ctx, conn, id, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

// flatten is shared by Create, Read, Update and the list resource.
func (r *portalResource) flatten(ctx context.Context, portal *apigatewayv2.GetPortalOutput, data *portalResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, portal, data)...)
	if diags.HasError() {
		return diags
	}

	// The response endpoint configuration is flat; these values are returned for every endpoint type.
	if v := portal.EndpointConfiguration; v != nil {
		data.PortalDefaultDomainName = fwflex.StringToFramework(ctx, v.PortalDefaultDomainName)
		data.PortalDomainHostedZoneID = fwflex.StringToFramework(ctx, v.PortalDomainHostedZoneId)
	}

	data.Publish = types.BoolValue(portalIsPublished(portal.PublishStatus))

	setTagsOut(ctx, portal.Tags)

	return diags
}

// portalProductARNInAccount reports whether a portal product ARN belongs to the specified account.
func portalProductARNInAccount(v, accountID string) bool {
	parsed, err := arn.Parse(v)
	if err != nil {
		return true
	}

	return parsed.AccountID == accountID
}

// portalIsPublished reports whether the portal is, or is becoming, publicly reachable.
// A failed disable leaves the portal published, and a failed publish leaves it disabled.
func portalIsPublished(status awstypes.PublishStatus) bool {
	switch status {
	case awstypes.PublishStatusPublished, awstypes.PublishStatusPublishInProgress, awstypes.PublishStatusDisableFailed:
		return true
	default:
		return false
	}
}

func publishPortal(ctx context.Context, conn *apigatewayv2.Client, id string, description *string, lastPublished *time.Time, timeout time.Duration) error {
	input := apigatewayv2.PublishPortalInput{
		Description: description,
		PortalId:    aws.String(id),
	}
	if _, err := conn.PublishPortal(ctx, &input); err != nil {
		return smarterr.NewError(err)
	}

	if _, err := waitPortalPublished(ctx, conn, id, lastPublished, timeout); err != nil {
		return smarterr.NewError(err)
	}

	return nil
}

func disablePortal(ctx context.Context, conn *apigatewayv2.Client, id string, timeout time.Duration) error {
	input := apigatewayv2.DisablePortalInput{
		PortalId: aws.String(id),
	}
	if _, err := conn.DisablePortal(ctx, &input); err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		return smarterr.NewError(err)
	}

	if _, err := waitPortalDisabled(ctx, conn, id, timeout); err != nil {
		return smarterr.NewError(err)
	}

	return nil
}

// Waiters.

// portalPublishStatusStale is a pseudo-status for a portal that still reports the
// previous publication while a new PublishPortal request is being processed.
const portalPublishStatusStale = "STALE"

// portalPublishStatusDeleted is returned by GetPortal for a short time after DeletePortal.
// It is not part of the PublishStatus enum.
const portalPublishStatusDeleted awstypes.PublishStatus = "DELETED"

// PublishPortal and DisablePortal both transition out of the other operation's
// statuses, including its failure status, so every status except this operation's
// own failure is pending. The failure status is deliberately left out of both
// lists so the wait ends with an UnexpectedStateError carrying StatusException.
func waitPortalPublished(ctx context.Context, conn *apigatewayv2.Client, id string, lastPublished *time.Time, timeout time.Duration) (*apigatewayv2.GetPortalOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: append(enum.Slice(
			awstypes.PublishStatusPublishInProgress,
			awstypes.PublishStatusDisabled,
			awstypes.PublishStatusDisableInProgress,
			awstypes.PublishStatusDisableFailed,
		), portalPublishStatusStale),
		Target:                    enum.Slice(awstypes.PublishStatusPublished),
		Refresh:                   statusPortalPublished(conn, id, lastPublished),
		Timeout:                   timeout,
		Delay:                     5 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*apigatewayv2.GetPortalOutput); ok {
		retry.SetLastError(err, portalStatusExceptionError(out.StatusException))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPortalDisabled(ctx context.Context, conn *apigatewayv2.Client, id string, timeout time.Duration) (*apigatewayv2.GetPortalOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.PublishStatusDisableInProgress,
			awstypes.PublishStatusPublished,
			awstypes.PublishStatusPublishInProgress,
			awstypes.PublishStatusPublishFailed,
		),
		Target:                    enum.Slice(awstypes.PublishStatusDisabled),
		Refresh:                   statusPortal(conn, id),
		Timeout:                   timeout,
		Delay:                     5 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*apigatewayv2.GetPortalOutput); ok {
		retry.SetLastError(err, portalStatusExceptionError(out.StatusException))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPortalDeleted(ctx context.Context, conn *apigatewayv2.Client, id string, timeout time.Duration) (*apigatewayv2.GetPortalOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PublishStatusDisabled),
		Target:  []string{},
		Refresh: statusPortal(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*apigatewayv2.GetPortalOutput); ok {
		retry.SetLastError(err, portalStatusExceptionError(out.StatusException))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusPortal(conn *apigatewayv2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findPortalByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.PublishStatus), nil
	}
}

// statusPortalPublished treats PUBLISHED as pending until the portal reports a
// publication newer than lastPublished, so that re-publishing an already published
// portal waits for the new publication to complete.
func statusPortalPublished(conn *apigatewayv2.Client, id string, lastPublished *time.Time) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findPortalByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		if out.PublishStatus == awstypes.PublishStatusPublished && lastPublished != nil {
			if v := out.LastPublished; v == nil || !v.After(*lastPublished) {
				return out, portalPublishStatusStale, nil
			}
		}

		return out, string(out.PublishStatus), nil
	}
}

func portalStatusExceptionError(apiObject *awstypes.StatusException) error {
	if apiObject == nil {
		return nil
	}

	return fmt.Errorf("%s: %s", aws.ToString(apiObject.Exception), aws.ToString(apiObject.Message))
}

// Finders.

func findPortalByID(ctx context.Context, conn *apigatewayv2.Client, id string) (*apigatewayv2.GetPortalOutput, error) {
	input := apigatewayv2.GetPortalInput{
		PortalId: aws.String(id),
	}

	return findPortal(ctx, conn, &input)
}

func findPortal(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetPortalInput) (*apigatewayv2.GetPortalOutput, error) {
	out, err := conn.GetPortal(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	// A deleted portal remains visible for a short time with an undocumented status.
	if out.PublishStatus == portalPublishStatusDeleted {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: string(out.PublishStatus),
		})
	}

	return out, nil
}

// Models.

type portalResourceModel struct {
	framework.WithRegionModel
	Authorization             fwtypes.ListNestedObjectValueOf[portalAuthorizationModel]         `tfsdk:"authorization"`
	EndpointConfiguration     fwtypes.ListNestedObjectValueOf[portalEndpointConfigurationModel] `tfsdk:"endpoint_configuration"`
	IncludedPortalProductARNs fwtypes.SetOfARN                                                  `tfsdk:"included_portal_product_arns" autoflex:",omitempty"`
	LastModified              timetypes.RFC3339                                                 `tfsdk:"last_modified"`
	LastPublished             timetypes.RFC3339                                                 `tfsdk:"last_published"`
	LastPublishedDescription  types.String                                                      `tfsdk:"last_published_description"`
	LogoURI                   types.String                                                      `tfsdk:"logo_uri"`
	PortalARN                 types.String                                                      `tfsdk:"portal_arn"`
	PortalContent             fwtypes.ListNestedObjectValueOf[portalContentModel]               `tfsdk:"portal_content"`
	PortalDefaultDomainName   types.String                                                      `tfsdk:"portal_default_domain_name" autoflex:"-"`
	PortalDomainHostedZoneID  types.String                                                      `tfsdk:"portal_domain_hosted_zone_id" autoflex:"-"`
	PortalID                  types.String                                                      `tfsdk:"portal_id"`
	Publish                   types.Bool                                                        `tfsdk:"publish" autoflex:"-"`
	PublishDescription        types.String                                                      `tfsdk:"publish_description" autoflex:"-"`
	PublishStatus             fwtypes.StringEnum[awstypes.PublishStatus]                        `tfsdk:"publish_status"`
	RumAppMonitorName         types.String                                                      `tfsdk:"rum_app_monitor_name" autoflex:",omitempty"`
	Tags                      tftags.Map                                                        `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                        `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                    `tfsdk:"timeouts"`
}

type portalAuthorizationModel struct {
	CognitoConfig fwtypes.ListNestedObjectValueOf[portalCognitoConfigModel] `tfsdk:"cognito_config"`
	None          fwtypes.ListNestedObjectValueOf[portalNoneModel]          `tfsdk:"none"`
}

type portalCognitoConfigModel struct {
	AppClientID    types.String `tfsdk:"app_client_id"`
	UserPoolARN    fwtypes.ARN  `tfsdk:"user_pool_arn"`
	UserPoolDomain types.String `tfsdk:"user_pool_domain"`
}

type portalNoneModel struct{}

type portalEndpointConfigurationModel struct {
	ACMManaged fwtypes.ListNestedObjectValueOf[portalACMManagedModel] `tfsdk:"acm_managed"`
	None       fwtypes.ListNestedObjectValueOf[portalNoneModel]       `tfsdk:"none"`
}

var _ fwflex.Flattener = &portalEndpointConfigurationModel{}

// Flatten rebuilds the request-shaped union from the flat EndpointConfigurationResponse.
// The certificate and domain name are only returned for an ACM-managed endpoint.
func (m *portalEndpointConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	var apiObject awstypes.EndpointConfigurationResponse
	switch t := v.(type) {
	case awstypes.EndpointConfigurationResponse:
		apiObject = t
	case *awstypes.EndpointConfigurationResponse:
		if t != nil {
			apiObject = *t
		}
	default:
		diags.AddError("Interface Conversion Error", fmt.Sprintf("cannot flatten %T into %T", v, m))
		return diags
	}

	m.ACMManaged = fwtypes.NewListNestedObjectValueOfNull[portalACMManagedModel](ctx)
	m.None = fwtypes.NewListNestedObjectValueOfNull[portalNoneModel](ctx)

	switch hasCertificate, hasDomainName := apiObject.CertificateArn != nil, apiObject.DomainName != nil; {
	case hasCertificate && hasDomainName:
		m.ACMManaged = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &portalACMManagedModel{
			CertificateARN: fwtypes.ARNValue(aws.ToString(apiObject.CertificateArn)),
			DomainName:     fwflex.StringToFramework(ctx, apiObject.DomainName),
		})
	case !hasCertificate && !hasDomainName:
		m.None = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &portalNoneModel{})
	default:
		// certificate_arn and domain_name are both required for an ACM-managed
		// endpoint. Writing a half-populated block would put an empty, unparseable
		// ARN into state and produce a permanent diff, so report the response instead.
		missing := names.AttrCertificateARN
		if hasCertificate {
			missing = names.AttrDomainName
		}
		diags.AddError(
			"Unexpected Endpoint Configuration",
			fmt.Sprintf("The portal endpoint configuration is missing %s. This is always returned for an ACM-managed endpoint.", missing),
		)
	}

	return diags
}

type portalACMManagedModel struct {
	CertificateARN fwtypes.ARN  `tfsdk:"certificate_arn"`
	DomainName     types.String `tfsdk:"domain_name"`
}

type portalContentModel struct {
	Description types.String                                      `tfsdk:"description" autoflex:",omitempty"`
	DisplayName types.String                                      `tfsdk:"display_name"`
	Theme       fwtypes.ListNestedObjectValueOf[portalThemeModel] `tfsdk:"theme"`
}

type portalThemeModel struct {
	CustomColors fwtypes.ListNestedObjectValueOf[portalCustomColorsModel] `tfsdk:"custom_colors"`
}

type portalCustomColorsModel struct {
	AccentColor          types.String `tfsdk:"accent_color"`
	BackgroundColor      types.String `tfsdk:"background_color"`
	ErrorValidationColor types.String `tfsdk:"error_validation_color"`
	HeaderColor          types.String `tfsdk:"header_color"`
	NavigationColor      types.String `tfsdk:"navigation_color"`
	TextColor            types.String `tfsdk:"text_color"`
}
