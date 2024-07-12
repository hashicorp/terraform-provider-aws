// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="TLS Inspection Configuration")
// @Tags(identifierAttribute="arn")
func newTLSInspectionConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &tlsInspectionConfigurationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type tlsInspectionConfigurationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (*tlsInspectionConfigurationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_networkfirewall_tls_inspection_configuration"
}

func (r *tlsInspectionConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"certificate_authority": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tlsCertificateDataModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[tlsCertificateDataModel](ctx),
				},
			},
			"certificates": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tlsCertificateDataModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[tlsCertificateDataModel](ctx),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 512),
				},
			},
			names.AttrEncryptionConfiguration: schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[encryptionConfigurationModel](ctx),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9-]+$`), "Must contain only a-z, A-Z, 0-9 and - (hyphen)"),
				},
			},
			"number_of_associations": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:                    tftags.TagsAttribute(),
			names.AttrTagsAll:                 tftags.TagsAttributeComputedOnly(),
			"tls_inspection_configuration_id": framework.IDAttribute(),
			"update_token": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"tls_inspection_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tlsInspectionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"server_certificate_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[serverCertificateConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"certificate_authority_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"check_certificate_revocation_status": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[checkCertificateRevocationStatusActionsModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"revoked_status_action": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.RevocationCheckAction](),
													Optional:   true,
												},
												"unknown_status_action": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.RevocationCheckAction](),
													Optional:   true,
												},
											},
										},
									},
									names.AttrScope: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[serverCertificateScopeModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"protocols": schema.SetAttribute{
													CustomType:  fwtypes.NewSetTypeOf[types.Int64](ctx),
													ElementType: types.Int64Type,
													Required:    true,
													Validators: []validator.Set{
														setvalidator.ValueInt64sAre(int64validator.Between(0, 255)),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"destination_ports": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[portRangeModel](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"from_port": schema.Int64Attribute{
																Required: true,
																Validators: []validator.Int64{
																	int64validator.Between(0, 65535),
																},
															},
															"to_port": schema.Int64Attribute{
																Required: true,
																Validators: []validator.Int64{
																	int64validator.Between(0, 65535),
																},
															},
														},
													},
												},
												names.AttrDestination: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[addressModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"address_definition": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 255),
																	stringvalidator.RegexMatches(regexache.MustCompile(`^([a-fA-F\d:\.]+($|/\d{1,3}))$`), "Must contain IP address or a block of IP addresses in Classless Inter-Domain Routing (CIDR) notation"),
																},
															},
														},
													},
												},
												"source_ports": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[portRangeModel](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"from_port": schema.Int64Attribute{
																Required: true,
																Validators: []validator.Int64{
																	int64validator.Between(0, 65535),
																},
															},
															"to_port": schema.Int64Attribute{
																Required: true,
																Validators: []validator.Int64{
																	int64validator.Between(0, 65535),
																},
															},
														},
													},
												},
												names.AttrSource: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[addressModel](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"address_definition": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 255),
																	stringvalidator.RegexMatches(regexache.MustCompile(`^([a-fA-F\d:\.]+($|/\d{1,3}))$`), "Must contain IP address or a block of IP addresses in Classless Inter-Domain Routing (CIDR) notation"),
																},
															},
														},
													},
												},
											},
										},
									},
									"server_certificate": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[serverCertificateModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrResourceARN: schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Optional:   true,
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

func (r *tlsInspectionConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data tlsInspectionConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFirewallClient(ctx)

	name := data.TLSInspectionConfigurationName.ValueString()
	input := &networkfirewall.CreateTLSInspectionConfigurationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	outputC, err := conn.CreateTLSInspectionConfiguration(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating NetworkFirewall TLS Inspection Configuration (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.TLSInspectionConfigurationARN = fwflex.StringToFramework(ctx, outputC.TLSInspectionConfigurationResponse.TLSInspectionConfigurationArn)
	data.TLSInspectionConfigurationID = fwflex.StringToFramework(ctx, outputC.TLSInspectionConfigurationResponse.TLSInspectionConfigurationId)
	data.UpdateToken = fwflex.StringToFramework(ctx, outputC.UpdateToken)
	data.setID()

	outputR, err := waitTLSInspectionConfigurationCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for NetworkFirewall TLS Inspection Configuration (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(flattenDescribeTLSInspectionConfigurationOutput(ctx, &data, outputR)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *tlsInspectionConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tlsInspectionConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().NetworkFirewallClient(ctx)

	output, err := findTLSInspectionConfigurationByARN(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading NetworkFirewall TLS Inspection Configuration (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(flattenDescribeTLSInspectionConfigurationOutput(ctx, &data, output)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, output.TLSInspectionConfigurationResponse.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *tlsInspectionConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new tlsInspectionConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFirewallClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.EncryptionConfiguration.Equal(old.EncryptionConfiguration) ||
		!new.TLSInspectionConfiguration.Equal(old.TLSInspectionConfiguration) {
		input := &networkfirewall.UpdateTLSInspectionConfigurationInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		input.UpdateToken = aws.String(old.UpdateToken.ValueString())

		output, err := conn.UpdateTLSInspectionConfiguration(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating NetworkFirewall TLS Inspection Configuration (%s)", new.ID.ValueString()), err.Error())

			return
		}

		new.UpdateToken = fwflex.StringToFramework(ctx, output.UpdateToken)

		outputR, err := waitTLSInspectionConfigurationUpdated(ctx, conn, new.ID.ValueString(), r.CreateTimeout(ctx, new.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for NetworkFirewall TLS Inspection Configuration (%s) update", new.ID.ValueString()), err.Error())

			return
		}

		// Set values for unknowns.
		response.Diagnostics.Append(flattenDescribeTLSInspectionConfigurationOutput(ctx, &new, outputR)...)
		if response.Diagnostics.HasError() {
			return
		}
	} else {
		new.CertificateAuthority = old.CertificateAuthority
		new.Certificates = old.Certificates
		new.UpdateToken = old.UpdateToken
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *tlsInspectionConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data tlsInspectionConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFirewallClient(ctx)

	_, err := conn.DeleteTLSInspectionConfiguration(ctx, &networkfirewall.DeleteTLSInspectionConfigurationInput{
		TLSInspectionConfigurationArn: aws.String(data.ID.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting NetworkFirewall TLS Inspection Configuration (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitTLSInspectionConfigurationDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for NetworkFirewall TLS Inspection Configuration (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *tlsInspectionConfigurationResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("tls_inspection_configuration").AtListIndex(0).AtName("server_certificate_configuration").AtListIndex(0).AtName("certificate_authority_arn"),
			path.MatchRoot("tls_inspection_configuration").AtListIndex(0).AtName("server_certificate_configuration").AtListIndex(0).AtName("server_certificate"),
		),
	}
}

func (r *tlsInspectionConfigurationResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findTLSInspectionConfigurationByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeTLSInspectionConfigurationOutput, error) {
	input := &networkfirewall.DescribeTLSInspectionConfigurationInput{
		TLSInspectionConfigurationArn: aws.String(arn),
	}

	output, err := conn.DescribeTLSInspectionConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TLSInspectionConfigurationResponse == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusTLSInspectionConfiguration(ctx context.Context, conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTLSInspectionConfigurationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus), nil
	}
}

const (
	resourceStatusPending = "PENDING"
)

func statusTLSInspectionConfigurationCertificates(ctx context.Context, conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTLSInspectionConfigurationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		certificates := output.TLSInspectionConfigurationResponse.Certificates
		certificateAuthority := output.TLSInspectionConfigurationResponse.CertificateAuthority

		// The API does not immediately return data for certificates and certificate authority even when the resource status is "ACTIVE",
		// which causes unexpected diffs when reading. This sets the status to "PENDING" until either the certificates or the certificate
		// authority is populated (the API will always return at least one of the two).
		if status := output.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus; status == awstypes.ResourceStatusActive && (certificates != nil || certificateAuthority != nil) {
			return output, string(status), nil
		}

		return output, resourceStatusPending, nil
	}
}

func waitTLSInspectionConfigurationCreated(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeTLSInspectionConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{resourceStatusPending},
		Target:  enum.Slice(awstypes.ResourceStatusActive),
		Refresh: statusTLSInspectionConfigurationCertificates(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeTLSInspectionConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}

func waitTLSInspectionConfigurationUpdated(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeTLSInspectionConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{resourceStatusPending},
		Target:  enum.Slice(awstypes.ResourceStatusActive),
		Refresh: statusTLSInspectionConfigurationCertificates(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeTLSInspectionConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}

func waitTLSInspectionConfigurationDeleted(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeTLSInspectionConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusActive, awstypes.ResourceStatusDeleting),
		Target:  []string{},
		Refresh: statusTLSInspectionConfiguration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkfirewall.DescribeTLSInspectionConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}

func flattenDescribeTLSInspectionConfigurationOutput(ctx context.Context, data *tlsInspectionConfigurationResourceModel, apiObject *networkfirewall.DescribeTLSInspectionConfigurationOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	d := fwflex.Flatten(ctx, apiObject.TLSInspectionConfigurationResponse, data)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	d = fwflex.Flatten(ctx, apiObject.TLSInspectionConfiguration, &data.TLSInspectionConfiguration)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	return diags
}

type tlsInspectionConfigurationResourceModel struct {
	CertificateAuthority           fwtypes.ListNestedObjectValueOf[tlsCertificateDataModel]         `tfsdk:"certificate_authority"`
	Certificates                   fwtypes.ListNestedObjectValueOf[tlsCertificateDataModel]         `tfsdk:"certificates"`
	Description                    types.String                                                     `tfsdk:"description"`
	EncryptionConfiguration        fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel]    `tfsdk:"encryption_configuration"`
	ID                             types.String                                                     `tfsdk:"id"`
	NumberOfAssociations           types.Int64                                                      `tfsdk:"number_of_associations"`
	Tags                           types.Map                                                        `tfsdk:"tags"`
	TagsAll                        types.Map                                                        `tfsdk:"tags_all"`
	Timeouts                       timeouts.Value                                                   `tfsdk:"timeouts"`
	TLSInspectionConfiguration     fwtypes.ListNestedObjectValueOf[tlsInspectionConfigurationModel] `tfsdk:"tls_inspection_configuration"`
	TLSInspectionConfigurationARN  types.String                                                     `tfsdk:"arn"`
	TLSInspectionConfigurationID   types.String                                                     `tfsdk:"tls_inspection_configuration_id"`
	TLSInspectionConfigurationName types.String                                                     `tfsdk:"name"`
	UpdateToken                    types.String                                                     `tfsdk:"update_token"`
}

func (model *tlsInspectionConfigurationResourceModel) InitFromID() error {
	model.TLSInspectionConfigurationARN = model.ID

	return nil
}

func (model *tlsInspectionConfigurationResourceModel) setID() {
	model.ID = model.TLSInspectionConfigurationARN
}

type encryptionConfigurationModel struct {
	KeyID types.String `tfsdk:"key_id"`
	Type  types.String `tfsdk:"type"`
}

type tlsInspectionConfigurationModel struct {
	ServerCertificateConfigurations fwtypes.ListNestedObjectValueOf[serverCertificateConfigurationModel] `tfsdk:"server_certificate_configuration"`
}

type serverCertificateConfigurationModel struct {
	CertificateAuthorityARN           fwtypes.ARN                                                                   `tfsdk:"certificate_authority_arn"`
	CheckCertificateRevocationsStatus fwtypes.ListNestedObjectValueOf[checkCertificateRevocationStatusActionsModel] `tfsdk:"check_certificate_revocation_status"`
	Scopes                            fwtypes.ListNestedObjectValueOf[serverCertificateScopeModel]                  `tfsdk:"scope"`
	ServerCertificates                fwtypes.ListNestedObjectValueOf[serverCertificateModel]                       `tfsdk:"server_certificate"`
}

type checkCertificateRevocationStatusActionsModel struct {
	RevokedStatusAction fwtypes.StringEnum[awstypes.RevocationCheckAction] `tfsdk:"revoked_status_action"`
	UnknownStatusAction fwtypes.StringEnum[awstypes.RevocationCheckAction] `tfsdk:"unknown_status_action"`
}

type serverCertificateScopeModel struct {
	DestinationPorts fwtypes.ListNestedObjectValueOf[portRangeModel] `tfsdk:"destination_ports"`
	Destinations     fwtypes.ListNestedObjectValueOf[addressModel]   `tfsdk:"destination"`
	SourcePorts      fwtypes.ListNestedObjectValueOf[portRangeModel] `tfsdk:"source_ports"`
	Protocols        fwtypes.SetValueOf[types.Int64]                 `tfsdk:"protocols"`
	Sources          fwtypes.ListNestedObjectValueOf[addressModel]   `tfsdk:"source"`
}

type portRangeModel struct {
	FromPort types.Int64 `tfsdk:"from_port"`
	ToPort   types.Int64 `tfsdk:"to_port"`
}

type addressModel struct {
	AddressDefinition types.String `tfsdk:"address_definition"`
}

type serverCertificateModel struct {
	ResourceARN fwtypes.ARN `tfsdk:"resource_arn"`
}

type tlsCertificateDataModel struct {
	CertificateARN    types.String `tfsdk:"certificate_arn"`
	CertificateSerial types.String `tfsdk:"certificate_serial"`
	Status            types.String `tfsdk:"status"`
	StatusMessage     types.String `tfsdk:"status_message"`
}
