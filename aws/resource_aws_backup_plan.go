package aws

import (
	"bytes"
	"fmt"

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
							Type:     schema.TypeMap,
							Optional: true,
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
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsBackupPlanCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	plan := &backup.PlanInput{
		BackupPlanName: aws.String(d.Get("name").(string)),
	}

	rules := gatherPlanRules(d)

	plan.Rules = rules

	input := &backup.CreateBackupPlanInput{
		BackupPlan: plan,
	}

	if v, ok := d.GetOk("tags"); ok {
		input.BackupPlanTags = v.(map[string]*string)
	}

	resp, err := conn.CreateBackupPlan(input)
	if err != nil {
		return err
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
	if err != nil {
		return err
	}

	rule := &schema.Set{F: resourceAwsPlanRuleHash}

	for _, r := range resp.BackupPlan.Rules {
		m := make(map[string]interface{})

		if r.CompletionWindowMinutes != nil {
			m["completion_window"] = *r.CompletionWindowMinutes
		}
		if r.Lifecycle != nil {
			l := map[string]int64{}
			if r.Lifecycle.DeleteAfterDays != nil {
				l["delete_after"] = *r.Lifecycle.DeleteAfterDays
			}
			if r.Lifecycle.MoveToColdStorageAfterDays != nil {
				l["cold_storage_after"] = *r.Lifecycle.MoveToColdStorageAfterDays
			}
			m["lifecycle"] = l
		}
		if r.RecoveryPointTags != nil {
			m["recovery_point_tags"] = r.RecoveryPointTags
		}
		if r.RuleName != nil {
			m["rule_name"] = *r.RuleName
		}
		if r.ScheduleExpression != nil {
			m["schedule"] = *r.ScheduleExpression
		}
		if r.StartWindowMinutes != nil {
			m["start_window"] = *r.StartWindowMinutes
		}
		if r.TargetBackupVaultName != nil {
			m["target_vault_name"] = *r.TargetBackupVaultName
		}

		rule.Add(m)
	}
	d.Set("rule", rule)

	d.Set("arn", resp.BackupPlanArn)
	d.Set("version", resp.VersionId)

	return nil
}

func resourceAwsBackupPlanUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	plan := &backup.PlanInput{
		BackupPlanName: aws.String(d.Get("name").(string)),
	}

	rules := gatherPlanRules(d)

	plan.Rules = rules

	input := &backup.UpdateBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
		BackupPlan:   plan,
	}

	_, err := conn.UpdateBackupPlan(input)
	if err != nil {
		return err
	}

	return resourceAwsBackupPlanRead(d, meta)
}

func resourceAwsBackupPlanDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.DeleteBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
	}

	_, err := conn.DeleteBackupPlan(input)
	if err != nil {
		return err
	}

	return nil
}

func gatherPlanRules(d *schema.ResourceData) []*backup.RuleInput {
	rules := []*backup.RuleInput{}
	planRules := d.Get("rule").(*schema.Set).List()

	for _, i := range planRules {
		item := i.(map[string]interface{})
		lifecycle := i.(map[string]interface{})["lifecycle"].(map[string]interface{})
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
		if lifecycle["delete_after"] != nil {
			rule.Lifecycle.DeleteAfterDays = aws.Int64(int64(lifecycle["delete_after"].(int)))
		}
		if lifecycle["cold_storage_after"] != nil {
			rule.Lifecycle.MoveToColdStorageAfterDays = aws.Int64(int64(lifecycle["cold_storage_after"].(int)))
		}
		if item["recovery_point_tags"] != nil {
			tagsUnwrapped := make(map[string]*string)
			for key, value := range item["recovery_point_tags"].(map[string]interface{}) {
				tagsUnwrapped[key] = aws.String(value.(string))
			}
			rule.RecoveryPointTags = tagsUnwrapped
		}

		rules = append(rules, rule)
	}

	return rules
}

func resourceAwsPlanRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v.(map[string]interface{})["lifecycle"] != nil {
		l := v.(map[string]interface{})["lifecycle"].(map[string]interface{})
		if v, ok := l["delete_after"]; ok {
			buf.WriteString(fmt.Sprintf("%d-", v.(int)))
		}

		if v, ok := l["cold_storage_after"]; ok {
			buf.WriteString(fmt.Sprintf("%d-", v.(int)))
		}
	}

	if v, ok := m["completion_window"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(interface{})))
	}

	if v, ok := m["recovery_point_tags"]; ok {
		switch t := v.(type) {
		case map[string]*string:
			buf.WriteString(fmt.Sprintf("%v-", v.(map[string]*string)))
		case map[string]interface{}:
			tagsUnwrapped := make(map[string]*string)
			for key, value := range m["recovery_point_tags"].(map[string]interface{}) {
				tagsUnwrapped[key] = aws.String(value.(string))
			}
			buf.WriteString(fmt.Sprintf("%v-", tagsUnwrapped))
		default:
			fmt.Println("invalid type: ", t)
		}
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
