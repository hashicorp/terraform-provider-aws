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
)

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
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"hsm_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"hsm_state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"hsm_eni_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHSMCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Conn()

	input := &cloudhsmv2.CreateHsmInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		input.AvailabilityZone = aws.String(v.(string))
	} else {
		cluster, err := FindCluster(ctx, conn, d.Get("cluster_id").(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudHSMv2 Cluster (%s): %s", d.Id(), err)
		}

		if cluster == nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudHSMv2 Cluster (%s): not found for subnet mappings", d.Id())
		}

		subnetId := d.Get("subnet_id").(string)
		for az, sn := range cluster.SubnetMapping {
			if aws.StringValue(sn) == subnetId {
				input.AvailabilityZone = aws.String(az)
			}
		}
	}

	if v, ok := d.GetOk("ip_address"); ok {
		input.IpAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] CloudHSMv2 HSM create %s", input)

	output, err := conn.CreateHsmWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudHSMv2 HSM: %s", err)
	}

	d.SetId(aws.StringValue(output.Hsm.HsmId))

	if _, err := waitHSMActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudHSMv2 HSM (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceHSMRead(ctx, d, meta)...)
}

func resourceHSMRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Conn()

	hsm, err := FindHSM(ctx, conn, d.Id(), d.Get("hsm_eni_id").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudHSMv2 HSM (%s): %s", d.Id(), err)
	}

	if hsm == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading CloudHSMv2 HSM (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] CloudHSMv2 HSM (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	// When matched by ENI ID, the ID should updated.
	if aws.StringValue(hsm.HsmId) != d.Id() {
		d.SetId(aws.StringValue(hsm.HsmId))
	}

	log.Printf("[INFO] Reading CloudHSMv2 HSM Information: %s", d.Id())

	d.Set("cluster_id", hsm.ClusterId)
	d.Set("subnet_id", hsm.SubnetId)
	d.Set("availability_zone", hsm.AvailabilityZone)
	d.Set("ip_address", hsm.EniIp)
	d.Set("hsm_id", hsm.HsmId)
	d.Set("hsm_state", hsm.State)
	d.Set("hsm_eni_id", hsm.EniId)

	return diags
}

func resourceHSMDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudHSMV2Conn()
	clusterId := d.Get("cluster_id").(string)

	log.Printf("[DEBUG] CloudHSMv2 HSM delete %s %s", clusterId, d.Id())
	input := &cloudhsmv2.DeleteHsmInput{
		ClusterId: aws.String(clusterId),
		HsmId:     aws.String(d.Id()),
	}

	_, err := conn.DeleteHsmWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudhsmv2.ErrCodeCloudHsmResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudHSMv2 HSM (%s): %s", d.Id(), err)
	}

	if _, err := waitHSMDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudHSMv2 HSM (%s) deletion: %s", d.Id(), err)
	}

	return diags
}
