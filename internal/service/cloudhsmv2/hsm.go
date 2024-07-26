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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudhsm_v2_hsm", name="HSM")
func resourceHSM() *schema.Resource {
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
			names.AttrAvailabilityZone: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrAvailabilityZone, names.AttrSubnetID},
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
			names.AttrIPAddress: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrSubnetID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrAvailabilityZone, names.AttrSubnetID},
			},
		},
	}
}

func resourceHSMCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Client(ctx)

	clusterID := d.Get("cluster_id").(string)
	input := &cloudhsmv2.CreateHsmInput{
		ClusterId: aws.String(clusterID),
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.AvailabilityZone = aws.String(v.(string))
	} else {
		cluster, err := findClusterByID(ctx, conn, clusterID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudHSMv2 Cluster (%s): %s", clusterID, err)
		}

		subnetID := d.Get(names.AttrSubnetID).(string)
		for az, sn := range cluster.SubnetMapping {
			if sn == subnetID {
				input.AvailabilityZone = aws.String(az)
			}
		}
	}

	if v, ok := d.GetOk(names.AttrIPAddress); ok {
		input.IpAddress = aws.String(v.(string))
	}

	output, err := conn.CreateHsm(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudHSMv2 HSM: %s", err)
	}

	d.SetId(aws.ToString(output.Hsm.HsmId))

	if _, err := waitHSMCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudHSMv2 HSM (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceHSMRead(ctx, d, meta)...)
}

func resourceHSMRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Client(ctx)

	hsm, err := findHSMByTwoPartKey(ctx, conn, d.Id(), d.Get("hsm_eni_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudHSMv2 HSM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudHSMv2 HSM (%s): %s", d.Id(), err)
	}

	// When matched by ENI ID, the ID should updated.
	if aws.ToString(hsm.HsmId) != d.Id() {
		d.SetId(aws.ToString(hsm.HsmId))
	}

	d.Set(names.AttrAvailabilityZone, hsm.AvailabilityZone)
	d.Set("cluster_id", hsm.ClusterId)
	d.Set("hsm_eni_id", hsm.EniId)
	d.Set("hsm_id", hsm.HsmId)
	d.Set("hsm_state", hsm.State)
	d.Set(names.AttrIPAddress, hsm.EniIp)
	d.Set(names.AttrSubnetID, hsm.SubnetId)

	return diags
}

func resourceHSMDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Client(ctx)

	log.Printf("[INFO] Deleting CloudHSMv2 HSM: %s", d.Id())
	_, err := conn.DeleteHsm(ctx, &cloudhsmv2.DeleteHsmInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
		HsmId:     aws.String(d.Id()),
	})

	if errs.IsA[*types.CloudHsmResourceNotFoundException](err) {
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

func findHSMByTwoPartKey(ctx context.Context, conn *cloudhsmv2.Client, hsmID, eniID string) (*types.Hsm, error) {
	input := &cloudhsmv2.DescribeClustersInput{}

	output, err := findClusters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		for _, v := range v.Hsms {
			v := v

			// CloudHSMv2 HSM instances can be recreated, but the ENI ID will
			// remain consistent. Without this ENI matching, HSM instances
			// instances can become orphaned.
			if aws.ToString(v.HsmId) == hsmID || aws.ToString(v.EniId) == eniID {
				return &v, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

func statusHSM(ctx context.Context, conn *cloudhsmv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findHSMByTwoPartKey(ctx, conn, id, "")

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), err
	}
}

func waitHSMCreated(ctx context.Context, conn *cloudhsmv2.Client, id string, timeout time.Duration) (*types.Hsm, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.HsmStateCreateInProgress),
		Target:     enum.Slice(types.HsmStateActive),
		Refresh:    statusHSM(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Hsm); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func waitHSMDeleted(ctx context.Context, conn *cloudhsmv2.Client, id string, timeout time.Duration) (*types.Hsm, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.HsmStateDeleteInProgress),
		Target:     []string{},
		Refresh:    statusHSM(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Hsm); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))

		return output, err
	}

	return nil, err
}
