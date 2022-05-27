package redshift

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSnapshotScheduleAssociation() *schema.Resource {

	return &schema.Resource{
		Create: resourceSnapshotScheduleAssociationCreate,
		Read:   resourceSnapshotScheduleAssociationRead,
		Delete: resourceSnapshotScheduleAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"schedule_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSnapshotScheduleAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	clusterIdentifier := d.Get("cluster_identifier").(string)
	scheduleIdentifier := d.Get("schedule_identifier").(string)

	_, err := conn.ModifyClusterSnapshotSchedule(&redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
		DisassociateSchedule: aws.Bool(false),
	})

	if err != nil {
		return fmt.Errorf("Error associating Redshift Cluster (%s) and Snapshot Schedule (%s): %s", clusterIdentifier, scheduleIdentifier, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", clusterIdentifier, scheduleIdentifier))

	if _, err := WaitScheduleAssociationActive(conn, d.Id()); err != nil {
		return err
	}

	return resourceSnapshotScheduleAssociationRead(d, meta)
}

func resourceSnapshotScheduleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	scheduleIdentifier, assoicatedCluster, err := FindScheduleAssociationById(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Schedule Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift Schedule Association (%s): %w", d.Id(), err)
	}
	d.Set("cluster_identifier", assoicatedCluster.ClusterIdentifier)
	d.Set("schedule_identifier", scheduleIdentifier)

	return nil
}

func resourceSnapshotScheduleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	clusterIdentifier, scheduleIdentifier, err := SnapshotScheduleAssociationParseID(d.Id())
	if err != nil {
		return fmt.Errorf("Error parse Redshift Cluster Snapshot Schedule Association ID %s: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Redshift Cluster Snapshot Schedule Association: %s", d.Id())
	_, err = conn.ModifyClusterSnapshotSchedule(&redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
		DisassociateSchedule: aws.Bool(true),
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error disassociate Redshift Cluster (%s) and Snapshot Schedule (%s) Association: %s", clusterIdentifier, scheduleIdentifier, err)
	}

	if _, err := waitScheduleAssociationDeleted(conn, d.Id()); err != nil {
		return err
	}

	return nil
}

func SnapshotScheduleAssociationParseID(id string) (clusterIdentifier, scheduleIdentifier string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = fmt.Errorf("aws_redshift_snapshot_schedule_association id must be of the form <ClusterIdentifier>/<ScheduleIdentifier>")
		return
	}

	clusterIdentifier = parts[0]
	scheduleIdentifier = parts[1]
	return
}
