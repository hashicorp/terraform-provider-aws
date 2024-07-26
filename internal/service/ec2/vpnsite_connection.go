// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpn_connection", name="VPN Connection")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceVPNConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPNConnectionCreate,
		ReadWithoutTimeout:   resourceVPNConnectionRead,
		UpdateWithoutTimeout: resourceVPNConnectionUpdate,
		DeleteWithoutTimeout: resourceVPNConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_attachment_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_gateway_configuration": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"customer_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enable_acceleration": {
				Type:         schema.TypeBool,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				RequiredWith: []string{names.AttrTransitGatewayID},
			},
			"local_ipv4_network_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsCIDRNetwork(0, 32),
			},
			"local_ipv6_network_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsCIDRNetwork(0, 128),
			},
			"outside_ip_address_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(outsideIPAddressType_Values(), false),
			},
			"remote_ipv4_network_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsCIDRNetwork(0, 32),
			},
			"remote_ipv6_network_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsCIDRNetwork(0, 128),
			},
			"routes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrSource: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrState: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"static_routes_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTransitGatewayAttachmentID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTransitGatewayID: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"vpn_gateway_id"},
			},
			"transport_transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tunnel_inside_ip_version": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TunnelInsideIpVersion](),
			},
			"tunnel1_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel1_bgp_asn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel1_bgp_holdtime": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tunnel1_cgw_inside_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel1_dpd_timeout_action": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(vpnTunnelOptionsDPDTimeoutAction_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == defaultVPNTunnelOptionsDPDTimeoutAction && new == "" {
						return true
					}
					return false
				},
			},
			"tunnel1_dpd_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(30),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsDPDTimeoutSeconds) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel1_enable_tunnel_lifecycle_control": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"tunnel1_ike_versions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(vpnTunnelOptionsIKEVersion_Values(), false),
				},
			},
			"tunnel1_inside_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validVPNConnectionTunnelInsideCIDR(),
			},
			"tunnel1_inside_ipv6_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validVPNConnectionTunnelInsideIPv6CIDR(),
			},
			"tunnel1_log_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_log_options": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"log_group_arn": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"log_output_format": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(vpnTunnelCloudWatchLogOutputFormat_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			"tunnel1_phase1_dh_group_numbers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"tunnel1_phase1_encryption_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(vpnTunnelOptionsPhase1EncryptionAlgorithm_Values(), false),
				},
			},
			"tunnel1_phase1_integrity_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(vpnTunnelOptionsPhase1IntegrityAlgorithm_Values(), false),
				},
			},
			"tunnel1_phase1_lifetime_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(900, 28800),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsPhase1LifetimeSeconds) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel1_phase2_dh_group_numbers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"tunnel1_phase2_encryption_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(vpnTunnelOptionsPhase2EncryptionAlgorithm_Values(), false),
				},
			},
			"tunnel1_phase2_integrity_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(vpnTunnelOptionsPhase2IntegrityAlgorithm_Values(), false),
				},
			},
			"tunnel1_phase2_lifetime_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(900, 3600),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsPhase2LifetimeSeconds) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel1_preshared_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				Computed:     true,
				ValidateFunc: validVPNConnectionTunnelPreSharedKey(),
			},
			"tunnel1_rekey_fuzz_percentage": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsRekeyFuzzPercentage) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel1_rekey_margin_time_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(60, 1800),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsRekeyMarginTimeSeconds) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel1_replay_window_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(64, 2048),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsReplayWindowSize) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel1_startup_action": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(vpnTunnelOptionsStartupAction_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == defaultVPNTunnelOptionsStartupAction && new == "" {
						return true
					}
					return false
				},
			},
			"tunnel1_vgw_inside_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel2_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel2_bgp_asn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel2_bgp_holdtime": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tunnel2_cgw_inside_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tunnel2_dpd_timeout_action": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(vpnTunnelOptionsDPDTimeoutAction_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == defaultVPNTunnelOptionsDPDTimeoutAction && new == "" {
						return true
					}
					return false
				},
			},
			"tunnel2_dpd_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(30),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsDPDTimeoutSeconds) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel2_enable_tunnel_lifecycle_control": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"tunnel2_ike_versions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(vpnTunnelOptionsIKEVersion_Values(), false),
				},
			},
			"tunnel2_inside_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validVPNConnectionTunnelInsideCIDR(),
			},
			"tunnel2_inside_ipv6_cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validVPNConnectionTunnelInsideIPv6CIDR(),
			},
			"tunnel2_log_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_log_options": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"log_group_arn": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"log_output_format": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(vpnTunnelCloudWatchLogOutputFormat_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			"tunnel2_phase1_dh_group_numbers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"tunnel2_phase1_encryption_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(vpnTunnelOptionsPhase1EncryptionAlgorithm_Values(), false),
				},
			},
			"tunnel2_phase1_integrity_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(vpnTunnelOptionsPhase1IntegrityAlgorithm_Values(), false),
				},
			},
			"tunnel2_phase1_lifetime_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(900, 28800),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsPhase1LifetimeSeconds) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel2_phase2_dh_group_numbers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"tunnel2_phase2_encryption_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(vpnTunnelOptionsPhase2EncryptionAlgorithm_Values(), false),
				},
			},
			"tunnel2_phase2_integrity_algorithms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(vpnTunnelOptionsPhase2IntegrityAlgorithm_Values(), false),
				},
			},
			"tunnel2_phase2_lifetime_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(900, 3600),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsPhase2LifetimeSeconds) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel2_preshared_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				Computed:     true,
				ValidateFunc: validVPNConnectionTunnelPreSharedKey(),
			},
			"tunnel2_rekey_fuzz_percentage": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsRekeyFuzzPercentage) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel2_rekey_margin_time_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(60, 1800),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsRekeyMarginTimeSeconds) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel2_replay_window_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(64, 2048),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == strconv.Itoa(defaultVPNTunnelOptionsReplayWindowSize) && new == "0" {
						return true
					}
					return false
				},
			},
			"tunnel2_startup_action": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(vpnTunnelOptionsStartupAction_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == defaultVPNTunnelOptionsStartupAction && new == "" {
						return true
					}
					return false
				},
			},
			"tunnel2_vgw_inside_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrType: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(vpnConnectionType_Values(), false),
			},
			"vgw_telemetry": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accepted_route_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrCertificateARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_status_change": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"outside_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatusMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"vpn_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrTransitGatewayID},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			customizeDiffValidateOutsideIPAddressType,
			verify.SetTagsDiff,
		),
	}
}

// https://docs.aws.amazon.com/vpn/latest/s2svpn/VPNTunnels.html.
var (
	defaultVPNTunnelOptionsDPDTimeoutAction             = vpnTunnelOptionsDPDTimeoutActionClear
	defaultVPNTunnelOptionsDPDTimeoutSeconds            = 30
	defaultVPNTunnelOptionsEnableTunnelLifecycleControl = false
	defaultVPNTunnelOptionsIKEVersions                  = []string{vpnTunnelOptionsIKEVersion1, vpnTunnelOptionsIKEVersion2}
	defaultVPNTunnelOptionsPhase1DHGroupNumbers         = []int{2, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}
	defaultVPNTunnelOptionsPhase1EncryptionAlgorithms   = []string{
		vpnTunnelOptionsPhase1EncryptionAlgorithmAES128,
		vpnTunnelOptionsPhase1EncryptionAlgorithmAES256,
		vpnTunnelOptionsPhase1EncryptionAlgorithmAES128_GCM_16,
		vpnTunnelOptionsPhase1EncryptionAlgorithmAES256_GCM_16,
	}
	defaultVPNTunnelOptionsPhase1IntegrityAlgorithms = []string{
		vpnTunnelOptionsPhase1IntegrityAlgorithmSHA1,
		vpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_256,
		vpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_384,
		vpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_512,
	}
	defaultVPNTunnelOptionsPhase1LifetimeSeconds      = 28800
	defaultVPNTunnelOptionsPhase2DHGroupNumbers       = []int{2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}
	defaultVPNTunnelOptionsPhase2EncryptionAlgorithms = []string{
		vpnTunnelOptionsPhase2EncryptionAlgorithmAES128,
		vpnTunnelOptionsPhase2EncryptionAlgorithmAES256,
		vpnTunnelOptionsPhase2EncryptionAlgorithmAES128_GCM_16,
		vpnTunnelOptionsPhase2EncryptionAlgorithmAES256_GCM_16,
	}
	defaultVPNTunnelOptionsPhase2IntegrityAlgorithms = []string{
		vpnTunnelOptionsPhase2IntegrityAlgorithmSHA1,
		vpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_256,
		vpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_384,
		vpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_512,
	}
	defaultVPNTunnelOptionsPhase2LifetimeSeconds  = 3600
	defaultVPNTunnelOptionsRekeyFuzzPercentage    = 100
	defaultVPNTunnelOptionsRekeyMarginTimeSeconds = 540
	defaultVPNTunnelOptionsReplayWindowSize       = 1024
	defaultVPNTunnelOptionsStartupAction          = vpnTunnelOptionsStartupActionAdd
)

func resourceVPNConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVpnConnectionInput{
		CustomerGatewayId: aws.String(d.Get("customer_gateway_id").(string)),
		Options:           expandVPNConnectionOptionsSpecification(d),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpnConnection),
		Type:              aws.String(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrTransitGatewayID); ok {
		input.TransitGatewayId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpn_gateway_id"); ok {
		input.VpnGatewayId = aws.String(v.(string))
	}

	output, err := conn.CreateVpnConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPN Connection: %s", err)
	}

	d.SetId(aws.ToString(output.VpnConnection.VpnConnectionId))

	if _, err := waitVPNConnectionCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPN Connection (%s) create: %s", d.Id(), err)
	}

	// Read off the API to populate our RO fields.
	return append(diags, resourceVPNConnectionRead(ctx, d, meta)...)
}

func resourceVPNConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpnConnection, err := findVPNConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPN Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPN Connection (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpn-connection/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("core_network_arn", vpnConnection.CoreNetworkArn)
	d.Set("core_network_attachment_arn", vpnConnection.CoreNetworkAttachmentArn)
	d.Set("customer_gateway_id", vpnConnection.CustomerGatewayId)
	d.Set(names.AttrType, vpnConnection.Type)
	d.Set("vpn_gateway_id", vpnConnection.VpnGatewayId)

	if v := vpnConnection.TransitGatewayId; v != nil {
		input := &ec2.DescribeTransitGatewayAttachmentsInput{
			Filters: newAttributeFilterList(map[string]string{
				"resource-id":        d.Id(),
				"resource-type":      string(awstypes.TransitGatewayAttachmentResourceTypeVpn),
				"transit-gateway-id": aws.ToString(v),
			}),
		}

		output, err := findTransitGatewayAttachment(ctx, conn, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 VPN Connection (%s) Transit Gateway Attachment: %s", d.Id(), err)
		}

		d.Set(names.AttrTransitGatewayAttachmentID, output.TransitGatewayAttachmentId)
		d.Set(names.AttrTransitGatewayID, v)
	} else {
		d.Set(names.AttrTransitGatewayAttachmentID, nil)
		d.Set(names.AttrTransitGatewayID, nil)
	}

	if err := d.Set("routes", flattenVPNStaticRoutes(vpnConnection.Routes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting routes: %s", err)
	}

	if err := d.Set("vgw_telemetry", flattenVGWTelemetries(vpnConnection.VgwTelemetry)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vgw_telemetry: %s", err)
	}

	setTagsOut(ctx, vpnConnection.Tags)

	if v := vpnConnection.Options; v != nil {
		d.Set("enable_acceleration", v.EnableAcceleration)
		d.Set("local_ipv4_network_cidr", v.LocalIpv4NetworkCidr)
		d.Set("local_ipv6_network_cidr", v.LocalIpv6NetworkCidr)
		d.Set("outside_ip_address_type", v.OutsideIpAddressType)
		d.Set("remote_ipv4_network_cidr", v.RemoteIpv4NetworkCidr)
		d.Set("remote_ipv6_network_cidr", v.RemoteIpv6NetworkCidr)
		d.Set("static_routes_only", v.StaticRoutesOnly)
		d.Set("transport_transit_gateway_attachment_id", v.TransportTransitGatewayAttachmentId)
		d.Set("tunnel_inside_ip_version", v.TunnelInsideIpVersion)

		for i, prefix := range []string{"tunnel1_", "tunnel2_"} {
			if len(v.TunnelOptions) > i {
				if err := flattenTunnelOption(d, prefix, v.TunnelOptions[i]); err != nil {
					return sdkdiag.AppendErrorf(diags, "reading EC2 VPN Connection (%s): %s", d.Id(), err)
				}
			}
		}
	} else {
		d.Set("enable_acceleration", nil)
		d.Set("local_ipv4_network_cidr", nil)
		d.Set("local_ipv6_network_cidr", nil)
		d.Set("outside_ip_address_type", nil)
		d.Set("remote_ipv4_network_cidr", nil)
		d.Set("remote_ipv6_network_cidr", nil)
		d.Set("static_routes_only", nil)
		d.Set("transport_transit_gateway_attachment_id", nil)
		d.Set("tunnel_inside_ip_version", nil)
	}

	d.Set("customer_gateway_configuration", vpnConnection.CustomerGatewayConfiguration)

	tunnelInfo, err := customerGatewayConfigurationToTunnelInfo(
		aws.ToString(vpnConnection.CustomerGatewayConfiguration),
		d.Get("tunnel1_preshared_key").(string), // Not currently available during import
		d.Get("tunnel1_inside_cidr").(string),
		d.Get("tunnel1_inside_ipv6_cidr").(string),
	)

	if err == nil {
		d.Set("tunnel1_address", tunnelInfo.Tunnel1Address)
		d.Set("tunnel1_bgp_asn", tunnelInfo.Tunnel1BGPASN)
		d.Set("tunnel1_bgp_holdtime", tunnelInfo.Tunnel1BGPHoldTime)
		d.Set("tunnel1_cgw_inside_address", tunnelInfo.Tunnel1CgwInsideAddress)
		d.Set("tunnel1_preshared_key", tunnelInfo.Tunnel1PreSharedKey)
		d.Set("tunnel1_vgw_inside_address", tunnelInfo.Tunnel1VgwInsideAddress)
		d.Set("tunnel2_address", tunnelInfo.Tunnel2Address)
		d.Set("tunnel2_bgp_asn", tunnelInfo.Tunnel2BGPASN)
		d.Set("tunnel2_bgp_holdtime", tunnelInfo.Tunnel2BGPHoldTime)
		d.Set("tunnel2_cgw_inside_address", tunnelInfo.Tunnel2CgwInsideAddress)
		d.Set("tunnel2_preshared_key", tunnelInfo.Tunnel2PreSharedKey)
		d.Set("tunnel2_vgw_inside_address", tunnelInfo.Tunnel2VgwInsideAddress)
	} else {
		// This element is present in the DescribeVpnConnections response only if the VPN connection is in the pending or available state.
		if vpnConnection.CustomerGatewayConfiguration != nil {
			log.Printf("[ERROR] Error unmarshaling Customer Gateway XML configuration for (%s): %s", d.Id(), err)
		}

		d.Set("tunnel1_address", nil)
		d.Set("tunnel1_bgp_asn", nil)
		d.Set("tunnel1_bgp_holdtime", nil)
		d.Set("tunnel1_cgw_inside_address", nil)
		d.Set("tunnel1_preshared_key", nil)
		d.Set("tunnel1_vgw_inside_address", nil)
		d.Set("tunnel2_address", nil)
		d.Set("tunnel2_bgp_asn", nil)
		d.Set("tunnel2_bgp_holdtime", nil)
		d.Set("tunnel2_cgw_inside_address", nil)
		d.Set("tunnel2_preshared_key", nil)
		d.Set("tunnel2_vgw_inside_address", nil)
	}

	return diags
}

func resourceVPNConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChanges("customer_gateway_id", names.AttrTransitGatewayID, "vpn_gateway_id") {
		input := &ec2.ModifyVpnConnectionInput{
			VpnConnectionId: aws.String(d.Id()),
		}

		if d.HasChange("customer_gateway_id") {
			input.CustomerGatewayId = aws.String(d.Get("customer_gateway_id").(string))
		}

		if hasChange, v := d.HasChange(names.AttrTransitGatewayID), d.Get(names.AttrTransitGatewayID).(string); hasChange && v != "" {
			input.TransitGatewayId = aws.String(v)
		}

		if hasChange, v := d.HasChange("vpn_gateway_id"), d.Get("vpn_gateway_id").(string); hasChange && v != "" {
			input.VpnGatewayId = aws.String(v)
		}

		_, err := conn.ModifyVpnConnection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 VPN Connection (%s): %s", d.Id(), err)
		}

		if _, err := waitVPNConnectionUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPN Connection (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChanges("local_ipv4_network_cidr", "local_ipv6_network_cidr", "remote_ipv4_network_cidr", "remote_ipv6_network_cidr") {
		input := &ec2.ModifyVpnConnectionOptionsInput{
			VpnConnectionId: aws.String(d.Id()),
		}

		if d.HasChange("local_ipv4_network_cidr") {
			input.LocalIpv4NetworkCidr = aws.String(d.Get("local_ipv4_network_cidr").(string))
		}

		if d.HasChange("local_ipv6_network_cidr") {
			input.LocalIpv6NetworkCidr = aws.String(d.Get("local_ipv6_network_cidr").(string))
		}

		if d.HasChange("remote_ipv4_network_cidr") {
			input.RemoteIpv4NetworkCidr = aws.String(d.Get("remote_ipv4_network_cidr").(string))
		}

		if d.HasChange("remote_ipv6_network_cidr") {
			input.RemoteIpv6NetworkCidr = aws.String(d.Get("remote_ipv6_network_cidr").(string))
		}

		_, err := conn.ModifyVpnConnectionOptions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 VPN Connection (%s) connection options: %s", d.Id(), err)
		}

		if _, err := waitVPNConnectionUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPN Connection (%s) connection options update: %s", d.Id(), err)
		}
	}

	for i, prefix := range []string{"tunnel1_", "tunnel2_"} {
		if options, address := expandModifyVPNTunnelOptionsSpecification(d, prefix), d.Get(prefix+names.AttrAddress).(string); options != nil && address != "" {
			input := &ec2.ModifyVpnTunnelOptionsInput{
				TunnelOptions:             options,
				VpnConnectionId:           aws.String(d.Id()),
				VpnTunnelOutsideIpAddress: aws.String(address),
			}

			_, err := conn.ModifyVpnTunnelOptions(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying EC2 VPN Connection (%s) tunnel (%d) options: %s", d.Id(), i+1, err)
			}

			if _, err := waitVPNConnectionUpdated(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPN Connection (%s) tunnel (%d) options update: %s", d.Id(), i+1, err)
			}
		}
	}

	return append(diags, resourceVPNConnectionRead(ctx, d, meta)...)
}

func resourceVPNConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 VPN Connection: %s", d.Id())
	_, err := conn.DeleteVpnConnection(ctx, &ec2.DeleteVpnConnectionInput{
		VpnConnectionId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNConnectionIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPN Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitVPNConnectionDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPN Connection (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandVPNConnectionOptionsSpecification(d *schema.ResourceData) *awstypes.VpnConnectionOptionsSpecification {
	apiObject := &awstypes.VpnConnectionOptionsSpecification{}

	if v, ok := d.GetOk("enable_acceleration"); ok {
		apiObject.EnableAcceleration = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("outside_ip_address_type"); ok {
		apiObject.OutsideIpAddressType = aws.String(v.(string))
	}

	if v := d.Get("tunnel_inside_ip_version").(string); v == string(awstypes.TunnelInsideIpVersionIpv6) {
		if v, ok := d.GetOk("local_ipv6_network_cidr"); ok {
			apiObject.LocalIpv6NetworkCidr = aws.String(v.(string))
		}

		if v, ok := d.GetOk("remote_ipv6_network_cidr"); ok {
			apiObject.RemoteIpv6NetworkCidr = aws.String(v.(string))
		}

		apiObject.TunnelInsideIpVersion = awstypes.TunnelInsideIpVersion(v)
	} else {
		if v, ok := d.GetOk("local_ipv4_network_cidr"); ok {
			apiObject.LocalIpv4NetworkCidr = aws.String(v.(string))
		}

		if v, ok := d.GetOk("remote_ipv4_network_cidr"); ok {
			apiObject.RemoteIpv4NetworkCidr = aws.String(v.(string))
		}
	}

	if v, ok := d.GetOk("static_routes_only"); ok {
		apiObject.StaticRoutesOnly = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("transport_transit_gateway_attachment_id"); ok {
		apiObject.TransportTransitGatewayAttachmentId = aws.String(v.(string))
	}

	apiObject.TunnelOptions = []awstypes.VpnTunnelOptionsSpecification{
		expandVPNTunnelOptionsSpecification(d, "tunnel1_"),
		expandVPNTunnelOptionsSpecification(d, "tunnel2_"),
	}

	return apiObject
}

func expandVPNTunnelOptionsSpecification(d *schema.ResourceData, prefix string) awstypes.VpnTunnelOptionsSpecification {
	apiObject := awstypes.VpnTunnelOptionsSpecification{}

	if v, ok := d.GetOk(prefix + "dpd_timeout_action"); ok {
		apiObject.DPDTimeoutAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk(prefix + "dpd_timeout_seconds"); ok {
		apiObject.DPDTimeoutSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "enable_tunnel_lifecycle_control"); ok {
		apiObject.EnableTunnelLifecycleControl = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(prefix + "ike_versions"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.IKEVersions = append(apiObject.IKEVersions, awstypes.IKEVersionsRequestListValue{Value: aws.String(v.(string))})
		}
	}

	if v, ok := d.GetOk(prefix + "log_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.LogOptions = expandVPNTunnelLogOptionsSpecification(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(prefix + "phase1_dh_group_numbers"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase1DHGroupNumbers = append(apiObject.Phase1DHGroupNumbers, awstypes.Phase1DHGroupNumbersRequestListValue{Value: aws.Int32(int32(v.(int)))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase1_encryption_algorithms"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase1EncryptionAlgorithms = append(apiObject.Phase1EncryptionAlgorithms, awstypes.Phase1EncryptionAlgorithmsRequestListValue{Value: aws.String(v.(string))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase1_integrity_algorithms"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase1IntegrityAlgorithms = append(apiObject.Phase1IntegrityAlgorithms, awstypes.Phase1IntegrityAlgorithmsRequestListValue{Value: aws.String(v.(string))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase1_lifetime_seconds"); ok {
		apiObject.Phase1LifetimeSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "phase2_dh_group_numbers"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase2DHGroupNumbers = append(apiObject.Phase2DHGroupNumbers, awstypes.Phase2DHGroupNumbersRequestListValue{Value: aws.Int32(int32(v.(int)))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase2_encryption_algorithms"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase2EncryptionAlgorithms = append(apiObject.Phase2EncryptionAlgorithms, awstypes.Phase2EncryptionAlgorithmsRequestListValue{Value: aws.String(v.(string))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase2_integrity_algorithms"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase2IntegrityAlgorithms = append(apiObject.Phase2IntegrityAlgorithms, awstypes.Phase2IntegrityAlgorithmsRequestListValue{Value: aws.String(v.(string))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase2_lifetime_seconds"); ok {
		apiObject.Phase2LifetimeSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "preshared_key"); ok {
		apiObject.PreSharedKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk(prefix + "rekey_fuzz_percentage"); ok {
		apiObject.RekeyFuzzPercentage = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "rekey_margin_time_seconds"); ok {
		apiObject.RekeyMarginTimeSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "replay_window_size"); ok {
		apiObject.ReplayWindowSize = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "startup_action"); ok {
		apiObject.StartupAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk(prefix + "inside_cidr"); ok {
		apiObject.TunnelInsideCidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk(prefix + "inside_ipv6_cidr"); ok {
		apiObject.TunnelInsideIpv6Cidr = aws.String(v.(string))
	}

	return apiObject
}

func expandVPNTunnelLogOptionsSpecification(tfMap map[string]interface{}) *awstypes.VpnTunnelLogOptionsSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.VpnTunnelLogOptionsSpecification{}

	if v, ok := tfMap["cloudwatch_log_options"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CloudWatchLogOptions = expandCloudWatchLogOptionsSpecification(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudWatchLogOptionsSpecification(tfMap map[string]interface{}) *awstypes.CloudWatchLogOptionsSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CloudWatchLogOptionsSpecification{}

	if v, ok := tfMap["log_enabled"].(bool); ok {
		apiObject.LogEnabled = aws.Bool(v)
	}

	// No ARN or format if not enabled.
	if aws.ToBool(apiObject.LogEnabled) {
		if v, ok := tfMap["log_group_arn"].(string); ok && v != "" {
			apiObject.LogGroupArn = aws.String(v)
		}

		if v, ok := tfMap["log_output_format"].(string); ok && v != "" {
			apiObject.LogOutputFormat = aws.String(v)
		}
	}

	return apiObject
}

func expandModifyVPNTunnelOptionsSpecification(d *schema.ResourceData, prefix string) *awstypes.ModifyVpnTunnelOptionsSpecification {
	apiObject := &awstypes.ModifyVpnTunnelOptionsSpecification{}
	hasChange := false

	if key := prefix + "dpd_timeout_action"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.DPDTimeoutAction = aws.String(v.(string))
		} else {
			apiObject.DPDTimeoutAction = aws.String(defaultVPNTunnelOptionsDPDTimeoutAction)
		}

		hasChange = true
	}

	if key := prefix + "dpd_timeout_seconds"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.DPDTimeoutSeconds = aws.Int32(int32(v.(int)))
		} else {
			apiObject.DPDTimeoutSeconds = aws.Int32(int32(defaultVPNTunnelOptionsDPDTimeoutSeconds))
		}

		hasChange = true
	}

	if key := prefix + "enable_tunnel_lifecycle_control"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.EnableTunnelLifecycleControl = aws.Bool(v.(bool))
		} else {
			apiObject.EnableTunnelLifecycleControl = aws.Bool(defaultVPNTunnelOptionsEnableTunnelLifecycleControl)
		}

		hasChange = true
	}

	if key := prefix + "ike_versions"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.IKEVersions = append(apiObject.IKEVersions, awstypes.IKEVersionsRequestListValue{Value: aws.String(v.(string))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsIKEVersions {
				apiObject.IKEVersions = append(apiObject.IKEVersions, awstypes.IKEVersionsRequestListValue{Value: aws.String(v)})
			}
		}

		hasChange = true
	}

	if key := prefix + "log_options"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			apiObject.LogOptions = expandVPNTunnelLogOptionsSpecification(v.([]interface{})[0].(map[string]interface{}))
		}

		hasChange = true
	}

	if key := prefix + "phase1_dh_group_numbers"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase1DHGroupNumbers = append(apiObject.Phase1DHGroupNumbers, awstypes.Phase1DHGroupNumbersRequestListValue{Value: aws.Int32(int32(v.(int)))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase1DHGroupNumbers {
				apiObject.Phase1DHGroupNumbers = append(apiObject.Phase1DHGroupNumbers, awstypes.Phase1DHGroupNumbersRequestListValue{Value: aws.Int32(int32(v))})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase1_encryption_algorithms"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase1EncryptionAlgorithms = append(apiObject.Phase1EncryptionAlgorithms, awstypes.Phase1EncryptionAlgorithmsRequestListValue{Value: aws.String(v.(string))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase1EncryptionAlgorithms {
				apiObject.Phase1EncryptionAlgorithms = append(apiObject.Phase1EncryptionAlgorithms, awstypes.Phase1EncryptionAlgorithmsRequestListValue{Value: aws.String(v)})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase1_integrity_algorithms"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase1IntegrityAlgorithms = append(apiObject.Phase1IntegrityAlgorithms, awstypes.Phase1IntegrityAlgorithmsRequestListValue{Value: aws.String(v.(string))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase1IntegrityAlgorithms {
				apiObject.Phase1IntegrityAlgorithms = append(apiObject.Phase1IntegrityAlgorithms, awstypes.Phase1IntegrityAlgorithmsRequestListValue{Value: aws.String(v)})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase1_lifetime_seconds"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.Phase1LifetimeSeconds = aws.Int32(int32(v.(int)))
		} else {
			apiObject.Phase1LifetimeSeconds = aws.Int32(int32(defaultVPNTunnelOptionsPhase1LifetimeSeconds))
		}

		hasChange = true
	}

	if key := prefix + "phase2_dh_group_numbers"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase2DHGroupNumbers = append(apiObject.Phase2DHGroupNumbers, awstypes.Phase2DHGroupNumbersRequestListValue{Value: aws.Int32(int32(v.(int)))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase2DHGroupNumbers {
				apiObject.Phase2DHGroupNumbers = append(apiObject.Phase2DHGroupNumbers, awstypes.Phase2DHGroupNumbersRequestListValue{Value: aws.Int32(int32(v))})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase2_encryption_algorithms"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase2EncryptionAlgorithms = append(apiObject.Phase2EncryptionAlgorithms, awstypes.Phase2EncryptionAlgorithmsRequestListValue{Value: aws.String(v.(string))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase2EncryptionAlgorithms {
				apiObject.Phase2EncryptionAlgorithms = append(apiObject.Phase2EncryptionAlgorithms, awstypes.Phase2EncryptionAlgorithmsRequestListValue{Value: aws.String(v)})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase2_integrity_algorithms"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase2IntegrityAlgorithms = append(apiObject.Phase2IntegrityAlgorithms, awstypes.Phase2IntegrityAlgorithmsRequestListValue{Value: aws.String(v.(string))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase2IntegrityAlgorithms {
				apiObject.Phase2IntegrityAlgorithms = append(apiObject.Phase2IntegrityAlgorithms, awstypes.Phase2IntegrityAlgorithmsRequestListValue{Value: aws.String(v)})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase2_lifetime_seconds"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.Phase2LifetimeSeconds = aws.Int32(int32(v.(int)))
		} else {
			apiObject.Phase2LifetimeSeconds = aws.Int32(int32(defaultVPNTunnelOptionsPhase2LifetimeSeconds))
		}

		hasChange = true
	}

	if key := prefix + "preshared_key"; d.HasChange(key) {
		apiObject.PreSharedKey = aws.String(d.Get(key).(string))

		hasChange = true
	}

	if key := prefix + "rekey_fuzz_percentage"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.RekeyFuzzPercentage = aws.Int32(int32(v.(int)))
		} else {
			apiObject.RekeyFuzzPercentage = aws.Int32(int32(defaultVPNTunnelOptionsRekeyFuzzPercentage))
		}

		hasChange = true
	}

	if key := prefix + "rekey_margin_time_seconds"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.RekeyMarginTimeSeconds = aws.Int32(int32(v.(int)))
		} else {
			apiObject.RekeyMarginTimeSeconds = aws.Int32(int32(defaultVPNTunnelOptionsRekeyMarginTimeSeconds))
		}

		hasChange = true
	}

	if key := prefix + "replay_window_size"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.ReplayWindowSize = aws.Int32(int32(v.(int)))
		} else {
			apiObject.ReplayWindowSize = aws.Int32(int32(defaultVPNTunnelOptionsReplayWindowSize))
		}

		hasChange = true
	}

	if key := prefix + "startup_action"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.StartupAction = aws.String(v.(string))
		} else {
			apiObject.StartupAction = aws.String(defaultVPNTunnelOptionsStartupAction)
		}

		hasChange = true
	}

	if !hasChange {
		return nil
	}

	return apiObject
}

func flattenTunnelOption(d *schema.ResourceData, prefix string, apiObject awstypes.TunnelOption) error {
	var s []*string
	var i []*int32

	d.Set(prefix+"dpd_timeout_action", apiObject.DpdTimeoutAction)
	d.Set(prefix+"dpd_timeout_seconds", apiObject.DpdTimeoutSeconds)
	d.Set(prefix+"enable_tunnel_lifecycle_control", apiObject.EnableTunnelLifecycleControl)

	for _, v := range apiObject.IkeVersions {
		s = append(s, v.Value)
	}
	d.Set(prefix+"ike_versions", aws.ToStringSlice(s))
	s = nil

	if apiObject.LogOptions != nil {
		if err := d.Set(prefix+"log_options", []interface{}{flattenVPNTunnelLogOptions(apiObject.LogOptions)}); err != nil {
			return fmt.Errorf("setting %s: %w", prefix+"log_options", err)
		}
	} else {
		d.Set(prefix+"log_options", nil)
	}

	for _, v := range apiObject.Phase1DHGroupNumbers {
		i = append(i, v.Value)
	}
	d.Set(prefix+"phase1_dh_group_numbers", aws.ToInt32Slice(i))
	i = nil

	for _, v := range apiObject.Phase1EncryptionAlgorithms {
		s = append(s, v.Value)
	}
	d.Set(prefix+"phase1_encryption_algorithms", aws.ToStringSlice(s))
	s = nil

	for _, v := range apiObject.Phase1IntegrityAlgorithms {
		s = append(s, v.Value)
	}
	d.Set(prefix+"phase1_integrity_algorithms", aws.ToStringSlice(s))
	s = nil

	d.Set(prefix+"phase1_lifetime_seconds", apiObject.Phase1LifetimeSeconds)

	for _, v := range apiObject.Phase2DHGroupNumbers {
		i = append(i, v.Value)
	}
	d.Set(prefix+"phase2_dh_group_numbers", aws.ToInt32Slice(i))

	for _, v := range apiObject.Phase2EncryptionAlgorithms {
		s = append(s, v.Value)
	}
	d.Set(prefix+"phase2_encryption_algorithms", aws.ToStringSlice(s))
	s = nil

	for _, v := range apiObject.Phase2IntegrityAlgorithms {
		s = append(s, v.Value)
	}
	d.Set(prefix+"phase2_integrity_algorithms", aws.ToStringSlice(s))

	d.Set(prefix+"phase2_lifetime_seconds", apiObject.Phase2LifetimeSeconds)
	d.Set(prefix+"rekey_fuzz_percentage", apiObject.RekeyFuzzPercentage)
	d.Set(prefix+"rekey_margin_time_seconds", apiObject.RekeyMarginTimeSeconds)
	d.Set(prefix+"replay_window_size", apiObject.ReplayWindowSize)
	d.Set(prefix+"startup_action", apiObject.StartupAction)
	d.Set(prefix+"inside_cidr", apiObject.TunnelInsideCidr)
	d.Set(prefix+"inside_ipv6_cidr", apiObject.TunnelInsideIpv6Cidr)

	return nil
}

func flattenVPNStaticRoute(apiObject awstypes.VpnStaticRoute) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.DestinationCidrBlock; v != nil {
		tfMap["destination_cidr_block"] = aws.ToString(v)
	}

	tfMap[names.AttrSource] = apiObject.Source
	tfMap[names.AttrState] = apiObject.State

	return tfMap
}

func flattenVPNStaticRoutes(apiObjects []awstypes.VpnStaticRoute) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenVPNStaticRoute(apiObject))
	}

	return tfList
}

func flattenVPNTunnelLogOptions(apiObject *awstypes.VpnTunnelLogOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CloudWatchLogOptions; v != nil {
		tfMap["cloudwatch_log_options"] = []interface{}{flattenCloudWatchLogOptions(v)}
	}

	return tfMap
}

func flattenCloudWatchLogOptions(apiObject *awstypes.CloudWatchLogOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogEnabled; v != nil {
		enabled := aws.ToBool(v)
		tfMap["log_enabled"] = enabled

		// No ARN or format if not enabled.
		if enabled {
			if v := apiObject.LogGroupArn; v != nil {
				tfMap["log_group_arn"] = aws.ToString(v)
			}

			if v := apiObject.LogOutputFormat; v != nil {
				tfMap["log_output_format"] = aws.ToString(v)
			}
		}
	}

	return tfMap
}

func flattenVGWTelemetry(apiObject awstypes.VgwTelemetry) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.AcceptedRouteCount; v != nil {
		tfMap["accepted_route_count"] = aws.ToInt32(v)
	}

	if v := apiObject.CertificateArn; v != nil {
		tfMap[names.AttrCertificateARN] = aws.ToString(v)
	}

	if v := apiObject.LastStatusChange; v != nil {
		tfMap["last_status_change"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := apiObject.OutsideIpAddress; v != nil {
		tfMap["outside_ip_address"] = aws.ToString(v)
	}

	tfMap[names.AttrStatus] = apiObject.Status

	if v := apiObject.StatusMessage; v != nil {
		tfMap[names.AttrStatusMessage] = aws.ToString(v)
	}

	return tfMap
}

func flattenVGWTelemetries(apiObjects []awstypes.VgwTelemetry) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenVGWTelemetry(apiObject))
	}

	return tfList
}

type xmlVpnConnectionConfig struct {
	Tunnels []xmlIpsecTunnel `xml:"ipsec_tunnel"`
}

type xmlIpsecTunnel struct {
	BGPASN           string `xml:"vpn_gateway>bgp>asn"`
	BGPHoldTime      int    `xml:"vpn_gateway>bgp>hold_time"`
	CgwInsideAddress string `xml:"customer_gateway>tunnel_inside_address>ip_address"`
	OutsideAddress   string `xml:"vpn_gateway>tunnel_outside_address>ip_address"`
	PreSharedKey     string `xml:"ike>pre_shared_key"`
	VgwInsideAddress string `xml:"vpn_gateway>tunnel_inside_address>ip_address"`
}

type tunnelInfo struct {
	Tunnel1Address          string
	Tunnel1BGPASN           string
	Tunnel1BGPHoldTime      int
	Tunnel1CgwInsideAddress string
	Tunnel1PreSharedKey     string
	Tunnel1VgwInsideAddress string
	Tunnel2Address          string
	Tunnel2BGPASN           string
	Tunnel2BGPHoldTime      int
	Tunnel2CgwInsideAddress string
	Tunnel2PreSharedKey     string
	Tunnel2VgwInsideAddress string
}

func (slice xmlVpnConnectionConfig) Len() int {
	return len(slice.Tunnels)
}

func (slice xmlVpnConnectionConfig) Less(i, j int) bool {
	return slice.Tunnels[i].OutsideAddress < slice.Tunnels[j].OutsideAddress
}

func (slice xmlVpnConnectionConfig) Swap(i, j int) {
	slice.Tunnels[i], slice.Tunnels[j] = slice.Tunnels[j], slice.Tunnels[i]
}

// customerGatewayConfigurationToTunnelInfo converts the configuration information for the
// VPN connection's customer gateway (in the native XML format) to a tunnelInfo structure.
// The tunnel1 parameters are optionally used to correctly order tunnel configurations.
func customerGatewayConfigurationToTunnelInfo(xmlConfig string, tunnel1PreSharedKey string, tunnel1InsideCidr string, tunnel1InsideIpv6Cidr string) (*tunnelInfo, error) {
	var vpnConfig xmlVpnConnectionConfig

	if err := xml.Unmarshal([]byte(xmlConfig), &vpnConfig); err != nil {
		return nil, err
	}

	// XML tunnel ordering was commented here as being inconsistent since
	// this logic was originally added. The original sorting is based on
	// outside address. Given potential tunnel identifying configuration,
	// we try to correctly align the tunnel ordering before preserving the
	// original outside address sorting fallback for backwards compatibility
	// as to not inadvertently flip existing configurations.
	if tunnel1PreSharedKey != "" {
		if tunnel1PreSharedKey != vpnConfig.Tunnels[0].PreSharedKey && tunnel1PreSharedKey == vpnConfig.Tunnels[1].PreSharedKey {
			vpnConfig.Tunnels[0], vpnConfig.Tunnels[1] = vpnConfig.Tunnels[1], vpnConfig.Tunnels[0]
		}
	} else if cidr := tunnel1InsideCidr; cidr != "" {
		if _, ipNet, err := net.ParseCIDR(cidr); err == nil && ipNet != nil {
			vgwInsideAddressIP1 := net.ParseIP(vpnConfig.Tunnels[0].VgwInsideAddress)
			vgwInsideAddressIP2 := net.ParseIP(vpnConfig.Tunnels[1].VgwInsideAddress)

			if !ipNet.Contains(vgwInsideAddressIP1) && ipNet.Contains(vgwInsideAddressIP2) {
				vpnConfig.Tunnels[0], vpnConfig.Tunnels[1] = vpnConfig.Tunnels[1], vpnConfig.Tunnels[0]
			}
		}
	} else if cidr := tunnel1InsideIpv6Cidr; cidr != "" {
		if _, ipNet, err := net.ParseCIDR(cidr); err == nil && ipNet != nil {
			vgwInsideAddressIP1 := net.ParseIP(vpnConfig.Tunnels[0].VgwInsideAddress)
			vgwInsideAddressIP2 := net.ParseIP(vpnConfig.Tunnels[1].VgwInsideAddress)

			if !ipNet.Contains(vgwInsideAddressIP1) && ipNet.Contains(vgwInsideAddressIP2) {
				vpnConfig.Tunnels[0], vpnConfig.Tunnels[1] = vpnConfig.Tunnels[1], vpnConfig.Tunnels[0]
			}
		}
	} else {
		sort.Sort(vpnConfig)
	}

	tunnelInfo := &tunnelInfo{
		Tunnel1Address:          vpnConfig.Tunnels[0].OutsideAddress,
		Tunnel1BGPASN:           vpnConfig.Tunnels[0].BGPASN,
		Tunnel1BGPHoldTime:      vpnConfig.Tunnels[0].BGPHoldTime,
		Tunnel1CgwInsideAddress: vpnConfig.Tunnels[0].CgwInsideAddress,
		Tunnel1PreSharedKey:     vpnConfig.Tunnels[0].PreSharedKey,
		Tunnel1VgwInsideAddress: vpnConfig.Tunnels[0].VgwInsideAddress,
		Tunnel2Address:          vpnConfig.Tunnels[1].OutsideAddress,
		Tunnel2BGPASN:           vpnConfig.Tunnels[1].BGPASN,
		Tunnel2BGPHoldTime:      vpnConfig.Tunnels[1].BGPHoldTime,
		Tunnel2CgwInsideAddress: vpnConfig.Tunnels[1].CgwInsideAddress,
		Tunnel2PreSharedKey:     vpnConfig.Tunnels[1].PreSharedKey,
		Tunnel2VgwInsideAddress: vpnConfig.Tunnels[1].VgwInsideAddress,
	}

	return tunnelInfo, nil
}

func validVPNConnectionTunnelPreSharedKey() schema.SchemaValidateFunc {
	return validation.All(
		validation.StringLenBetween(8, 64),
		validation.StringDoesNotMatch(regexache.MustCompile(`^0`), "cannot start with zero character"),
		validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.]+$`), "can only contain alphanumeric, period and underscore characters"),
	)
}

// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_VpnTunnelOptionsSpecification.html
// https://docs.aws.amazon.com/vpn/latest/s2svpn/VPNTunnels.html
func validVPNConnectionTunnelInsideCIDR() schema.SchemaValidateFunc {
	disallowedCidrs := []string{
		"169.254.0.0/30",
		"169.254.1.0/30",
		"169.254.2.0/30",
		"169.254.3.0/30",
		"169.254.4.0/30",
		"169.254.5.0/30",
		"169.254.169.252/30",
	}

	return validation.All(
		validation.IsCIDRNetwork(30, 30),
		validation.StringMatch(regexache.MustCompile(`^169\.254\.`), "must be within 169.254.0.0/16"),
		validation.StringNotInSlice(disallowedCidrs, false),
	)
}

func validVPNConnectionTunnelInsideIPv6CIDR() schema.SchemaValidateFunc {
	return validation.All(
		validation.IsCIDRNetwork(126, 126),
		validation.StringMatch(regexache.MustCompile(`^fd`), "must be within fd00::/8"),
	)
}

// customizeDiffValidateOutsideIPAddressType validates that if provided `outside_ip_address_type` is `PrivateIpv4` then `transport_transit_gateway_attachment_id` must be provided
func customizeDiffValidateOutsideIPAddressType(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk("outside_ip_address_type"); !ok || v.(string) == outsideIPAddressTypePublicIPv4 {
		return nil
	}

	if v, ok := diff.GetOk("transport_transit_gateway_attachment_id"); !ok || v.(string) != "" {
		return nil
	}
	return fmt.Errorf("`transport_transit_gateway_attachment_id` must be provided if `outside_ip_address_type` is `PrivateIpv4`")
}
