// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoverycontrolconfig_cluster")
func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		DeleteWithoutTimeout: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEndpoint: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrRegion: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	input := &r53rcc.CreateClusterInput{
		ClientToken: aws.String(id.UniqueId()),
		ClusterName: aws.String(d.Get(names.AttrName).(string)),
	}

	output, err := conn.CreateClusterWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Cluster: %s", err)
	}

	if output == nil || output.Cluster == nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Cluster: empty response")
	}

	result := output.Cluster
	d.SetId(aws.StringValue(result.ClusterArn))

	if _, err := waitClusterCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Cluster (%s) to be Deployed: %s", d.Id(), err)
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	input := &r53rcc.DescribeClusterInput{
		ClusterArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeClusterWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53 Recovery Control Config Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Cluster: %s", err)
	}

	if output == nil || output.Cluster == nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Cluster: %s", "empty response")
	}

	result := output.Cluster
	d.Set(names.AttrARN, result.ClusterArn)
	d.Set(names.AttrName, result.Name)
	d.Set(names.AttrStatus, result.Status)

	if err := d.Set("cluster_endpoints", flattenClusterEndpoints(result.ClusterEndpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_endpoints: %s", err)
	}

	return diags
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigConn(ctx)

	log.Printf("[INFO] Deleting Route53 Recovery Control Config Cluster: %s", d.Id())
	_, err := conn.DeleteClusterWithContext(ctx, &r53rcc.DeleteClusterInput{
		ClusterArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Control Config Cluster: %s", err)
	}

	_, err = waitClusterDeleted(ctx, conn, d.Id())

	if tfawserr.ErrCodeEquals(err, r53rcc.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config  Cluster (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func flattenClusterEndpoints(endpoints []*r53rcc.ClusterEndpoint) []interface{} {
	if len(endpoints) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}

		tfList = append(tfList, flattenClusterEndpoint(endpoint))
	}

	return tfList
}

func flattenClusterEndpoint(ce *r53rcc.ClusterEndpoint) map[string]interface{} {
	if ce == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := ce.Endpoint; v != nil {
		tfMap[names.AttrEndpoint] = aws.StringValue(v)
	}

	if v := ce.Region; v != nil {
		tfMap[names.AttrRegion] = aws.StringValue(v)
	}

	return tfMap
}
