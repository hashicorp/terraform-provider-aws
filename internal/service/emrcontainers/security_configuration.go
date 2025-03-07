// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrcontainers

import (
	"context"
	"errors"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"

	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emrcontainers"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emrcontainers/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_emrcontainers_security_configuration", name="Security Configuration")
// @Tags(identifierAttribute="arn")
func newResourceSecurityConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceSecurityConfiguration{}, nil
}

const (
	ResNameSecurityConfiguration = "Security Configuration"
)

type resourceSecurityConfiguration struct {
	framework.ResourceWithConfigure
}

func (r *resourceSecurityConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	optionalBlock := []validator.List{
		listvalidator.SizeAtMost(1),
	}
	requiredBlock := []validator.List{
		listvalidator.SizeAtLeast(1),
		listvalidator.SizeAtMost(1),
	}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("[.-_/#A-Za-z0-9]+"), ""),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"security_configuration_data": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[securityConfigurationDataModel](ctx),
				Validators: requiredBlock,
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"authorization_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authorizationConfigurationModel](ctx),
							Validators: optionalBlock,
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"lake_formation_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lakeFormationConfigurationModel](ctx),
										Validators: optionalBlock,
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"authorized_session_tag_value": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthAtLeast(1),
														stringvalidator.LengthAtMost(512),
													},
												},
												"query_engine_role_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
												},
											},
											Blocks: map[string]schema.Block{
												"secure_namespace_info": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[secureNamespaceInfoModel](ctx),
													Validators: requiredBlock,
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"cluster_id": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthAtLeast(1),
																	stringvalidator.LengthAtMost(100),
																},
															},
															"namespace": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthAtLeast(1),
																	stringvalidator.LengthAtMost(63),
																},
															},
														},
													},
												},
											},
										},
									},
									"encryption_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationModel](ctx),
										Validators: optionalBlock,
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"in_transit_encryption_configuration": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[inTransitEncryptionConfigurationModel](ctx),
													Validators: requiredBlock,
													NestedObject: schema.NestedBlockObject{
														Blocks: map[string]schema.Block{
															"tls_certificate_configuration": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[tlsCertificateConfigurationModel](ctx),
																Validators: requiredBlock,
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"certificate_provider_type": schema.StringAttribute{
																			Required: true,
																		},
																		"private_certificate_secret_arn": schema.StringAttribute{
																			CustomType: fwtypes.ARNType,
																			Required:   true,
																		},
																		"public_certificate_secret_arn": schema.StringAttribute{
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

func (r *resourceSecurityConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	conn := r.Meta().EMRContainersClient(ctx)

	var plan resourceSecurityConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input emrcontainers.CreateSecurityConfigurationInput

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("SecurityConfiguration"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateSecurityConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EMRContainers, create.ErrActionCreating, ResNameSecurityConfiguration, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Id == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EMRContainers, create.ErrActionCreating, ResNameSecurityConfiguration, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSecurityConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	conn := r.Meta().EMRContainersClient(ctx)

	var state resourceSecurityConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSecurityConfigurationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EMRContainers, create.ErrActionSetting, ResNameSecurityConfiguration, state.ID.String(), err),
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

func (r *resourceSecurityConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var new resourceSecurityConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Tags only.

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceSecurityConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Noop (not implemented by AWS)
	return
}

func (r *resourceSecurityConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findSecurityConfigurationByID(ctx context.Context, conn *emrcontainers.Client, id string) (*awstypes.SecurityConfiguration, error) {
	input := emrcontainers.DescribeSecurityConfigurationInput{
		Id: aws.String(id),
	}

	out, err := conn.DescribeSecurityConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.SecurityConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.SecurityConfiguration, nil
}

type resourceSecurityConfigurationModel struct {
	ARN                       types.String                                                    `tfsdk:"arn"`
	ID                        types.String                                                    `tfsdk:"id"`
	Name                      types.String                                                    `tfsdk:"name"`
	SecurityConfigurationData fwtypes.ListNestedObjectValueOf[securityConfigurationDataModel] `tfsdk:"security_configuration_data"`
	Tags                      tftags.Map                                                      `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                      `tfsdk:"tags_all"`
}

type securityConfigurationDataModel struct {
	AuthorizationConfiguration fwtypes.ListNestedObjectValueOf[authorizationConfigurationModel] `tfsdk:"authorization_configuration"`
}

type authorizationConfigurationModel struct {
	EncryptionConfiguration    fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel]    `tfsdk:"encryption_configuration"`
	LakeFormationConfiguration fwtypes.ListNestedObjectValueOf[lakeFormationConfigurationModel] `tfsdk:"lake_formation_configuration"`
}
type encryptionConfigurationModel struct {
	InTransitEncryptionConfiguration fwtypes.ListNestedObjectValueOf[inTransitEncryptionConfigurationModel] `tfsdk:"in_transit_encryption_configuration"`
}

type inTransitEncryptionConfigurationModel struct {
	TlsCertificateConfiguration fwtypes.ListNestedObjectValueOf[tlsCertificateConfigurationModel] `tfsdk:"tls_certificate_configuration"`
}

type tlsCertificateConfigurationModel struct {
	CertificateProviderType     fwtypes.StringEnum[awstypes.CertificateProviderType] `tfsdk:"certificate_provider_type"`
	PrivateCertificateSecretArn fwtypes.ARN                                          `tfsdk:"private_certificate_secret_arn"`
	PublicCertificateSecretArn  fwtypes.ARN                                          `tfsdk:"public_certificate_secret_arn"`
}
type lakeFormationConfigurationModel struct {
	AuthorizedSessionTagValue types.String                                              `tfsdk:"authorized_session_tag_value"`
	QueryEngineRoleArn        fwtypes.ARN                                               `tfsdk:"query_engine_role_arn"`
	SecureNamespaceInfo       fwtypes.ListNestedObjectValueOf[secureNamespaceInfoModel] `tfsdk:"secure_namespace_info"`
}

type secureNamespaceInfoModel struct {
	ClusterId types.String `tfsdk:"cluster_id"`
	Namespace types.String `tfsdk:"namespace"`
}
