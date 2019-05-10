package aws

import (
	"bytes"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsCodeBuildWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeBuildWebhookCreate,
		Read:   resourceAwsCodeBuildWebhookRead,
		Delete: resourceAwsCodeBuildWebhookDelete,
		Update: resourceAwsCodeBuildWebhookUpdate,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"branch_filter": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filter_group": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.WebhookFilterTypeEvent,
								codebuild.WebhookFilterTypeActorAccountId,
								codebuild.WebhookFilterTypeBaseRef,
								codebuild.WebhookFilterTypeFilePath,
								codebuild.WebhookFilterTypeHeadRef,
							}, false),
						},
						"exclude_matched_pattern": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"pattern": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: resourceAwsCodeBuildWebhookFilterHash,
			},
			"payload_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCodeBuildWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	webhookFilterGroups := expandWebhookFilterGroup(d)

	input := &codebuild.CreateWebhookInput{
		ProjectName:  aws.String(d.Get("project_name").(string)),
		FilterGroups: webhookFilterGroups,
	}

	// The CodeBuild API requires this to be non-empty if defined
	if v, ok := d.GetOk("branch_filter"); ok {
		input.BranchFilter = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating CodeBuild Webhook: %s", input)
	resp, err := conn.CreateWebhook(input)

	if err != nil {
		return fmt.Errorf("error creating CodeBuild Webhook: %s", err)
	}

	// Secret is only returned on create, so capture it at the start
	d.Set("secret", resp.Webhook.Secret)
	d.SetId(d.Get("project_name").(string))

	return resourceAwsCodeBuildWebhookRead(d, meta)
}

func expandWebhookFilterGroup(d *schema.ResourceData) [][]*codebuild.WebhookFilter {
	webhookFilters := make([]*codebuild.WebhookFilter, 0)

	configsList := d.Get("filter_group").(*schema.Set).List()

	if len(configsList) == 0 {
		return nil
	}

	for _, config := range configsList {
		filter := expandWebhookFilterData(config.(map[string]interface{}))
		webhookFilters = append(webhookFilters, &filter)
	}

	return [][]*codebuild.WebhookFilter{webhookFilters}
}

func expandWebhookFilterData(data map[string]interface{}) codebuild.WebhookFilter {
	filter := codebuild.WebhookFilter{
		Type:                  aws.String(data["type"].(string)),
		ExcludeMatchedPattern: aws.Bool(data["exclude_matched_pattern"].(bool)),
	}

	if v := data["pattern"]; v != nil {
		filter.Pattern = aws.String(v.(string))
	}

	return filter
}

func resourceAwsCodeBuildWebhookRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	resp, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
		Names: []*string{
			aws.String(d.Id()),
		},
	})

	if err != nil {
		return err
	}

	if len(resp.Projects) == 0 {
		log.Printf("[WARN] CodeBuild Project %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	project := resp.Projects[0]

	if project.Webhook == nil {
		log.Printf("[WARN] CodeBuild Project %q webhook not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("branch_filter", project.Webhook.BranchFilter)
	d.Set("filter_group", flattenAwsCodeBuildWebhookFilterGroups(project.Webhook.FilterGroups[0]))
	d.Set("payload_url", project.Webhook.PayloadUrl)
	d.Set("project_name", project.Name)
	d.Set("url", project.Webhook.Url)
	// The secret is never returned after creation, so don't set it here

	return nil
}

func resourceAwsCodeBuildWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	var err error
	filterGroups := expandWebhookFilterGroup(d)

	if len(filterGroups) >= 1 {
		_, err = conn.UpdateWebhook(&codebuild.UpdateWebhookInput{
			ProjectName:  aws.String(d.Id()),
			FilterGroups: filterGroups,
			RotateSecret: aws.Bool(false),
		})
	} else {
		_, err = conn.UpdateWebhook(&codebuild.UpdateWebhookInput{
			ProjectName:  aws.String(d.Id()),
			BranchFilter: aws.String(d.Get("branch_filter").(string)),
			RotateSecret: aws.Bool(false),
		})
	}

	if err != nil {
		return err
	}

	return resourceAwsCodeBuildWebhookRead(d, meta)
}

func resourceAwsCodeBuildWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	_, err := conn.DeleteWebhook(&codebuild.DeleteWebhookInput{
		ProjectName: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, codebuild.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}

func flattenAwsCodeBuildWebhookFilterGroups(filterList []*codebuild.WebhookFilter) *schema.Set {
	filterSet := schema.Set{
		F: resourceAwsCodeBuildWebhookFilterHash,
	}

	for _, filters := range filterList {
		filterSet.Add(flattenAwsCodeBuildWebhookFilterData(*filters))
	}
	return &filterSet
}

func resourceAwsCodeBuildWebhookFilterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m["type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["pattern"].(string)))
	buf.WriteString(fmt.Sprintf("%q", m["exclude_matched_pattern"]))

	return hashcode.String(buf.String())
}

func flattenAwsCodeBuildWebhookFilterData(filter codebuild.WebhookFilter) map[string]interface{} {
	values := map[string]interface{}{}

	values["type"] = *filter.Type
	values["pattern"] = *filter.Pattern
	values["exclude_matched_pattern"] = *filter.ExcludeMatchedPattern

	return values
}
