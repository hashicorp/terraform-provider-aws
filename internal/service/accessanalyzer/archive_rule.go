package accessanalyzer

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceArchiveRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceArchiveRuleCreate,
		ReadWithoutTimeout:   resourceArchiveRuleRead,
		UpdateWithoutTimeout: resourceArchiveRuleUpdate,
		DeleteWithoutTimeout: resourceArchiveRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"analyzer_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"filter": {
				Type:     schema.TypeMap,
				Required: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"contains": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     schema.TypeString,
						},
						"eq": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							Elem:     schema.TypeString,
						},
						"exists": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"neq": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     schema.TypeString,
						},
					},
				},
			},
			"rule_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceArchiveRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccessAnalyzerConn

	analyzerName := d.Get("analyzer_name").(string)
	ruleName := d.Get("rule_name").(string)

	in := &accessanalyzer.CreateArchiveRuleInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(resource.UniqueId()),
		RuleName:     aws.String(ruleName),
	}

	if v, ok := d.GetOk("filter"); ok {
		in.Filter = expandFilter(v.(map[string]interface{}))
	}

	_, err := conn.CreateArchiveRuleWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("creating AWS IAM Access Analyzer ArchiveRule (%s): %s", d.Get("rule_name").(string), err)
	}

	id := EncodeRuleID(analyzerName, ruleName)
	d.SetId(id)

	return resourceArchiveRuleRead(ctx, d, meta)
}

func resourceArchiveRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccessAnalyzerConn

	analyzerName, ruleName := DecodeRuleID(d.Id())
	out, err := FindArchiveRule(ctx, conn, analyzerName, ruleName)

	// TIP: -- 3. Set ID to empty where resource is not new and not found
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AccessAnalyzer ArchiveRule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading AccessAnalyzer ArchiveRule (%s): %s", d.Id(), err)
	}

	d.Set("filter", flattenFilter(out.Filter))

	return nil
}

func resourceArchiveRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccessAnalyzerConn

	analyzerName, ruleName := DecodeRuleID(d.Id())
	in := &accessanalyzer.UpdateArchiveRuleInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(resource.UniqueId()),
		RuleName:     aws.String(ruleName),
	}

	if d.HasChanges("filter") {
		in.Filter = expandFilter(d.Get("filter").(map[string]interface{}))

	}

	log.Printf("[DEBUG] Updating AccessAnalyzer ArchiveRule (%s): %#v", d.Id(), in)
	_, err := conn.UpdateArchiveRuleWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("updating AccessAnalyzer ArchiveRule (%s): %s", d.Id(), err)
	}

	return resourceArchiveRuleRead(ctx, d, meta)
}

func resourceArchiveRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccessAnalyzerConn

	log.Printf("[INFO] Deleting AccessAnalyzer ArchiveRule %s", d.Id())

	analyzerName, ruleName := DecodeRuleID(d.Id())
	_, err := conn.DeleteArchiveRuleWithContext(ctx, &accessanalyzer.DeleteArchiveRuleInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(resource.UniqueId()),
		RuleName:     aws.String(ruleName),
	})

	if tfawserr.ErrCodeEquals(err, accessanalyzer.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting AccessAnalyzer ArchiveRule (%s): %s", d.Id(), err)
	}

	return nil
}

func FindArchiveRule(ctx context.Context, conn *accessanalyzer.AccessAnalyzer, analyzerName, ruleName string) (*accessanalyzer.ArchiveRuleSummary, error) {
	in := &accessanalyzer.GetArchiveRuleInput{
		AnalyzerName: aws.String(analyzerName),
		RuleName:     aws.String(ruleName),
	}

	out, err := conn.GetArchiveRuleWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, accessanalyzer.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.ArchiveRule == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.ArchiveRule, nil
}

func flattenFilter(filter map[string]*accessanalyzer.Criterion) map[string]interface{} {
	if filter == nil {
		return nil
	}

	m := map[string]interface{}{}

	for key, value := range filter {
		val := make(map[string]interface{})
		val["contains"] = aws.ToStringSlice(value.Contains)
		val["eq"] = aws.ToStringSlice(value.Eq)
		val["exists"] = aws.ToBool(value.Exists)
		val["neq"] = aws.ToStringSlice(value.Neq)

		m[key] = val
	}

	return m
}

func expandFilter(tfMap map[string]interface{}) map[string]*accessanalyzer.Criterion {
	if tfMap == nil {
		return nil
	}

	a := make(map[string]*accessanalyzer.Criterion)

	for key, value := range tfMap {
		c := &accessanalyzer.Criterion{}
		if v, ok := value.(map[string]interface{})["contains"]; ok {
			c.Contains = aws.StringSlice(v.([]string))
		}
		if v, ok := value.(map[string]interface{})["eq"]; ok {
			c.Eq = aws.StringSlice(v.([]string))
		}
		if v, ok := value.(map[string]interface{})["exists"]; ok {
			c.Exists = aws.Bool(v.(bool))
		}
		if v, ok := value.(map[string]interface{})["neq"]; ok {
			c.Neq = aws.StringSlice(v.([]string))
		}

		a[key] = c
	}

	return a
}

func EncodeRuleID(analyzerName, ruleName string) string {
	return fmt.Sprintf("%s/%s", analyzerName, ruleName)
}

func DecodeRuleID(id string) (string, string) {
	parts := strings.Split(id, "/")

	return parts[0], parts[1]
}
