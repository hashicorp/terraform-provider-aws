package aws

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsBackupSelection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBackupSelectionCreate,
		Read:   resourceAwsBackupSelectionRead,
		Update: nil,
		Delete: resourceAwsBackupSelectionDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
		},

		Schema: map[string]*schema.Schema{
			"backup_plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"iam_role_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resources": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
			"tag_condition": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"test": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "STRINGEQUALS",
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"variable": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsBackupSelectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.CreateBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("backup_plan_id").(string)),
		BackupSelection: &backup.Selection{
			IamRoleArn:    aws.String(d.Get("iam_role_arn").(string)),
			SelectionName: aws.String(d.Get("name").(string)),
		},
	}

	if v, ok := d.GetOk("resources"); ok {
		input.BackupSelection.SetResources(expandStringSet(v.(*schema.Set)))
	}

	if v, ok := d.GetOk("tag_condition"); ok {
		input.BackupSelection.SetListOfTags(expandTagConditionSet(v.(*schema.Set)))
	}

	resp, err := retryOnAwsCode("InvalidParameterValueException", func() (interface{}, error) {
		return conn.CreateBackupSelection(input)
	})
	if err != nil {
		return err
	}

	output := resp.(*backup.CreateBackupSelectionOutput)

	d.SetId(*output.BackupPlanId + "/" + *output.SelectionId)

	return resourceAwsBackupSelectionRead(d, meta)
}

func resourceAwsBackupSelectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	ids := strings.Split(d.Id(), "/")
	backupPlanId, selectionId := ids[0], ids[1]

	input := &backup.GetBackupSelectionInput{
		BackupPlanId: aws.String(backupPlanId),
		SelectionId:  aws.String(selectionId),
	}

	resp, err := conn.GetBackupSelection(input)
	if err != nil {
		return err
	}

	d.Set("backup_plan_id", *resp.BackupPlanId)

	resources := make([]string, 0, len(resp.BackupSelection.Resources))
	for _, resource := range resp.BackupSelection.Resources {
		resources = append(resources, aws.StringValue(resource))
	}
	d.Set("resources", resources)

	tagConditions := make([]map[string]string, 0, len(resp.BackupSelection.ListOfTags))
	for _, condition := range resp.BackupSelection.ListOfTags {
		tagConditions = append(tagConditions, map[string]string{
			"test":     aws.StringValue(condition.ConditionType),
			"variable": aws.StringValue(condition.ConditionKey),
			"value":    aws.StringValue(condition.ConditionValue),
		})
	}
	d.Set("tag_condition", tagConditions)

	return nil
}

func resourceAwsBackupSelectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	ids := strings.Split(d.Id(), "/")
	backupPlanId, selectionId := ids[0], ids[1]

	input := &backup.DeleteBackupSelectionInput{
		BackupPlanId: aws.String(backupPlanId),
		SelectionId:  aws.String(selectionId),
	}

	_, err := conn.DeleteBackupSelection(input)
	if err != nil {
		return err
	}

	return nil
}

func expandTagConditionSet(v *schema.Set) []*backup.Condition {
	conditionList := v.List()
	conditions := make([]*backup.Condition, 0, len(conditionList))

	for _, c := range conditionList {
		sourceMap := c.(map[string]interface{})
		condition := &backup.Condition{
			ConditionType:  aws.String(sourceMap["test"].(string)),
			ConditionKey:   aws.String(sourceMap["variable"].(string)),
			ConditionValue: aws.String(sourceMap["value"].(string)),
		}
		conditions = append(conditions, condition)
	}

	return conditions
}
