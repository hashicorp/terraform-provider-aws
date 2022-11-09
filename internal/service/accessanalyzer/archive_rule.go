package accessanalyzer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criteria": {
							Type:     schema.TypeString,
							Required: true,
						},
						"contains": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"eq": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"exists": {
							Type:         nullable.TypeNullableBool,
							Optional:     true,
							Computed:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableBool,
						},
						"neq": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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
		in.Filter = expandFilter(v.(*schema.Set))
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

	analyzerName, ruleName, err := DecodeRuleID(d.Id())
	if err != nil {
		return diag.Errorf("unable to decode AccessAnalyzer ArchiveRule ID (%s): %s", d.Id(), err)
	}

	out, err := FindArchiveRule(ctx, conn, analyzerName, ruleName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AccessAnalyzer ArchiveRule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading AccessAnalyzer ArchiveRule (%s): %s", d.Id(), err)
	}

	d.Set("analyzer_name", analyzerName)
	d.Set("filter", flattenFilter(out.Filter))
	d.Set("rule_name", out.RuleName)

	return nil
}

func resourceArchiveRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccessAnalyzerConn

	analyzerName, ruleName, err := DecodeRuleID(d.Id())
	if err != nil {
		return diag.Errorf("unable to decode AccessAnalyzer ArchiveRule ID (%s): %s", d.Id(), err)
	}

	in := &accessanalyzer.UpdateArchiveRuleInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(resource.UniqueId()),
		RuleName:     aws.String(ruleName),
	}

	if d.HasChanges("filter") {
		in.Filter = expandFilter(d.Get("filter").(*schema.Set))
	}

	log.Printf("[DEBUG] Updating AccessAnalyzer ArchiveRule (%s): %#v", d.Id(), in)
	_, err = conn.UpdateArchiveRuleWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("updating AccessAnalyzer ArchiveRule (%s): %s", d.Id(), err)
	}

	return resourceArchiveRuleRead(ctx, d, meta)
}

func resourceArchiveRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AccessAnalyzerConn

	log.Printf("[INFO] Deleting AccessAnalyzer ArchiveRule %s", d.Id())

	analyzerName, ruleName, err := DecodeRuleID(d.Id())
	if err != nil {
		return diag.Errorf("unable to decode AccessAnalyzer ArchiveRule ID (%s): %s", d.Id(), err)
	}

	_, err = conn.DeleteArchiveRuleWithContext(ctx, &accessanalyzer.DeleteArchiveRuleInput{
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

func flattenFilter(filter map[string]*accessanalyzer.Criterion) []interface{} {
	if filter == nil {
		return nil
	}

	l := make([]interface{}, 0)

	for key, value := range filter {
		val := make(map[string]interface{})
		val["criteria"] = key
		val["contains"] = aws.ToStringSlice(value.Contains)
		val["eq"] = aws.ToStringSlice(value.Eq)

		if value.Exists != nil {
			val["exists"] = strconv.FormatBool(aws.ToBool(value.Exists))
		}

		val["neq"] = aws.ToStringSlice(value.Neq)

		l = append(l, val)
	}

	return l
}

func expandFilter(l *schema.Set) map[string]*accessanalyzer.Criterion {
	if len(l.List()) == 0 || l.List()[0] == nil {
		return nil
	}

	a := make(map[string]*accessanalyzer.Criterion)

	for _, value := range l.List() {
		c := &accessanalyzer.Criterion{}
		if v, ok := value.(map[string]interface{})["contains"]; ok {
			if len(v.([]interface{})) > 0 {
				c.Contains = flex.ExpandStringList(v.([]interface{}))
			}
		}
		if v, ok := value.(map[string]interface{})["eq"]; ok {
			if len(v.([]interface{})) > 0 {
				c.Eq = flex.ExpandStringList(v.([]interface{}))
			}
		}
		if v, ok := value.(map[string]interface{})["neq"]; ok {
			if len(v.([]interface{})) > 0 {
				c.Neq = flex.ExpandStringList(v.([]interface{}))
			}
		}
		if v, ok := value.(map[string]interface{})["exists"]; ok {
			if val, null, _ := nullable.Bool(v.(string)).Value(); !null {
				c.Exists = aws.Bool(val)
			}
		}

		a[value.(map[string]interface{})["criteria"].(string)] = c
	}

	return a
}

func EncodeRuleID(analyzerName, ruleName string) string {
	return fmt.Sprintf("%s/%s", analyzerName, ruleName)
}

func DecodeRuleID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID to be the form analyzer_name/rule_name, given: %s", id)
	}

	return idParts[0], idParts[1], nil
}
