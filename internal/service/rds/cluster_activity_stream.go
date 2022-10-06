package rds

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceClusterActivityStream() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterActivityStreamCreate,
		ReadContext:   resourceClusterActivityStreamRead,
		DeleteContext: resourceClusterActivityStreamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mode": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(rds.ActivityStreamMode_Values(), false),
			},
			"kinesis_stream_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_native_audit_fields_included": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func resourceClusterActivityStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn

	resourceArn := d.Get("resource_arn").(string)

	startActivityStreamInput := &rds.StartActivityStreamInput{
		ResourceArn:                     aws.String(resourceArn),
		ApplyImmediately:                aws.Bool(true),
		KmsKeyId:                        aws.String(d.Get("kms_key_id").(string)),
		Mode:                            aws.String(d.Get("mode").(string)),
		EngineNativeAuditFieldsIncluded: aws.Bool(d.Get("engine_native_audit_fields_included").(bool)),
	}

	log.Printf("[DEBUG] RDS Cluster start activity stream input: %s", startActivityStreamInput)

	resp, err := conn.StartActivityStream(startActivityStreamInput)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating RDS Cluster Activity Stream: %s", err))
	}

	log.Printf("[DEBUG]: RDS Cluster start activity stream response: %s", resp)

	d.SetId(resourceArn)

	err = waitActivityStreamStarted(ctx, conn, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceClusterActivityStreamRead(ctx, d, meta)
}

func resourceClusterActivityStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn

	log.Printf("[DEBUG] Finding DB Cluster (%s)", d.Id())
	resp, err := FindDBClusterWithActivityStream(conn, d.Id())

	if tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error describing RDS Cluster (%s): %s", d.Id(), err))
	}

	d.Set("resource_arn", resp.DBClusterArn)
	d.Set("kms_key_id", resp.ActivityStreamKmsKeyId)
	d.Set("kinesis_stream_name", resp.ActivityStreamKinesisStreamName)
	d.Set("mode", resp.ActivityStreamMode)

	return nil
}

func resourceClusterActivityStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn

	stopActivityStreamInput := &rds.StopActivityStreamInput{
		ApplyImmediately: aws.Bool(true),
		ResourceArn:      aws.String(d.Id()),
	}

	log.Printf("[DEBUG] RDS Cluster stop activity stream input: %s", stopActivityStreamInput)

	resp, err := conn.StopActivityStream(stopActivityStreamInput)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error stopping RDS Cluster Activity Stream: %w", err))
	}

	log.Printf("[DEBUG] RDS Cluster stop activity stream response: %s", resp)

	err = waitActivityStreamStopped(ctx, conn, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
