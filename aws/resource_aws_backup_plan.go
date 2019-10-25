package aws

import (
	"bytes"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsBackupPlan() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBackupPlanCreate,
		Read:   resourceAwsBackupPlanRead,
		Update: resourceAwsBackupPlanUpdate,
		Delete: resourceAwsBackupPlanDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rule": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target_vault_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"schedule": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"start_window": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  60,
						},
						"completion_window": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  180,
						},
						"lifecycle": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cold_storage_after": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"delete_after": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"recovery_point_tags": tagsSchema(),
					},
				},
				Set: backupBackuPlanHash,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsBackupPlanCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.CreateBackupPlanInput{
		BackupPlan: &backup.PlanInput{
			BackupPlanName: aws.String(d.Get("name").(string)),
			Rules:          expandBackupPlanRules(d.Get("rule").(*schema.Set)),
		},
	}
	if v, ok := d.GetOk("tags"); ok {
		input.BackupPlanTags = tagsFromMapGeneric(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating Backup Plan: %#v", input)
	resp, err := conn.CreateBackupPlan(input)
	if err != nil {
		return fmt.Errorf("error creating Backup Plan: %s", err)
	}

	d.SetId(aws.StringValue(resp.BackupPlanId))

	return resourceAwsBackupPlanRead(d, meta)
}

func resourceAwsBackupPlanRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	resp, err := conn.GetBackupPlan(&backup.GetBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
	})
	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Backup Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading Backup Plan (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.BackupPlan.BackupPlanName)
	if err := d.Set("rule", flattenBackupPlanRules(resp.BackupPlan.Rules)); err != nil {
		return fmt.Errorf("error setting rule: %s", err)
	}

	tagsOutput, err := conn.ListTags(&backup.ListTagsInput{
		ResourceArn: resp.BackupPlanArn,
	})
	if err != nil {
		return fmt.Errorf("error listing tags AWS Backup plan %s: %s", d.Id(), err)
	}

	if err := d.Set("tags", tagsToMapGeneric(tagsOutput.Tags)); err != nil {
		return fmt.Errorf("error setting tags on AWS Backup plan %s: %s", d.Id(), err)
	}

	d.Set("arn", resp.BackupPlanArn)
	d.Set("version", resp.VersionId)

	return nil
}

func resourceAwsBackupPlanUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.UpdateBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
		BackupPlan: &backup.PlanInput{
			BackupPlanName: aws.String(d.Get("name").(string)),
			Rules:          expandBackupPlanRules(d.Get("rule").(*schema.Set)),
		},
	}

	log.Printf("[DEBUG] Updating Backup Plan: %#v", input)
	_, err := conn.UpdateBackupPlan(input)
	if err != nil {
		return fmt.Errorf("error updating Backup Plan (%s): %s", d.Id(), err)
	}

	if d.HasChange("tags") {
		resourceArn := d.Get("arn").(string)
		oraw, nraw := d.GetChange("tags")
		create, remove := diffTagsGeneric(oraw.(map[string]interface{}), nraw.(map[string]interface{}))

		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			keys := make([]*string, 0, len(remove))
			for k := range remove {
				keys = append(keys, aws.String(k))
			}

			_, err := conn.UntagResource(&backup.UntagResourceInput{
				ResourceArn: aws.String(resourceArn),
				TagKeyList:  keys,
			})
			if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
				log.Printf("[WARN] Backup Plan %s not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			if err != nil {
				return fmt.Errorf("Error removing tags for (%s): %s", d.Id(), err)
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&backup.TagResourceInput{
				ResourceArn: aws.String(resourceArn),
				Tags:        create,
			})
			if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
				log.Printf("[WARN] Backup Plan %s not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			if err != nil {
				return fmt.Errorf("Error setting tags for (%s): %s", d.Id(), err)
			}
		}
	}

	return resourceAwsBackupPlanRead(d, meta)
}

func resourceAwsBackupPlanDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	log.Printf("[DEBUG] Deleting Backup Plan: %s", d.Id())
	_, err := conn.DeleteBackupPlan(&backup.DeleteBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
	})
	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Backup Plan (%s): %s", d.Id(), err)
	}

	return nil
}

func expandBackupPlanRules(vRules *schema.Set) []*backup.RuleInput {
	rules := []*backup.RuleInput{}

	for _, vRule := range vRules.List() {
		rule := &backup.RuleInput{}

		mRule := vRule.(map[string]interface{})

		if vRuleName, ok := mRule["rule_name"].(string); ok && vRuleName != "" {
			rule.RuleName = aws.String(vRuleName)
		}
		if vTargetVaultName, ok := mRule["target_vault_name"].(string); ok && vTargetVaultName != "" {
			rule.TargetBackupVaultName = aws.String(vTargetVaultName)
		}
		if vSchedule, ok := mRule["schedule"].(string); ok && vSchedule != "" {
			rule.ScheduleExpression = aws.String(vSchedule)
		}
		if vStartWindow, ok := mRule["start_window"].(int); ok {
			rule.StartWindowMinutes = aws.Int64(int64(vStartWindow))
		}
		if vCompletionWindow, ok := mRule["completion_window"].(int); ok {
			rule.CompletionWindowMinutes = aws.Int64(int64(vCompletionWindow))
		}

		if vRecoveryPointTags, ok := mRule["recovery_point_tags"].(map[string]interface{}); ok && len(vRecoveryPointTags) > 0 {
			rule.RecoveryPointTags = tagsFromMapGeneric(vRecoveryPointTags)
		}

		if vLifecycle, ok := mRule["lifecycle"].([]interface{}); ok && len(vLifecycle) > 0 && vLifecycle[0] != nil {
			lifecycle := &backup.Lifecycle{}

			mLifecycle := vLifecycle[0].(map[string]interface{})

			if vDeleteAfter, ok := mLifecycle["delete_after"].(int); ok && vDeleteAfter > 0 {
				lifecycle.DeleteAfterDays = aws.Int64(int64(vDeleteAfter))
			}
			if vColdStorageAfter, ok := mLifecycle["cold_storage_after"].(int); ok && vColdStorageAfter > 0 {
				lifecycle.MoveToColdStorageAfterDays = aws.Int64(int64(vColdStorageAfter))
			}

			rule.Lifecycle = lifecycle
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenBackupPlanRules(rules []*backup.Rule) *schema.Set {
	vRules := []interface{}{}

	for _, rule := range rules {
		mRule := map[string]interface{}{
			"rule_name":           aws.StringValue(rule.RuleName),
			"target_vault_name":   aws.StringValue(rule.TargetBackupVaultName),
			"schedule":            aws.StringValue(rule.ScheduleExpression),
			"start_window":        int(aws.Int64Value(rule.StartWindowMinutes)),
			"completion_window":   int(aws.Int64Value(rule.CompletionWindowMinutes)),
			"recovery_point_tags": tagsToMapGeneric(rule.RecoveryPointTags),
		}

		if lifecycle := rule.Lifecycle; lifecycle != nil {
			mRule["lifecycle"] = []interface{}{
				map[string]interface{}{
					"delete_after":       int(aws.Int64Value(lifecycle.DeleteAfterDays)),
					"cold_storage_after": int(aws.Int64Value(lifecycle.MoveToColdStorageAfterDays)),
				},
			}
		}

		vRules = append(vRules, mRule)
	}

	return schema.NewSet(backupBackuPlanHash, vRules)
}

func backupBackuPlanHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v.(map[string]interface{})["lifecycle"] != nil {
		lcRaw := v.(map[string]interface{})["lifecycle"].([]interface{})
		if len(lcRaw) == 1 {
			l := lcRaw[0].(map[string]interface{})
			if w, ok := l["delete_after"]; ok {
				buf.WriteString(fmt.Sprintf("%v-", w))
			}

			if w, ok := l["cold_storage_after"]; ok {
				buf.WriteString(fmt.Sprintf("%v-", w))
			}
		}
	}

	if v, ok := m["completion_window"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(interface{})))
	}

	if v, ok := m["recovery_point_tags"]; ok {
		buf.WriteString(fmt.Sprintf("%v-", v))
	}

	if v, ok := m["rule_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["schedule"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["start_window"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(interface{})))
	}

	if v, ok := m["target_vault_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}
