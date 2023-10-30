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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_cloudhsm_v2_hsm")
func ResourceHSM() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHSMCreate,
		ReadWithoutTimeout:   resourceHSMRead,
		DeleteWithoutTimeout: resourceHSMDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"availability_zone", "subnet_id"},
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_eni_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hsm_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hsm_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"availability_zone", "subnet_id"},
			},
		},
	}
}

func resourceHSMCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Conn(ctx)

	clusterID := d.Get("cluster_id").(string)
	input := &cloudhsmv2.CreateHsmInput{
		ClusterId: aws.String(clusterID),
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		input.AvailabilityZone = aws.String(v.(string))
	} else {
		cluster, err := FindClusterByID(ctx, conn, clusterID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudHSMv2 Cluster (%s): %s", clusterID, err)
		}

		subnetID := d.Get("subnet_id").(string)
		for az, sn := range cluster.SubnetMapping {
			if aws.StringValue(sn) == subnetID {
				input.AvailabilityZone = aws.String(az)
			}
		}
	}

	if v, ok := d.GetOk("ip_address"); ok {
		input.IpAddress = aws.String(v.(string))
	}

	output, err := conn.CreateHsmWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudHSMv2 HSM: %s", err)
	}

	d.SetId(aws.StringValue(output.Hsm.HsmId))

	if _, err := waitHSMCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudHSMv2 HSM (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceHSMRead(ctx, d, meta)...)
}

func resourceHSMRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Conn(ctx)

	hsm, err := FindHSMByTwoPartKey(ctx, conn, d.Id(), d.Get("hsm_eni_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudHSMv2 HSM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudHSMv2 HSM (%s): %s", d.Id(), err)
	}

	// When matched by ENI ID, the ID should updated.
	if aws.StringValue(hsm.HsmId) != d.Id() {
		d.SetId(aws.StringValue(hsm.HsmId))
	}

	d.Set("availability_zone", hsm.AvailabilityZone)
	d.Set("cluster_id", hsm.ClusterId)
	d.Set("hsm_eni_id", hsm.EniId)
	d.Set("hsm_id", hsm.HsmId)
	d.Set("hsm_state", hsm.State)
	d.Set("ip_address", hsm.EniIp)
	d.Set("subnet_id", hsm.SubnetId)

	return diags
}

func resourceHSMDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Conn(ctx)

	log.Printf("[INFO] Deleting CloudHSMv2 HSM: %s", d.Id())
	_, err := conn.DeleteHsmWithContext(ctx, &cloudhsmv2.DeleteHsmInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
		HsmId:     aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudhsmv2.ErrCodeCloudHsmResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudHSMv2 HSM (%s): %s", d.Id(), err)
	}

	if _, err := waitHSMDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudHSMv2 HSM (%s) delete: %s", d.Id(), err)
	}

	return diags
}
