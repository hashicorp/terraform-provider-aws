package efs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	//"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceReplicationConfigurationCreate,
		Read:   resourceReplicationConfigurationRead,
		Update: resourceReplicationConfigurationUpdate,
		Delete: resourceReplicationConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destinations": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						//TODO looks like you must specify either AZ or region, can we validate for that?
						"availability_zone_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"file_system_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kms_key_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						// TODO do we really want this one? Causes TF to grump
						// that it changed outside of Terraform and I'm not sure
						// why we'd ever need it in the state
						"last_replicated_timestamp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Computed:     true,
							ValidateFunc: verify.ValidRegionName,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"original_source_file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_file_system_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceReplicationConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	fsId := d.Get("source_file_system_id").(string)

	input := &efs.CreateReplicationConfigurationInput{
		Destinations:       expandEfsReplicationConfigurationDestinations(d.Get("destinations").([]interface{})),
		SourceFileSystemId: aws.String(fsId),
	}

	_, err := conn.CreateReplicationConfiguration(input)

	if err != nil {
		return fmt.Errorf("error creating EFS Replication Configuration for File System (%s): %w", fsId, err)
	}

	d.SetId(fsId)

	if _, err := waitReplicationConfigurationEnabled(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EFS replication configuration (%s) to be enabled: %w", d.Id(), err)
	}

	return resourceReplicationConfigurationRead(d, meta)
}

func resourceReplicationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	output, err := FindReplicationConfigurationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS Replication Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EFS Replication Configuration (%s): %w", d.Id(), err)
	}

	replication := output.Replications[0] //TODO error checking and such

	if err := d.Set("destinations", flattenEfsReplicationConfigurationDestinations(replication.Destinations)); err != nil {
		return fmt.Errorf("error setting destinations: %w", err)
	}

	d.Set("creation_time", aws.TimeValue(replication.CreationTime).String())
	d.Set("original_source_file_system_arn", replication.OriginalSourceFileSystemArn)
	d.Set("source_file_system_region", replication.SourceFileSystemRegion)
	d.Set("source_file_system_id", replication.SourceFileSystemId)

	return nil
}

func resourceReplicationConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	// I don't think you can update a replication configuration... TODO

	return resourceReplicationConfigurationRead(d, meta)
}

func resourceReplicationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	input := &efs.DeleteReplicationConfigurationInput{
		SourceFileSystemId: aws.String(d.Id()),
	}

	// deletion of the replication configuration must be done from the
	// region in which the destination file system is located.
	destination := expandEfsReplicationConfigurationDestinations(d.Get("destinations").([]interface{}))[0]
	session, sessionErr := conns.NewSessionForRegion(&conn.Config, *destination.Region, meta.(*conns.AWSClient).TerraformVersion)

	if sessionErr != nil {
		return fmt.Errorf("error creating AWS session: %w", sessionErr)
	}

	deleteConn := efs.New(session)

	_, err := deleteConn.DeleteReplicationConfiguration(input)

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound, efs.ErrCodeReplicationNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EFS replication configuration for (%s): %w", d.Id(), err)
	}

	if _, err := waitReplicationConfigurationDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EFS replication configuration (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func expandEfsReplicationConfigurationDestinations(l []interface{}) []*efs.DestinationToCreate {
	destination := &efs.DestinationToCreate{}

	// TODO this should be an error. Either AZ or Region must be specified.
	if len(l) == 0 || l[0] == nil {
		return []*efs.DestinationToCreate{destination} //TODO return error instead?
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["availability_zone_name"]; ok && v != "" {
		log.Printf("[DEBUG] Setting destination.AvailabilityZoneName to %s", aws.String(v.(string)))
		destination.AvailabilityZoneName = aws.String(v.(string))
	}

	if v, ok := m["kms_key_id"]; ok && v != "" {
		log.Printf("[DEBUG] Setting destination.KmsKeyId to %s", aws.String(v.(string)))
		destination.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := m["region"]; ok && v != "" {
		log.Printf("[DEBUG] Setting destination.Region to %s", aws.String(v.(string)))
		destination.Region = aws.String(v.(string))
	}

	return []*efs.DestinationToCreate{destination}
}

func flattenEfsReplicationConfigurationDestinations(destinations []*efs.Destination) []interface{} {
	if len(destinations) == 0 || destinations[0] == nil {
		return []interface{}{}
	}

	destination := destinations[0]
	m := map[string]interface{}{}

	if destination.FileSystemId != nil {
		m["file_system_id"] = destination.FileSystemId
	}

	if destination.LastReplicatedTimestamp != nil {
		m["last_replicated_timestamp"] = aws.TimeValue(destination.LastReplicatedTimestamp).String()
	}

	if destination.Region != nil {
		m["region"] = destination.Region
	}

	if destination.Status != nil {
		m["status"] = destination.Status
	}

	return []interface{}{m}
}
