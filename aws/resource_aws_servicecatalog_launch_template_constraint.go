package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsServiceCatalogLaunchTemplateConstraint() *schema.Resource {
	var awsResourceIdPattern = regexp.MustCompile("^[a-zA-Z0-9_\\-]*")
	return &schema.Resource{
		Create: resourceAwsServiceCatalogLaunchTemplateConstraintCreate,
		Read:   resourceAwsServiceCatalogLaunchTemplateConstraintRead,
		Update: resourceAwsServiceCatalogLaunchTemplateConstraintUpdate,
		Delete: resourceAwsServiceCatalogLaunchTemplateConstraintDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rule_condition": {
							Type:     schema.TypeString, //JSON: Rule-specific intrinsic function
							Optional: true,
						},
						"assertion": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"assert": {
										Type:     schema.TypeString, //JSON: Rule-specific intrinsic function
										Required: true,
									},
									"assert_description": {
										Type:     schema.TypeString,
										Optional: true, //might be required
									},
								},
							},
						},
					},
				},
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					awsResourceIdPattern,
					"invalid id format"),
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					awsResourceIdPattern,
					"invalid id format"),
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameters": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsServiceCatalogLaunchTemplateConstraintCreate(d *schema.ResourceData, meta interface{}) error {
	jsonDoc, errJson := resourceAwsServiceCatalogLaunchTemplateConstraintJsonParameters(d)
	if errJson != nil {
		return errJson
	}
	log.Printf("[DEBUG] constraint Parameters (JSON): %s\n", jsonDoc)
	errCreate := resourceAwsServiceCatalogConstraintCreateFromJson(d, meta, jsonDoc, "TEMPLATE")
	if errCreate != nil {
		return errCreate
	}
	return resourceAwsServiceCatalogLaunchTemplateConstraintRead(d, meta)
}

type awsServiceCatalogLaunchTemplateConstraintRuleAssert struct {
	Assert            interface{} //JSON
	AssertDescription string
}

type awsServiceCatalogLaunchTemplateConstraintRule struct {
	RuleCondition interface{} //JSON
	Assertions    []awsServiceCatalogLaunchTemplateConstraintRuleAssert
}

type awsServiceCatalogLaunchTemplateConstraint struct {
	Description string
	PortfolioId string
	ProductId   string
	Rules       map[string]awsServiceCatalogLaunchTemplateConstraintRule
}

func resourceAwsServiceCatalogLaunchTemplateConstraintJsonParameters(d *schema.ResourceData) (string, error) {
	constraint := awsServiceCatalogLaunchTemplateConstraint{}
	if description, ok := d.GetOk("description"); ok && description != "" {
		constraint.Description = description.(string)
	}
	constraint.PortfolioId = d.Get("portfolio_id").(string)
	constraint.ProductId = d.Get("product_id").(string)
	if err := resourceAwsServiceCatalogLaunchTemplateConstraintParseRules(d, &constraint); err != nil {
		return "", err
	}
	marshal, err := json.Marshal(constraint)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

func resourceAwsServiceCatalogLaunchTemplateConstraintParseRules(d *schema.ResourceData, constraint *awsServiceCatalogLaunchTemplateConstraint) error {
	constraint.Rules = map[string]awsServiceCatalogLaunchTemplateConstraintRule{}
	if rules, ok := d.GetOk("rule"); ok {
		for _, ruleMap := range rules.([]interface{}) {
			name, rule, err := resourceAwsServiceCatalogLaunchTemplateConstraintParseRule(ruleMap.(map[string]interface{}))
			if err != nil {
				return err
			}
			constraint.Rules[name] = *rule
		}
	}
	return nil
}

func resourceAwsServiceCatalogLaunchTemplateConstraintParseRule(m map[string]interface{}) (string, *awsServiceCatalogLaunchTemplateConstraintRule, error) {
	rule := awsServiceCatalogLaunchTemplateConstraintRule{}
	name := ""
	for k, v := range m {
		if k == "name" {
			name = v.(string)
		} else if k == "rule_condition" {
			err := json.Unmarshal([]byte(v.(string)), &rule.RuleCondition)
			if err != nil {
				log.Printf("[ERROR] rule.RuleCondition Unmarshal error: %s\n%s\n", err.Error(), rule.RuleCondition)
				return "", nil, err
			}
		} else if k == "assertion" {
			for _, assertMap := range v.([]interface{}) {
				assert, err := resourceAwsServiceCatalogLaunchTemplateConstraintParseRuleAsserts(assertMap.(map[string]interface{}))
				if err != nil {
					return "", nil, err
				}
				rule.Assertions = append(rule.Assertions, *assert)
			}
		}
	}
	if name == "" {
		return "", nil, fmt.Errorf("rule name is missing")
	}
	return name, &rule, nil
}

func resourceAwsServiceCatalogLaunchTemplateConstraintParseRuleAsserts(m map[string]interface{}) (*awsServiceCatalogLaunchTemplateConstraintRuleAssert, error) {
	assert := awsServiceCatalogLaunchTemplateConstraintRuleAssert{}
	for k, v := range m {
		if k == "assert" {
			err := json.Unmarshal([]byte(v.(string)), &assert.Assert)
			if err != nil {
				log.Printf("[ERROR] assert.Assert Unmarshal error: %s\n%s\n", err.Error(), assert.Assert)
				return nil, err
			}
		} else if k == "assert_description" {
			assert.AssertDescription = v.(string)
		}
	}
	return &assert, nil
}

func resourceAwsServiceCatalogLaunchTemplateConstraintRead(d *schema.ResourceData, meta interface{}) error {
	constraint, err := resourceAwsServiceCatalogConstraintReadBase(d, meta)
	if err != nil {
		return err
	}
	if constraint == nil {
		return nil
	}
	rule, err := flattenAwsServiceCatalogLaunchTemplateConstraintReadParameters(constraint.ConstraintParameters)
	if err != nil {
		return err
	}
	if err := d.Set("rule", rule); err != nil {
		return fmt.Errorf("error setting rule: %s", err)
	}

	return nil
}

func flattenAwsServiceCatalogLaunchTemplateConstraintReadParameters(configured *string) ([]map[string]interface{}, error) {
	// ConstraintParameters is returned from AWS as a JSON string
	var bytes = []byte(*configured)
	var constraint awsServiceCatalogLaunchTemplateConstraint
	err := json.Unmarshal(bytes, &constraint)
	if err != nil {
		log.Printf("[ERROR] parameters Unmarshal error: %s\n%s\n", err.Error(), constraint)
		return nil, err
	}
	rules, err := flattenAwsServiceCatalogLaunchTemplateConstraintReadRules(constraint.Rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func flattenAwsServiceCatalogLaunchTemplateConstraintReadRules(configured map[string]awsServiceCatalogLaunchTemplateConstraintRule) ([]map[string]interface{}, error) {
	rules := make([]map[string]interface{}, 0, len(configured))
	for name, item := range configured {
		rule, err := flattenAwsServiceCatalogLaunchTemplateConstraintReadRule(name, item)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func flattenAwsServiceCatalogLaunchTemplateConstraintReadRule(name string, configured awsServiceCatalogLaunchTemplateConstraintRule) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	m["name"] = name
	if configured.RuleCondition != nil {
		jsonDoc, err := json.Marshal(configured.RuleCondition)
		if err != nil {
			return nil, err
		}
		m["rule_condition"] = string(jsonDoc)
	}
	assertions, err := flattenAwsServiceCatalogLaunchTemplateConstraintReadRuleAssertions(configured.Assertions)
	if err != nil {
		return nil, err
	}
	m["assertion"] = assertions
	return m, nil
}

func flattenAwsServiceCatalogLaunchTemplateConstraintReadRuleAssertions(assertions []awsServiceCatalogLaunchTemplateConstraintRuleAssert) ([]map[string]interface{}, error) {
	m := make([]map[string]interface{}, 0, len(assertions))
	for _, assertion := range assertions {
		ruleAssertion, err := flattenAwsServiceCatalogLaunchTemplateConstraintReadRuleAssertion(assertion)
		if err != nil {
			return nil, err
		}
		m = append(m, ruleAssertion)
	}
	return m, nil
}

func flattenAwsServiceCatalogLaunchTemplateConstraintReadRuleAssertion(assertion awsServiceCatalogLaunchTemplateConstraintRuleAssert) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	if assertion.AssertDescription != "" {
		m["assert_description"] = assertion.AssertDescription
	}
	jsonDoc, err := json.Marshal(assertion.Assert)
	if err != nil {
		return nil, err
	}
	m["assert"] = string(jsonDoc)
	return m, nil
}

func resourceAwsServiceCatalogLaunchTemplateConstraintUpdate(d *schema.ResourceData, meta interface{}) error {
	input := servicecatalog.UpdateConstraintInput{}
	if d.HasChanges("launch_role_arn", "role_arn") {
		parameters, err := resourceAwsServiceCatalogLaunchTemplateConstraintJsonParameters(d)
		if err != nil {
			return err
		}
		input.Parameters = aws.String(parameters)
	}
	err := resourceAwsServiceCatalogConstraintUpdateBase(d, meta, input)
	if err != nil {
		return err
	}
	return resourceAwsServiceCatalogLaunchTemplateConstraintRead(d, meta)
}

func resourceAwsServiceCatalogLaunchTemplateConstraintDelete(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsServiceCatalogConstraintDelete(d, meta)
}
