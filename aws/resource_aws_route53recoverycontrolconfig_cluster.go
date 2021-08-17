package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53recoverycontrolconfig"
)

func resourceAwsRoute53RecoveryControlConfigCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53RecoveryControlConfigClusterCreate,
		Read:   resourceAwsRoute53RecoveryControlConfigClusterRead,
		Delete: resourceAwsRoute53RecoveryControlConfigClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_endpoint": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"endpoint": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"region": {
										Type:     schema.TypeString,
										Computed: true,
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

func resourceAwsRoute53RecoveryControlConfigClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.CreateClusterInput{
		ClientToken: aws.String(resource.UniqueId()),
		ClusterName: aws.String(d.Get("name").(string)),
	}

	output, err := conn.CreateCluster(input)
	result := output.Cluster

	if err != nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Cluster: %w", err)
	}

	if result == nil {
		return fmt.Errorf("Error creating Route53 Recovery Control Config Cluster empty response")
	}

	d.SetId(aws.StringValue(result.ClusterArn))

	if _, err := waiter.Route53RecoveryControlConfigClusterCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config Cluster (%s) to be Deployed: %w", d.Id(), err)
	}

	return resourceAwsRoute53RecoveryControlConfigClusterRead(d, meta)
}

func resourceAwsRoute53RecoveryControlConfigClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.DescribeClusterInput{
		ClusterArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeCluster(input)
	result := output.Cluster

	if err != nil {
		return fmt.Errorf("Error describing Route53 Recovery Control Config Cluster: %s", err)
	}

	if !d.IsNewResource() && result == nil {
		log.Printf("[WARN] SRoute53 Recovery Control Config Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("cluster_arn", result.ClusterArn)
	d.Set("cluster_enpoints", result.ClusterEndpoints)
	d.Set("name", result.Name)
	d.Set("status", result.Status)

	if err := d.Set("cluster_enpoints", flattenRoute53RecoveryControlConfigClusterEndpoints(result.ClusterEndpoints)); err != nil {
		return fmt.Errorf("Error setting cluster_endpoints: %w", err)
	}

	return nil
}

func resourceAwsRoute53RecoveryControlConfigClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53recoverycontrolconfigconn

	input := &route53recoverycontrolconfig.DeleteClusterInput{
		ClusterArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteCluster(input)

	if err != nil {
		if isAWSErr(err, route53recoverycontrolconfig.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Route53 Recovery Control Config Cluster: %s", err)
	}

	if _, err := waiter.Route53RecoveryControlConfigClusterDeleted(conn, d.Id()); err != nil {
		if isResourceNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("Error waiting for Route53 Recovery Control Config  Cluster (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func flattenRoute53RecoveryControlConfigClusterEndpoints(endpoints []*route53recoverycontrolconfig.ClusterEndpoint) []interface{} {
	if len(endpoints) == 0 {
        return nil
    }

    var tfList []interface{}

    for _, endpoint := range endpoints {
        if endpoint == nil {
            continue
        }

        tfList = append(tfList, flattenRoute53RecoveryControlConfigClusterEndpoint(endpoint))
    }

    return tfList
}

func flattenRoute53RecoveryControlConfigClusterEndpoint(ce *route53recoverycontrolconfig.ClusterEndpoint) map[string]interface{} {
	if ce == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := ce.Endpoint; v != nil {
		tfMap["endpoint"] = aws.StringValue(v)
	}

	if v := ce.Region; v != nil {
		tfMap["region"] = aws.StringValue(v)
	}

	return tfMap
}