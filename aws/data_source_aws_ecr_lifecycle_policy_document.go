package aws

import (
	"encoding/json"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

type ECRLifecyclePolicyDocument struct {
	Id    string                    `json:",omitempty"`
	Rules []*ECRLifecyclePolicyRule `json:"rules"`
}

type ECRLifecyclePolicyRule struct {
	Priority    int                            `json:"rulePriority,omitempty"`
	Description string                         `json:"description,omitempty"`
	Selection   ECRLifecyclePolicySelectionSet `json:"selection,omitempty"`
	Action      ECRLifecyclePolicyActionSet    `json:"action,omitempty"`
}

type ECRLifecyclePolicySelectionSet struct {
	TagStatus     string      `json:"tagStatus,omitempty"`
	TagPrefixList interface{} `json:"tagPrefixList,omitempty"`
	CountType     string      `json:"countType,omitempty"`
	CountUnit     string      `json:"countUnit,omitempty"`
	CountNumber   int         `json:"countNumber,omitempty"`
}

type ECRLifecyclePolicyActionSet struct {
	Type string `json:"type,omitempty"`
}

func (self *ECRLifecyclePolicyDocument) merge(newDoc *ECRLifecyclePolicyDocument) {
	// adopt newDoc's Id
	if len(newDoc.Id) > 0 {
		self.Id = newDoc.Id
	}

	// merge in newDoc's statements, overwriting any existing Sids
	var seen bool
	for _, newRule := range newDoc.Rules {
		seen = false
		for _, existingRule := range self.Rules {
			if existingRule.Priority == newRule.Priority {
				panic("Unsupported merge of rules with the same rule priority")
			}
		}

		if !seen {
			self.Rules = append(self.Rules, newRule)
		}
	}
}

func dataSourceAwsEcrLifecyclePolicyDocument() *schema.Resource {

	return &schema.Resource{
		Read: dataSourceAwsEcrLifecyclePolicyDocumentRead,

		Schema: map[string]*schema.Schema{
			"source_json": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"priority": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"selection": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tag_status": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"untagged", "tagged"}, false),
									},
									"tag_prefixes": &schema.Schema{
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"count_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"imageCountMoreThan", "sinceImagePushed"}, false),
									},
									"count_unit": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"days"}, false),
									},
									"count_number": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"action": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"expire"}, false),
									},
								},
							},
						},
					},
				},
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEcrLifecyclePolicyDocumentRead(d *schema.ResourceData, meta interface{}) error {
	mergedDoc := &ECRLifecyclePolicyDocument{}

	// populate mergedDoc directly with any source_json
	if sourceJson, hasSourceJson := d.GetOk("source_json"); hasSourceJson {
		if err := json.Unmarshal([]byte(sourceJson.(string)), mergedDoc); err != nil {
			return err
		}
	}

	// process the current document
	doc := &ECRLifecyclePolicyDocument{}

	var cfgRules = d.Get("rule").([]interface{})
	rules := make([]*ECRLifecyclePolicyRule, len(cfgRules))
	for i, ruleI := range cfgRules {
		cfgRule := ruleI.(map[string]interface{})
		rule := &ECRLifecyclePolicyRule{}

		if priority, ok := cfgRule["priority"]; ok {
			rule.Priority, _ = strconv.Atoi(priority.(string))
		}

		if description, ok := cfgRule["description"]; ok {
			rule.Description = description.(string)
		}

		rule.Selection = ecrLifecyclePolicyMakeSelection(cfgRule["selection"].(*schema.Set).List())
		rule.Action = ecrLifecyclePolicyMakeAction(cfgRule["action"].(*schema.Set).List())

		rules[i] = rule
	}

	doc.Rules = rules

	// merge our current document into mergedDoc
	mergedDoc.merge(doc)

	jsonDoc, err := json.MarshalIndent(mergedDoc, "", "  ")
	if err != nil {
		// should never happen if the above code is correct
		return err
	}
	jsonString := string(jsonDoc)

	d.Set("json", jsonString)
	d.SetId(strconv.Itoa(hashcode.String(jsonString)))

	return nil
}

func ecrLifecyclePolicyMakeSelection(in []interface{}) ECRLifecyclePolicySelectionSet {
	if len(in) == 1 {
		selection := in[0].(map[string]interface{})
		out := &ECRLifecyclePolicySelectionSet{}

		if tagStatus, ok := selection["tag_status"]; ok {
			out.TagStatus = tagStatus.(string)
		}

		if tagPrefixList := selection["tag_prefixes"].(*schema.Set).List(); len(tagPrefixList) > 0 {
			out.TagPrefixList = ecrLifecyclePolicyDecodeConfigStringList(tagPrefixList)
		}

		if countType, ok := selection["count_type"]; ok {
			out.CountType = countType.(string)
		}

		if countUnit, ok := selection["count_unit"]; ok {
			out.CountUnit = countUnit.(string)
		}

		if countNumber, ok := selection["count_number"]; ok {
			out.CountNumber, _ = strconv.Atoi(countNumber.(string))
		}

		return *out
	} else {
		panic("Cannot specify more than one selection per rule")
	}
}

func ecrLifecyclePolicyMakeAction(in []interface{}) ECRLifecyclePolicyActionSet {
	if len(in) == 0 {
		return ECRLifecyclePolicyActionSet{
			Type: "expire",
		}
	} else if len(in) == 1 {
		action := in[0].(map[string]interface{})

		return ECRLifecyclePolicyActionSet{
			Type: action["type"].(string),
		}
	} else {
		panic("Cannot specify more than one action per rule")
	}
}

func ecrLifecyclePolicyDecodeConfigStringList(lI []interface{}) interface{} {
	ret := make([]string, len(lI))
	for i, vI := range lI {
		ret[i] = vI.(string)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ret)))
	return ret
}
