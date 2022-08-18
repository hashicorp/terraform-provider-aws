package drs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/drs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceReplicationConfigurationTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceReplicationConfigurationTemplateCreate,
		Read:   resourceReplicationConfigurationTemplateRead,
		Update: resourceReplicationConfigurationTemplateUpdate,
		Delete: resourceReplicationConfigurationTemplateDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associate_default_security_group": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"bandwidth_throttling": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"create_public_ip": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"data_plane_routing": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_large_staging_disk_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ebs_encryption": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ebs_encryption_key_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pit_policy": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"interval": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"retention_duration": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"rule_id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"units": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"replication_server_instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"replication_servers_security_groups_ids": {
				Type:     schema.TypeList,
				Required: true,
			},
			"staging_area_subnet_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"staging_area_tags": tftags.TagsSchema(),
			"tags":              tftags.TagsSchema(),
			"use_dedicated_replication_server": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceReplicationConfigurationTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DRSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	staging_tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("staging_area_tags").(map[string]interface{})))
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &drs.CreateReplicationConfigurationTemplateInput{
		AssociateDefaultSecurityGroup:       aws.Bool(d.Get("associate_default_security_group").(bool)),
		BandwidthThrottling:                 aws.Int64(d.Get("bandwidth_throttling").(int64)),
		CreatePublicIP:                      aws.Bool(d.Get("create_public_ip").(bool)),
		DataPlaneRouting:                    aws.String(d.Get("data_plane_routing").(string)),
		DefaultLargeStagingDiskType:         aws.String(d.Get("default_large_staging_disk_type").(string)),
		EbsEncryption:                       aws.String(d.Get("ebs_encryption").(string)),
		EbsEncryptionKeyArn:                 aws.String(d.Get("ebs_encryption_key_arn").(string)),
		PitPolicy:                           expandPitPolicy(d.Get("pit_policy").(*schema.Set).List()),
		ReplicationServerInstanceType:       aws.String(d.Get("replication_server_instance_type").(string)),
		ReplicationServersSecurityGroupsIDs: flex.ExpandStringList(d.Get("replication_servers_security_groups_ids").([]interface{})),
		StagingAreaSubnetId:                 aws.String(d.Get("staging_area_subnet_id").(string)),
		StagingAreaTags:                     Tags(tags.IgnoreAWS()),
		Tags:                                Tags(staging_tags.IgnoreAWS()),
		UseDedicatedReplicationServer:       aws.Bool(d.Get("use_dedicated_replication_server").(bool)),
	}

	log.Printf("[DEBUG] Creating DRS Replication Configuration Template: %s", input)
	output, err := conn.CreateReplicationConfigurationTemplate(input)

	if err != nil {
		return fmt.Errorf("error creating DRS Replication Configuration Template: %w", err)
	}

	d.SetId(aws.ToString(output.ReplicationConfigurationTemplateID))

	return resourceReplicationConfigurationTemplateRead(d, meta)
}

func resourceReplicationConfigurationTemplateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DRSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	replicationConfigurationTemplate, err := FindReplicationConfigurationTemplateByID(conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Replication Configuration Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Replication Configuration Template (%s): %w", d.Id(), err)
	}

	d.Set("arn", replicationConfigurationTemplate.Arn)
	d.Set("associate_default_security_group", replicationConfigurationTemplate.AssociateDefaultSecurityGroup)
	d.Set("bandwidth_throttling", replicationConfigurationTemplate.BandwidthThrottling)
	d.Set("create_public_ip", replicationConfigurationTemplate.CreatePublicIP)
	d.Set("data_plane_routing", replicationConfigurationTemplate.DataPlaneRouting)
	d.Set("default_large_staging_disk_type", replicationConfigurationTemplate.DefaultLargeStagingDiskType)
	d.Set("ebs_encryption", replicationConfigurationTemplate.EbsEncryption)
	d.Set("ebs_encryption_key_arn", replicationConfigurationTemplate.EbsEncryptionKeyArn)
	d.Set("pit_policy", replicationConfigurationTemplate.PitPolicy)
	d.Set("replication_server_instance_type", replicationConfigurationTemplate.ReplicationServerInstanceType)
	d.Set("replication_servers_security_groups_ids", replicationConfigurationTemplate.ReplicationServersSecurityGroupsIDs)
	d.Set("staging_area_subnet_id", replicationConfigurationTemplate.StagingAreaSubnetId)
	d.Set("use_dedicated_replication_server", replicationConfigurationTemplate.UseDedicatedReplicationServer)

	tags := KeyValueTags(replicationConfigurationTemplate.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("staging_area_tags", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceReplicationConfigurationTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DRSConn

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	staging_tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("staging_area_tags").(map[string]interface{})))

	input := &drs.UpdateReplicationConfigurationTemplateInput{
		ReplicationConfigurationTemplateID: aws.String(d.Id()),
	}

	if d.HasChange("associate_default_security_group") {
		input.AssociateDefaultSecurityGroup = aws.Bool(d.Get("associate_default_security_group").(bool))
	}

	if d.HasChange("bandwidth_throttling") {
		input.BandwidthThrottling = aws.Int64(d.Get("bandwidth_throttling").(int64))
	}

	if d.HasChange("create_public_ip") {
		input.CreatePublicIP = aws.Bool(d.Get("create_public_ip").(bool))
	}

	if d.HasChange("data_plane_routing") {
		input.DataPlaneRouting = aws.String(d.Get("data_plane_routing").(string))
	}

	if d.HasChange("default_large_staging_disk_type") {
		input.DefaultLargeStagingDiskType = aws.String(d.Get("default_large_staging_disk_type").(string))
	}

	if d.HasChange("ebs_encryption") {
		input.EbsEncryption = aws.String(d.Get("ebs_encryption").(string))
	}

	if d.HasChange("ebs_encryption_key_arn") {
		input.EbsEncryptionKeyArn = aws.String(d.Get("ebs_encryption_key_arn").(string))
	}

	if d.HasChange("pit_policy") {
		input.PitPolicy = expandPitPolicy(d.Get("pit_policy").(*schema.Set).List())
	}

	if d.HasChange("replication_server_instance_type") {
		input.ReplicationServerInstanceType = aws.String(d.Get("replication_server_instance_type").(string))
	}

	if d.HasChange("replication_servers_security_groups_ids") {
		input.ReplicationServersSecurityGroupsIDs = flex.ExpandStringList(d.Get("replication_servers_security_groups_ids").([]interface{}))
	}

	if d.HasChange("staging_area_subnet_id") {
		input.StagingAreaSubnetId = aws.String(d.Get("staging_area_subnet_id").(string))
	}

	if d.HasChange("staging_area_tags") {
		input.StagingAreaTags = Tags(staging_tags.IgnoreAWS())
	}

	if d.HasChange("use_dedicated_replication_server") {
		input.UseDedicatedReplicationServer = aws.Bool(d.Get("use_dedicated_replication_server").(bool))
	}

	_, err := conn.UpdateReplicationConfigurationTemplate(input)

	if err != nil {
		return fmt.Errorf("error updating replication configuration template (%s): : %w", d.Id(), err)
	}

	return resourceReplicationConfigurationTemplateRead(d, meta)

}

func resourceReplicationConfigurationTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DRSConn

	log.Printf("[DEBUG] Deleting Replication Configuration Template: %s", d.Id())
	_, err := conn.DeleteReplicationConfigurationTemplate(&drs.DeleteReplicationConfigurationTemplateInput{
		ReplicationConfigurationTemplateID: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, drs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Replication Configuration Tempalte (%s): %w", d.Id(), err)
	}

	return nil
}

func expandPitPolicy(p []interface{}) []*drs.PITPolicyRule {
	pitPolicyRules := make([]*drs.PITPolicyRule, len(p))
	for i, pitPolicy := range p {
		pitPol := pitPolicy.(map[string]interface{})
		pitPolicyRules[i] = &drs.PITPolicyRule{
			Enabled:           aws.Bool(pitPol["enabled"].(bool)),
			Interval:          aws.Int64(pitPol["interval"].(int64)),
			RetentionDuration: aws.Int64(pitPol["retention_duration"].(int64)),
			RuleID:            aws.Int64(pitPol["rule_id"].(int64)),
			Units:             aws.String(pitPol["units"].(string)),
		}
	}
	return pitPolicyRules
}

func Tags(tags tftags.KeyValueTags) map[string]*string {
	return aws.StringMap(tags.Map())
}

func KeyValueTags(tags map[string]*string) tftags.KeyValueTags {
	return tftags.New(tags)
}
