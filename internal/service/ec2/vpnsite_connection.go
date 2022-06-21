package ec2

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPNConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPNConnectionCreate,
		Read:   resourceVPNConnectionRead,
		Update: resourceVPNConnectionUpdate,
		Delete: resourceVPNConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_attachment_arn": {
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
				RequiredWith: []string{"transit_gateway_id"},
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
				RequiredWith: []string{"transit_gateway_id"},
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
				RequiredWith: []string{"transit_gateway_id"},
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
						"source": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"transit_gateway_id", "vpn_gateway_id"},
			},
			"tunnel_inside_ip_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.TunnelInsideIpVersion_Values(), false),
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
				RequiredWith: []string{"transit_gateway_id"},
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
				RequiredWith: []string{"transit_gateway_id"},
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
			"type": {
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
						"certificate_arn": {
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
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"vpn_gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"transit_gateway_id", "vpn_gateway_id"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

// https://docs.aws.amazon.com/vpn/latest/s2svpn/VPNTunnels.html.
var (
	defaultVPNTunnelOptionsDPDTimeoutAction           = vpnTunnelOptionsDPDTimeoutActionClear
	defaultVPNTunnelOptionsDPDTimeoutSeconds          = 30
	defaultVPNTunnelOptionsIKEVersions                = []string{vpnTunnelOptionsIKEVersion1, vpnTunnelOptionsIKEVersion2}
	defaultVPNTunnelOptionsPhase1DHGroupNumbers       = []int{2, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}
	defaultVPNTunnelOptionsPhase1EncryptionAlgorithms = []string{
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

func resourceVPNConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateVpnConnectionInput{
		CustomerGatewayId: aws.String(d.Get("customer_gateway_id").(string)),
		Options:           expandVPNConnectionOptionsSpecification(d),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVpnConnection),
		Type:              aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("transit_gateway_id"); ok {
		input.TransitGatewayId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpn_gateway_id"); ok {
		input.VpnGatewayId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 VPN Connection: %s", input)
	output, err := conn.CreateVpnConnection(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 VPN Connection: %w", err)
	}

	d.SetId(aws.StringValue(output.VpnConnection.VpnConnectionId))

	if _, err := WaitVPNConnectionCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 VPN Connection (%s) create: %w", d.Id(), err)
	}

	// Read off the API to populate our RO fields.
	return resourceVPNConnectionRead(d, meta)
}

func resourceVPNConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	vpnConnection, err := FindVPNConnectionByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPN Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPN Connection (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpn-connection/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("core_network_arn", vpnConnection.CoreNetworkArn)
	d.Set("core_network_attachment_arn", vpnConnection.CoreNetworkAttachmentArn)
	d.Set("customer_gateway_id", vpnConnection.CustomerGatewayId)
	d.Set("type", vpnConnection.Type)
	d.Set("vpn_gateway_id", vpnConnection.VpnGatewayId)

	if v := vpnConnection.TransitGatewayId; v != nil {
		input := &ec2.DescribeTransitGatewayAttachmentsInput{
			Filters: BuildAttributeFilterList(map[string]string{
				"resource-id":        d.Id(),
				"resource-type":      ec2.TransitGatewayAttachmentResourceTypeVpn,
				"transit-gateway-id": aws.StringValue(v),
			}),
		}

		output, err := FindTransitGatewayAttachment(conn, input)

		if err != nil {
			return fmt.Errorf("error reading EC2 VPN Connection (%s) Transit Gateway Attachment: %s", d.Id(), err)
		}

		d.Set("transit_gateway_attachment_id", output.TransitGatewayAttachmentId)
		d.Set("transit_gateway_id", v)
	} else {
		d.Set("transit_gateway_attachment_id", nil)
		d.Set("transit_gateway_id", nil)
	}

	if err := d.Set("routes", flattenVPNStaticRoutes(vpnConnection.Routes)); err != nil {
		return fmt.Errorf("error setting routes: %w", err)
	}

	if err := d.Set("vgw_telemetry", flattenVGWTelemetries(vpnConnection.VgwTelemetry)); err != nil {
		return fmt.Errorf("error setting vgw_telemetry: %w", err)
	}

	tags := KeyValueTags(vpnConnection.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if v := vpnConnection.Options; v != nil {
		d.Set("enable_acceleration", v.EnableAcceleration)
		d.Set("local_ipv4_network_cidr", v.LocalIpv4NetworkCidr)
		d.Set("local_ipv6_network_cidr", v.LocalIpv6NetworkCidr)
		d.Set("remote_ipv4_network_cidr", v.RemoteIpv4NetworkCidr)
		d.Set("remote_ipv6_network_cidr", v.RemoteIpv6NetworkCidr)
		d.Set("static_routes_only", v.StaticRoutesOnly)
		d.Set("tunnel_inside_ip_version", v.TunnelInsideIpVersion)

		for i, prefix := range []string{"tunnel1_", "tunnel2_"} {
			if len(v.TunnelOptions) > i {
				flattenTunnelOption(d, prefix, v.TunnelOptions[i])
			}
		}
	} else {
		d.Set("enable_acceleration", nil)
		d.Set("local_ipv4_network_cidr", nil)
		d.Set("local_ipv6_network_cidr", nil)
		d.Set("remote_ipv4_network_cidr", nil)
		d.Set("remote_ipv6_network_cidr", nil)
		d.Set("static_routes_only", nil)
		d.Set("tunnel_inside_ip_version", nil)
	}

	d.Set("customer_gateway_configuration", vpnConnection.CustomerGatewayConfiguration)

	tunnelInfo, err := CustomerGatewayConfigurationToTunnelInfo(
		aws.StringValue(vpnConnection.CustomerGatewayConfiguration),
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

	return nil
}

func resourceVPNConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChanges("customer_gateway_id", "transit_gateway_id", "vpn_gateway_id") {
		input := &ec2.ModifyVpnConnectionInput{
			VpnConnectionId: aws.String(d.Id()),
		}

		if d.HasChange("customer_gateway_id") {
			input.CustomerGatewayId = aws.String(d.Get("customer_gateway_id").(string))
		}

		if hasChange, v := d.HasChange("transit_gateway_id"), d.Get("transit_gateway_id").(string); hasChange && v != "" {
			input.TransitGatewayId = aws.String(v)
		}

		if hasChange, v := d.HasChange("vpn_gateway_id"), d.Get("vpn_gateway_id").(string); hasChange && v != "" {
			input.VpnGatewayId = aws.String(v)
		}

		_, err := conn.ModifyVpnConnection(input)

		if err != nil {
			return fmt.Errorf("error modifying EC2 VPN Connection (%s): %w", d.Id(), err)
		}

		if _, err := WaitVPNConnectionUpdated(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for EC2 VPN Connection (%s) update: %w", d.Id(), err)
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

		_, err := conn.ModifyVpnConnectionOptions(input)

		if err != nil {
			return fmt.Errorf("error modifying EC2 VPN Connection (%s) connection options: %w", d.Id(), err)
		}

		if _, err := WaitVPNConnectionUpdated(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for EC2 VPN Connection (%s) connection options update: %w", d.Id(), err)
		}
	}

	for i, prefix := range []string{"tunnel1_", "tunnel2_"} {
		if options, address := expandModifyVPNTunnelOptionsSpecification(d, prefix), d.Get(prefix+"address").(string); options != nil && address != "" {
			input := &ec2.ModifyVpnTunnelOptionsInput{
				TunnelOptions:             options,
				VpnConnectionId:           aws.String(d.Id()),
				VpnTunnelOutsideIpAddress: aws.String(address),
			}

			log.Printf("[DEBUG] Modifying EC2 VPN Connection tunnel (%d) options: %s", i+1, input)
			_, err := conn.ModifyVpnTunnelOptions(input)

			if err != nil {
				return fmt.Errorf("error modifying EC2 VPN Connection (%s) tunnel (%d) options: %w", d.Id(), i+1, err)
			}

			if _, err := WaitVPNConnectionUpdated(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for EC2 VPN Connection (%s) tunnel (%d) options update: %w", d.Id(), i+1, err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 VPN Connection (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceVPNConnectionRead(d, meta)
}

func resourceVPNConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EC2 VPN Connection: %s", d.Id())
	_, err := conn.DeleteVpnConnection(&ec2.DeleteVpnConnectionInput{
		VpnConnectionId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNConnectionIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 VPN Connection (%s): %w", d.Id(), err)
	}

	if _, err := WaitVPNConnectionDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 VPN Connection (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandVPNConnectionOptionsSpecification(d *schema.ResourceData) *ec2.VpnConnectionOptionsSpecification {
	apiObject := &ec2.VpnConnectionOptionsSpecification{}

	if v, ok := d.GetOk("enable_acceleration"); ok {
		apiObject.EnableAcceleration = aws.Bool(v.(bool))
	}

	if v := d.Get("tunnel_inside_ip_version").(string); v == ec2.TunnelInsideIpVersionIpv6 {
		if v, ok := d.GetOk("local_ipv6_network_cidr"); ok {
			apiObject.LocalIpv6NetworkCidr = aws.String(v.(string))
		}

		if v, ok := d.GetOk("remote_ipv6_network_cidr"); ok {
			apiObject.RemoteIpv6NetworkCidr = aws.String(v.(string))
		}

		apiObject.TunnelInsideIpVersion = aws.String(v)
	} else {
		if v, ok := d.GetOk("local_ipv4_network_cidr"); ok {
			apiObject.LocalIpv4NetworkCidr = aws.String(v.(string))
		}

		if v, ok := d.GetOk("remote_ipv4_network_cidr"); ok {
			apiObject.RemoteIpv4NetworkCidr = aws.String(v.(string))
		}

		apiObject.TunnelInsideIpVersion = aws.String(ec2.TunnelInsideIpVersionIpv4)
	}

	if v, ok := d.GetOk("static_routes_only"); ok {
		apiObject.StaticRoutesOnly = aws.Bool(v.(bool))
	}

	apiObject.TunnelOptions = []*ec2.VpnTunnelOptionsSpecification{
		expandVPNTunnelOptionsSpecification(d, "tunnel1_"),
		expandVPNTunnelOptionsSpecification(d, "tunnel2_"),
	}

	return apiObject
}

func expandVPNTunnelOptionsSpecification(d *schema.ResourceData, prefix string) *ec2.VpnTunnelOptionsSpecification {
	apiObject := &ec2.VpnTunnelOptionsSpecification{}

	if v, ok := d.GetOk(prefix + "dpd_timeout_action"); ok {
		apiObject.DPDTimeoutAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk(prefix + "dpd_timeout_seconds"); ok {
		apiObject.DPDTimeoutSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "ike_versions"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.IKEVersions = append(apiObject.IKEVersions, &ec2.IKEVersionsRequestListValue{Value: aws.String(v.(string))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase1_dh_group_numbers"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase1DHGroupNumbers = append(apiObject.Phase1DHGroupNumbers, &ec2.Phase1DHGroupNumbersRequestListValue{Value: aws.Int64(int64(v.(int)))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase1_encryption_algorithms"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase1EncryptionAlgorithms = append(apiObject.Phase1EncryptionAlgorithms, &ec2.Phase1EncryptionAlgorithmsRequestListValue{Value: aws.String(v.(string))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase1_integrity_algorithms"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase1IntegrityAlgorithms = append(apiObject.Phase1IntegrityAlgorithms, &ec2.Phase1IntegrityAlgorithmsRequestListValue{Value: aws.String(v.(string))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase1_lifetime_seconds"); ok {
		apiObject.Phase1LifetimeSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "phase2_dh_group_numbers"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase2DHGroupNumbers = append(apiObject.Phase2DHGroupNumbers, &ec2.Phase2DHGroupNumbersRequestListValue{Value: aws.Int64(int64(v.(int)))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase2_encryption_algorithms"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase2EncryptionAlgorithms = append(apiObject.Phase2EncryptionAlgorithms, &ec2.Phase2EncryptionAlgorithmsRequestListValue{Value: aws.String(v.(string))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase2_integrity_algorithms"); ok {
		for _, v := range v.(*schema.Set).List() {
			apiObject.Phase2IntegrityAlgorithms = append(apiObject.Phase2IntegrityAlgorithms, &ec2.Phase2IntegrityAlgorithmsRequestListValue{Value: aws.String(v.(string))})
		}
	}

	if v, ok := d.GetOk(prefix + "phase2_lifetime_seconds"); ok {
		apiObject.Phase2LifetimeSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "preshared_key"); ok {
		apiObject.PreSharedKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk(prefix + "rekey_fuzz_percentage"); ok {
		apiObject.RekeyFuzzPercentage = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "rekey_margin_time_seconds"); ok {
		apiObject.RekeyMarginTimeSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk(prefix + "replay_window_size"); ok {
		apiObject.ReplayWindowSize = aws.Int64(int64(v.(int)))
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

func expandModifyVPNTunnelOptionsSpecification(d *schema.ResourceData, prefix string) *ec2.ModifyVpnTunnelOptionsSpecification {
	apiObject := &ec2.ModifyVpnTunnelOptionsSpecification{}
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
			apiObject.DPDTimeoutSeconds = aws.Int64(int64(v.(int)))
		} else {
			apiObject.DPDTimeoutSeconds = aws.Int64(int64(defaultVPNTunnelOptionsDPDTimeoutSeconds))
		}

		hasChange = true
	}

	if key := prefix + "ike_versions"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.IKEVersions = append(apiObject.IKEVersions, &ec2.IKEVersionsRequestListValue{Value: aws.String(v.(string))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsIKEVersions {
				apiObject.IKEVersions = append(apiObject.IKEVersions, &ec2.IKEVersionsRequestListValue{Value: aws.String(v)})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase1_dh_group_numbers"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase1DHGroupNumbers = append(apiObject.Phase1DHGroupNumbers, &ec2.Phase1DHGroupNumbersRequestListValue{Value: aws.Int64(int64(v.(int)))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase1DHGroupNumbers {
				apiObject.Phase1DHGroupNumbers = append(apiObject.Phase1DHGroupNumbers, &ec2.Phase1DHGroupNumbersRequestListValue{Value: aws.Int64(int64(v))})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase1_encryption_algorithms"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase1EncryptionAlgorithms = append(apiObject.Phase1EncryptionAlgorithms, &ec2.Phase1EncryptionAlgorithmsRequestListValue{Value: aws.String(v.(string))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase1EncryptionAlgorithms {
				apiObject.Phase1EncryptionAlgorithms = append(apiObject.Phase1EncryptionAlgorithms, &ec2.Phase1EncryptionAlgorithmsRequestListValue{Value: aws.String(v)})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase1_integrity_algorithms"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase1IntegrityAlgorithms = append(apiObject.Phase1IntegrityAlgorithms, &ec2.Phase1IntegrityAlgorithmsRequestListValue{Value: aws.String(v.(string))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase1IntegrityAlgorithms {
				apiObject.Phase1IntegrityAlgorithms = append(apiObject.Phase1IntegrityAlgorithms, &ec2.Phase1IntegrityAlgorithmsRequestListValue{Value: aws.String(v)})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase1_lifetime_seconds"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.Phase1LifetimeSeconds = aws.Int64(int64(v.(int)))
		} else {
			apiObject.Phase1LifetimeSeconds = aws.Int64(int64(defaultVPNTunnelOptionsPhase1LifetimeSeconds))
		}

		hasChange = true
	}

	if key := prefix + "phase2_dh_group_numbers"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase2DHGroupNumbers = append(apiObject.Phase2DHGroupNumbers, &ec2.Phase2DHGroupNumbersRequestListValue{Value: aws.Int64(int64(v.(int)))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase2DHGroupNumbers {
				apiObject.Phase2DHGroupNumbers = append(apiObject.Phase2DHGroupNumbers, &ec2.Phase2DHGroupNumbersRequestListValue{Value: aws.Int64(int64(v))})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase2_encryption_algorithms"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase2EncryptionAlgorithms = append(apiObject.Phase2EncryptionAlgorithms, &ec2.Phase2EncryptionAlgorithmsRequestListValue{Value: aws.String(v.(string))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase2EncryptionAlgorithms {
				apiObject.Phase2EncryptionAlgorithms = append(apiObject.Phase2EncryptionAlgorithms, &ec2.Phase2EncryptionAlgorithmsRequestListValue{Value: aws.String(v)})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase2_integrity_algorithms"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok && v.(*schema.Set).Len() > 0 {
			for _, v := range d.Get(key).(*schema.Set).List() {
				apiObject.Phase2IntegrityAlgorithms = append(apiObject.Phase2IntegrityAlgorithms, &ec2.Phase2IntegrityAlgorithmsRequestListValue{Value: aws.String(v.(string))})
			}
		} else {
			for _, v := range defaultVPNTunnelOptionsPhase2IntegrityAlgorithms {
				apiObject.Phase2IntegrityAlgorithms = append(apiObject.Phase2IntegrityAlgorithms, &ec2.Phase2IntegrityAlgorithmsRequestListValue{Value: aws.String(v)})
			}
		}

		hasChange = true
	}

	if key := prefix + "phase2_lifetime_seconds"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.Phase2LifetimeSeconds = aws.Int64(int64(v.(int)))
		} else {
			apiObject.Phase2LifetimeSeconds = aws.Int64(int64(defaultVPNTunnelOptionsPhase2LifetimeSeconds))
		}

		hasChange = true
	}

	if key := prefix + "preshared_key"; d.HasChange(key) {
		apiObject.PreSharedKey = aws.String(d.Get(key).(string))

		hasChange = true
	}

	if key := prefix + "rekey_fuzz_percentage"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.RekeyFuzzPercentage = aws.Int64(int64(v.(int)))
		} else {
			apiObject.RekeyFuzzPercentage = aws.Int64(int64(defaultVPNTunnelOptionsRekeyFuzzPercentage))
		}

		hasChange = true
	}

	if key := prefix + "rekey_margin_time_seconds"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.RekeyMarginTimeSeconds = aws.Int64(int64(v.(int)))
		} else {
			apiObject.RekeyMarginTimeSeconds = aws.Int64(int64(defaultVPNTunnelOptionsRekeyMarginTimeSeconds))
		}

		hasChange = true
	}

	if key := prefix + "replay_window_size"; d.HasChange(key) {
		if v, ok := d.GetOk(key); ok {
			apiObject.ReplayWindowSize = aws.Int64(int64(v.(int)))
		} else {
			apiObject.ReplayWindowSize = aws.Int64(int64(defaultVPNTunnelOptionsReplayWindowSize))
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

func flattenTunnelOption(d *schema.ResourceData, prefix string, apiObject *ec2.TunnelOption) {
	if apiObject == nil {
		return
	}

	var s []*string
	var i []*int64

	d.Set(prefix+"dpd_timeout_action", apiObject.DpdTimeoutAction)
	d.Set(prefix+"dpd_timeout_seconds", apiObject.DpdTimeoutSeconds)

	for _, v := range apiObject.IkeVersions {
		s = append(s, v.Value)
	}
	d.Set(prefix+"ike_versions", aws.StringValueSlice(s))
	s = nil

	for _, v := range apiObject.Phase1DHGroupNumbers {
		i = append(i, v.Value)
	}
	d.Set(prefix+"phase1_dh_group_numbers", aws.Int64ValueSlice(i))
	i = nil

	for _, v := range apiObject.Phase1EncryptionAlgorithms {
		s = append(s, v.Value)
	}
	d.Set(prefix+"phase1_encryption_algorithms", aws.StringValueSlice(s))
	s = nil

	for _, v := range apiObject.Phase1IntegrityAlgorithms {
		s = append(s, v.Value)
	}
	d.Set(prefix+"phase1_integrity_algorithms", aws.StringValueSlice(s))
	s = nil

	d.Set(prefix+"phase1_lifetime_seconds", apiObject.Phase1LifetimeSeconds)

	for _, v := range apiObject.Phase2DHGroupNumbers {
		i = append(i, v.Value)
	}
	d.Set(prefix+"phase2_dh_group_numbers", aws.Int64ValueSlice(i))

	for _, v := range apiObject.Phase2EncryptionAlgorithms {
		s = append(s, v.Value)
	}
	d.Set(prefix+"phase2_encryption_algorithms", aws.StringValueSlice(s))
	s = nil

	for _, v := range apiObject.Phase2IntegrityAlgorithms {
		s = append(s, v.Value)
	}
	d.Set(prefix+"phase2_integrity_algorithms", aws.StringValueSlice(s))

	d.Set(prefix+"phase2_lifetime_seconds", apiObject.Phase2LifetimeSeconds)

	d.Set(prefix+"rekey_fuzz_percentage", apiObject.RekeyFuzzPercentage)
	d.Set(prefix+"rekey_margin_time_seconds", apiObject.RekeyMarginTimeSeconds)
	d.Set(prefix+"replay_window_size", apiObject.ReplayWindowSize)
	d.Set(prefix+"startup_action", apiObject.StartupAction)
	d.Set(prefix+"inside_cidr", apiObject.TunnelInsideCidr)
	d.Set(prefix+"inside_ipv6_cidr", apiObject.TunnelInsideIpv6Cidr)
}

func flattenVPNStaticRoute(apiObject *ec2.VpnStaticRoute) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DestinationCidrBlock; v != nil {
		tfMap["destination_cidr_block"] = aws.StringValue(v)
	}

	if v := apiObject.Source; v != nil {
		tfMap["source"] = aws.StringValue(v)
	}

	if v := apiObject.State; v != nil {
		tfMap["state"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenVPNStaticRoutes(apiObjects []*ec2.VpnStaticRoute) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenVPNStaticRoute(apiObject))
	}

	return tfList
}

func flattenVGWTelemetry(apiObject *ec2.VgwTelemetry) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AcceptedRouteCount; v != nil {
		tfMap["accepted_route_count"] = aws.Int64Value(v)
	}

	if v := apiObject.CertificateArn; v != nil {
		tfMap["certificate_arn"] = aws.StringValue(v)
	}

	if v := apiObject.LastStatusChange; v != nil {
		tfMap["last_status_change"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := apiObject.OutsideIpAddress; v != nil {
		tfMap["outside_ip_address"] = aws.StringValue(v)
	}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	if v := apiObject.StatusMessage; v != nil {
		tfMap["status_message"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenVGWTelemetries(apiObjects []*ec2.VgwTelemetry) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenVGWTelemetry(apiObject))
	}

	return tfList
}

type XmlVpnConnectionConfig struct {
	Tunnels []XmlIpsecTunnel `xml:"ipsec_tunnel"`
}

type XmlIpsecTunnel struct {
	BGPASN           string `xml:"vpn_gateway>bgp>asn"`
	BGPHoldTime      int    `xml:"vpn_gateway>bgp>hold_time"`
	CgwInsideAddress string `xml:"customer_gateway>tunnel_inside_address>ip_address"`
	OutsideAddress   string `xml:"vpn_gateway>tunnel_outside_address>ip_address"`
	PreSharedKey     string `xml:"ike>pre_shared_key"`
	VgwInsideAddress string `xml:"vpn_gateway>tunnel_inside_address>ip_address"`
}

type TunnelInfo struct {
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

func (slice XmlVpnConnectionConfig) Len() int {
	return len(slice.Tunnels)
}

func (slice XmlVpnConnectionConfig) Less(i, j int) bool {
	return slice.Tunnels[i].OutsideAddress < slice.Tunnels[j].OutsideAddress
}

func (slice XmlVpnConnectionConfig) Swap(i, j int) {
	slice.Tunnels[i], slice.Tunnels[j] = slice.Tunnels[j], slice.Tunnels[i]
}

// CustomerGatewayConfigurationToTunnelInfo converts the configuration information for the
// VPN connection's customer gateway (in the native XML format) to a TunnelInfo structure.
// The tunnel1 parameters are optionally used to correctly order tunnel configurations.
func CustomerGatewayConfigurationToTunnelInfo(xmlConfig string, tunnel1PreSharedKey string, tunnel1InsideCidr string, tunnel1InsideIpv6Cidr string) (*TunnelInfo, error) {
	var vpnConfig XmlVpnConnectionConfig

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

	tunnelInfo := &TunnelInfo{
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
		validation.StringDoesNotMatch(regexp.MustCompile(`^0`), "cannot start with zero character"),
		validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z_.]+$`), "can only contain alphanumeric, period and underscore characters"),
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
		validation.StringMatch(regexp.MustCompile(`^169\.254\.`), "must be within 169.254.0.0/16"),
		validation.StringNotInSlice(disallowedCidrs, false),
	)
}

func validVPNConnectionTunnelInsideIPv6CIDR() schema.SchemaValidateFunc {
	return validation.All(
		validation.IsCIDRNetwork(126, 126),
		validation.StringMatch(regexp.MustCompile(`^fd00:`), "must be within fd00::/8"),
	)
}
