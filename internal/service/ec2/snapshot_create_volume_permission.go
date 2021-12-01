package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceSnapshotCreateVolumePermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceSnapshotCreateVolumePermissionCreate,
		Read:   resourceSnapshotCreateVolumePermissionRead,
		Delete: resourceSnapshotCreateVolumePermissionDelete,

		Schema: map[string]*schema.Schema{
			"snapshot_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSnapshotCreateVolumePermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	snapshot_id := d.Get("snapshot_id").(string)
	account_id := d.Get("account_id").(string)

	_, err := conn.ModifySnapshotAttribute(&ec2.ModifySnapshotAttributeInput{
		SnapshotId: aws.String(snapshot_id),
		Attribute:  aws.String("createVolumePermission"),
		CreateVolumePermission: &ec2.CreateVolumePermissionModifications{
			Add: []*ec2.CreateVolumePermission{
				{UserId: aws.String(account_id)},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("Error adding snapshot createVolumePermission: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s", snapshot_id, account_id))

	// Wait for the account to appear in the permission list
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"denied"},
		Target:     []string{"granted"},
		Refresh:    resourceSnapshotCreateVolumePermissionStateRefreshFunc(conn, snapshot_id, account_id),
		Timeout:    20 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for snapshot createVolumePermission (%s) to be added: %s",
			d.Id(), err)
	}

	return nil
}

func resourceSnapshotCreateVolumePermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	snapshotID, accountID, err := SnapshotCreateVolumePermissionParseID(d.Id())
	if err != nil {
		return err
	}

	exists, err := HasCreateVolumePermission(conn, snapshotID, accountID)
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("[WARN] snapshot createVolumePermission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func resourceSnapshotCreateVolumePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	snapshotID, accountID, err := SnapshotCreateVolumePermissionParseID(d.Id())
	if err != nil {
		return err
	}

	_, err = conn.ModifySnapshotAttribute(&ec2.ModifySnapshotAttributeInput{
		SnapshotId: aws.String(snapshotID),
		Attribute:  aws.String("createVolumePermission"),
		CreateVolumePermission: &ec2.CreateVolumePermissionModifications{
			Remove: []*ec2.CreateVolumePermission{
				{UserId: aws.String(accountID)},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("Error removing snapshot createVolumePermission: %s", err)
	}

	// Wait for the account to disappear from the permission list
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"granted"},
		Target:     []string{"denied"},
		Refresh:    resourceSnapshotCreateVolumePermissionStateRefreshFunc(conn, snapshotID, accountID),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for snapshot createVolumePermission (%s) to be removed: %s",
			d.Id(), err)
	}

	return nil
}

func HasCreateVolumePermission(conn *ec2.EC2, snapshot_id string, account_id string) (bool, error) {
	_, state, err := resourceSnapshotCreateVolumePermissionStateRefreshFunc(conn, snapshot_id, account_id)()
	if err != nil {
		return false, err
	}
	if state == "granted" {
		return true, nil
	} else {
		return false, nil
	}
}

func resourceSnapshotCreateVolumePermissionStateRefreshFunc(conn *ec2.EC2, snapshot_id string, account_id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attrs, err := conn.DescribeSnapshotAttribute(&ec2.DescribeSnapshotAttributeInput{
			SnapshotId: aws.String(snapshot_id),
			Attribute:  aws.String("createVolumePermission"),
		})
		if err != nil {
			return nil, "", fmt.Errorf("Error refreshing snapshot createVolumePermission state: %s", err)
		}

		for _, vp := range attrs.CreateVolumePermissions {
			if aws.StringValue(vp.UserId) == account_id {
				return attrs, "granted", nil
			}
		}
		return attrs, "denied", nil
	}
}

func SnapshotCreateVolumePermissionParseID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "-", 3)
	if len(idParts) != 3 || idParts[0] != "snap" || idParts[1] == "" || idParts[2] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected SNAPSHOT_ID-ACCOUNT_ID", id)
	}
	return fmt.Sprintf("%s-%s", idParts[0], idParts[1]), idParts[2], nil
}
