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

// @SDKDataSource("aws_msk_serverless_bootstrap_brokers")
func DataSourceServerlessBootstrapBrokers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBootstrapBrokers,

		Schema: map[string]*schema.Schema{
			"BootstrapBrokerStringPublicSaslIam": {
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

	d.Set("BootstrapBrokerStringPublicSaslIam", resp.BootstrapBrokerStringPublicSaslIam)

	return diags
}
