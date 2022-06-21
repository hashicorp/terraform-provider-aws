package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSnapshotCreateVolumePermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceSnapshotCreateVolumePermissionCreate,
		Read:   resourceSnapshotCreateVolumePermissionRead,
		Delete: resourceSnapshotCreateVolumePermissionDelete,

		CustomizeDiff: resourceSnapshotCreateVolumePermissionCustomizeDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSnapshotCreateVolumePermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	snapshotID := d.Get("snapshot_id").(string)
	accountID := d.Get("account_id").(string)
	id := EBSSnapshotCreateVolumePermissionCreateResourceID(snapshotID, accountID)
	input := &ec2.ModifySnapshotAttributeInput{
		Attribute: aws.String(ec2.SnapshotAttributeNameCreateVolumePermission),
		CreateVolumePermission: &ec2.CreateVolumePermissionModifications{
			Add: []*ec2.CreateVolumePermission{
				{UserId: aws.String(accountID)},
			},
		},
		SnapshotId: aws.String(snapshotID),
	}

	log.Printf("[DEBUG] Creating EBS Snapshot CreateVolumePermission: %s", input)
	_, err := conn.ModifySnapshotAttribute(input)

	if err != nil {
		return fmt.Errorf("creating EBS Snapshot CreateVolumePermission (%s): %w", id, err)
	}

	d.SetId(id)

	_, err = tfresource.RetryWhenNotFound(d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return FindCreateSnapshotCreateVolumePermissionByTwoPartKey(conn, snapshotID, accountID)
	})

	if err != nil {
		return fmt.Errorf("waiting for EBS Snapshot CreateVolumePermission create (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceSnapshotCreateVolumePermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	snapshotID, accountID, err := EBSSnapshotCreateVolumePermissionParseResourceID(d.Id())

	if err != nil {
		return err
	}

	_, err = FindCreateSnapshotCreateVolumePermissionByTwoPartKey(conn, snapshotID, accountID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS Snapshot CreateVolumePermission %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EBS Snapshot CreateVolumePermission (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceSnapshotCreateVolumePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	snapshotID, accountID, err := EBSSnapshotCreateVolumePermissionParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting EBS Snapshot CreateVolumePermission: %s", d.Id())
	_, err = conn.ModifySnapshotAttribute(&ec2.ModifySnapshotAttributeInput{
		Attribute: aws.String(ec2.SnapshotAttributeNameCreateVolumePermission),
		CreateVolumePermission: &ec2.CreateVolumePermissionModifications{
			Remove: []*ec2.CreateVolumePermission{
				{UserId: aws.String(accountID)},
			},
		},
		SnapshotId: aws.String(snapshotID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSnapshotNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EBS Snapshot CreateVolumePermission (%s): %w", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return FindCreateSnapshotCreateVolumePermissionByTwoPartKey(conn, snapshotID, accountID)
	})

	if err != nil {
		return fmt.Errorf("waiting for EBS Snapshot CreateVolumePermission delete (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceSnapshotCreateVolumePermissionCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if diff.Id() == "" {
		if snapshotID := diff.Get("snapshot_id").(string); snapshotID != "" {
			conn := meta.(*conns.AWSClient).EC2Conn

			snapshot, err := FindSnapshotByID(conn, snapshotID)

			if err != nil {
				return fmt.Errorf("reading EBS Snapshot (%s): %w", snapshotID, err)
			}

			if accountID := diff.Get("account_id").(string); aws.StringValue(snapshot.OwnerId) == accountID {
				return fmt.Errorf("AWS Account (%s) owns EBS Snapshot (%s)", accountID, snapshotID)
			}
		}
	}

	return nil
}

const ebsSnapshotCreateVolumePermissionIDSeparator = "-"

func EBSSnapshotCreateVolumePermissionCreateResourceID(snapshotID, accountID string) string {
	parts := []string{snapshotID, accountID}
	id := strings.Join(parts, ebsSnapshotCreateVolumePermissionIDSeparator)

	return id
}

func EBSSnapshotCreateVolumePermissionParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, ebsSnapshotCreateVolumePermissionIDSeparator, 3)

	if len(parts) != 3 || parts[0] != "snap" || parts[1] == "" || parts[2] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected SNAPSHOT_ID%[2]sACCOUNT_ID", id, ebsSnapshotCreateVolumePermissionIDSeparator)
	}

	return strings.Join([]string{parts[0], parts[1]}, ebsSnapshotCreateVolumePermissionIDSeparator), parts[2], nil
}
