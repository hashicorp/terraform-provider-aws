// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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

// @FrameworkResource("aws_bedrockagentcore_browser", name="Browser")
// @Tags(identifierAttribute="browser_arn")
// @Testing(tagsTest=false)
func newBrowserResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &browserResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type browserResource struct {
	framework.ResourceWithModel[browserResourceModel]
	framework.WithTimeouts
}

func (r *browserResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"browser_arn": framework.ARNAttributeComputedOnly(),
			"browser_id":  framework.IDAttribute(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrExecutionRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"browser_signing": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[browserSigningConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrEnabled: schema.BoolAttribute{
							Required: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			names.AttrCertificate: certificateSchema(ctx),
			"enterprise_policy": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[browserEnterprisePolicyModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(100),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.BrowserEnterprisePolicyType](),
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrLocation: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[resourceLocationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"s3": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[enterprisePolicyS3LocationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												// If another member is added to the union, this will need to be updated.
												path.MatchRelative().AtParent().AtName("s3"),
											),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucket: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`), "must be a valid S3 bucket name"),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												names.AttrPrefix: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 1024),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"version_id": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(3, 1024),
													},
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
			names.AttrNetworkConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[browserNetworkConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"network_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.BrowserNetworkMode](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrVPCConfig: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigNoS3EndpointModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrSecurityGroups: schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
										},
									},
									names.AttrSubnets: schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Required:   true,
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
			"recording": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[recordingConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrEnabled: schema.BoolAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"s3_location": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3LocationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucket: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`), "must be a valid S3 bucket name"),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									names.AttrPrefix: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
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
				Delete: true,
			}),
		},
	}
}

func certificateSchema(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[certificateModel](ctx),
		// SizeAtLeast(1) enforces the SDK's @length(min: 1) only when the
		// user provides the block; the validator skips null/unknown, so
		// omitting `certificates` entirely is still valid (block is optional).
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(200),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.RequiresReplace(),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				names.AttrLocation: schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[certificateLocationModel](ctx),
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(1),
					},
					PlanModifiers: []planmodifier.List{
						listplanmodifier.RequiresReplace(),
					},
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"secrets_manager": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[secretsManagerLocationModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
									listvalidator.ExactlyOneOf(
										// If another member is added to the union, this will need to be updated.
										path.MatchRelative().AtParent().AtName("secrets_manager"),
									),
								},
								PlanModifiers: []planmodifier.List{
									listplanmodifier.RequiresReplace(),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"secret_arn": schema.StringAttribute{
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
	}
}

func (r *browserResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data browserResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateBrowserInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	var (
		out *bedrockagentcorecontrol.CreateBrowserOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateBrowser(ctx, &input)

		// IAM propagation.
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Browser role validation failed") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	browserID := aws.ToString(out.BrowserId)

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := waitBrowserCreated(ctx, conn, browserID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root("browser_id"), browserID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, browserID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *browserResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data browserResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	browserID := fwflex.StringValueFromFramework(ctx, data.BrowserID)
	out, err := findBrowserByID(ctx, conn, browserID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, browserID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *browserResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data browserResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	browserID := fwflex.StringValueFromFramework(ctx, data.BrowserID)
	input := bedrockagentcorecontrol.DeleteBrowserInput{
		BrowserId: aws.String(browserID),
	}

	_, err := conn.DeleteBrowser(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, browserID)
		return
	}

	if _, err := waitBrowserDeleted(ctx, conn, browserID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, browserID)
		return
	}
}

func (r *browserResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("browser_id"), request, response)
}

func waitBrowserCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetBrowserOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BrowserStatusCreating),
		Target:                    enum.Slice(awstypes.BrowserStatusReady),
		Refresh:                   statusBrowser(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetBrowserOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitBrowserDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetBrowserOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BrowserStatusDeleting, awstypes.BrowserStatusReady),
		Target:  []string{},
		Refresh: statusBrowser(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetBrowserOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusBrowser(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findBrowserByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findBrowserByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetBrowserOutput, error) {
	input := bedrockagentcorecontrol.GetBrowserInput{
		BrowserId: aws.String(id),
	}

	return findBrowser(ctx, conn, &input)
}

func findBrowser(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetBrowserInput) (*bedrockagentcorecontrol.GetBrowserOutput, error) {
	out, err := conn.GetBrowser(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

	return out, nil
}

type browserResourceModel struct {
	framework.WithRegionModel
	BrowserARN           types.String                                                      `tfsdk:"browser_arn"`
	BrowserID            types.String                                                      `tfsdk:"browser_id"`
	BrowserSigning       fwtypes.ListNestedObjectValueOf[browserSigningConfigModel]        `tfsdk:"browser_signing"`
	Certificates         fwtypes.ListNestedObjectValueOf[certificateModel]                 `tfsdk:"certificate"`
	Description          types.String                                                      `tfsdk:"description"`
	EnterprisePolicies   fwtypes.ListNestedObjectValueOf[browserEnterprisePolicyModel]     `tfsdk:"enterprise_policy"`
	ExecutionRoleARN     fwtypes.ARN                                                       `tfsdk:"execution_role_arn"`
	Name                 types.String                                                      `tfsdk:"name"`
	NetworkConfiguration fwtypes.ListNestedObjectValueOf[browserNetworkConfigurationModel] `tfsdk:"network_configuration"`
	Recording            fwtypes.ListNestedObjectValueOf[recordingConfigModel]             `tfsdk:"recording"`
	Tags                 tftags.Map                                                        `tfsdk:"tags"`
	TagsAll              tftags.Map                                                        `tfsdk:"tags_all"`
	Timeouts             timeouts.Value                                                    `tfsdk:"timeouts"`
}

type browserNetworkConfigurationModel struct {
	NetworkMode fwtypes.StringEnum[awstypes.BrowserNetworkMode]             `tfsdk:"network_mode"`
	VPCConfig   fwtypes.ListNestedObjectValueOf[vpcConfigNoS3EndpointModel] `tfsdk:"vpc_config"`
}

type recordingConfigModel struct {
	Enabled    types.Bool                                       `tfsdk:"enabled"`
	S3Location fwtypes.ListNestedObjectValueOf[s3LocationModel] `tfsdk:"s3_location"`
}

type s3LocationModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Prefix types.String `tfsdk:"prefix"`
}

type browserSigningConfigModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type browserEnterprisePolicyModel struct {
	Location fwtypes.ListNestedObjectValueOf[resourceLocationModel]   `tfsdk:"location"`
	Type     fwtypes.StringEnum[awstypes.BrowserEnterprisePolicyType] `tfsdk:"type"`
}

type resourceLocationModel struct {
	S3 fwtypes.ListNestedObjectValueOf[enterprisePolicyS3LocationModel] `tfsdk:"s3"`
}

type enterprisePolicyS3LocationModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	Prefix    types.String `tfsdk:"prefix"`
	VersionID types.String `tfsdk:"version_id"`
}

var (
	_ fwflex.Expander  = resourceLocationModel{}
	_ fwflex.Flattener = &resourceLocationModel{}
)

func (m *resourceLocationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.ResourceLocationMemberS3:
		var data enterprisePolicyS3LocationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("resource location flatten: %T", v),
		)
	}
	return diags
}

func (m resourceLocationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.S3.IsNull():
		data, d := m.S3.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ResourceLocationMemberS3
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type certificateModel struct {
	Location fwtypes.ListNestedObjectValueOf[certificateLocationModel] `tfsdk:"location"`
}

type certificateLocationModel struct {
	SecretsManager fwtypes.ListNestedObjectValueOf[secretsManagerLocationModel] `tfsdk:"secrets_manager"`
}

type secretsManagerLocationModel struct {
	SecretARN fwtypes.ARN `tfsdk:"secret_arn"`
}

var (
	_ fwflex.Expander  = certificateLocationModel{}
	_ fwflex.Flattener = &certificateLocationModel{}
)

func (m *certificateLocationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.CertificateLocationMemberSecretsManager:
		var data secretsManagerLocationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.SecretsManager = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("certificate location flatten: %T", v),
		)
	}
	return diags
}

func (m certificateLocationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.SecretsManager.IsNull():
		data, d := m.SecretsManager.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.CertificateLocationMemberSecretsManager
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}
