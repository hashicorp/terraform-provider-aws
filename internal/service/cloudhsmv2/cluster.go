// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudhsmv2

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudhsmv2"
	"github.com/aws/aws-sdk-go-v2/service/cloudhsmv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudhsm_v2_cluster", name="Cluster")
// @Tags(identifierAttribute="id")
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_certificates": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_hardware_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_csr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hsm_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"manufacturer_hardware_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hsm_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"hsm1.medium"}, false),
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_backup_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Client(ctx)

	input := &cloudhsmv2.CreateClusterInput{
		HsmType:   aws.String(d.Get("hsm_type").(string)),
		SubnetIds: flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
		TagList:   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("source_backup_identifier"); ok {
		input.SourceBackupId = aws.String(v.(string))
	}

	output, err := conn.CreateCluster(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudHSMv2 Cluster: %s", err)
	}

	d.SetId(aws.ToString(output.Cluster.ClusterId))

	f := waitClusterUninitialized
	if input.SourceBackupId != nil {
		f = waitClusterActive
	}

	if _, err := f(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudHSMv2 Cluster (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Client(ctx)

	cluster, err := findClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudHSMv2 Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudHSMv2 Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("cluster_certificates", flattenCertificates(cluster)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_certificates: %s", err)
	}
	d.Set("cluster_id", cluster.ClusterId)
	d.Set("cluster_state", cluster.State)
	d.Set("hsm_type", cluster.HsmType)
	d.Set("security_group_id", cluster.SecurityGroup)
	d.Set("source_backup_identifier", cluster.SourceBackupId)
	d.Set(names.AttrSubnetIDs, tfmaps.Values(cluster.SubnetMapping))
	d.Set(names.AttrVPCID, cluster.VpcId)

	setTagsOut(ctx, cluster.TagList)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Client(ctx)

	log.Printf("[INFO] Deleting CloudHSMv2 Cluster: %s", d.Id())
	_, err := conn.DeleteCluster(ctx, &cloudhsmv2.DeleteClusterInput{
		ClusterId: aws.String(d.Id()),
	})

	if errs.IsA[*types.CloudHsmResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudHSMv2 Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudHSMv2 Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findClusterByID(ctx context.Context, conn *cloudhsmv2.Client, id string) (*types.Cluster, error) {
	input := &cloudhsmv2.DescribeClustersInput{
		Filters: map[string][]string{
			"clusterIds": {id},
		},
	}

	output, err := findCluster(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == types.ClusterStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.ClusterId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findCluster(ctx context.Context, conn *cloudhsmv2.Client, input *cloudhsmv2.DescribeClustersInput) (*types.Cluster, error) {
	output, err := findClusters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClusters(ctx context.Context, conn *cloudhsmv2.Client, input *cloudhsmv2.DescribeClustersInput) ([]types.Cluster, error) {
	var output []types.Cluster

	pages := cloudhsmv2.NewDescribeClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Clusters...)
	}

	return output, nil
}

func statusCluster(ctx context.Context, conn *cloudhsmv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), err
	}
}

func waitClusterActive(ctx context.Context, conn *cloudhsmv2.Client, id string, timeout time.Duration) (*types.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ClusterStateCreateInProgress, types.ClusterStateInitializeInProgress),
		Target:     enum.Slice(types.ClusterStateActive),
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *cloudhsmv2.Client, id string, timeout time.Duration) (*types.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ClusterStateDeleteInProgress),
		Target:     []string{},
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func waitClusterUninitialized(ctx context.Context, conn *cloudhsmv2.Client, id string, timeout time.Duration) (*types.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ClusterStateCreateInProgress, types.ClusterStateInitializeInProgress),
		Target:     enum.Slice(types.ClusterStateUninitialized),
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func flattenCertificates(apiObject *types.Cluster) []map[string]interface{} {
	tfMap := map[string]interface{}{}

	if apiObject, clusterState := apiObject.Certificates, apiObject.State; apiObject != nil {
		if clusterState == types.ClusterStateUninitialized {
			tfMap["cluster_csr"] = aws.ToString(apiObject.ClusterCsr)
			tfMap["aws_hardware_certificate"] = aws.ToString(apiObject.AwsHardwareCertificate)
			tfMap["hsm_certificate"] = aws.ToString(apiObject.HsmCertificate)
			tfMap["manufacturer_hardware_certificate"] = aws.ToString(apiObject.ManufacturerHardwareCertificate)
		} else if clusterState == types.ClusterStateActive {
			tfMap["cluster_certificate"] = aws.ToString(apiObject.ClusterCertificate)
		}
	}

	if len(tfMap) > 0 {
		return []map[string]interface{}{tfMap}
	}

	return []map[string]interface{}{}
}
