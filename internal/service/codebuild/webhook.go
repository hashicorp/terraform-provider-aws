package codebuild

import (
	"bytes"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceWebhookCreate,
		Read:   resourceWebhookRead,
		Delete: resourceWebhookDelete,
		Update: resourceWebhookUpdate,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"build_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(codebuild.WebhookBuildType_Values(), false),
			},
			"branch_filter": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"filter_group"},
			},
			"filter_group": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(codebuild.WebhookFilterType_Values(), false),
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
						},
					},
				},
				Set:           resourceWebhookFilterHash,
				ConflictsWith: []string{"branch_filter"},
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

func resourceWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	input := &codebuild.CreateWebhookInput{
		ProjectName:  aws.String(d.Get("project_name").(string)),
		FilterGroups: expandWebhookFilterGroups(d),
	}

	if v, ok := d.GetOk("build_type"); ok {
		input.BuildType = aws.String(v.(string))
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

	return resourceWebhookRead(d, meta)
}

func expandWebhookFilterGroups(d *schema.ResourceData) [][]*codebuild.WebhookFilter {
	configs := d.Get("filter_group").(*schema.Set).List()

	webhookFilters := make([][]*codebuild.WebhookFilter, 0)

	if len(configs) == 0 {
		return nil
	}

	for _, config := range configs {
		filters := expandWebhookFilterData(config.(map[string]interface{}))
		webhookFilters = append(webhookFilters, filters)
	}

	return webhookFilters
}

func expandWebhookFilterData(data map[string]interface{}) []*codebuild.WebhookFilter {
	filters := make([]*codebuild.WebhookFilter, 0)

	filterConfigs := data["filter"].([]interface{})

	for i, filterConfig := range filterConfigs {
		filter := filterConfig.(map[string]interface{})
		filters = append(filters, &codebuild.WebhookFilter{
			Type:                  aws.String(filter["type"].(string)),
			ExcludeMatchedPattern: aws.Bool(filter["exclude_matched_pattern"].(bool)),
		})
		if v := filter["pattern"]; v != nil {
			filters[i].Pattern = aws.String(v.(string))
		}
	}

	return filters
}

func resourceWebhookRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	resp, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
		Names: []*string{
			aws.String(d.Id()),
		},
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codebuild.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CodeBuild, names.ErrActionReading, ResWebhook, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CodeBuild, names.ErrActionReading, ResWebhook, d.Id(), err)
	}

	if d.IsNewResource() && len(resp.Projects) == 0 {
		return names.Error(names.CodeBuild, names.ErrActionReading, ResWebhook, d.Id(), errors.New("no project found after create"))
	}

	if !d.IsNewResource() && len(resp.Projects) == 0 {
		names.LogNotFoundRemoveState(names.CodeBuild, names.ErrActionReading, ResWebhook, d.Id())
		d.SetId("")
		return nil
	}

	project := resp.Projects[0]

	if d.IsNewResource() && project.Webhook == nil {
		return names.Error(names.CodeBuild, names.ErrActionReading, ResWebhook, d.Id(), errors.New("no webhook after creation"))
	}

	if !d.IsNewResource() && project.Webhook == nil {
		names.LogNotFoundRemoveState(names.CodeBuild, names.ErrActionReading, ResWebhook, d.Id())
		d.SetId("")
		return nil
	}

	d.Set("build_type", project.Webhook.BuildType)
	d.Set("branch_filter", project.Webhook.BranchFilter)
	d.Set("filter_group", flattenWebhookFilterGroups(project.Webhook.FilterGroups))
	d.Set("payload_url", project.Webhook.PayloadUrl)
	d.Set("project_name", project.Name)
	d.Set("url", project.Webhook.Url)
	// The secret is never returned after creation, so don't set it here

	return nil
}

func resourceWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	var err error
	filterGroups := expandWebhookFilterGroups(d)

	var buildType *string = nil
	if v, ok := d.GetOk("build_type"); ok {
		buildType = aws.String(v.(string))
	}

	if len(filterGroups) >= 1 {
		_, err = conn.UpdateWebhook(&codebuild.UpdateWebhookInput{
			ProjectName:  aws.String(d.Id()),
			BuildType:    buildType,
			FilterGroups: filterGroups,
			RotateSecret: aws.Bool(false),
		})
	} else {
		_, err = conn.UpdateWebhook(&codebuild.UpdateWebhookInput{
			ProjectName:  aws.String(d.Id()),
			BuildType:    buildType,
			BranchFilter: aws.String(d.Get("branch_filter").(string)),
			RotateSecret: aws.Bool(false),
		})
	}

	if err != nil {
		return err
	}

	return resourceWebhookRead(d, meta)
}

func resourceWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	_, err := conn.DeleteWebhook(&codebuild.DeleteWebhookInput{
		ProjectName: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, codebuild.ErrCodeResourceNotFoundException) {
			return nil
		}
		return err
	}

	return nil
}

func flattenWebhookFilterGroups(filterList [][]*codebuild.WebhookFilter) *schema.Set {
	filterSet := schema.Set{
		F: resourceWebhookFilterHash,
	}

	for _, filters := range filterList {
		filterSet.Add(flattenWebhookFilterData(filters))
	}
	return &filterSet
}

func resourceWebhookFilterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	for _, g := range m {
		for _, f := range g.([]interface{}) {
			r := f.(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%s-", r["type"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", r["pattern"].(string)))
			buf.WriteString(fmt.Sprintf("%q", r["exclude_matched_pattern"]))
		}
	}

	return create.StringHashcode(buf.String())
}

func flattenWebhookFilterData(filters []*codebuild.WebhookFilter) map[string]interface{} {
	values := map[string]interface{}{}
	ff := make([]interface{}, 0)

	for _, f := range filters {
		ff = append(ff, map[string]interface{}{
			"type":                    *f.Type,
			"pattern":                 *f.Pattern,
			"exclude_matched_pattern": *f.ExcludeMatchedPattern,
		})
	}

	values["filter"] = ff

	return values
}
