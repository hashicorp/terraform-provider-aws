// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
func resourceServer() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCertificate: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDomain: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.DomainS3,
				ValidateDiagFunc: enum.Validate[awstypes.Domain](),
			},
			names.AttrEndpoint: {
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
						names.AttrSecurityGroupIDs: {
							Type:          schema.TypeSet,
							Optional:      true,
							Computed:      true,
							Elem:          &schema.Schema{Type: schema.TypeString},
							ConflictsWith: []string{"endpoint_details.0.vpc_endpoint_id"},
						},
						names.AttrSubnetIDs: {
							Type:          schema.TypeSet,
							Optional:      true,
							Elem:          &schema.Schema{Type: schema.TypeString},
							ConflictsWith: []string{"endpoint_details.0.vpc_endpoint_id"},
						},
						names.AttrVPCEndpointID: {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"endpoint_details.0.address_allocation_ids", "endpoint_details.0.security_group_ids", "endpoint_details.0.subnet_ids", "endpoint_details.0.vpc_id"},
						},
						names.AttrVPCID: {
							Type:          schema.TypeString,
							Optional:      true,
							ValidateFunc:  validation.NoZeroValues,
							ConflictsWith: []string{"endpoint_details.0.vpc_endpoint_id"},
						},
					},
				},
			},
			names.AttrEndpointType: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.EndpointTypePublic,
				ValidateDiagFunc: enum.Validate[awstypes.EndpointType](),
			},
			names.AttrForceDestroy: {
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.IdentityProviderTypeServiceManaged,
				ValidateDiagFunc: enum.Validate[awstypes.IdentityProviderType](),
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
				ValidateFunc: validation.StringLenBetween(0, 4096),
			},
			"pre_authentication_login_banner": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(0, 4096),
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
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.As2Transport](),
							},
						},
						"passive_ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 15),
						},
						"set_stat_option": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SetStatOption](),
						},
						"tls_session_resumption_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TlsSessionResumptionMode](),
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
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.Protocol](),
				},
			},
			"s3_storage_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"directory_listing_optimization": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DirectoryListingOptimization](),
						},
					},
				},
			},
			"security_policy_name": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          securityPolicyName2018_11,
				ValidateDiagFunc: enum.Validate[securityPolicyName](),
			},
			"sftp_authentication_methods": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.SftpAuthenticationMethods](),
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
			names.AttrURL: {
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
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	input := &transfer.CreateServerInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrCertificate); ok {
		input.Certificate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("directory_id"); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &awstypes.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.DirectoryId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDomain); ok {
		input.Domain = awstypes.Domain(v.(string))
	}

	var addressAllocationIDs []string

	if v, ok := d.GetOk("endpoint_details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EndpointDetails = expandEndpointDetails(v.([]interface{})[0].(map[string]interface{}))

		// Prevent the following error: InvalidRequestException: AddressAllocationIds cannot be set in CreateServer
		// Reference: https://docs.aws.amazon.com/transfer/latest/userguide/API_EndpointDetails.html#TransferFamily-Type-EndpointDetails-AddressAllocationIds
		addressAllocationIDs = input.EndpointDetails.AddressAllocationIds
		input.EndpointDetails.AddressAllocationIds = nil
	}

	if v, ok := d.GetOk(names.AttrEndpointType); ok {
		input.EndpointType = awstypes.EndpointType(v.(string))
	}

	if v, ok := d.GetOk("function"); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &awstypes.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.Function = aws.String(v.(string))
	}

	if v, ok := d.GetOk("host_key"); ok {
		input.HostKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("identity_provider_type"); ok {
		input.IdentityProviderType = awstypes.IdentityProviderType(v.(string))
	}

	if v, ok := d.GetOk("invocation_role"); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &awstypes.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.InvocationRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sftp_authentication_methods"); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &awstypes.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.SftpAuthenticationMethods = awstypes.SftpAuthenticationMethods(v.(string))
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
		input.Protocols = flex.ExpandStringyValueSet[awstypes.Protocol](d.Get("protocols").(*schema.Set))
	}

	if v, ok := d.GetOk("s3_storage_options"); ok && len(v.([]interface{})) > 0 {
		input.S3StorageOptions = expandS3StorageOptions(v.([]interface{}))
	}

	if v, ok := d.GetOk("security_policy_name"); ok {
		input.SecurityPolicyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("structured_log_destinations"); ok && v.(*schema.Set).Len() > 0 {
		input.StructuredLogDestinations = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrURL); ok {
		if input.IdentityProviderDetails == nil {
			input.IdentityProviderDetails = &awstypes.IdentityProviderDetails{}
		}

		input.IdentityProviderDetails.Url = aws.String(v.(string))
	}

	if v, ok := d.GetOk("workflow_details"); ok && len(v.([]interface{})) > 0 {
		input.WorkflowDetails = expandWorkflowDetails(v.([]interface{}))
	}

	output, err := conn.CreateServer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Server: %s", err)
	}

	d.SetId(aws.ToString(output.ServerId))

	if _, err := waitServerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Transfer Server (%s) create: %s", d.Id(), err)
	}

	// AddressAllocationIds is only valid in the UpdateServer API.
	if len(addressAllocationIDs) > 0 {
		if err := stopServer(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &transfer.UpdateServerInput{
			EndpointDetails: &awstypes.EndpointDetails{
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
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	output, err := findServerByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Server (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Server (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrCertificate, output.Certificate)
	if output.IdentityProviderDetails != nil {
		d.Set("directory_id", output.IdentityProviderDetails.DirectoryId)
	} else {
		d.Set("directory_id", "")
	}
	d.Set(names.AttrDomain, output.Domain)
	d.Set(names.AttrEndpoint, meta.(*conns.AWSClient).RegionalHostname(ctx, fmt.Sprintf("%s.server.transfer", d.Id())))
	if output.EndpointDetails != nil {
		securityGroupIDs := make([]*string, 0)

		// Security Group IDs are not returned for VPC endpoints.
		if output.EndpointType == awstypes.EndpointTypeVpc && len(output.EndpointDetails.SecurityGroupIds) == 0 {
			vpcEndpointID := aws.ToString(output.EndpointDetails.VpcEndpointId)
			output, err := tfec2.FindVPCEndpointByID(ctx, meta.(*conns.AWSClient).EC2Client(ctx), vpcEndpointID)

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
	d.Set(names.AttrEndpointType, output.EndpointType)
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
	if output.IdentityProviderDetails != nil {
		d.Set("sftp_authentication_methods", output.IdentityProviderDetails.SftpAuthenticationMethods)
	} else {
		d.Set("sftp_authentication_methods", "")
	}
	d.Set("logging_role", output.LoggingRole)
	d.Set("post_authentication_login_banner", output.PostAuthenticationLoginBanner)
	d.Set("pre_authentication_login_banner", output.PreAuthenticationLoginBanner)
	if err := d.Set("protocol_details", flattenProtocolDetails(output.ProtocolDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting protocol_details: %s", err)
	}
	d.Set("protocols", output.Protocols)
	if err := d.Set("s3_storage_options", flattenS3StorageOptions(output.S3StorageOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3_storage_options: %s", err)
	}
	d.Set("security_policy_name", output.SecurityPolicyName)
	d.Set("structured_log_destinations", output.StructuredLogDestinations)
	if output.IdentityProviderDetails != nil {
		d.Set(names.AttrURL, output.IdentityProviderDetails.Url)
	} else {
		d.Set(names.AttrURL, "")
	}
	if err := d.Set("workflow_details", flattenWorkflowDetails(output.WorkflowDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workflow_details: %s", err)
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		var newEndpointTypeVpc bool
		var oldEndpointTypeVpc bool

		old, new := d.GetChange(names.AttrEndpointType)

		if old, new := old.(string), new.(string); new == string(awstypes.EndpointTypeVpc) {
			newEndpointTypeVpc = true
			oldEndpointTypeVpc = old == new
		}

		var addressAllocationIDs []string
		var offlineUpdate bool
		var removeAddressAllocationIDs bool

		input := &transfer.UpdateServerInput{
			ServerId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrCertificate) {
			input.Certificate = aws.String(d.Get(names.AttrCertificate).(string))
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
						input.EndpointDetails.AddressAllocationIds = []string{}
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
						input.EndpointDetails.SubnetIds = []string{}
					}
				}
			}

			// You can edit the SecurityGroupIds property in the UpdateServer API only if you are changing the EndpointType from PUBLIC or VPC_ENDPOINT to VPC.
			// To change security groups associated with your server's VPC endpoint after creation, use the Amazon EC2 ModifyVpcEndpoint API.
			if d.HasChange("endpoint_details.0.security_group_ids") && newEndpointTypeVpc && oldEndpointTypeVpc {
				conn := meta.(*conns.AWSClient).EC2Client(ctx)

				vpcEndpointID := d.Get("endpoint_details.0.vpc_endpoint_id").(string)
				input := &ec2.ModifyVpcEndpointInput{
					VpcEndpointId: aws.String(vpcEndpointID),
				}

				old, new := d.GetChange("endpoint_details.0.security_group_ids")

				if add := flex.ExpandStringValueSet(new.(*schema.Set).Difference(old.(*schema.Set))); len(add) > 0 {
					input.AddSecurityGroupIds = add
				}

				if del := flex.ExpandStringValueSet(old.(*schema.Set).Difference(new.(*schema.Set))); len(del) > 0 {
					input.RemoveSecurityGroupIds = del
				}

				_, err := conn.ModifyVpcEndpoint(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying Transfer Server (%s) VPC Endpoint (%s): %s", d.Id(), vpcEndpointID, err)
				}

				if _, err := tfec2.WaitVPCEndpointAvailable(ctx, conn, vpcEndpointID, tfec2.VPCEndpointCreationTimeout); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Transfer Server (%s) VPC Endpoint (%s) update: %s", d.Id(), vpcEndpointID, err)
				}
			}
		}

		if d.HasChange(names.AttrEndpointType) {
			input.EndpointType = awstypes.EndpointType(d.Get(names.AttrEndpointType).(string))

			// Prevent the following error: InvalidRequestException: Server must be OFFLINE to change EndpointType
			offlineUpdate = true
		}

		if d.HasChange("host_key") {
			if attr, ok := d.GetOk("host_key"); ok {
				input.HostKey = aws.String(attr.(string))
			}
		}

		if d.HasChanges("directory_id", "function", "invocation_role", "sftp_authentication_methods", names.AttrURL) {
			identityProviderDetails := &awstypes.IdentityProviderDetails{}

			if attr, ok := d.GetOk("directory_id"); ok {
				identityProviderDetails.DirectoryId = aws.String(attr.(string))
			}

			if attr, ok := d.GetOk("function"); ok {
				identityProviderDetails.Function = aws.String(attr.(string))
			}

			if attr, ok := d.GetOk("invocation_role"); ok {
				identityProviderDetails.InvocationRole = aws.String(attr.(string))
			}

			if attr, ok := d.GetOk("sftp_authentication_methods"); ok {
				identityProviderDetails.SftpAuthenticationMethods = awstypes.SftpAuthenticationMethods(attr.(string))
			}

			if attr, ok := d.GetOk(names.AttrURL); ok {
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
			input.Protocols = flex.ExpandStringyValueSet[awstypes.Protocol](d.Get("protocols").(*schema.Set))
		}

		if d.HasChange("s3_storage_options") {
			input.S3StorageOptions = expandS3StorageOptions(d.Get("s3_storage_options").([]interface{}))
		}

		if d.HasChange("security_policy_name") {
			input.SecurityPolicyName = aws.String(d.Get("security_policy_name").(string))
		}

		// Per the docs it does not matter if this field has changed,
		// if the update passes this as empty the structured logging will be turned off,
		// so we need to always pass the new.
		input.StructuredLogDestinations = flex.ExpandStringValueSet(d.Get("structured_log_destinations").(*schema.Set))

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
				EndpointDetails: &awstypes.EndpointDetails{
					AddressAllocationIds: []string{},
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
				EndpointDetails: &awstypes.EndpointDetails{
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
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	if d.Get(names.AttrForceDestroy).(bool) && d.Get("identity_provider_type").(string) == string(awstypes.IdentityProviderTypeServiceManaged) {
		input := &transfer.ListUsersInput{
			ServerId: aws.String(d.Id()),
		}
		var errs []error

		pages := transfer.NewListUsersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil {
				errs = append(errs, fmt.Errorf("listing Transfer Server (%s) Users: %w", d.Id(), err))
				continue
			}

			for _, user := range page.Users {
				err := userDelete(ctx, conn, d.Id(), aws.ToString(user.UserName), d.Timeout(schema.TimeoutDelete))

				if err != nil {
					errs = append(errs, err)
					continue
				}
			}
		}

		if errs != nil {
			return sdkdiag.AppendFromErr(diags, errors.Join(errs...))
		}
	}

	log.Printf("[DEBUG] Deleting Transfer Server: (%s)", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidRequestException](ctx, 1*time.Minute,
		func() (interface{}, error) {
			return conn.DeleteServer(ctx, &transfer.DeleteServerInput{
				ServerId: aws.String(d.Id()),
			})
		}, "Unable to delete VPC endpoint")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func stopServer(ctx context.Context, conn *transfer.Client, serverID string, timeout time.Duration) error {
	input := &transfer.StopServerInput{
		ServerId: aws.String(serverID),
	}

	if _, err := conn.StopServer(ctx, input); err != nil {
		return fmt.Errorf("stopping Transfer Server (%s): %w", serverID, err)
	}

	if _, err := waitServerStopped(ctx, conn, serverID, timeout); err != nil {
		return fmt.Errorf("waiting for Transfer Server (%s) stop: %w", serverID, err)
	}

	return nil
}

func startServer(ctx context.Context, conn *transfer.Client, serverID string, timeout time.Duration) error {
	input := &transfer.StartServerInput{
		ServerId: aws.String(serverID),
	}

	if _, err := conn.StartServer(ctx, input); err != nil {
		return fmt.Errorf("starting Transfer Server (%s): %w", serverID, err)
	}

	if _, err := waitServerStarted(ctx, conn, serverID, timeout); err != nil {
		return fmt.Errorf("waiting for Transfer Server (%s) start: %w", serverID, err)
	}

	return nil
}

func updateServer(ctx context.Context, conn *transfer.Client, input *transfer.UpdateServerInput) error {
	// The Transfer API will return a state of ONLINE for a server before the
	// underlying VPC Endpoint is available and attempting to update the server
	// will return an error until that EC2 API process is complete:
	//   ConflictException: VPC Endpoint state is not yet available
	// To prevent accessing the EC2 API directly to check the VPC Endpoint
	// state, which can require confusing IAM permissions and have other
	// eventual consistency consideration, we retry only via the Transfer API.
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ConflictException](ctx, tfec2.VPCEndpointCreationTimeout, func() (interface{}, error) {
		return conn.UpdateServer(ctx, input)
	}, "VPC Endpoint state is not yet available")

	if err != nil {
		return fmt.Errorf("updating Transfer Server (%s): %w", aws.ToString(input.ServerId), err)
	}

	return nil
}

func findServerByID(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedServer, error) {
	input := &transfer.DescribeServerInput{
		ServerId: aws.String(id),
	}

	output, err := conn.DescribeServer(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Server == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Server, nil
}

func statusServer(ctx context.Context, conn *transfer.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findServerByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitServerCreated(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedServer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StateStarting),
		Target:  enum.Slice(awstypes.StateOnline),
		Refresh: statusServer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func waitServerDeleted(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedServer, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StateOffline, awstypes.StateOnline, awstypes.StateStarting, awstypes.StateStopping, awstypes.StateStartFailed, awstypes.StateStopFailed),
		Target:  []string{},
		Refresh: statusServer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func waitServerStarted(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedServer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StateStarting, awstypes.StateOffline, awstypes.StateStopping),
		Target:  enum.Slice(awstypes.StateOnline),
		Refresh: statusServer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func waitServerStopped(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedServer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StateStarting, awstypes.StateOnline, awstypes.StateStopping),
		Target:  enum.Slice(awstypes.StateOffline),
		Refresh: statusServer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func expandEndpointDetails(tfMap map[string]interface{}) *awstypes.EndpointDetails {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EndpointDetails{}

	if v, ok := tfMap["address_allocation_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AddressAllocationIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrVPCEndpointID].(string); ok && v != "" {
		apiObject.VpcEndpointId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVPCID].(string); ok && v != "" {
		apiObject.VpcId = aws.String(v)
	}

	return apiObject
}

func flattenEndpointDetails(apiObject *awstypes.EndpointDetails, securityGroupIDs []*string) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AddressAllocationIds; v != nil {
		tfMap["address_allocation_ids"] = v
	}

	if v := apiObject.SecurityGroupIds; len(v) > 0 {
		tfMap[names.AttrSecurityGroupIDs] = v
	} else if len(securityGroupIDs) > 0 {
		tfMap[names.AttrSecurityGroupIDs] = aws.ToStringSlice(securityGroupIDs)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = v
	}

	if v := apiObject.VpcEndpointId; v != nil {
		tfMap[names.AttrVPCEndpointID] = aws.ToString(v)
	}

	if v := apiObject.VpcId; v != nil {
		tfMap[names.AttrVPCID] = aws.ToString(v)
	}

	return tfMap
}

func expandProtocolDetails(tfList []interface{}) *awstypes.ProtocolDetails {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &awstypes.ProtocolDetails{}

	if v, ok := tfMap["as2_transports"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.As2Transports = flex.ExpandStringyValueSet[awstypes.As2Transport](v)
	}

	if v, ok := tfMap["passive_ip"].(string); ok && len(v) > 0 {
		apiObject.PassiveIp = aws.String(v)
	}

	if v, ok := tfMap["set_stat_option"].(string); ok && len(v) > 0 {
		apiObject.SetStatOption = awstypes.SetStatOption(v)
	}

	if v, ok := tfMap["tls_session_resumption_mode"].(string); ok && len(v) > 0 {
		apiObject.TlsSessionResumptionMode = awstypes.TlsSessionResumptionMode(v)
	}

	return apiObject
}

func flattenProtocolDetails(apiObject *awstypes.ProtocolDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"set_stat_option":             apiObject.SetStatOption,
		"tls_session_resumption_mode": apiObject.TlsSessionResumptionMode,
	}

	if v := apiObject.As2Transports; v != nil {
		tfMap["as2_transports"] = v
	}

	if v := apiObject.PassiveIp; v != nil {
		tfMap["passive_ip"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func expandS3StorageOptions(tfList []interface{}) *awstypes.S3StorageOptions {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &awstypes.S3StorageOptions{}

	if v, ok := tfMap["directory_listing_optimization"].(string); ok && len(v) > 0 {
		apiObject.DirectoryListingOptimization = awstypes.DirectoryListingOptimization(v)
	}

	return apiObject
}

func flattenS3StorageOptions(apiObject *awstypes.S3StorageOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"directory_listing_optimization": apiObject.DirectoryListingOptimization,
	}

	return []interface{}{tfMap}
}

func expandWorkflowDetails(tfList []interface{}) *awstypes.WorkflowDetails {
	apiObject := &awstypes.WorkflowDetails{
		OnPartialUpload: []awstypes.WorkflowDetail{},
		OnUpload:        []awstypes.WorkflowDetail{},
	}

	if len(tfList) == 0 || tfList[0] == nil {
		return apiObject
	}

	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["on_upload"].([]interface{}); ok && len(v) > 0 {
		apiObject.OnUpload = expandWorkflowDetail(v)
	}

	if v, ok := tfMap["on_partial_upload"].([]interface{}); ok && len(v) > 0 {
		apiObject.OnPartialUpload = expandWorkflowDetail(v)
	}

	return apiObject
}

func flattenWorkflowDetails(apiObject *awstypes.WorkflowDetails) []interface{} {
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

func expandWorkflowDetail(tfList []interface{}) []awstypes.WorkflowDetail {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.WorkflowDetail

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.WorkflowDetail{}

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

func flattenWorkflowDetail(apiObjects []awstypes.WorkflowDetail) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if v := apiObject.ExecutionRole; v != nil {
			tfMap["execution_role"] = aws.ToString(v)
		}

		if v := apiObject.WorkflowId; v != nil {
			tfMap["workflow_id"] = aws.ToString(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

type securityPolicyName string

const (
	securityPolicyName2018_11             securityPolicyName = "TransferSecurityPolicy-2018-11"
	securityPolicyName2020_06             securityPolicyName = "TransferSecurityPolicy-2020-06"
	securityPolicyName2022_03             securityPolicyName = "TransferSecurityPolicy-2022-03"
	securityPolicyName2023_05             securityPolicyName = "TransferSecurityPolicy-2023-05"
	securityPolicyName2024_01             securityPolicyName = "TransferSecurityPolicy-2024-01"
	securityPolicyNameFIPS_2020_06        securityPolicyName = "TransferSecurityPolicy-FIPS-2020-06"
	securityPolicyNameFIPS_2023_05        securityPolicyName = "TransferSecurityPolicy-FIPS-2023-05"
	securityPolicyNameFIPS_2024_01        securityPolicyName = "TransferSecurityPolicy-FIPS-2024-01"
	securityPolicyNameFIPS_2024_05        securityPolicyName = "TransferSecurityPolicy-FIPS-2024-05"
	securityPolicyNamePQ_SSH_2023_04      securityPolicyName = "TransferSecurityPolicy-PQ-SSH-Experimental-2023-04"
	securityPolicyNamePQ_SSH_FIPS_2023_04 securityPolicyName = "TransferSecurityPolicy-PQ-SSH-FIPS-Experimental-2023-04"
	securityPolicyNameRestricted_2018_11  securityPolicyName = "TransferSecurityPolicy-Restricted-2018-11"
	securityPolicyNameRestricted_2020_06  securityPolicyName = "TransferSecurityPolicy-Restricted-2020-06"
)

func (securityPolicyName) Values() []securityPolicyName {
	return []securityPolicyName{
		securityPolicyName2018_11,
		securityPolicyName2020_06,
		securityPolicyName2022_03,
		securityPolicyName2023_05,
		securityPolicyName2024_01,
		securityPolicyNameFIPS_2020_06,
		securityPolicyNameFIPS_2023_05,
		securityPolicyNameFIPS_2024_01,
		securityPolicyNameFIPS_2024_05,
		securityPolicyNamePQ_SSH_2023_04,
		securityPolicyNamePQ_SSH_FIPS_2023_04,
		securityPolicyNameRestricted_2018_11,
		securityPolicyNameRestricted_2020_06,
	}
}
