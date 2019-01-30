package aws

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsBackupSelection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBackupSelectionCreate,
		Read:   resourceAwsBackupSelectionRead,
		Delete: resourceAwsBackupSelectionDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"iam_role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tag": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: resourceAwsConditionTagHash,
			},
			"resources": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsBackupSelectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	selection := &backup.Selection{}

	selection.SelectionName = aws.String(d.Get("name").(string))
	selection.IamRoleArn = aws.String(d.Get("iam_role").(string))
	selection.ListOfTags = gatherConditionTags(d)
	selection.Resources = expandStringList(d.Get("resources").([]interface{}))

	input := &backup.CreateBackupSelectionInput{
		BackupPlanId:    aws.String(d.Get("plan_id").(string)),
		BackupSelection: selection,
	}

	resp, err := conn.CreateBackupSelection(input)
	if err != nil {
		return err
	}

	d.SetId(*resp.SelectionId)

	return resourceAwsBackupSelectionRead(d, meta)
}

func resourceAwsBackupSelectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.GetBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("plan_id").(string)),
		SelectionId:  aws.String(d.Id()),
	}

	resp, err := conn.GetBackupSelection(input)
	if err != nil {
		return err
	}

	d.Set("plan_id", resp.BackupPlanId)

	s := make(map[string]interface{})

	s["name"] = *resp.BackupSelection.SelectionName
	s["iam_role"] = *resp.BackupSelection.IamRoleArn
	if resp.BackupSelection.ListOfTags != nil {
		tag := &schema.Set{F: resourceAwsConditionTagHash}

		for _, r := range resp.BackupSelection.ListOfTags {
			m := make(map[string]interface{})

			m["type"] = *r.ConditionType
			m["key"] = *r.ConditionKey
			m["value"] = *r.ConditionValue

			tag.Add(m)
		}

		s["tag"] = tag
	}
	if resp.BackupSelection.Resources != nil {
		s["resources"] = resp.BackupSelection.Resources
	}

	d.Set("selection", s)

	return nil
}

func resourceAwsBackupSelectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.DeleteBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("plan_id").(string)),
		SelectionId:  aws.String(d.Id()),
	}

	_, err := conn.DeleteBackupSelection(input)
	if err != nil {
		return err
	}

	return nil
}

func gatherConditionTags(d *schema.ResourceData) []*backup.Condition {
	conditions := []*backup.Condition{}
	selectionTags := d.Get("tag").(*schema.Set).List()

	for _, i := range selectionTags {
		item := i.(map[string]interface{})
		tag := &backup.Condition{}

		tag.ConditionType = aws.String(item["type"].(string))
		tag.ConditionKey = aws.String(item["key"].(string))
		tag.ConditionValue = aws.String(item["value"].(string))

		conditions = append(conditions, tag)
	}

	return conditions
}

func resourceAwsConditionTagHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["key"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["value"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}
