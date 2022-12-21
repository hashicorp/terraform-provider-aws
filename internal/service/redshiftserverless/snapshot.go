package redshiftserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceSnapshotCreate,
		Read:   resourceSnapshotRead,
		Update: resourceSnapshotUpdate,
		Delete: resourceSnapshotDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"accounts_with_provisioned_restore_access": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"accounts_with_restore_access": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"admin_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"namespace_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"snapshot_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	input := redshiftserverless.CreateSnapshotInput{
		NamespaceName: aws.String(d.Get("namespace_name").(string)),
		SnapshotName:  aws.String(d.Get("snapshot_name").(string)),
	}

	if v, ok := d.GetOk("retention_period"); ok {
		input.RetentionPeriod = aws.Int64(int64(v.(int)))
	}

	out, err := conn.CreateSnapshot(&input)

	if err != nil {
		return fmt.Errorf("error creating Redshift Serverless Snapshot : %w", err)
	}

	d.SetId(aws.StringValue(out.Snapshot.SnapshotName))

	if _, err := waitSnapshotAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Redshift Serverless Snapshot (%s) to be Available: %w", d.Id(), err)
	}

	return resourceSnapshotRead(d, meta)
}

func resourceSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	out, err := FindSnapshotByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift Serverless Snapshot (%s): %w", d.Id(), err)
	}

	d.Set("arn", out.SnapshotArn)
	d.Set("snapshot_name", out.SnapshotName)
	d.Set("namespace_name", out.NamespaceName)
	d.Set("namespace_arn", out.NamespaceArn)
	d.Set("retention_period", out.SnapshotRetentionPeriod)
	d.Set("admin_username", out.AdminUsername)
	d.Set("kms_key_id", out.KmsKeyId)
	d.Set("owner_account", out.OwnerAccount)
	d.Set("accounts_with_provisioned_restore_access", flex.FlattenStringSet(out.AccountsWithRestoreAccess))
	d.Set("accounts_with_restore_access", flex.FlattenStringSet(out.AccountsWithRestoreAccess))

	return nil
}

func resourceSnapshotUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	input := &redshiftserverless.UpdateSnapshotInput{
		SnapshotName:    aws.String(d.Id()),
		RetentionPeriod: aws.Int64(int64(d.Get("retention_period").(int))),
	}

	_, err := conn.UpdateSnapshot(input)
	if err != nil {
		return fmt.Errorf("error updating Redshift Serverless Snapshot (%s): %w", d.Id(), err)
	}

	return resourceSnapshotRead(d, meta)
}

func resourceSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	log.Printf("[DEBUG] Deleting Redshift Serverless Snapshot: %s", d.Id())
	_, err := conn.DeleteSnapshot(&redshiftserverless.DeleteSnapshotInput{
		SnapshotName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Redshift Serverless Snapshot (%s): %w", d.Id(), err)
	}

	if _, err := waitSnapshotDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Redshift Serverless Snapshot (%s) to be Deleted: %w", d.Id(), err)
	}

	return nil
}
