package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
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
						"target_backup_vault": {
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
						},
						"completion_window": {
							Type:     schema.TypeInt,
							Optional: true,
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
		input.BackupPlanTags = v.(map[string]interface{})
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

	rule := &schema.Set{R: resourceAwsPlanRuleHash}

	for _, r := range resp.BackupPlan.Rules {
		m := make(map[string]interface{})

		if r.CompletionWindowMinutes != nil {
			m["completion_window"] = *r.CompletionWindowMinutes
		}
		if r.Lifecycle.DeleteAfterDays != nil {
			m["lifecycle"]["delete_after"] = *r.Lifecycle.DeleteAfterDays
		}
		if r.Lifecycle.MoveToColdStorageAfterDays != nil {
			m["lifecycle"]["cold_storage_after"] = *r.Lifecycle.MoveToColdStorageAfterDays
		}
		if r.RecoveryPointTags != nil {
			m["recovery_point_tags"] = *r.RecoveryPointTags
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
		BackupPlan: plan,
	}

	resp, err := conn.UpdateBackupPlan(input)
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

	resp, err := conn.DeleteBackupPlan(input)
	if err != nil {
		return err
	}

	return nil
}

func gatherPlanRules(d *schema.ResourceData) []*backup.RuleInput {
	rules := []*backup.RuleInput
	planRules := d.Get("rule").(*schema.Set).List()
	
	for _, i := range planRules {
		item := i.(map[string]interface{})
		rule := &backup.RuleInput{}
	
		if item["rule_name"] {
			rule.RuleName = item["rule_name"]
		}
		if item["target_vault_name"] {
			rule.TargetBackupVaultName = item["target_vault_name"]
		}
		if item["schedule"] {
			rule.ScheduleExpression = item["schedule"]
		}
		if item["start_window"] {
			rule.StartWindowMinutes = item["start_window"]
		}
		if item["completion_window"] {
			rule.CompletionWindowMinutes = item["completion_window"]
		}
		if item["lifecycle"]["delete_after"] {
			rule.Lifecycle.DeleteAfterDays = item["lifecycle"]["delete_after"]
		}
		if item["lifecycle"]["cold_storage_after"] {
			rule.Lifecycle.MoveToColdStorageAfterDays = item["lifecycle"]["cold_storage_after"]
		}
		if item["recovery_point_tags"] {
			rule.RecoveryPointTags = item["recovery_point_tags"]
		}
	
		rules = append(rules, rule)
	}

	return rules
}

func resourceAwsPlanRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m, castOk := v.(map[string]interface{})
	if !castOk {
		return 0
	}

	if v, ok := m["completion_window"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}

	if v, ok := m["lifecycle"]["delete_after"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}

	if v, ok := m["lifecycle"]["cold_storage_after"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}

	if v, ok := m["recovery_point_tags"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(map[string]interface{})))
	}

	if v, ok := m["rule_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["schedule"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["start_window"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}

	if v, ok := m["target_vault_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}