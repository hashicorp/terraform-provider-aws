package redshift

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSnapshotCopyGrant() *schema.Resource {
	return &schema.Resource{
		Create: resourceSnapshotCopyGrantCreate,
		Read:   resourceSnapshotCopyGrantRead,
		Update: resourceSnapshotCopyGrantUpdate,
		Delete: resourceSnapshotCopyGrantDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_copy_grant_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSnapshotCopyGrantCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	grantName := d.Get("snapshot_copy_grant_name").(string)

	input := redshift.CreateSnapshotCopyGrantInput{
		SnapshotCopyGrantName: aws.String(grantName),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	input.Tags = Tags(tags.IgnoreAWS())

	log.Printf("[DEBUG]: Adding new Redshift SnapshotCopyGrant: %s", input)

	var out *redshift.CreateSnapshotCopyGrantOutput
	var err error

	out, err = conn.CreateSnapshotCopyGrant(&input)

	if err != nil {
		return fmt.Errorf("creating Redshift Snapshot Copy Grant (%s): %s", grantName, err)
	}

	log.Printf("[DEBUG] Created new Redshift SnapshotCopyGrant: %s", *out.SnapshotCopyGrant.SnapshotCopyGrantName)
	d.SetId(grantName)

	_, err = tfresource.RetryWhenNotFound(3*time.Minute, func() (any, error) {
		return findSnapshotCopyGrant(conn, grantName)
	})
	if err != nil {
		return fmt.Errorf("creating Redshift Snapshot Copy Grant (%s): waiting for completion: %s", grantName, err)
	}

	return resourceSnapshotCopyGrantRead(d, meta)
}

func resourceSnapshotCopyGrantRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	grantName := d.Id()

	grant, err := findSnapshotCopyGrant(conn, grantName)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Snapshot Copy Grant (%s) not found, removing from state", grantName)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading Redshift Snapshot Copy Grant (%s): %s", grantName, err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("snapshotcopygrant:%s", grantName),
	}.String()

	d.Set("arn", arn)

	d.Set("kms_key_id", grant.KmsKeyId)
	d.Set("snapshot_copy_grant_name", grant.SnapshotCopyGrantName)
	tags := KeyValueTags(grant.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceSnapshotCopyGrantUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("updating Redshift Snapshot Copy Grant (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceSnapshotCopyGrantRead(d, meta)
}

func resourceSnapshotCopyGrantDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn()

	grantName := d.Id()

	deleteInput := redshift.DeleteSnapshotCopyGrantInput{
		SnapshotCopyGrantName: aws.String(grantName),
	}

	log.Printf("[DEBUG] Deleting snapshot copy grant: %s", grantName)
	_, err := conn.DeleteSnapshotCopyGrant(&deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault) {
			return nil
		}
		return fmt.Errorf("deleting Redshift Snapshot Copy Grant (%s): %s", d.Id(), err)
	}

	if err := WaitForSnapshotCopyGrantToBeDeleted(conn, grantName); err != nil {
		return fmt.Errorf("deleting Redshift Snapshot Copy Grant (%s): waiting for completion: %s", d.Id(), err)
	}

	return nil
}

// Used by the tests as well
func WaitForSnapshotCopyGrantToBeDeleted(conn *redshift.Redshift, grantName string) error {
	_, err := tfresource.RetryUntilNotFound(3*time.Minute, func() (any, error) {
		return findSnapshotCopyGrant(conn, grantName)
	})
	return err
}

func findSnapshotCopyGrant(conn *redshift.Redshift, grantName string) (*redshift.SnapshotCopyGrant, error) {
	input := redshift.DescribeSnapshotCopyGrantsInput{
		SnapshotCopyGrantName: aws.String(grantName),
	}

	out, err := conn.DescribeSnapshotCopyGrants(&input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if out == nil || len(out.SnapshotCopyGrants) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}
	if l := len(out.SnapshotCopyGrants); l > 1 {
		return nil, tfresource.NewTooManyResultsError(1, nil)
	}

	return out.SnapshotCopyGrants[0], nil
}
