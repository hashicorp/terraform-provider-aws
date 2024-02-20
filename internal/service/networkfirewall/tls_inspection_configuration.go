// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="TLS Inspection Configuration")
func newResourceTLSInspectionConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTLSInspectionConfiguration{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameTLSInspectionConfiguration = "TLS Inspection Configuration"
)

type resourceTLSInspectionConfiguration struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceTLSInspectionConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_networkfirewall_tls_inspection_configuration"
}

func (r *resourceTLSInspectionConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 512),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^.*$`), "Must provide a valid ARN",
					),
				},
			},
			"certificate_authority": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"certificate_arn":    types.StringType,
						"certificate_serial": types.StringType,
						"status":             types.StringType,
						"status_message":     types.StringType,
					},
				},
			},
			"certificates": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[certificatesData](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"certificate_arn":    types.StringType,
						"certificate_serial": types.StringType,
						"status":             types.StringType,
						"status_message":     types.StringType,
					},
				},
			},
			"id": framework.IDAttribute(),
			// Map name to TLSInspectionConfigurationName
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9-]+$`), "Must contain only alphanumeric characters and dash '-'"),
				},
			},
			"last_modified_time": schema.StringAttribute{
				Computed: true,
			},
			"number_of_associations": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"update_token": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"encryption_configuration": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("AWS_OWNED_KMS_KEY"),
						},
						"type": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("AWS_OWNED_KMS_KEY"),
							Validators: []validator.String{
								stringvalidator.OneOf(networkfirewall.EncryptionTypeAwsOwnedKmsKey, networkfirewall.EncryptionTypeCustomerKms),
							},
						},
					},
				},
			},
			"tls_inspection_configuration": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"server_certificate_configurations": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"certificate_authority_arn": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 256),
											stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws.*`), "Must provide a valid ARN"),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"check_certificate_revocation_status": schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"revoked_status_action": schema.StringAttribute{
													Optional: true,
												},
												"unknown_status_action": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
									"server_certificates": schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"resource_arn": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 256),
														stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws.*`), "Must provide a valid ARN"),
													},
												},
											},
										},
									},
									"scopes": schema.ListNestedBlock{
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"protocols": schema.ListAttribute{
													ElementType: types.Int64Type,
													Required:    true,
												},
											},
											Blocks: map[string]schema.Block{
												"destination_ports": schema.ListNestedBlock{
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
												"destinations": schema.ListNestedBlock{
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
												"sources": schema.ListNestedBlock{
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
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceTLSInspectionConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkFirewallConn(ctx)

	var plan resourceTLSInspectionConfigurationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &networkfirewall.CreateTLSInspectionConfigurationInput{
		// NOTE: Name is mandatory
		TLSInspectionConfigurationName: aws.String(plan.Name.ValueString()),
	}

	if !plan.Description.IsNull() {
		// NOTE: Description is optional
		in.Description = aws.String(plan.Description.ValueString())
	}

	if !plan.TLSInspectionConfiguration.IsNull() {
		var tfList []tlsInspectionConfigurationData
		resp.Diagnostics.Append(plan.TLSInspectionConfiguration.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.TLSInspectionConfiguration = expandTLSInspectionConfiguration(ctx, tfList)
	}

	if !plan.EncryptionConfiguration.IsNull() {
		var tfList []encryptionConfigurationData
		resp.Diagnostics.Append(plan.EncryptionConfiguration.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.EncryptionConfiguration = expandTLSEncryptionConfiguration(tfList)
	}

	out, err := conn.CreateTLSInspectionConfigurationWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionCreating, ResNameTLSInspectionConfiguration, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationArn)
	// Set ID to ARN since ID value is not used for describe, update, delete or list API calls
	plan.ID = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationArn)
	plan.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitTLSInspectionConfigurationCreated(ctx, conn, plan.ARN.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionWaitingForCreation, ResNameTLSInspectionConfiguration, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Read to get computed attributes not returned from create
	readComputed, _ := findTLSInspectionConfigurationByNameAndARN(ctx, conn, plan.ARN.ValueString())

	// Set computed attributes
	plan.LastModifiedTime = flex.StringValueToFramework(ctx, readComputed.TLSInspectionConfigurationResponse.LastModifiedTime.Format(time.RFC3339))
	plan.NumberOfAssociations = flex.Int64ToFramework(ctx, readComputed.TLSInspectionConfigurationResponse.NumberOfAssociations)
	plan.Status = flex.StringToFramework(ctx, readComputed.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus)

	resp.Diagnostics.Append(flex.Flatten(ctx, readComputed.TLSInspectionConfigurationResponse.Certificates, &plan.Certificates)...)

	certificateAuthority, d := flattenTLSCertificate(ctx, readComputed.TLSInspectionConfigurationResponse.CertificateAuthority)
	resp.Diagnostics.Append(d...)
	plan.CertificateAuthority = certificateAuthority

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTLSInspectionConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkFirewallConn(ctx)

	var state resourceTLSInspectionConfigurationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTLSInspectionConfigurationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionSetting, ResNameTLSInspectionConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationArn)
	state.Description = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.Description)

	// Set ID to ARN since ID value is not used for Describe, Update, Delete or List calls
	state.ID = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationArn)
	state.Name = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationName)

	state.LastModifiedTime = flex.StringValueToFramework(ctx, out.TLSInspectionConfigurationResponse.LastModifiedTime.Format(time.RFC3339))
	state.NumberOfAssociations = flex.Int64ToFramework(ctx, out.TLSInspectionConfigurationResponse.NumberOfAssociations)
	state.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)
	state.Status = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus)

	// Complex types
	encryptionConfiguration, d := flattenTLSEncryptionConfiguration(ctx, out.TLSInspectionConfigurationResponse.EncryptionConfiguration)
	resp.Diagnostics.Append(d...)
	state.EncryptionConfiguration = encryptionConfiguration

	certificateAuthority, d := flattenTLSCertificate(ctx, out.TLSInspectionConfigurationResponse.CertificateAuthority)
	resp.Diagnostics.Append(d...)
	state.CertificateAuthority = certificateAuthority

	resp.Diagnostics.Append(flex.Flatten(ctx, out.TLSInspectionConfigurationResponse.Certificates, &state.Certificates)...)

	tlsInspectionConfiguration, d := flattenTLSInspectionConfiguration(ctx, out.TLSInspectionConfiguration)
	resp.Diagnostics.Append(d...)
	state.TLSInspectionConfiguration = tlsInspectionConfiguration

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTLSInspectionConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkFirewallConn(ctx)

	var plan, state resourceTLSInspectionConfigurationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.TLSInspectionConfiguration.Equal(state.TLSInspectionConfiguration) ||
		!plan.EncryptionConfiguration.Equal(state.EncryptionConfiguration) {
		in := &networkfirewall.UpdateTLSInspectionConfigurationInput{
			TLSInspectionConfigurationArn:  aws.String(plan.ARN.ValueString()),
			TLSInspectionConfigurationName: aws.String(plan.Name.ValueString()),
			UpdateToken:                    aws.String(state.UpdateToken.ValueString()),
		}

		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}

		if !plan.TLSInspectionConfiguration.IsNull() {
			var tfList []tlsInspectionConfigurationData
			resp.Diagnostics.Append(plan.TLSInspectionConfiguration.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.TLSInspectionConfiguration = expandTLSInspectionConfiguration(ctx, tfList)
		}

		if !plan.EncryptionConfiguration.IsNull() {
			var tfList []encryptionConfigurationData
			resp.Diagnostics.Append(plan.EncryptionConfiguration.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.EncryptionConfiguration = expandTLSEncryptionConfiguration(tfList)
		}

		out, err := conn.UpdateTLSInspectionConfigurationWithContext(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionUpdating, ResNameTLSInspectionConfiguration, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.TLSInspectionConfigurationResponse == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionUpdating, ResNameTLSInspectionConfiguration, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationArn)
		// Set ID to ARN since ID value is not used for describe, update, delete or list API calls
		plan.ID = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationArn)
		plan.LastModifiedTime = flex.StringValueToFramework(ctx, out.TLSInspectionConfigurationResponse.LastModifiedTime.Format(time.RFC3339))
		plan.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)
		plan.Status = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus)

		encryptionConfiguration, d := flattenTLSEncryptionConfiguration(ctx, out.TLSInspectionConfigurationResponse.EncryptionConfiguration)
		resp.Diagnostics.Append(d...)
		plan.EncryptionConfiguration = encryptionConfiguration

		// Update does not certificates and CA, so read to backfill the missing attributes
		// NOTE: number of associations should be returned according to the API docs, but isn't!
		readComputed, _ := findTLSInspectionConfigurationByNameAndARN(ctx, conn, plan.ARN.ValueString())
		plan.NumberOfAssociations = flex.Int64ToFramework(ctx, readComputed.TLSInspectionConfigurationResponse.NumberOfAssociations)

		// Complex types
		certificateAuthority, d := flattenTLSCertificate(ctx, readComputed.TLSInspectionConfigurationResponse.CertificateAuthority)
		resp.Diagnostics.Append(d...)
		plan.CertificateAuthority = certificateAuthority

		resp.Diagnostics.Append(flex.Flatten(ctx, readComputed.TLSInspectionConfigurationResponse.Certificates, &plan.Certificates)...)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitTLSInspectionConfigurationUpdated(ctx, conn, plan.ARN.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionWaitingForUpdate, ResNameTLSInspectionConfiguration, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTLSInspectionConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkFirewallConn(ctx)

	var state resourceTLSInspectionConfigurationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &networkfirewall.DeleteTLSInspectionConfigurationInput{
		TLSInspectionConfigurationArn: aws.String(state.ARN.ValueString()),
	}

	_, err := conn.DeleteTLSInspectionConfigurationWithContext(ctx, in)
	if err != nil {
		if errs.IsA[*networkfirewall.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionDeleting, ResNameTLSInspectionConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitTLSInspectionConfigurationDeleted(ctx, conn, state.ARN.ValueString(), deleteTimeout)
	if err != nil {
		if errs.IsA[*networkfirewall.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionWaitingForDeletion, ResNameTLSInspectionConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTLSInspectionConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitTLSInspectionConfigurationCreated(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string, timeout time.Duration) (*networkfirewall.TLSInspectionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{networkfirewall.ResourceStatusActive},
		Refresh:                   statusTLSInspectionConfigurationCertificates(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 5,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.TLSInspectionConfiguration); ok {
		return out, err
	}

	return nil, err
}

func statusTLSInspectionConfigurationCertificates(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findTLSInspectionConfigurationByNameAndARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		certificates := out.TLSInspectionConfigurationResponse.Certificates
		certificateAuthority := out.TLSInspectionConfigurationResponse.CertificateAuthority

		// The API does not immediately return data for certificates and certificate authority even when the resource status is "ACTIVE",
		// which causes unexpected diffs when reading. This sets the status to "PENDING" until either the certificates or the certificate
		// authority is populated (the API will always return at least one of the two)
		if aws.ToString(out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus) == networkfirewall.ResourceStatusActive &&
			(certificates != nil || certificateAuthority != nil) {
			return out, aws.ToString(out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus), nil
		} else {
			return out, statusChangePending, nil
		}
	}
}

func waitTLSInspectionConfigurationUpdated(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string, timeout time.Duration) (*networkfirewall.TLSInspectionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{networkfirewall.ResourceStatusActive},
		Refresh:                   statusTLSInspectionConfigurationCertificates(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.TLSInspectionConfiguration); ok {
		return out, err
	}

	return nil, err
}

func waitTLSInspectionConfigurationDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string, timeout time.Duration) (*networkfirewall.TLSInspectionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkfirewall.ResourceStatusDeleting, networkfirewall.ResourceStatusActive},
		Target:  []string{},
		Refresh: statusTLSInspectionConfiguration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.TLSInspectionConfiguration); ok {
		return out, err
	}

	return nil, err
}

func statusTLSInspectionConfiguration(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findTLSInspectionConfigurationByNameAndARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus), nil
	}
}

func findTLSInspectionConfigurationByNameAndARN(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeTLSInspectionConfigurationOutput, error) {
	in := &networkfirewall.DescribeTLSInspectionConfigurationInput{
		TLSInspectionConfigurationArn: aws.String(arn),
	}

	out, err := conn.DescribeTLSInspectionConfigurationWithContext(ctx, in)
	if err != nil {
		if errs.IsA[*networkfirewall.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.TLSInspectionConfigurationResponse == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func findTLSInspectionConfigurationByID(ctx context.Context, conn *networkfirewall.NetworkFirewall, id string) (*networkfirewall.DescribeTLSInspectionConfigurationOutput, error) {
	in := &networkfirewall.DescribeTLSInspectionConfigurationInput{
		TLSInspectionConfigurationArn: aws.String(id),
	}

	out, err := conn.DescribeTLSInspectionConfigurationWithContext(ctx, in)
	if err != nil {
		if errs.IsA[*networkfirewall.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.TLSInspectionConfigurationResponse == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenTLSInspectionConfiguration(ctx context.Context, tlsInspectionConfiguration *networkfirewall.TLSInspectionConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: tlsInspectionConfigurationAttrTypes}

	if tlsInspectionConfiguration == nil {
		return types.ListNull(elemType), diags
	}

	flattenedConfig, d := flattenServerCertificateConfigurations(ctx, tlsInspectionConfiguration.ServerCertificateConfigurations)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"server_certificate_configurations": flattenedConfig,
	}
	objVal, d := types.ObjectValue(tlsInspectionConfigurationAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenServerCertificateConfigurations(ctx context.Context, serverCertificateConfigurations []*networkfirewall.ServerCertificateConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: serverCertificateConfigurationAttrTypes}

	if serverCertificateConfigurations == nil {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, serverCertificateConfiguration := range serverCertificateConfigurations {
		checkCertRevocationStatus, d := flattenCheckCertificateRevocationStatus(ctx, serverCertificateConfiguration.CheckCertificateRevocationStatus)
		diags.Append(d...)
		scopes, d := flattenScopes(ctx, serverCertificateConfiguration.Scopes)
		diags.Append(d...)
		serverCertificates, d := flattenServerCertificates(ctx, serverCertificateConfiguration.ServerCertificates)
		diags.Append(d...)

		obj := map[string]attr.Value{
			"certificate_authority_arn":           flex.StringToFramework(ctx, serverCertificateConfiguration.CertificateAuthorityArn),
			"check_certificate_revocation_status": checkCertRevocationStatus,
			"scopes":                              scopes,
			"server_certificates":                 serverCertificates,
		}

		objVal, d := types.ObjectValue(serverCertificateConfigurationAttrTypes, obj)
		diags.Append(d...)
		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenCheckCertificateRevocationStatus(ctx context.Context, checkCertificateRevocationStatus *networkfirewall.CheckCertificateRevocationStatusActions) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: checkCertificateRevocationStatusAttrTypes}

	if checkCertificateRevocationStatus == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"revoked_status_action": flex.StringToFramework(ctx, checkCertificateRevocationStatus.RevokedStatusAction),
		"unknown_status_action": flex.StringToFramework(ctx, checkCertificateRevocationStatus.UnknownStatusAction),
	}

	flattenedCheckCertificateRevocationStatus, d := types.ObjectValue(checkCertificateRevocationStatusAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{flattenedCheckCertificateRevocationStatus})
	diags.Append(d...)

	return listVal, diags
}

func flattenServerCertificates(ctx context.Context, serverCertificateList []*networkfirewall.ServerCertificate) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: serverCertificatesAttrTypes}

	if len(serverCertificateList) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, serverCertificate := range serverCertificateList {
		if serverCertificate == nil {
			continue
		}
		obj := map[string]attr.Value{
			"resource_arn": flex.StringToFramework(ctx, serverCertificate.ResourceArn),
		}

		flattenedServerCertificate, d := types.ObjectValue(serverCertificatesAttrTypes, obj)

		diags.Append(d...)
		elems = append(elems, flattenedServerCertificate)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenTLSCertificate(ctx context.Context, certificate *networkfirewall.TlsCertificateData) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: certificatesAttrTypes}

	if certificate == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"certificate_arn":    flex.StringToFramework(ctx, certificate.CertificateArn),
		"certificate_serial": flex.StringToFramework(ctx, certificate.CertificateSerial),
		"status":             flex.StringToFramework(ctx, certificate.Status),
		"status_message":     flex.StringToFramework(ctx, certificate.StatusMessage),
	}
	objVal, d := types.ObjectValue(certificatesAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenScopes(ctx context.Context, scopes []*networkfirewall.ServerCertificateScope) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: scopeAttrTypes}

	if len(scopes) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, scope := range scopes {
		if scope == nil {
			continue
		}

		destinationPorts, d := flattenPortRange(ctx, scope.DestinationPorts)
		diags.Append(d...)
		destinations, d := flattenSourceDestinations(ctx, scope.Destinations)
		diags.Append(d...)
		protocols, d := flattenProtocols(scope.Protocols)
		diags.Append(d...)
		sourcePorts, d := flattenPortRange(ctx, scope.SourcePorts)
		diags.Append(d...)
		sources, d := flattenSourceDestinations(ctx, scope.Sources)
		diags.Append(d...)

		obj := map[string]attr.Value{
			"destination_ports": destinationPorts,
			"destinations":      destinations,
			"protocols":         protocols,
			"source_ports":      sourcePorts,
			"sources":           sources,
		}
		objVal, d := types.ObjectValue(scopeAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenProtocols(list []*int64) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.Int64Type

	if len(list) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, item := range list {
		if item == nil {
			continue
		}

		objVal := types.Int64Value(*item)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenSourceDestinations(ctx context.Context, destinations []*networkfirewall.Address) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: sourceDestinationAttrTypes}

	if len(destinations) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, destination := range destinations {
		if destination == nil {
			continue
		}

		obj := map[string]attr.Value{
			"address_definition": flex.StringToFramework(ctx, destination.AddressDefinition),
		}
		objVal, d := types.ObjectValue(sourceDestinationAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenPortRange(ctx context.Context, ranges []*networkfirewall.PortRange) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: portRangeAttrTypes}

	if len(ranges) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, portRange := range ranges {
		if portRange == nil {
			continue
		}

		obj := map[string]attr.Value{
			"from_port": flex.Int64ToFramework(ctx, portRange.FromPort),
			"to_port":   flex.Int64ToFramework(ctx, portRange.ToPort),
		}
		objVal, d := types.ObjectValue(portRangeAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenTLSEncryptionConfiguration(ctx context.Context, encryptionConfiguration *networkfirewall.EncryptionConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: encryptionConfigurationAttrTypes}

	if encryptionConfiguration == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"key_id": flex.StringToFramework(ctx, encryptionConfiguration.KeyId),
		"type":   flex.StringToFramework(ctx, encryptionConfiguration.Type),
	}
	objVal, d := types.ObjectValue(encryptionConfigurationAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

// TODO: add note explaining why not using existing expandEncryptionConfiguration()
func expandTLSEncryptionConfiguration(tfList []encryptionConfigurationData) *networkfirewall.EncryptionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &networkfirewall.EncryptionConfiguration{
		KeyId: aws.String(tfObj.KeyId.ValueString()),
		Type:  aws.String(tfObj.Type.ValueString()),
	}

	return apiObject
}

func expandTLSInspectionConfiguration(ctx context.Context, tfList []tlsInspectionConfigurationData) *networkfirewall.TLSInspectionConfiguration {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]

	var serverCertConfig []serverCertificateConfigurationsData
	diags.Append(tfObj.ServerCertificateConfiguration.ElementsAs(ctx, &serverCertConfig, false)...)

	apiObject := &networkfirewall.TLSInspectionConfiguration{
		ServerCertificateConfigurations: expandServerCertificateConfigurations(ctx, serverCertConfig),
	}

	return apiObject
}

func expandServerCertificateConfigurations(ctx context.Context, tfList []serverCertificateConfigurationsData) []*networkfirewall.ServerCertificateConfiguration {
	var diags diag.Diagnostics

	var apiObject []*networkfirewall.ServerCertificateConfiguration

	for _, item := range tfList {
		conf := &networkfirewall.ServerCertificateConfiguration{}

		// Configure CertificateAuthorityArn for outbound SSL/TLS inspection
		if !item.CertificateAuthorityArn.IsNull() {
			conf.CertificateAuthorityArn = aws.String(item.CertificateAuthorityArn.ValueString())
		}
		if !item.CheckCertificateRevocationsStatus.IsNull() {
			var certificateRevocationStatus []checkCertificateRevocationStatusData
			diags.Append(item.CheckCertificateRevocationsStatus.ElementsAs(ctx, &certificateRevocationStatus, false)...)
			conf.CheckCertificateRevocationStatus = expandCheckCertificateRevocationStatus(certificateRevocationStatus)
		}
		if !item.Scope.IsNull() {
			var scopesList []scopeData
			diags.Append(item.Scope.ElementsAs(ctx, &scopesList, false)...)
			conf.Scopes = expandScopes(ctx, scopesList)
		}
		// Configure ServerCertificates for inbound SSL/TLS inspection
		if !item.ServerCertificates.IsNull() {
			var serverCertificates []serverCertificatesData
			diags.Append(item.ServerCertificates.ElementsAs(ctx, &serverCertificates, false)...)
			conf.ServerCertificates = expandServerCertificates(serverCertificates)
		}

		apiObject = append(apiObject, conf)
	}

	return apiObject
}

func expandCheckCertificateRevocationStatus(tfList []checkCertificateRevocationStatusData) *networkfirewall.CheckCertificateRevocationStatusActions {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &networkfirewall.CheckCertificateRevocationStatusActions{
		RevokedStatusAction: aws.String(tfObj.RevokedStatusAction.ValueString()),
		UnknownStatusAction: aws.String(tfObj.UnknownStatusAction.ValueString()),
	}
	return apiObject
}

func expandServerCertificates(tfList []serverCertificatesData) []*networkfirewall.ServerCertificate {
	var apiObject []*networkfirewall.ServerCertificate

	for _, item := range tfList {
		conf := &networkfirewall.ServerCertificate{
			ResourceArn: aws.String(item.ResourceARN.ValueString()),
		}

		apiObject = append(apiObject, conf)
	}
	return apiObject
}

func expandScopes(ctx context.Context, tfList []scopeData) []*networkfirewall.ServerCertificateScope {
	var diags diag.Diagnostics
	var apiObject []*networkfirewall.ServerCertificateScope

	for _, tfObj := range tfList {
		item := &networkfirewall.ServerCertificateScope{}
		if !tfObj.Protocols.IsNull() {
			protocols := []*int64{}
			diags.Append(tfObj.Protocols.ElementsAs(ctx, &protocols, false)...)
			item.Protocols = protocols
		}
		if !tfObj.DestinationPorts.IsNull() {
			var destinationPorts []portRangeData
			diags.Append(tfObj.DestinationPorts.ElementsAs(ctx, &destinationPorts, false)...)
			item.DestinationPorts = expandPortRange(destinationPorts)
		}
		if !tfObj.Destinations.IsNull() {
			var destinations []sourceDestinationData
			diags.Append(tfObj.Destinations.ElementsAs(ctx, &destinations, false)...)
			item.Destinations = expandSourceDestinations(destinations)
		}
		if !tfObj.SourcePorts.IsNull() {
			var sourcePorts []portRangeData
			diags.Append(tfObj.SourcePorts.ElementsAs(ctx, &sourcePorts, false)...)
			item.SourcePorts = expandPortRange(sourcePorts)
		}
		if !tfObj.Sources.IsNull() {
			var sources []sourceDestinationData
			diags.Append(tfObj.Sources.ElementsAs(ctx, &sources, false)...)
			item.Sources = expandSourceDestinations(sources)
		}
		apiObject = append(apiObject, item)
	}

	return apiObject
}

func expandPortRange(tfList []portRangeData) []*networkfirewall.PortRange {
	var apiObject []*networkfirewall.PortRange

	for _, tfObj := range tfList {
		item := &networkfirewall.PortRange{
			FromPort: aws.Int64(tfObj.FromPort.ValueInt64()),
			ToPort:   aws.Int64(tfObj.ToPort.ValueInt64()),
		}
		apiObject = append(apiObject, item)
	}

	return apiObject
}

func expandSourceDestinations(tfList []sourceDestinationData) []*networkfirewall.Address {
	var apiObject []*networkfirewall.Address

	for _, tfObj := range tfList {
		item := &networkfirewall.Address{
			AddressDefinition: aws.String(tfObj.AddressDefinition.ValueString()),
		}
		apiObject = append(apiObject, item)
	}

	return apiObject
}

type resourceTLSInspectionConfigurationData struct {
	ARN                        types.String                                      `tfsdk:"arn"`
	EncryptionConfiguration    types.List                                        `tfsdk:"encryption_configuration"`
	Certificates               fwtypes.ListNestedObjectValueOf[certificatesData] `tfsdk:"certificates"`
	CertificateAuthority       types.List                                        `tfsdk:"certificate_authority"`
	Description                types.String                                      `tfsdk:"description"`
	ID                         types.String                                      `tfsdk:"id"`
	LastModifiedTime           types.String                                      `tfsdk:"last_modified_time"`
	Name                       types.String                                      `tfsdk:"name"`
	NumberOfAssociations       types.Int64                                       `tfsdk:"number_of_associations"`
	Status                     types.String                                      `tfsdk:"status"`
	TLSInspectionConfiguration types.List                                        `tfsdk:"tls_inspection_configuration"`
	Timeouts                   timeouts.Value                                    `tfsdk:"timeouts"`
	UpdateToken                types.String                                      `tfsdk:"update_token"`
}

type encryptionConfigurationData struct {
	Type  types.String `tfsdk:"type"`
	KeyId types.String `tfsdk:"key_id"`
}

type certificatesData struct {
	CertificateArn    types.String `tfsdk:"certificate_arn"`
	CertificateSerial types.String `tfsdk:"certificate_serial"`
	Status            types.String `tfsdk:"status"`
	StatusMessage     types.String `tfsdk:"status_message"`
}

type tlsInspectionConfigurationData struct {
	ServerCertificateConfiguration types.List `tfsdk:"server_certificate_configurations"`
}

type serverCertificateConfigurationsData struct {
	CertificateAuthorityArn           types.String `tfsdk:"certificate_authority_arn"`
	CheckCertificateRevocationsStatus types.List   `tfsdk:"check_certificate_revocation_status"`
	Scope                             types.List   `tfsdk:"scopes"`
	ServerCertificates                types.List   `tfsdk:"server_certificates"`
}

type scopeData struct {
	DestinationPorts types.List `tfsdk:"destination_ports"`
	Destinations     types.List `tfsdk:"destinations"`
	Protocols        types.List `tfsdk:"protocols"`
	SourcePorts      types.List `tfsdk:"source_ports"`
	Sources          types.List `tfsdk:"sources"`
}

type sourceDestinationData struct {
	AddressDefinition types.String `tfsdk:"address_definition"`
}

type portRangeData struct {
	FromPort types.Int64 `tfsdk:"from_port"`
	ToPort   types.Int64 `tfsdk:"to_port"`
}

type checkCertificateRevocationStatusData struct {
	RevokedStatusAction types.String `tfsdk:"revoked_status_action"`
	UnknownStatusAction types.String `tfsdk:"unknown_status_action"`
}

type serverCertificatesData struct {
	ResourceARN types.String `tfsdk:"resource_arn"`
}

var certificatesAttrTypes = map[string]attr.Type{
	"certificate_arn":    types.StringType,
	"certificate_serial": types.StringType,
	"status":             types.StringType,
	"status_message":     types.StringType,
}

var encryptionConfigurationAttrTypes = map[string]attr.Type{
	"type":   types.StringType,
	"key_id": types.StringType,
}

var tlsInspectionConfigurationAttrTypes = map[string]attr.Type{
	"server_certificate_configurations": types.ListType{ElemType: types.ObjectType{AttrTypes: serverCertificateConfigurationAttrTypes}},
}

var serverCertificateConfigurationAttrTypes = map[string]attr.Type{
	"certificate_authority_arn":           types.StringType,
	"check_certificate_revocation_status": types.ListType{ElemType: types.ObjectType{AttrTypes: checkCertificateRevocationStatusAttrTypes}},
	"scopes":                              types.ListType{ElemType: types.ObjectType{AttrTypes: scopeAttrTypes}},
	"server_certificates":                 types.ListType{ElemType: types.ObjectType{AttrTypes: serverCertificatesAttrTypes}},
}

var checkCertificateRevocationStatusAttrTypes = map[string]attr.Type{
	"revoked_status_action": types.StringType,
	"unknown_status_action": types.StringType,
}

var (
	scopeAttrTypes = map[string]attr.Type{
		"destination_ports": types.ListType{ElemType: types.ObjectType{AttrTypes: portRangeAttrTypes}},
		"destinations":      types.ListType{ElemType: types.ObjectType{AttrTypes: sourceDestinationAttrTypes}},
		"protocols":         types.ListType{ElemType: types.Int64Type},
		"source_ports":      types.ListType{ElemType: types.ObjectType{AttrTypes: portRangeAttrTypes}},
		"sources":           types.ListType{ElemType: types.ObjectType{AttrTypes: sourceDestinationAttrTypes}},
	}

	sourceDestinationAttrTypes = map[string]attr.Type{
		"address_definition": types.StringType,
	}

	portRangeAttrTypes = map[string]attr.Type{
		"from_port": types.Int64Type,
		"to_port":   types.Int64Type,
	}
)

var serverCertificatesAttrTypes = map[string]attr.Type{
	"resource_arn": types.StringType,
}
