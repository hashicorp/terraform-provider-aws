package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceRegexPatternSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceRegexPatternSetCreate,
		Read:   resourceRegexPatternSetRead,
		Update: resourceRegexPatternSetUpdate,
		Delete: resourceRegexPatternSetDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ID/NAME/SCOPE", d.Id())
				}
				id := idParts[0]
				name := idParts[1]
				scope := idParts[2]
				d.SetId(id)
				d.Set("name", name)
				d.Set("scope", scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"lock_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"regular_expression": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"regex_string": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 200),
						},
					},
				},
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRegexPatternSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	params := &wafv2.CreateRegexPatternSetInput{
		Name:                  aws.String(d.Get("name").(string)),
		Scope:                 aws.String(d.Get("scope").(string)),
		RegularExpressionList: []*wafv2.Regex{},
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("regular_expression"); ok && v.(*schema.Set).Len() > 0 {
		params.RegularExpressionList = expandWafv2RegexPatternSet(v.(*schema.Set).List())
	}

	if len(tags) > 0 {
		params.Tags = tags.IgnoreAws().Wafv2Tags()
	}

	resp, err := conn.CreateRegexPatternSet(params)

	if err != nil {
		return fmt.Errorf("Error creating WAFv2 RegexPatternSet: %s", err)
	}

	if resp == nil || resp.Summary == nil {
		return fmt.Errorf("Error creating WAFv2 RegexPatternSet")
	}

	d.SetId(aws.StringValue(resp.Summary.Id))

	return resourceRegexPatternSetRead(d, meta)
}

func resourceRegexPatternSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	params := &wafv2.GetRegexPatternSetInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetRegexPatternSet(params)
	if err != nil {
		if tfawserr.ErrMessageContains(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAFv2 RegexPatternSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.RegexPatternSet == nil {
		return fmt.Errorf("Error getting WAFv2 RegexPatternSet")
	}

	d.Set("name", resp.RegexPatternSet.Name)
	d.Set("description", resp.RegexPatternSet.Description)
	d.Set("arn", resp.RegexPatternSet.ARN)
	d.Set("lock_token", resp.LockToken)

	if err := d.Set("regular_expression", flattenWafv2RegexPatternSet(resp.RegexPatternSet.RegularExpressionList)); err != nil {
		return fmt.Errorf("Error setting regular_expression: %s", err)
	}

	tags, err := tftags.Wafv2ListTags(conn, aws.StringValue(resp.RegexPatternSet.ARN))
	if err != nil {
		return fmt.Errorf("Error listing tags for WAFv2 RegexPatternSet (%s): %s", aws.StringValue(resp.RegexPatternSet.ARN), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceRegexPatternSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn

	log.Printf("[INFO] Updating WAFv2 RegexPatternSet %s", d.Id())

	params := &wafv2.UpdateRegexPatternSetInput{
		Id:                    aws.String(d.Id()),
		Name:                  aws.String(d.Get("name").(string)),
		Scope:                 aws.String(d.Get("scope").(string)),
		LockToken:             aws.String(d.Get("lock_token").(string)),
		RegularExpressionList: []*wafv2.Regex{},
	}

	if v, ok := d.GetOk("regular_expression"); ok && v.(*schema.Set).Len() > 0 {
		params.RegularExpressionList = expandWafv2RegexPatternSet(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateRegexPatternSet(params)

	if err != nil {
		return fmt.Errorf("Error updating WAFv2 RegexPatternSet: %s", err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.Wafv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("Error updating tags: %s", err)
		}
	}

	return resourceRegexPatternSetRead(d, meta)
}

func resourceRegexPatternSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn

	log.Printf("[INFO] Deleting WAFv2 RegexPatternSet %s", d.Id())
	params := &wafv2.DeleteRegexPatternSetInput{
		Id:        aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
		LockToken: aws.String(d.Get("lock_token").(string)),
	}
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteRegexPatternSet(params)
		if err != nil {
			if tfawserr.ErrMessageContains(err, wafv2.ErrCodeWAFAssociatedItemException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteRegexPatternSet(params)
	}

	if err != nil {
		return fmt.Errorf("Error deleting WAFv2 RegexPatternSet: %s", err)
	}

	return nil
}

func expandWafv2RegexPatternSet(l []interface{}) []*wafv2.Regex {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	regexPatterns := make([]*wafv2.Regex, 0)
	for _, regexPattern := range l {
		if regexPattern == nil {
			continue
		}
		regexPatterns = append(regexPatterns, expandWafv2Regex(regexPattern.(map[string]interface{})))
	}

	return regexPatterns
}

func expandWafv2Regex(m map[string]interface{}) *wafv2.Regex {
	if m == nil {
		return nil
	}

	return &wafv2.Regex{
		RegexString: aws.String(m["regex_string"].(string)),
	}
}

func flattenWafv2RegexPatternSet(r []*wafv2.Regex) interface{} {
	if r == nil {
		return []interface{}{}
	}

	regexPatterns := make([]interface{}, 0)

	for _, regexPattern := range r {
		if regexPattern == nil {
			continue
		}
		d := map[string]interface{}{
			"regex_string": aws.StringValue(regexPattern.RegexString),
		}
		regexPatterns = append(regexPatterns, d)
	}

	return regexPatterns
}
