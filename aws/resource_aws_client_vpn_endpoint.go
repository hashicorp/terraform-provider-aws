package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsClientVpnEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsClientVpnEndpointCreate,
		Read:   resourceAwsClientVpnEndpointRead,
		Delete: resourceAwsClientVpnEndpointDelete,
		Update: resourceAwsClientVpnEndpointUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceAwsClientVpnEndpointImportState,
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "udp",
			},
			"authentication_options": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
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

func resourceAwsClientVpnEndpointCreate(d *schema.ResourceData, meta interface{}) error {
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
	d.Set("dns_name", resp.DnsName)
	d.Set("status", resp.Status)

	return resourceAwsClientVpnEndpointRead(d, meta)
}

func resourceAwsClientVpnEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	result, err := conn.DescribeClientVpnEndpoints(&ec2.DescribeClientVpnEndpointsInput{
		ClientVpnEndpointIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		return fmt.Errorf("Error reading Client VPN endpoint: %s", err)
	}

	d.Set("dns_name", result.ClientVpnEndpoints[0].DnsName)
	d.Set("status", result.ClientVpnEndpoints[0].Status)

	return nil
}

func resourceAwsClientVpnEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.DeleteClientVpnEndpoint(&ec2.DeleteClientVpnEndpointInput{
		ClientVpnEndpointId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Client VPN endpoint: %s", err)
	}

	return nil
}

func resourceAwsClientVpnEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
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

	return resourceAwsClientVpnEndpointRead(d, meta)
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
