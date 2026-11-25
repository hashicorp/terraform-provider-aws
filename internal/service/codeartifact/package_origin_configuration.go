// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codeartifact

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	awstypes "github.com/aws/aws-sdk-go-v2/service/codeartifact/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_codeartifact_package_origin_configuration", name="Package Origin Configuration")
// @IdentityAttribute("domain")
// @IdentityAttribute("repository")
// @IdentityAttribute("format")
// @IdentityAttribute("namespace", optional="true")
// @IdentityAttribute("package")
// @ImportIDHandler("packageOriginConfigurationImportID")
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttributes="domain;repository;format;namespace;package", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/codeartifact/types;awstypes;awstypes.PackageOriginConfiguration")
func newPackageOriginConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &packageOriginConfigurationResource{}, nil
}

type packageOriginConfigurationResource struct {
	framework.ResourceWithModel[packageOriginConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *packageOriginConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDomain: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_owner": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrFormat: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PackageFormat](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrNamespace: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"package": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"restrictions": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[restrictionsModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"publish": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AllowPublish](),
							Required:   true,
						},
						"upstream": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AllowUpstream](),
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func (r *packageOriginConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CodeArtifactClient(ctx)

	var plan packageOriginConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.DomainOwner.IsNull() || plan.DomainOwner.IsUnknown() {
		plan.DomainOwner = types.StringValue(r.Meta().AccountID(ctx))
	}

	var input codeartifact.PutPackageOriginConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutPackageOriginConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Package.String())
		return
	}
	if out == nil || out.OriginConfiguration == nil || out.OriginConfiguration.Restrictions == nil {
		smerr.AddError(ctx, &resp.Diagnostics, tfresource.NewEmptyResultError(), smerr.ID, plan.Package.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.OriginConfiguration.Restrictions, &plan.Restrictions))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *packageOriginConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CodeArtifactClient(ctx)

	var state packageOriginConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPackageOriginConfiguration(ctx, conn, state.Domain.ValueString(), state.DomainOwner.ValueString(), state.Repository.ValueString(), state.Format.ValueEnum(), state.Namespace.ValueString(), state.Package.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Package.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.Restrictions, &state.Restrictions))
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DomainOwner.IsNull() {
		state.DomainOwner = types.StringValue(r.Meta().AccountID(ctx))
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *packageOriginConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CodeArtifactClient(ctx)

	var plan packageOriginConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input codeartifact.PutPackageOriginConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutPackageOriginConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Package.String())
		return
	}
	if out == nil || out.OriginConfiguration == nil || out.OriginConfiguration.Restrictions == nil {
		smerr.AddError(ctx, &resp.Diagnostics, tfresource.NewEmptyResultError(), smerr.ID, plan.Package.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.OriginConfiguration.Restrictions, &plan.Restrictions))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

// Package origin configurations cannot be deleted. Deleting this resource
// resets the package's origin restrictions to the CodeArtifact defaults
// (publish and upstream both ALLOW).
func (r *packageOriginConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CodeArtifactClient(ctx)

	var state packageOriginConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := codeartifact.PutPackageOriginConfigurationInput{
		Domain:     state.Domain.ValueStringPointer(),
		Format:     state.Format.ValueEnum(),
		Package:    state.Package.ValueStringPointer(),
		Repository: state.Repository.ValueStringPointer(),
		Restrictions: &awstypes.PackageOriginRestrictions{
			Publish:  awstypes.AllowPublishAllow,
			Upstream: awstypes.AllowUpstreamAllow,
		},
	}
	if !state.DomainOwner.IsNull() {
		input.DomainOwner = state.DomainOwner.ValueStringPointer()
	}
	if !state.Namespace.IsNull() {
		input.Namespace = state.Namespace.ValueStringPointer()
	}

	_, err := conn.PutPackageOriginConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Package.String())
	}
}

func findPackageOriginConfiguration(ctx context.Context, conn *codeartifact.Client, domain, domainOwner, repository string, format awstypes.PackageFormat, namespace, pkg string) (*awstypes.PackageOriginConfiguration, error) {
	input := codeartifact.DescribePackageInput{
		Domain:     aws.String(domain),
		Format:     format,
		Package:    aws.String(pkg),
		Repository: aws.String(repository),
	}
	if domainOwner != "" {
		input.DomainOwner = aws.String(domainOwner)
	}
	if namespace != "" {
		input.Namespace = aws.String(namespace)
	}

	out, err := conn.DescribePackage(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Package == nil || out.Package.OriginConfiguration == nil || out.Package.OriginConfiguration.Restrictions == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Package.OriginConfiguration, nil
}

type packageOriginConfigurationResourceModel struct {
	framework.WithRegionModel
	Domain       types.String                                       `tfsdk:"domain"`
	DomainOwner  types.String                                       `tfsdk:"domain_owner"`
	Format       fwtypes.StringEnum[awstypes.PackageFormat]         `tfsdk:"format"`
	Namespace    types.String                                       `tfsdk:"namespace"`
	Package      types.String                                       `tfsdk:"package"`
	Repository   types.String                                       `tfsdk:"repository"`
	Restrictions fwtypes.ListNestedObjectValueOf[restrictionsModel] `tfsdk:"restrictions"`
}

type restrictionsModel struct {
	Publish  fwtypes.StringEnum[awstypes.AllowPublish]  `tfsdk:"publish"`
	Upstream fwtypes.StringEnum[awstypes.AllowUpstream] `tfsdk:"upstream"`
}

var (
	_ inttypes.ImportIDParser = packageOriginConfigurationImportID{}
)

type packageOriginConfigurationImportID struct{}

func (packageOriginConfigurationImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, intflex.ResourceIdSeparator)
	if len(parts) != 5 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[4] == "" {
		return "", nil, fmt.Errorf("id %q should be in the format <domain>%[2]s<repository>%[2]s<format>%[2]s<namespace>%[2]s<package>, where <namespace> may be empty", id, intflex.ResourceIdSeparator)
	}

	result := map[string]any{
		names.AttrDomain: parts[0],
		"repository":     parts[1],
		names.AttrFormat: parts[2],
		"package":        parts[4],
	}
	if parts[3] != "" {
		result[names.AttrNamespace] = parts[3]
	}

	return id, result, nil
}
