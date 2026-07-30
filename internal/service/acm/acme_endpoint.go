// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acm

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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

// @FrameworkResource("aws_acm_acme_endpoint", name="ACME Endpoint")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/acm/types;types.AcmeEndpoint")
// @Testing(generator=false)
// @Testing(hasNoPreExistingResource=true)
func newACMEEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &acmeEndpointResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameACMEEndpoint = "ACME Endpoint"

	// certificateTagsFieldName is handled outside of AutoFlex as the AWS API models it
	// as a tag slice rather than a map.
	certificateTagsFieldName = "CertificateTags"
)

type acmeEndpointResource struct {
	framework.ResourceWithModel[acmeEndpointResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *acmeEndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"authorization_behavior": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AcmeAuthorizationBehavior](),
				Required:   true,
			},
			"certificate_tags": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"contact": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AcmeContact](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"failure_reason": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AcmeEndpointStatus](),
				Computed:   true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"certificate_authority": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[acmeCertificateAuthorityModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"public_certificate_authority": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[acmePublicCertificateAuthorityModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"allowed_key_algorithms": schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringEnumType[awstypes.PublicKeyAlgorithm](),
										Optional:    true,
										Computed:    true,
										ElementType: fwtypes.StringEnumType[awstypes.PublicKeyAlgorithm](),
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.UseStateForUnknown(),
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

func (r *acmeEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ACMClient(ctx)

	var plan acmeEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input acm.CreateAcmeEndpointInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("AcmeEndpoint"), fwflex.WithIgnoredFieldNamesAppend(certificateTagsFieldName)))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields not covered by AutoFlex.
	input.IdempotencyToken = aws.String(create.UniqueId(ctx))
	input.CertificateTags = expandCertificateTags(ctx, plan.CertificateTags)

	// An ACME endpoint has no user-supplied name, so there is no identifier to report until the ARN comes back.
	out, err := conn.CreateAcmeEndpoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	arn := aws.ToString(out.AcmeEndpointArn)
	plan.ARN = fwflex.StringValueToFramework(ctx, arn)

	endpoint, err := waitACMEEndpointActive(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.State.SetAttribute(ctx, path.Root(names.AttrARN), plan.ARN)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, endpoint, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *acmeEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ACMClient(ctx)

	var state acmeEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findACMEEndpointByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *acmeEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ACMClient(ctx)

	var plan, state acmeEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input acm.UpdateAcmeEndpointInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("AcmeEndpoint"), fwflex.WithIgnoredFieldNamesAppend(certificateTagsFieldName)))
		if resp.Diagnostics.HasError() {
			return
		}

		if _, err := conn.UpdateAcmeEndpoint(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
			return
		}

		endpoint, err := waitACMEEndpointActive(ctx, conn, plan.ARN.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, endpoint, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *acmeEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ACMClient(ctx)

	var state acmeEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := acm.DeleteAcmeEndpointInput{
		AcmeEndpointArn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteAcmeEndpoint(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	if _, err := waitACMEEndpointDeleted(ctx, conn, state.ARN.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
}

// flatten maps an AcmeEndpoint onto the resource model. certificate_tags is handled
// here because the AWS API returns a tag slice where the schema exposes a map.
func (r *acmeEndpointResource) flatten(ctx context.Context, endpoint *awstypes.AcmeEndpoint, data *acmeEndpointResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, endpoint, data, fwflex.WithFieldNamePrefix("AcmeEndpoint"), fwflex.WithIgnoredFieldNamesAppend(certificateTagsFieldName))...)
	if diags.HasError() {
		return diags
	}

	data.CertificateTags = flattenCertificateTags(ctx, endpoint.CertificateTags)

	return diags
}

func expandCertificateTags(ctx context.Context, tags fwtypes.MapOfString) []awstypes.Tag {
	if tags.IsNull() || tags.IsUnknown() {
		return nil
	}

	return svcTags(tftags.New(ctx, fwflex.ExpandFrameworkStringValueMap(ctx, tags)))
}

func flattenCertificateTags(ctx context.Context, tags []awstypes.Tag) fwtypes.MapOfString {
	if len(tags) == 0 {
		return fwtypes.NewMapValueOfNull[types.String](ctx)
	}

	return fwflex.FlattenFrameworkStringValueMapOfString(ctx, keyValueTags(ctx, tags).Map())
}

func findACMEEndpointByARN(ctx context.Context, conn *acm.Client, arn string) (*awstypes.AcmeEndpoint, error) {
	input := acm.DescribeAcmeEndpointInput{
		AcmeEndpointArn: aws.String(arn),
	}

	out, err := conn.DescribeAcmeEndpoint(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.AcmeEndpoint == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.AcmeEndpoint, nil
}

func statusACMEEndpoint(conn *acm.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findACMEEndpointByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func waitACMEEndpointActive(ctx context.Context, conn *acm.Client, arn string, timeout time.Duration) (*awstypes.AcmeEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AcmeEndpointStatusCreating),
		Target:                    enum.Slice(awstypes.AcmeEndpointStatusActive),
		Refresh:                   statusACMEEndpoint(conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AcmeEndpoint); ok {
		if out.Status == awstypes.AcmeEndpointStatusFailed {
			retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		}

		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitACMEEndpointDeleted(ctx context.Context, conn *acm.Client, arn string, timeout time.Duration) (*awstypes.AcmeEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AcmeEndpointStatusActive, awstypes.AcmeEndpointStatusDeleting),
		Target:  []string{},
		Refresh: statusACMEEndpoint(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.AcmeEndpoint); ok {
		if out.Status == awstypes.AcmeEndpointStatusFailed {
			retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		}

		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type acmeEndpointResourceModel struct {
	framework.WithRegionModel
	ARN                   types.String                                                   `tfsdk:"arn"`
	AuthorizationBehavior fwtypes.StringEnum[awstypes.AcmeAuthorizationBehavior]         `tfsdk:"authorization_behavior"`
	CertificateAuthority  fwtypes.ListNestedObjectValueOf[acmeCertificateAuthorityModel] `tfsdk:"certificate_authority"`
	CertificateTags       fwtypes.MapOfString                                            `tfsdk:"certificate_tags"`
	Contact               fwtypes.StringEnum[awstypes.AcmeContact]                       `tfsdk:"contact"`
	CreatedAt             timetypes.RFC3339                                              `tfsdk:"created_at"`
	EndpointURL           types.String                                                   `tfsdk:"endpoint_url"`
	FailureReason         types.String                                                   `tfsdk:"failure_reason"`
	Status                fwtypes.StringEnum[awstypes.AcmeEndpointStatus]                `tfsdk:"status"`
	Timeouts              timeouts.Value                                                 `tfsdk:"timeouts"`
	UpdatedAt             timetypes.RFC3339                                              `tfsdk:"updated_at"`
}

var (
	_ fwflex.Expander  = acmeCertificateAuthorityModel{}
	_ fwflex.Flattener = &acmeCertificateAuthorityModel{}
)

type acmeCertificateAuthorityModel struct {
	PublicCertificateAuthority fwtypes.ListNestedObjectValueOf[acmePublicCertificateAuthorityModel] `tfsdk:"public_certificate_authority"`
}

func (m acmeCertificateAuthorityModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.PublicCertificateAuthority.IsNull():
		data, d := m.PublicCertificateAuthority.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CertificateAuthorityMemberPublicCertificateAuthority
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

func (m *acmeCertificateAuthorityModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.CertificateAuthorityMemberPublicCertificateAuthority:
		var data acmePublicCertificateAuthorityModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}

		m.PublicCertificateAuthority = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	}

	return diags
}

type acmePublicCertificateAuthorityModel struct {
	AllowedKeyAlgorithms fwtypes.SetOfStringEnum[awstypes.PublicKeyAlgorithm] `tfsdk:"allowed_key_algorithms"`
}
