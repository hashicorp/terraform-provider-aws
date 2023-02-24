package redshift

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceClusterSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterSnapshotCreate,
		ReadWithoutTimeout:   resourceClusterSnapshotRead,
		UpdateWithoutTimeout: resourceClusterSnapshotUpdate,
		DeleteWithoutTimeout: resourceClusterSnapshotDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manual_snapshot_retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"owner_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	input := redshift.CreateClusterSnapshotInput{
		ClusterIdentifier:  aws.String(d.Get("cluster_identifier").(string)),
		SnapshotIdentifier: aws.String(d.Get("snapshot_identifier").(string)),
		Tags:               Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("manual_snapshot_retention_period"); ok {
		input.ManualSnapshotRetentionPeriod = aws.Int64(int64(v.(int)))
	}

	out, err := conn.CreateClusterSnapshotWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Cluster Snapshot : %s", err)
	}

	d.SetId(aws.StringValue(out.Snapshot.SnapshotIdentifier))

	if _, err := waitClusterSnapshotAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster Snapshot (%s) to be Available: %s", d.Id(), err)
	}

	return append(diags, resourceClusterSnapshotRead(ctx, d, meta)...)
}

func resourceClusterSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	out, err := FindClusterSnapshotById(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Cluster Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Cluster Snapshot (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("snapshot:%s/%s", aws.StringValue(out.ClusterIdentifier), d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("snapshot_identifier", out.SnapshotIdentifier)
	d.Set("cluster_identifier", out.ClusterIdentifier)
	d.Set("manual_snapshot_retention_period", out.ManualSnapshotRetentionPeriod)
	d.Set("kms_key_id", out.KmsKeyId)
	d.Set("owner_account", out.OwnerAccount)

	tags := KeyValueTags(ctx, rsc.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceClusterSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &redshift.ModifyClusterSnapshotInput{
			SnapshotIdentifier:            aws.String(d.Id()),
			ManualSnapshotRetentionPeriod: aws.Int64(int64(d.Get("manual_snapshot_retention_period").(int))),
		}

		_, err := conn.ModifyClusterSnapshotWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Cluster Snapshot (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Cluster Snapshot (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterSnapshotRead(ctx, d, meta)...)
}

func resourceClusterSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()

	log.Printf("[DEBUG] Deleting Redshift Cluster Snapshot: %s", d.Id())
	_, err := conn.DeleteClusterSnapshotWithContext(ctx, &redshift.DeleteClusterSnapshotInput{
		SnapshotIdentifier: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterSnapshotNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Cluster Snapshot (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterSnapshotDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster Snapshot (%s) to be Deleted: %s", d.Id(), err)
	}

	return diags
}
