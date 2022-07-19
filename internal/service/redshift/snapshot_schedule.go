package redshift

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSnapshotSchedule() *schema.Resource {
	return &schema.Resource{
		Create: resourceSnapshotScheduleCreate,
		Read:   resourceSnapshotScheduleRead,
		Update: resourceSnapshotScheduleUpdate,
		Delete: resourceSnapshotScheduleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
			},
			"identifier_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"definitions": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}

}

func resourceSnapshotScheduleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var identifier string
	if v, ok := d.GetOk("identifier"); ok {
		identifier = v.(string)
	} else {
		if v, ok := d.GetOk("identifier_prefix"); ok {
			identifier = resource.PrefixedUniqueId(v.(string))
		} else {
			identifier = resource.UniqueId()
		}
	}
	createOpts := &redshift.CreateSnapshotScheduleInput{
		ScheduleIdentifier:  aws.String(identifier),
		ScheduleDefinitions: flex.ExpandStringSet(d.Get("definitions").(*schema.Set)),
		Tags:                Tags(tags.IgnoreAWS()),
	}
	if attr, ok := d.GetOk("description"); ok {
		createOpts.ScheduleDescription = aws.String(attr.(string))
	}

	resp, err := conn.CreateSnapshotSchedule(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Redshift Snapshot Schedule: %s", err)
	}

	d.SetId(aws.StringValue(resp.ScheduleIdentifier))

	return resourceSnapshotScheduleRead(d, meta)
}

func resourceSnapshotScheduleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	descOpts := &redshift.DescribeSnapshotSchedulesInput{
		ScheduleIdentifier: aws.String(d.Id()),
	}

	resp, err := conn.DescribeSnapshotSchedules(descOpts)
	if err != nil {
		return fmt.Errorf("error describing Redshift Cluster Snapshot Schedule %s: %w", d.Id(), err)
	}

	if !d.IsNewResource() && (resp.SnapshotSchedules == nil || len(resp.SnapshotSchedules) != 1) {
		log.Printf("[WARN] Redshift Cluster Snapshot Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	snapshotSchedule := resp.SnapshotSchedules[0]

	d.Set("identifier", snapshotSchedule.ScheduleIdentifier)
	d.Set("description", snapshotSchedule.ScheduleDescription)
	if err := d.Set("definitions", flex.FlattenStringList(snapshotSchedule.ScheduleDefinitions)); err != nil {
		return fmt.Errorf("error setting definitions: %w", err)
	}

	tags := KeyValueTags(snapshotSchedule.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("snapshotschedule:%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceSnapshotScheduleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Snapshot Schedule (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	if d.HasChange("definitions") {
		modifyOpts := &redshift.ModifySnapshotScheduleInput{
			ScheduleIdentifier:  aws.String(d.Id()),
			ScheduleDefinitions: flex.ExpandStringSet(d.Get("definitions").(*schema.Set)),
		}
		_, err := conn.ModifySnapshotSchedule(modifyOpts)
		if err != nil {
			return fmt.Errorf("error modifying Redshift Snapshot Schedule %s: %w", d.Id(), err)
		}
	}

	return resourceSnapshotScheduleRead(d, meta)
}

func resourceSnapshotScheduleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	if d.Get("force_destroy").(bool) {
		if err := resourceSnapshotScheduleDeleteAllAssociatedClusters(conn, d.Id()); err != nil {
			return err
		}
	}

	_, err := conn.DeleteSnapshotSchedule(&redshift.DeleteSnapshotScheduleInput{
		ScheduleIdentifier: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting Redshift Snapshot Schedule %s: %s", d.Id(), err)
	}

	return nil
}

func resourceSnapshotScheduleDeleteAllAssociatedClusters(conn *redshift.Redshift, scheduleIdentifier string) error {

	resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
		ScheduleIdentifier: aws.String(scheduleIdentifier),
	})
	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing Redshift Cluster Snapshot Schedule %s: %s", scheduleIdentifier, err)
	}
	if resp.SnapshotSchedules == nil || len(resp.SnapshotSchedules) != 1 {
		log.Printf("[WARN] Unable to find Redshift Cluster Snapshot Schedule (%s)", scheduleIdentifier)
		return nil
	}

	snapshotSchedule := resp.SnapshotSchedules[0]

	for _, associatedCluster := range snapshotSchedule.AssociatedClusters {
		_, err = conn.ModifyClusterSnapshotSchedule(&redshift.ModifyClusterSnapshotScheduleInput{
			ClusterIdentifier:    associatedCluster.ClusterIdentifier,
			ScheduleIdentifier:   aws.String(scheduleIdentifier),
			DisassociateSchedule: aws.Bool(true),
		})

		clusterId := aws.StringValue(associatedCluster.ClusterIdentifier)

		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault) {
			log.Printf("[WARN] Redshift Snapshot Cluster (%s) not found, removing from state", clusterId)
			continue
		}
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotScheduleNotFoundFault) {
			log.Printf("[WARN] Redshift Snapshot Schedule (%s) not found, removing from state", scheduleIdentifier)
			continue
		}
		if err != nil {
			return fmt.Errorf("Error disassociate Redshift Cluster (%s) and Snapshot Schedule (%s) Association: %s", clusterId, scheduleIdentifier, err)
		}
	}

	for _, associatedCluster := range snapshotSchedule.AssociatedClusters {
		id := fmt.Sprintf("%s/%s", aws.StringValue(associatedCluster.ClusterIdentifier), scheduleIdentifier)
		if _, err := waitScheduleAssociationDeleted(conn, id); err != nil {
			return err
		}
	}

	return nil
}
