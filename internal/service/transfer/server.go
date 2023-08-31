// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_server", name="Server")
// @Tags(identifierAttribute="arn")
func ResourceServer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServerCreate,
		ReadWithoutTimeout:   resourceServerRead,
		UpdateWithoutTimeout: resourceServerUpdate,
		DeleteWithoutTimeout: resourceServerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			customdiff.ForceNewIfChange("endpoint_details.0.vpc_id", func(_ context.Context, old, new, meta interface{}) bool {
				// "InvalidRequestException: Changing VpcId is not supported".
				if old, new := old.(string), new.(string); old != "" && new != old {
					return true
				}

				return false
			}),
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      transfer.DomainS3,
				ValidateFunc: validation.StringInSlice(transfer.Domain_Values(), false),
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_details": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address_allocation_ids": {
							Type:          schema.TypeSet,
							Optional:      true,
							Elem:          &schema.Schema{Type: schema.TypeString},
							ConflictsWith: []string{"endpoint_details.0.vpc_endpoint_id"},
						},
						"security_group_ids": {
							Type:          schema.TypeSet,
							Optional:      true,
							Computed:      true,
							Elem:          &schema.Schema{Type: schema.TypeString},
							ConflictsWith: []string{"endpoint_details.0.vpc_endpoint_id"},
						},
						"subnet_ids": {
							Type:          schema.TypeSet,
							Optional:      true,
							Elem:          &schema.Schema{Type: schema.TypeString},
							ConflictsWith: []string{"endpoint_details.0.vpc_endpoint_id"},
						},
						"vpc_endpoint_id": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"endpoint_details.0.address_allocation_ids", "endpoint_details.0.security_group_ids", "endpoint_details.0.subnet_ids", "endpoint_details.0.vpc_id"},
						},
						"vpc_id": {
							Type:          schema.TypeString,
							Optional:      true,
							ValidateFunc:  validation.NoZeroValues,
							ConflictsWith: []string{"endpoint_details.0.vpc_endpoint_id"},
						},
					},
				},
			},
			"endpoint_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      transfer.EndpointTypePublic,
				ValidateFunc: validation.StringInSlice(transfer.EndpointType_Values(), false),
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"function": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"host_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(0, 4096),
			},
			"host_key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_provider_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      transfer.IdentityProviderTypeServiceManaged,
				ValidateFunc: validation.StringInSlice(transfer.IdentityProviderType_Values(), false),
			},
			"invocation_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"logging_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"post_authentication_login_banner": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"pre_authentication_login_banner": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"protocol_details": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"as2_transports": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(transfer.As2Transport_Values(), false),
							},
						},
						"passive_ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 15),
						},
						"set_stat_option": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(transfer.SetStatOption_Values(), false),
						},
						"tls_session_resumption_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(transfer.TlsSessionResumptionMode_Values(), false),
						},
					},
				},
			},
			"protocols": {
				Type:     schema.TypeSet,
				MinItems: 1,
				MaxItems: 3,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(transfer.Protocol_Values(), false),
				},
			},
			"security_policy_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      SecurityPolicyName2018_11,
				ValidateFunc: validation.StringInSlice(SecurityPolicyName_Values(), false),
			},
			"structured_log_destinations": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
				Description: "This is a set of arns of destinations that will receive structured logs from the transfer server",
				Optional:    true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"workflow_details": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"on_partial_upload": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							AtLeastOneOf: []string{"workflow_details.0.on_upload", "workflow_details.0.on_partial_upload"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"execution_role": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"workflow_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"on_upload": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							AtLeastOneOf: []string{"workflow_details.0.on_upload", "workflow_details.0.on_partial_upload"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"execution_role": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"workflow_id": {
										Type:     schema.TypeString,
										Required: true,
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

func resourceServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	input := &transfer.CreateServerInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("certificate"); ok {
		input.Certificate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("directory_id"); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &transfer.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.DirectoryId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain"); ok {
		input.Domain = aws.String(v.(string))
	}

	var addressAllocationIDs []*string

	if v, ok := d.GetOk("endpoint_details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EndpointDetails = expandEndpointDetails(v.([]interface{})[0].(map[string]interface{}))

		// Prevent the following error: InvalidRequestException: AddressAllocationIds cannot be set in CreateServer
		// Reference: https://docs.aws.amazon.com/transfer/latest/userguide/API_EndpointDetails.html#TransferFamily-Type-EndpointDetails-AddressAllocationIds
		addressAllocationIDs = input.EndpointDetails.AddressAllocationIds
		input.EndpointDetails.AddressAllocationIds = nil
	}

	if v, ok := d.GetOk("endpoint_type"); ok {
		input.EndpointType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("function"); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &transfer.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.Function = aws.String(v.(string))
	}

	if v, ok := d.GetOk("host_key"); ok {
		input.HostKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("identity_provider_type"); ok {
		input.IdentityProviderType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invocation_role"); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &transfer.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.InvocationRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_role"); ok {
		input.LoggingRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("post_authentication_login_banner"); ok {
		input.PostAuthenticationLoginBanner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pre_authentication_login_banner"); ok {
		input.PreAuthenticationLoginBanner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("protocol_details"); ok && len(v.([]interface{})) > 0 {
		input.ProtocolDetails = expandProtocolDetails(v.([]interface{}))
	}

	if v, ok := d.GetOk("protocols"); ok && v.(*schema.Set).Len() > 0 {
		input.Protocols = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("security_policy_name"); ok {
		input.SecurityPolicyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("structured_log_destinations"); ok && v.(*schema.Set).Len() > 0 {
		input.StructuredLogDestinations = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("url"); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &transfer.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.Url = aws.String(v.(string))
	}

	if v, ok := d.GetOk("workflow_details"); ok && len(v.([]interface{})) > 0 {
		input.WorkflowDetails = expandWorkflowDetails(v.([]interface{}))
	}

	output, err := conn.CreateServerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Server: %s", err)
	}

	d.SetId(aws.StringValue(output.ServerId))

	if _, err := waitServerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Transfer Server (%s) create: %s", d.Id(), err)
	}

	// AddressAllocationIds is only valid in the UpdateServer API.
	if len(addressAllocationIDs) > 0 {
		if err := stopServer(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &transfer.UpdateServerInput{
			EndpointDetails: &transfer.EndpointDetails{
				AddressAllocationIds: addressAllocationIDs,
			},
			ServerId: aws.String(d.Id()),
		}

		if err := updateServer(ctx, conn, input); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if err := startServer(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceServerRead(ctx, d, meta)...)
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	output, err := FindServerByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Server (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Server (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.Arn)
	d.Set("certificate", output.Certificate)
	if output.IdentityProviderDetails != nil {
		d.Set("directory_id", output.IdentityProviderDetails.DirectoryId)
	} else {
		d.Set("directory_id", "")
	}
	d.Set("domain", output.Domain)
	d.Set("endpoint", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s.server.transfer", d.Id())))
	if output.EndpointDetails != nil {
		securityGroupIDs := make([]*string, 0)

		// Security Group IDs are not returned for VPC endpoints.
		if aws.StringValue(output.EndpointType) == transfer.EndpointTypeVpc && len(output.EndpointDetails.SecurityGroupIds) == 0 {
			vpcEndpointID := aws.StringValue(output.EndpointDetails.VpcEndpointId)
			output, err := tfec2.FindVPCEndpointByID(ctx, meta.(*conns.AWSClient).EC2Conn(ctx), vpcEndpointID)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading Transfer Server (%s) VPC Endpoint (%s): %s", d.Id(), vpcEndpointID, err)
			}

			for _, group := range output.Groups {
				securityGroupIDs = append(securityGroupIDs, group.GroupId)
			}
		}

		if err := d.Set("endpoint_details", []interface{}{flattenEndpointDetails(output.EndpointDetails, securityGroupIDs)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting endpoint_details: %s", err)
		}
	} else {
		d.Set("endpoint_details", nil)
	}
	d.Set("endpoint_type", output.EndpointType)
	if output.IdentityProviderDetails != nil {
		d.Set("function", output.IdentityProviderDetails.Function)
	} else {
		d.Set("function", "")
	}
	d.Set("host_key_fingerprint", output.HostKeyFingerprint)
	d.Set("identity_provider_type", output.IdentityProviderType)
	if output.IdentityProviderDetails != nil {
		d.Set("invocation_role", output.IdentityProviderDetails.InvocationRole)
	} else {
		d.Set("invocation_role", "")
	}
	d.Set("logging_role", output.LoggingRole)
	d.Set("post_authentication_login_banner", output.PostAuthenticationLoginBanner)
	d.Set("pre_authentication_login_banner", output.PreAuthenticationLoginBanner)
	if err := d.Set("protocol_details", flattenProtocolDetails(output.ProtocolDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting protocol_details: %s", err)
	}
	d.Set("protocols", aws.StringValueSlice(output.Protocols))
	d.Set("security_policy_name", output.SecurityPolicyName)
	d.Set("structured_log_destinations", aws.StringValueSlice(output.StructuredLogDestinations))
	if output.IdentityProviderDetails != nil {
		d.Set("url", output.IdentityProviderDetails.Url)
	} else {
		d.Set("url", "")
	}
	if err := d.Set("workflow_details", flattenWorkflowDetails(output.WorkflowDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workflow_details: %s", err)
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		var newEndpointTypeVpc bool
		var oldEndpointTypeVpc bool

		old, new := d.GetChange("endpoint_type")

		if old, new := old.(string), new.(string); new == transfer.EndpointTypeVpc {
			newEndpointTypeVpc = true
			oldEndpointTypeVpc = old == new
		}

		var addressAllocationIDs []*string
		var offlineUpdate bool
		var removeAddressAllocationIDs bool

		input := &transfer.UpdateServerInput{
			ServerId: aws.String(d.Id()),
		}

		if d.HasChange("certificate") {
			input.Certificate = aws.String(d.Get("certificate").(string))
		}

		if d.HasChange("endpoint_details") {
			if v, ok := d.GetOk("endpoint_details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.EndpointDetails = expandEndpointDetails(v.([]interface{})[0].(map[string]interface{}))

				if newEndpointTypeVpc && !oldEndpointTypeVpc {
					// Prevent the following error: InvalidRequestException: Cannot specify AddressAllocationids when updating server to EndpointType: VPC
					addressAllocationIDs = input.EndpointDetails.AddressAllocationIds
					input.EndpointDetails.AddressAllocationIds = nil

					// Prevent the following error: InvalidRequestException: VPC Endpoint ID unsupported for EndpointType: VPC
					input.EndpointDetails.VpcEndpointId = nil
				} else if newEndpointTypeVpc && oldEndpointTypeVpc {
					// Prevent the following error: InvalidRequestException: Server must be OFFLINE to change AddressAllocationIds
					if d.HasChange("endpoint_details.0.address_allocation_ids") {
						offlineUpdate = true
					}

					// Update to 0 AddressAllocationIds.
					if input.EndpointDetails.AddressAllocationIds == nil {
						input.EndpointDetails.AddressAllocationIds = []*string{}
					}

					// Prevent the following error: InvalidRequestException: AddressAllocationIds must be removed before SubnetIds can be modified
					if d.HasChange("endpoint_details.0.subnet_ids") {
						old, _ := d.GetChange("endpoint_details.0.address_allocation_ids")

						if old := old.(*schema.Set); old.Len() > 0 {
							offlineUpdate = true
							removeAddressAllocationIDs = true

							addressAllocationIDs = input.EndpointDetails.AddressAllocationIds
							input.EndpointDetails.AddressAllocationIds = nil
						}
					}

					// Prevent the following error: InvalidRequestException: Changing Security Group is not supported
					input.EndpointDetails.SecurityGroupIds = nil

					// Update to 0 SubnetIds.
					if input.EndpointDetails.SubnetIds == nil {
						input.EndpointDetails.SubnetIds = []*string{}
					}
				}
			}

			// You can edit the SecurityGroupIds property in the UpdateServer API only if you are changing the EndpointType from PUBLIC or VPC_ENDPOINT to VPC.
			// To change security groups associated with your server's VPC endpoint after creation, use the Amazon EC2 ModifyVpcEndpoint API.
			if d.HasChange("endpoint_details.0.security_group_ids") && newEndpointTypeVpc && oldEndpointTypeVpc {
				conn := meta.(*conns.AWSClient).EC2Conn(ctx)

				vpcEndpointID := d.Get("endpoint_details.0.vpc_endpoint_id").(string)
				input := &ec2.ModifyVpcEndpointInput{
					VpcEndpointId: aws.String(vpcEndpointID),
				}

				old, new := d.GetChange("endpoint_details.0.security_group_ids")

				if add := flex.ExpandStringSet(new.(*schema.Set).Difference(old.(*schema.Set))); len(add) > 0 {
					input.AddSecurityGroupIds = add
				}

				if del := flex.ExpandStringSet(old.(*schema.Set).Difference(new.(*schema.Set))); len(del) > 0 {
					input.RemoveSecurityGroupIds = del
				}

				_, err := conn.ModifyVpcEndpointWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying Transfer Server (%s) VPC Endpoint (%s): %s", d.Id(), vpcEndpointID, err)
				}

				if _, err := tfec2.WaitVPCEndpointAvailable(ctx, conn, vpcEndpointID, tfec2.VPCEndpointCreationTimeout); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Transfer Server (%s) VPC Endpoint (%s) update: %s", d.Id(), vpcEndpointID, err)
				}
			}
		}

		if d.HasChange("endpoint_type") {
			input.EndpointType = aws.String(d.Get("endpoint_type").(string))

			// Prevent the following error: InvalidRequestException: Server must be OFFLINE to change EndpointType
			offlineUpdate = true
		}

		if d.HasChange("host_key") {
			if attr, ok := d.GetOk("host_key"); ok {
				input.HostKey = aws.String(attr.(string))
			}
		}

		if d.HasChanges("directory_id", "function", "invocation_role", "url") {
			identityProviderDetails := &transfer.IdentityProviderDetails{}

			if attr, ok := d.GetOk("directory_id"); ok {
				identityProviderDetails.DirectoryId = aws.String(attr.(string))
			}

			if attr, ok := d.GetOk("function"); ok {
				identityProviderDetails.Function = aws.String(attr.(string))
			}

			if attr, ok := d.GetOk("invocation_role"); ok {
				identityProviderDetails.InvocationRole = aws.String(attr.(string))
			}

			if attr, ok := d.GetOk("url"); ok {
				identityProviderDetails.Url = aws.String(attr.(string))
			}

			input.IdentityProviderDetails = identityProviderDetails
		}

		if d.HasChange("logging_role") {
			input.LoggingRole = aws.String(d.Get("logging_role").(string))
		}

		if d.HasChange("post_authentication_login_banner") {
			input.PostAuthenticationLoginBanner = aws.String(d.Get("post_authentication_login_banner").(string))
		}

		if d.HasChange("pre_authentication_login_banner") {
			input.PreAuthenticationLoginBanner = aws.String(d.Get("pre_authentication_login_banner").(string))
		}

		if d.HasChange("protocol_details") {
			input.ProtocolDetails = expandProtocolDetails(d.Get("protocol_details").([]interface{}))
		}

		if d.HasChange("protocols") {
			input.Protocols = flex.ExpandStringSet(d.Get("protocols").(*schema.Set))
		}

		if d.HasChange("security_policy_name") {
			input.SecurityPolicyName = aws.String(d.Get("security_policy_name").(string))
		}

		// Per the docs it does not matter if this field has changed,
		// if the update passes this as empty the structured logging will be turned off,
		// so we need to always pass the new.
		input.StructuredLogDestinations = flex.ExpandStringSet(d.Get("structured_log_destinations").(*schema.Set))

		if d.HasChange("workflow_details") {
			input.WorkflowDetails = expandWorkflowDetails(d.Get("workflow_details").([]interface{}))
		}

		if offlineUpdate {
			if err := stopServer(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		if removeAddressAllocationIDs {
			input := &transfer.UpdateServerInput{
				EndpointDetails: &transfer.EndpointDetails{
					AddressAllocationIds: []*string{},
				},
				ServerId: aws.String(d.Id()),
			}

			if err := updateServer(ctx, conn, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "removing address allocation IDs: %s", err)
			}
		}

		if err := updateServer(ctx, conn, input); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if len(addressAllocationIDs) > 0 {
			input := &transfer.UpdateServerInput{
				EndpointDetails: &transfer.EndpointDetails{
					AddressAllocationIds: addressAllocationIDs,
				},
				ServerId: aws.String(d.Id()),
			}

			if err := updateServer(ctx, conn, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "adding address allocation IDs: %s", err)
			}
		}

		if offlineUpdate {
			if err := startServer(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceServerRead(ctx, d, meta)...)
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	if d.Get("force_destroy").(bool) && d.Get("identity_provider_type").(string) == transfer.IdentityProviderTypeServiceManaged {
		input := &transfer.ListUsersInput{
			ServerId: aws.String(d.Id()),
		}
		var errs *multierror.Error

		err := conn.ListUsersPagesWithContext(ctx, input, func(page *transfer.ListUsersOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, user := range page.Users {
				err := userDelete(ctx, conn, d.Id(), aws.StringValue(user.UserName), d.Timeout(schema.TimeoutDelete))

				if err != nil {
					errs = multierror.Append(errs, err)

					continue
				}
			}

			return !lastPage
		})

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("listing Transfer Server (%s) Users: %w", d.Id(), err))
		}

		err = errs.ErrorOrNil()

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[DEBUG] Deleting Transfer Server: (%s)", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 1*time.Minute,
		func() (interface{}, error) {
			return conn.DeleteServerWithContext(ctx, &transfer.DeleteServerInput{
				ServerId: aws.String(d.Id()),
			})
		},
		transfer.ErrCodeInvalidRequestException, "Unable to delete VPC endpoint")

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Server (%s): %s", d.Id(), err)
	}

	if _, err := waitServerDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Transfer Server (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandEndpointDetails(tfMap map[string]interface{}) *transfer.EndpointDetails {
	if tfMap == nil {
		return nil
	}

	apiObject := &transfer.EndpointDetails{}

	if v, ok := tfMap["address_allocation_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AddressAllocationIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["vpc_endpoint_id"].(string); ok && v != "" {
		apiObject.VpcEndpointId = aws.String(v)
	}

	if v, ok := tfMap["vpc_id"].(string); ok && v != "" {
		apiObject.VpcId = aws.String(v)
	}

	return apiObject
}

func flattenEndpointDetails(apiObject *transfer.EndpointDetails, securityGroupIDs []*string) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AddressAllocationIds; v != nil {
		tfMap["address_allocation_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.SecurityGroupIds; len(v) > 0 {
		tfMap["security_group_ids"] = aws.StringValueSlice(v)
	} else if len(securityGroupIDs) > 0 {
		tfMap["security_group_ids"] = aws.StringValueSlice(securityGroupIDs)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap["subnet_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.VpcEndpointId; v != nil {
		tfMap["vpc_endpoint_id"] = aws.StringValue(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap["vpc_id"] = aws.StringValue(v)
	}

	return tfMap
}

func expandProtocolDetails(m []interface{}) *transfer.ProtocolDetails {
	if len(m) < 1 || m[0] == nil {
		return nil
	}

	tfMap := m[0].(map[string]interface{})

	apiObject := &transfer.ProtocolDetails{}

	if v, ok := tfMap["as2_transports"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.As2Transports = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["passive_ip"].(string); ok && len(v) > 0 {
		apiObject.PassiveIp = aws.String(v)
	}

	if v, ok := tfMap["set_stat_option"].(string); ok && len(v) > 0 {
		apiObject.SetStatOption = aws.String(v)
	}

	if v, ok := tfMap["tls_session_resumption_mode"].(string); ok && len(v) > 0 {
		apiObject.TlsSessionResumptionMode = aws.String(v)
	}

	return apiObject
}

func flattenProtocolDetails(apiObject *transfer.ProtocolDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.As2Transports; v != nil {
		tfMap["as2_transports"] = aws.StringValueSlice(v)
	}

	if v := apiObject.PassiveIp; v != nil {
		tfMap["passive_ip"] = aws.StringValue(v)
	}

	if v := apiObject.SetStatOption; v != nil {
		tfMap["set_stat_option"] = aws.StringValue(v)
	}

	if v := apiObject.TlsSessionResumptionMode; v != nil {
		tfMap["tls_session_resumption_mode"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func expandWorkflowDetails(tfMap []interface{}) *transfer.WorkflowDetails {
	apiObject := &transfer.WorkflowDetails{
		OnPartialUpload: []*transfer.WorkflowDetail{},
		OnUpload:        []*transfer.WorkflowDetail{},
	}

	if len(tfMap) == 0 || tfMap[0] == nil {
		return apiObject
	}

	tfMapRaw := tfMap[0].(map[string]interface{})

	if v, ok := tfMapRaw["on_upload"].([]interface{}); ok && len(v) > 0 {
		apiObject.OnUpload = expandWorkflowDetail(v)
	}

	if v, ok := tfMapRaw["on_partial_upload"].([]interface{}); ok && len(v) > 0 {
		apiObject.OnPartialUpload = expandWorkflowDetail(v)
	}

	return apiObject
}

func flattenWorkflowDetails(apiObject *transfer.WorkflowDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.OnUpload; v != nil {
		tfMap["on_upload"] = flattenWorkflowDetail(v)
	}

	if v := apiObject.OnPartialUpload; v != nil {
		tfMap["on_partial_upload"] = flattenWorkflowDetail(v)
	}

	return []interface{}{tfMap}
}

func expandWorkflowDetail(tfList []interface{}) []*transfer.WorkflowDetail {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*transfer.WorkflowDetail

	for _, tfMapRaw := range tfList {
		tfMap, _ := tfMapRaw.(map[string]interface{})

		apiObject := &transfer.WorkflowDetail{}

		if v, ok := tfMap["execution_role"].(string); ok && v != "" {
			apiObject.ExecutionRole = aws.String(v)
		}

		if v, ok := tfMap["workflow_id"].(string); ok && v != "" {
			apiObject.WorkflowId = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenWorkflowDetail(apiObjects []*transfer.WorkflowDetail) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		flattenedObject := map[string]interface{}{}
		if v := apiObject.ExecutionRole; v != nil {
			flattenedObject["execution_role"] = aws.StringValue(v)
		}

		if v := apiObject.WorkflowId; v != nil {
			flattenedObject["workflow_id"] = aws.StringValue(v)
		}

		tfList = append(tfList, flattenedObject)
	}

	return tfList
}

func stopServer(ctx context.Context, conn *transfer.Transfer, serverID string, timeout time.Duration) error {
	input := &transfer.StopServerInput{
		ServerId: aws.String(serverID),
	}

	if _, err := conn.StopServerWithContext(ctx, input); err != nil {
		return fmt.Errorf("stopping Transfer Server (%s): %w", serverID, err)
	}

	if _, err := waitServerStopped(ctx, conn, serverID, timeout); err != nil {
		return fmt.Errorf("waiting for Transfer Server (%s) stop: %w", serverID, err)
	}

	return nil
}

func startServer(ctx context.Context, conn *transfer.Transfer, serverID string, timeout time.Duration) error {
	input := &transfer.StartServerInput{
		ServerId: aws.String(serverID),
	}

	if _, err := conn.StartServerWithContext(ctx, input); err != nil {
		return fmt.Errorf("starting Transfer Server (%s): %w", serverID, err)
	}

	if _, err := waitServerStarted(ctx, conn, serverID, timeout); err != nil {
		return fmt.Errorf("waiting for Transfer Server (%s) start: %w", serverID, err)
	}

	return nil
}

func updateServer(ctx context.Context, conn *transfer.Transfer, input *transfer.UpdateServerInput) error {
	// The Transfer API will return a state of ONLINE for a server before the
	// underlying VPC Endpoint is available and attempting to update the server
	// will return an error until that EC2 API process is complete:
	//   ConflictException: VPC Endpoint state is not yet available
	// To prevent accessing the EC2 API directly to check the VPC Endpoint
	// state, which can require confusing IAM permissions and have other
	// eventual consistency consideration, we retry only via the Transfer API.
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, tfec2.VPCEndpointCreationTimeout, func() (interface{}, error) {
		return conn.UpdateServerWithContext(ctx, input)
	}, transfer.ErrCodeConflictException, "VPC Endpoint state is not yet available")

	if err != nil {
		return fmt.Errorf("updating Transfer Server (%s): %w", aws.StringValue(input.ServerId), err)
	}

	return nil
}
