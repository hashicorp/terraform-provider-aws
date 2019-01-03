package aws

import (
	"bytes"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsEc2ClientVpnEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2ClientVpnEndpointCreate,
		Read:   resourceAwsEc2ClientVpnEndpointRead,
		Delete: resourceAwsEc2ClientVpnEndpointDelete,
		Update: resourceAwsEc2ClientVpnEndpointUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"client_cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dns_servers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"server_certificate_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"transport_protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "udp",
				ValidateFunc: validation.StringInSlice([]string{"udp", "tcp"}, false),
			},
			"authentication_options": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"certificate-authentication", "directory-service-authentication"}, false),
						},
						"active_directory_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"root_certificate_chain_arn": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"connection_log_options": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_log_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cloudwatch_log_stream": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2ClientVpnEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.CreateClientVpnEndpointInput{
		ClientCidrBlock:      aws.String(d.Get("client_cidr_block").(string)),
		ServerCertificateArn: aws.String(d.Get("server_certificate_arn").(string)),
		TransportProtocol:    aws.String(d.Get("transport_protocol").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dns_servers"); ok {
		req.DnsServers = expandStringList(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("authentication_options"); ok {
		vpnAuthRequest := []*ec2.ClientVpnAuthenticationRequest{}
		authSet := v.(*schema.Set)

		for _, authObject := range authSet.List() {
			auth := authObject.(map[string]interface{})

			authObject := expandEc2ClientVpnAuthenticationRequest(auth)
			vpnAuthRequest = append(vpnAuthRequest, authObject)
		}
		req.AuthenticationOptions = vpnAuthRequest
	}

	if v, ok := d.GetOk("connection_log_options"); ok {
		var connLogRequest *ec2.ConnectionLogOptions
		connSet := v.(*schema.Set)

		for _, connObject := range connSet.List() {
			connData := connObject.(map[string]interface{})

			connLogRequest = expandEc2ConnectionLogOptionsRequest(connData)
		}

		req.ConnectionLogOptions = connLogRequest
	}

	log.Printf("[DEBUG] Creating Client VPN endpoint: %#v", req)
	resp, err := conn.CreateClientVpnEndpoint(req)
	if err != nil {
		return fmt.Errorf("Error creating Client VPN endpoint: %s", err)
	}

	d.SetId(*resp.ClientVpnEndpointId)

	return resourceAwsEc2ClientVpnEndpointRead(d, meta)
}

func resourceAwsEc2ClientVpnEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	var err error

	result, err := conn.DescribeClientVpnEndpoints(&ec2.DescribeClientVpnEndpointsInput{
		ClientVpnEndpointIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		return fmt.Errorf("Error reading Client VPN endpoint: %s", err)
	}

	d.Set("description", result.ClientVpnEndpoints[0].Description)
	d.Set("client_cidr_block", result.ClientVpnEndpoints[0].ClientCidrBlock)
	d.Set("server_certificate_arn", result.ClientVpnEndpoints[0].ServerCertificateArn)
	d.Set("transport_protocol", result.ClientVpnEndpoints[0].TransportProtocol)
	d.Set("dns_name", result.ClientVpnEndpoints[0].DnsName)
	d.Set("status", result.ClientVpnEndpoints[0].Status)

	err = d.Set("authentication_options", flattenAuthOptsConfig(result.ClientVpnEndpoints[0].AuthenticationOptions))
	if err != nil {
		return err
	}

	err = d.Set("connection_log_options", flattenConnLoggingConfig(result.ClientVpnEndpoints[0].ConnectionLogOptions))
	if err != nil {
		return err
	}

	return nil
}

func resourceAwsEc2ClientVpnEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.DeleteClientVpnEndpoint(&ec2.DeleteClientVpnEndpointInput{
		ClientVpnEndpointId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Client VPN endpoint: %s", err)
	}

	return nil
}

func resourceAwsEc2ClientVpnEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.ModifyClientVpnEndpointInput{
		ClientVpnEndpointId: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		req.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("dns_servers") {
		dnsValue := expandStringList(d.Get("dns_servers").(*schema.Set).List())
		var enabledValue *bool

		if len(dnsValue) > 0 {
			enabledValue = aws.Bool(true)
		} else {
			enabledValue = aws.Bool(false)
		}

		dnsMod := &ec2.DnsServersOptionsModifyStructure{
			CustomDnsServers: dnsValue,
			Enabled:          enabledValue,
		}
		req.DnsServers = dnsMod
	}

	if d.HasChange("server_certificate_arn") {
		req.ServerCertificateArn = aws.String(d.Get("server_certificate_arn").(string))
	}

	if d.HasChange("connection_log_options") {
		if v, ok := d.GetOk("connection_log_options"); ok {
			var connLogRequest *ec2.ConnectionLogOptions
			connSet := v.(*schema.Set)

			for _, connObject := range connSet.List() {
				connData := connObject.(map[string]interface{})

				connLogRequest = expandEc2ConnectionLogOptionsRequest(connData)
			}

			req.ConnectionLogOptions = connLogRequest
		}
	}

	_, err := conn.ModifyClientVpnEndpoint(req)
	if err != nil {
		return fmt.Errorf("Error modifying Client VPN endpoint: %s", err)
	}

	return resourceAwsEc2ClientVpnEndpointRead(d, meta)
}

func expandEc2ConnectionLogOptionsRequest(data map[string]interface{}) *ec2.ConnectionLogOptions {
	req := &ec2.ConnectionLogOptions{
		Enabled: aws.Bool(data["enabled"].(bool)),
	}

	if data["enabled"].(bool) == true && data["cloudwatch_log_group"].(string) != "" {
		req.CloudwatchLogGroup = aws.String(data["cloudwatch_log_group"].(string))
	}

	if data["enabled"].(bool) == true && data["cloudwatch_log_stream"].(string) != "" {
		req.CloudwatchLogStream = aws.String(data["cloudwatch_log_stream"].(string))
	}

	return req
}

func expandEc2ClientVpnAuthenticationRequest(data map[string]interface{}) *ec2.ClientVpnAuthenticationRequest {
	req := &ec2.ClientVpnAuthenticationRequest{
		Type: aws.String(data["type"].(string)),
	}

	if data["type"].(string) == "certificate-authentication" {
		req.MutualAuthentication = &ec2.CertificateAuthenticationRequest{
			ClientRootCertificateChainArn: aws.String(data["root_certificate_chain_arn"].(string)),
		}
	}

	if data["type"].(string) == "directory-service-authentication" {
		req.ActiveDirectory = &ec2.DirectoryServiceAuthenticationRequest{
			DirectoryId: aws.String(data["active_directory_id"].(string)),
		}
	}

	return req
}

func flattenConnLoggingConfig(lopts *ec2.ConnectionLogResponseOptions) *schema.Set {
	m := make(map[string]interface{})
	if lopts.CloudwatchLogGroup != nil {
		m["cloudwatch_log_group"] = *lopts.CloudwatchLogGroup
	}
	if lopts.CloudwatchLogStream != nil {
		m["cloudwatch_log_stream"] = *lopts.CloudwatchLogStream
	}
	m["enabled"] = *lopts.Enabled
	return schema.NewSet(connLoggingConfigHash, []interface{}{m})
}

func connLoggingConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if m["cloudwatch_log_group"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["cloudwatch_log_group"].(string)))
	}
	if m["cloudwatch_log_stream"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["cloudwatch_log_stream"].(string)))
	}
	buf.WriteString(fmt.Sprintf("%t-", m["enabled"].(bool)))
	return hashcode.String(buf.String())
}

func flattenAuthOptsConfig(aopts []*ec2.ClientVpnAuthentication) *schema.Set {
	m := make(map[string]interface{})
	if aopts[0].MutualAuthentication != nil {
		m["root_certificate_chain_arn"] = *aopts[0].MutualAuthentication.ClientRootCertificateChain
	}
	if aopts[0].ActiveDirectory != nil {
		m["active_directory_id"] = *aopts[0].ActiveDirectory.DirectoryId
	}
	m["type"] = *aopts[0].Type
	return schema.NewSet(authOptsConfigHash, []interface{}{m})
}

func authOptsConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if m["root_certificate_chain_arn"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["root_certificate_chain_arn"].(string)))
	}
	if m["active_directory_id"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["active_directory_id"].(string)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["type"].(string)))
	return hashcode.String(buf.String())
}
