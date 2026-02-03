// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package route53recoverycontrolconfig

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	r53rcc "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoverycontrolconfig_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
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
			"network_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.NetworkType](),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	input := &r53rcc.CreateClusterInput{
		ClientToken: aws.String(id.UniqueId()),
		ClusterName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("network_type"); ok {
		input.NetworkType = awstypes.NetworkType(v.(string))
	}

	output, err := conn.CreateCluster(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Cluster: %s", err)
	}

	if output == nil || output.Cluster == nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Control Config Cluster: empty response")
	}

	result := output.Cluster
	d.SetId(aws.ToString(result.ClusterArn))

	if _, err := waitClusterCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Cluster (%s) to be Deployed: %s", d.Id(), err)
	}

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Recovery Control Config Cluster (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	output, err := findClusterByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Route53 Recovery Control Config Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Control Config Cluster: %s", err)
	}

	d.Set(names.AttrARN, output.ClusterArn)
	d.Set(names.AttrName, output.Name)
	d.Set("network_type", output.NetworkType)
	d.Set(names.AttrStatus, output.Status)

	if err := d.Set("cluster_endpoints", flattenClusterEndpoints(output.ClusterEndpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_endpoints: %s", err)
	}

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &r53rcc.UpdateClusterInput{
			ClusterArn: aws.String(d.Id()),
		}

		if d.HasChanges("network_type") {
			input.NetworkType = awstypes.NetworkType(d.Get("network_type").(string))
		}

		output, err := conn.UpdateCluster(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Control Config Cluster: %s", err)
		}

		if output == nil || output.Cluster == nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Control Config Cluster: empty response")
		}

		if _, err := waitClusterUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config Cluster (%s) to be Updated: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryControlConfigClient(ctx)

	log.Printf("[INFO] Deleting Route53 Recovery Control Config Cluster: %s", d.Id())
	_, err := conn.DeleteCluster(ctx, &r53rcc.DeleteClusterInput{
		ClusterArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Control Config Cluster: %s", err)
	}

	_, err = waitClusterDeleted(ctx, conn, d.Id())

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Control Config  Cluster (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func findClusterByARN(ctx context.Context, conn *r53rcc.Client, arn string) (*awstypes.Cluster, error) {
	input := &r53rcc.DescribeClusterInput{
		ClusterArn: aws.String(arn),
	}

	output, err := conn.DescribeCluster(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.Cluster == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Cluster, nil
}

func flattenClusterEndpoints(endpoints []awstypes.ClusterEndpoint) []any {
	if len(endpoints) == 0 {
		return nil
	}

	var tfList []any

	for _, endpoint := range endpoints {
		tfList = append(tfList, flattenClusterEndpoint(endpoint))
	}

	return tfList
}

func flattenClusterEndpoint(ce awstypes.ClusterEndpoint) map[string]any {
	tfMap := map[string]any{}

	if v := ce.Endpoint; v != nil {
		tfMap[names.AttrEndpoint] = aws.ToString(v)
	}

	if v := ce.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	return tfMap
}
