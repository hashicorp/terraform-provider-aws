package efs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
		Update: schema.Noop,
		Delete: resourceReplicationConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ExactlyOneOf: []string{"destination.0.availability_zone_name", "destination.0.region"},
						},
						"file_system_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kms_key_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"region": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidRegionName,
							ExactlyOneOf: []string{"destination.0.availability_zone_name", "destination.0.region"},
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
			"source_file_system_arn": {
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
		Destinations:       expandEfsReplicationConfigurationDestinations(d.Get("destination").([]interface{})),
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

	if output == nil || len(output.Replications) == 0 || output.Replications[0] == nil {
		return fmt.Errorf("error updating state of EFS replication configuration (%s)", d.Id())
	}

	replication := output.Replications[0]

	if replication == nil || len(replication.Destinations) == 0 || replication.Destinations[0] == nil {
		return fmt.Errorf("error updating state of EFS replication configuration (%s)", d.Id())
	}

	destination := flattenEfsReplicationConfigurationDestination(replication.Destinations[0])

	dest := make(map[string]interface{})
	if v, ok := d.GetOk("destinations"); ok {
		val := v.([]interface{})
		if len(val) > 0 {
			dest = val[0].(map[string]interface{})
		}
	}

	if v, ok := dest["availability_zone_name"]; ok {
		destination["availability_zone_name"] = v
	}

	if v, ok := dest["kms_key_id"]; ok {
		destination["kms_key_id"] = v
	}

	/*
		// Create new connection for the region of the destination file system
		session, sessionErr := conns.NewSessionForRegion(&conn.Config, *destination["region"].(*string), meta.(*conns.AWSClient).TerraformVersion)

		if sessionErr != nil {
			return fmt.Errorf("error creating AWS session: %w", sessionErr)
		}

		altConn := efs.New(session)
		destinationFsConfiguration, err := FindFileSystemByID(altConn, *destination["file_system_id"].(*string))


		if v := destinationFsConfiguration.AvailabilityZoneName; v != nil && len(v) > 0 {
			destination["availability_zone_name"] = v
		}

		if v := destinationFsConfiguration.KmsKeyId; v != nil && len(v) > 0 {
			// TODO logic to be able to handle the different formats that kms_key_id could be (arn, id, alias, alias arn)
			destination["kms_key_id"] = v
		}
	*/

	if err := d.Set("destination", []interface{}{destination}); err != nil {
		return fmt.Errorf("error setting destination: %w", err)
	}

	d.Set("creation_time", aws.TimeValue(replication.CreationTime).String())
	d.Set("original_source_file_system_arn", replication.OriginalSourceFileSystemArn)
	d.Set("source_file_system_arn", replication.SourceFileSystemArn)
	d.Set("source_file_system_id", replication.SourceFileSystemId)
	d.Set("source_file_system_region", replication.SourceFileSystemRegion)

	return nil
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

	m := l[0].(map[string]interface{})

	if v, ok := m["availability_zone_name"]; ok && v != "" {
		destination.AvailabilityZoneName = aws.String(v.(string))
	}

	if v, ok := m["kms_key_id"]; ok && v != "" {
		destination.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := m["region"]; ok && v != "" {
		destination.Region = aws.String(v.(string))
	}

	return []*efs.DestinationToCreate{destination}
}

func flattenEfsReplicationConfigurationDestination(destination *efs.Destination) map[string]interface{} {
	m := map[string]interface{}{}

	if destination.FileSystemId != nil {
		m["file_system_id"] = destination.FileSystemId
	}

	if destination.Region != nil {
		m["region"] = destination.Region
	}

	if destination.Status != nil {
		m["status"] = destination.Status
	}

	return m
}
