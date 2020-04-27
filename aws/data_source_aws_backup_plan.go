package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsBackupPlan() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsBackupPlanRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_vault_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"schedule": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"start_window": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"completion_window": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"lifecycle": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cold_storage_after": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"delete_after": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"copy_action": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"lifecycle": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cold_storage_after": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"delete_after": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
									"destination_vault_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"recovery_point_tags": tagsSchemaComputed(),
					},
				},
				Set: backupBackupPlanHash,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsBackupPlanRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn
	id := d.Get("plan_id").(string)

	resp, err := conn.GetBackupPlan(&backup.GetBackupPlanInput{
		BackupPlanId: aws.String(id),
	})
	if err != nil {
		return fmt.Errorf("Error getting Backup Plan: %v", err)
	}

	d.SetId(aws.StringValue(resp.BackupPlanId))
	d.Set("arn", resp.BackupPlanArn)
	d.Set("name", resp.BackupPlan.BackupPlanName)
	d.Set("version", resp.VersionId)

	if err := d.Set("rule", flattenBackupPlanRules(resp.BackupPlan.Rules)); err != nil {
		return fmt.Errorf("error setting rule: %s", err)
	}

	tags, err := keyvaluetags.BackupListTags(conn, aws.StringValue(resp.BackupPlanArn))
	if err != nil {
		return fmt.Errorf("error listing tags for Backup Plan (%s): %s", id, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
