// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudhsmv2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudhsm_v2_cluster", name="Cluster")
// @Tags(identifierAttribute="id")
func ResourceCluster() *schema.Resource {
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
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Conn(ctx)

	input := &cloudhsmv2.CreateClusterInput{
		HsmType:   aws.String(d.Get("hsm_type").(string)),
		SubnetIds: flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		TagList:   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("source_backup_identifier"); ok {
		input.SourceBackupId = aws.String(v.(string))
	}

	output, err := conn.CreateClusterWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudHSMv2 Cluster: %s", err)
	}

	d.SetId(aws.StringValue(output.Cluster.ClusterId))

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
	conn := meta.(*conns.AWSClient).CloudHSMV2Conn(ctx)

	cluster, err := FindClusterByID(ctx, conn, d.Id())

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
	var subnetIDs []string
	for _, v := range cluster.SubnetMapping {
		subnetIDs = append(subnetIDs, aws.StringValue(v))
	}
	d.Set("subnet_ids", subnetIDs)
	d.Set("vpc_id", cluster.VpcId)

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
	conn := meta.(*conns.AWSClient).CloudHSMV2Conn(ctx)

	log.Printf("[INFO] Deleting CloudHSMv2 Cluster: %s", d.Id())
	_, err := conn.DeleteClusterWithContext(ctx, &cloudhsmv2.DeleteClusterInput{
		ClusterId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudhsmv2.ErrCodeCloudHsmResourceNotFoundException) {
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

func flattenCertificates(apiObject *cloudhsmv2.Cluster) []map[string]interface{} {
	tfMap := map[string]interface{}{}

	if apiObject, clusterState := apiObject.Certificates, aws.StringValue(apiObject.State); apiObject != nil {
		if clusterState == cloudhsmv2.ClusterStateUninitialized {
			tfMap["cluster_csr"] = aws.StringValue(apiObject.ClusterCsr)
			tfMap["aws_hardware_certificate"] = aws.StringValue(apiObject.AwsHardwareCertificate)
			tfMap["hsm_certificate"] = aws.StringValue(apiObject.HsmCertificate)
			tfMap["manufacturer_hardware_certificate"] = aws.StringValue(apiObject.ManufacturerHardwareCertificate)
		} else if clusterState == cloudhsmv2.ClusterStateActive {
			tfMap["cluster_certificate"] = aws.StringValue(apiObject.ClusterCertificate)
		}
	}

	if len(tfMap) > 0 {
		return []map[string]interface{}{tfMap}
	}

	return []map[string]interface{}{}
}
