package aws

import (
	"bytes"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
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
						"recovery_point_tags": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
				Set: resourceAwsPlanRuleHash,
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

	plan := &backup.PlanInput{
		BackupPlanName: aws.String(d.Get("name").(string)),
	}

	rules := expandBackupPlanRules(d.Get("rule").(*schema.Set).List())

	plan.Rules = rules

	input := &backup.CreateBackupPlanInput{
		BackupPlan: plan,
	}

	if v, ok := d.GetOk("tags"); ok {
		input.BackupPlanTags = v.(map[string]*string)
	}

	resp, err := conn.CreateBackupPlan(input)
	if err != nil {
		return fmt.Errorf("error creating Backup Plan: %s", err)
	}

	d.SetId(*resp.BackupPlanId)

	return resourceAwsBackupPlanRead(d, meta)
}

func resourceAwsBackupPlanRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.GetBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
	}

	resp, err := conn.GetBackupPlan(input)
	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Backup Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Backup Plan: %s", err)
	}

	rule := &schema.Set{F: resourceAwsPlanRuleHash}

	for _, r := range resp.BackupPlan.Rules {
		m := make(map[string]interface{})

		m["completion_window"] = aws.Int64Value(r.CompletionWindowMinutes)
		m["recovery_point_tags"] = aws.StringValueMap(r.RecoveryPointTags)
		m["rule_name"] = aws.StringValue(r.RuleName)
		m["schedule"] = aws.StringValue(r.ScheduleExpression)
		m["start_window"] = aws.Int64Value(r.StartWindowMinutes)
		m["target_vault_name"] = aws.StringValue(r.TargetBackupVaultName)

		if r.Lifecycle != nil {
			l := map[string]int64{}
			l["delete_after"] = aws.Int64Value(r.Lifecycle.DeleteAfterDays)
			l["cold_storage_after"] = aws.Int64Value(r.Lifecycle.MoveToColdStorageAfterDays)
			m["lifecycle"] = l
		}

		rule.Add(m)
	}
	if err := d.Set("rule", rule); err != nil {
		return fmt.Errorf("error setting rule: %s", err)
	}

	d.Set("arn", resp.BackupPlanArn)
	d.Set("version", resp.VersionId)

	return nil
}

func resourceAwsBackupPlanUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	plan := &backup.PlanInput{
		BackupPlanName: aws.String(d.Get("name").(string)),
	}

	rules := expandBackupPlanRules(d.Get("rule").(*schema.Set).List())

	plan.Rules = rules

	input := &backup.UpdateBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
		BackupPlan:   plan,
	}

	_, err := conn.UpdateBackupPlan(input)
	if err != nil {
		return fmt.Errorf("error updating Backup Plan: %s", err)
	}

	return resourceAwsBackupPlanRead(d, meta)
}

func resourceAwsBackupPlanDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.DeleteBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
	}

	_, err := conn.DeleteBackupPlan(input)
	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Backup Plan: %s", err)
	}

	return nil
}

func expandBackupPlanRules(l []interface{}) []*backup.RuleInput {
	rules := []*backup.RuleInput{}

	for _, i := range l {
		item := i.(map[string]interface{})
		rule := &backup.RuleInput{}

		if item["rule_name"] != "" {
			rule.RuleName = aws.String(item["rule_name"].(string))
		}
		if item["target_vault_name"] != "" {
			rule.TargetBackupVaultName = aws.String(item["target_vault_name"].(string))
		}
		if item["schedule"] != "" {
			rule.ScheduleExpression = aws.String(item["schedule"].(string))
		}
		if item["start_window"] != nil {
			rule.StartWindowMinutes = aws.Int64(int64(item["start_window"].(int)))
		}
		if item["completion_window"] != nil {
			rule.CompletionWindowMinutes = aws.Int64(int64(item["completion_window"].(int)))
		}

		if item["recovery_point_tags"] != nil {
			rule.RecoveryPointTags = tagsFromMapGeneric(item["recovery_point_tags"].(map[string]interface{}))
		}

		var lifecycle map[string]interface{}
		if i.(map[string]interface{})["lifecycle"] != nil {
			lifecycleRaw := i.(map[string]interface{})["lifecycle"].([]interface{})
			if len(lifecycleRaw) == 1 {
				lifecycle = lifecycleRaw[0].(map[string]interface{})
				if lifecycle["delete_after"] != nil {
					rule.Lifecycle.DeleteAfterDays = aws.Int64(int64(lifecycle["delete_after"].(int)))
				}
				if lifecycle["cold_storage_after"] != nil {
					rule.Lifecycle.MoveToColdStorageAfterDays = aws.Int64(int64(lifecycle["cold_storage_after"].(int)))
				}
			}

		}

		rules = append(rules, rule)
	}

	return rules
}

func resourceAwsPlanRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v.(map[string]interface{})["lifecycle"] != nil {
		lcRaw := v.(map[string]interface{})["lifecycle"].([]interface{})
		if len(lcRaw) == 1 {
			l := lcRaw[0].(map[string]interface{})
			if w, ok := l["delete_after"]; ok {
				buf.WriteString(fmt.Sprintf("%d-", w.(int)))
			}

			if w, ok := l["cold_storage_after"]; ok {
				buf.WriteString(fmt.Sprintf("%d-", w.(int)))
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
