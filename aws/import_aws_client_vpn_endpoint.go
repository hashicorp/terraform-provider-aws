package aws

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

// Client VPN endpoint import also imports all assocations
func resourceAwsClientVpnEndpointImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*AWSClient).ec2conn

	id := d.Id()
	resp, err := conn.DescribeClientVpnEndpoints(&ec2.DescribeClientVpnEndpointsInput{
		ClientVpnEndpointIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		return nil, err
	}

	if len(resp.ClientVpnEndpoints) < 1 || resp.ClientVpnEndpoints[0] == nil {
		return nil, fmt.Errorf("Client VPN endpoint %s was not found", id)
	}
	results := make([]*schema.ResourceData, 1)

	clientVpnEndpointData := resp.ClientVpnEndpoints[0]

	err = flattenClientVpnEndpointConfig(d, clientVpnEndpointData)
	if err != nil {
		return nil, err
	}

	results[0] = d

	return results, nil
}

func flattenClientVpnEndpointConfig(d *schema.ResourceData, clientVpnEndpointConfig *ec2.ClientVpnEndpoint) error {
	var err error

	d.SetId(*clientVpnEndpointConfig.ClientVpnEndpointId)
	if clientVpnEndpointConfig.Description != nil {
		d.Set("description", clientVpnEndpointConfig.Description)
	}
	d.Set("client_cidr_block", clientVpnEndpointConfig.ClientCidrBlock)
	d.Set("server_certificate_arn", clientVpnEndpointConfig.ServerCertificateArn)
	d.Set("transport_protocol", clientVpnEndpointConfig.TransportProtocol)
	d.Set("dns_name", clientVpnEndpointConfig.DnsName)
	d.Set("status", clientVpnEndpointConfig.Status)

	err = d.Set("authentication_options", flattenAuthOptsConfig(clientVpnEndpointConfig.AuthenticationOptions))
	if err != nil {
		return err
	}

	err = d.Set("connection_log_options", flattenConnLoggingConfig(clientVpnEndpointConfig.ConnectionLogOptions))
	//err = d.Set("connection_log_options", schema.NewSet(connLoggingConfigHash, []interface{}{}))
	if err != nil {
		return err
	}

	return nil
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
