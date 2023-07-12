package kafka

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_msk_bootstrap_brokers", name="BootstrapBrokers")
func DataSourceBootstrapBrokers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBootstrapBrokers,

		Schema: map[string]*schema.Schema{

			"bootstrap_broker_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_public_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_public_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_public_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_vpc_connectivity_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_vpc_connectivity_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_broker_string_vpc_connectivity_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBootstrapBrokers(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn(ctx)

	clusterArn := d.Get("cluster_arn").(string)
	input := &kafka.GetBootstrapBrokersInput{
		ClusterArn: aws.String(clusterArn),
	}

	req, resp := conn.GetBootstrapBrokersRequest(input)
	err := req.Send()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Getting MSK Cluster (%s) Boostrap Brokers: %s", clusterArn, err)
	}

	d.Set("bootstrap_broker_string", resp.BootstrapBrokerString)
	d.Set("bootstrap_broker_string_tls", resp.BootstrapBrokerStringTls)
	d.Set("bootstrap_broker_string_sasl_scram", resp.BootstrapBrokerStringSaslScram)
	d.Set("bootstrap_broker_string_sasl_iam", resp.BootstrapBrokerStringSaslIam)
	d.Set("bootstrap_broker_string_public_tls", resp.BootstrapBrokerStringPublicTls)
	d.Set("bootstrap_broker_string_public_sasl_scram", resp.BootstrapBrokerStringPublicSaslScram)
	d.Set("bootstrap_broker_string_public_sasl_iam", resp.BootstrapBrokerStringPublicSaslIam)
	d.Set("bootstrap_broker_string_vpc_connectivity_tls", resp.BootstrapBrokerStringVpcConnectivityTls)
	d.Set("bootstrap_broker_string_vpc_connectivity_sasl_scram", resp.BootstrapBrokerStringVpcConnectivitySaslScram)
	d.Set("bootstrap_broker_string_public_sasl_iam", resp.BootstrapBrokerStringVpcConnectivitySaslIam)

	return diags
}
