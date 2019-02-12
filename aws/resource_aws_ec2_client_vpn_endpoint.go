package aws

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.TransportProtocolUdp,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.TransportProtocolTcp,
					ec2.TransportProtocolUdp,
				}, false),
			},
			"authentication_options": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								ec2.ClientVpnAuthenticationTypeCertificateAuthentication,
								ec2.ClientVpnAuthenticationTypeDirectoryServiceAuthentication,
							}, false),
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
				Type:     schema.TypeList,
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
			"authorization_rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourceAwsAuthRuleHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"target_network_cidr": {
							Type:     schema.TypeString,
							Required: true,
						},
						"access_group_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"authorize_all_groups": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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
		authOptsSet := v.([]interface{})
		attrs := authOptsSet[0].(map[string]interface{})

		authOptsReq := &ec2.ClientVpnAuthenticationRequest{
			Type: aws.String(attrs["type"].(string)),
		}

		if attrs["type"].(string) == "certificate-authentication" {
			authOptsReq.MutualAuthentication = &ec2.CertificateAuthenticationRequest{
				ClientRootCertificateChainArn: aws.String(attrs["root_certificate_chain_arn"].(string)),
			}
		}

		if attrs["type"].(string) == "directory-service-authentication" {
			authOptsReq.ActiveDirectory = &ec2.DirectoryServiceAuthenticationRequest{
				DirectoryId: aws.String(attrs["active_directory_id"].(string)),
			}
		}

		req.AuthenticationOptions = []*ec2.ClientVpnAuthenticationRequest{authOptsReq}
	}

	if v, ok := d.GetOk("connection_log_options"); ok {
		connLogSet := v.([]interface{})
		attrs := connLogSet[0].(map[string]interface{})

		connLogReq := &ec2.ConnectionLogOptions{
			Enabled: aws.Bool(attrs["enabled"].(bool)),
		}

		if attrs["enabled"].(bool) && attrs["cloudwatch_log_group"].(string) != "" {
			connLogReq.CloudwatchLogGroup = aws.String(attrs["cloudwatch_log_group"].(string))
		}

		if attrs["enabled"].(bool) && attrs["cloudwatch_log_stream"].(string) != "" {
			connLogReq.CloudwatchLogStream = aws.String(attrs["cloudwatch_log_stream"].(string))
		}

		req.ConnectionLogOptions = connLogReq
	}

	log.Printf("[DEBUG] Creating Client VPN endpoint: %#v", req)
	var resp *ec2.CreateClientVpnEndpointOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateClientVpnEndpoint(req)
		if isAWSErr(err, "OperationNotPermitted", "Endpoint cannot be created while another endpoint is being created") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error creating Client VPN endpoint: %s", err)
	}

	d.SetId(*resp.ClientVpnEndpointId)

	if _, ok := d.GetOk("authorization_rule"); ok {
		rules := addAuthorizationRules(d.Get("authorization_rule").(*schema.Set).List())

		for _, r := range rules {
			_, err := conn.AuthorizeClientVpnIngress(r)
			if err != nil {
				return fmt.Errorf("Failure adding new Client VPN authorization rules: %s", err)
			}
		}
	}

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

	if result == nil || len(result.ClientVpnEndpoints) == 0 || result.ClientVpnEndpoints[0] == nil {
		log.Printf("[WARN] EC2 Client VPN Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if result.ClientVpnEndpoints[0].Status != nil && aws.StringValue(result.ClientVpnEndpoints[0].Status.Code) == ec2.ClientVpnEndpointStatusCodeDeleted {
		log.Printf("[WARN] EC2 Client VPN Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
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

	authRuleResult, err := conn.DescribeClientVpnAuthorizationRules(&ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	d.Set("authorization_rule", flattenAuthRules(authRuleResult.AuthorizationRules))

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
			connSet := v.([]interface{})
			attrs := connSet[0].(map[string]interface{})

			connReq := &ec2.ConnectionLogOptions{
				Enabled: aws.Bool(attrs["enabled"].(bool)),
			}

			if attrs["enabled"].(bool) && attrs["cloudwatch_log_group"].(string) != "" {
				connReq.CloudwatchLogGroup = aws.String(attrs["cloudwatch_log_group"].(string))
			}

			if attrs["enabled"].(bool) && attrs["cloudwatch_log_stream"].(string) != "" {
				connReq.CloudwatchLogStream = aws.String(attrs["cloudwatch_log_stream"].(string))
			}

			req.ConnectionLogOptions = connReq
		}
	}

	_, err := conn.ModifyClientVpnEndpoint(req)
	if err != nil {
		return fmt.Errorf("Error modifying Client VPN endpoint: %s", err)
	}

	if d.HasChange("authorization_rule") {
		o, n := d.GetChange("authorization_rule")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		remove := removeAuthorizationRules(os.Difference(ns).List())
		add := addAuthorizationRules(ns.Difference(os).List())

		if len(remove) > 0 {
			for _, r := range remove {
				log.Printf("[DEBUG] Client VPN authorization rule opts: %s", r)
				_, err := conn.RevokeClientVpnIngress(r)
				if err != nil {
					return fmt.Errorf("Failure removing outdated Client VPN authorization rules: %s", err)
				}
			}
		}

		if len(add) > 0 {
			for _, r := range add {
				log.Printf("[DEBUG] Client VPN authorization rule opts: %s", r)
				_, err := conn.AuthorizeClientVpnIngress(r)
				if err != nil {
					return fmt.Errorf("Failure adding new Client VPN authorization rules: %s", err)
				}
			}
		}
	}

	return resourceAwsEc2ClientVpnEndpointRead(d, meta)
}

func flattenConnLoggingConfig(lopts *ec2.ConnectionLogResponseOptions) []map[string]interface{} {
	m := make(map[string]interface{})
	if lopts.CloudwatchLogGroup != nil {
		m["cloudwatch_log_group"] = *lopts.CloudwatchLogGroup
	}
	if lopts.CloudwatchLogStream != nil {
		m["cloudwatch_log_stream"] = *lopts.CloudwatchLogStream
	}
	m["enabled"] = *lopts.Enabled
	return []map[string]interface{}{m}
}

func flattenAuthOptsConfig(aopts []*ec2.ClientVpnAuthentication) []map[string]interface{} {
	m := make(map[string]interface{})
	if aopts[0].MutualAuthentication != nil {
		m["root_certificate_chain_arn"] = *aopts[0].MutualAuthentication.ClientRootCertificateChain
	}
	if aopts[0].ActiveDirectory != nil {
		m["active_directory_id"] = *aopts[0].ActiveDirectory.DirectoryId
	}
	m["type"] = *aopts[0].Type
	return []map[string]interface{}{m}
}

func flattenAuthRules(list []*ec2.AuthorizationRule) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		l := map[string]interface{}{
			"target_network_cidr": *i.DestinationCidr,
		}

		if i.Description != nil {
			l["description"] = aws.String(*i.Description)
		}
		if i.GroupId != nil {
			l["access_group_id"] = aws.String(*i.GroupId)
		}
		if i.AccessAll != nil {
			l["authorize_all_groups"] = aws.Bool(*i.AccessAll)
		}

		result = append(result, l)
	}
	return result
}

func addAuthorizationRules(configured []interface{}) []*ec2.AuthorizeClientVpnIngressInput {
	rules := make([]*ec2.AuthorizeClientVpnIngressInput, 0, len(configured))

	for _, i := range configured {
		item := i.(map[string]interface{})
		rule := &ec2.AuthorizeClientVpnIngressInput{}

		rule.ClientVpnEndpointId = aws.String(item["id"].(string))
		rule.TargetNetworkCidr = aws.String(item["target_network_cidr"].(string))

		if item["description"].(string) != "" {
			rule.Description = aws.String(item["description"].(string))
		}

		if item["access_group_id"].(string) != "" {
			rule.AccessGroupId = aws.String(item["access_group_id"].(string))
		}

		if item["authorize_all_groups"] != nil {
			rule.AuthorizeAllGroups = aws.Bool(item["authorize_all_groups"].(bool))
		}

		rules = append(rules, rule)
	}

	return rules
}

func removeAuthorizationRules(configured []interface{}) []*ec2.RevokeClientVpnIngressInput {
	rules := make([]*ec2.RevokeClientVpnIngressInput, 0, len(configured))

	for _, i := range configured {
		item := i.(map[string]interface{})
		rule := &ec2.RevokeClientVpnIngressInput{}

		rule.ClientVpnEndpointId = aws.String(item["id"].(string))
		rule.TargetNetworkCidr = aws.String(item["target_network_cidr"].(string))

		rules = append(rules, rule)
	}

	return rules
}

func resourceAwsAuthRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["description"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["target_network_cidr"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["access_group_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["authorize_all_groups"]; ok {
		buf.WriteString(fmt.Sprintf("%v-", v.(bool)))
	}

	return hashcode.String(buf.String())
}
