package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform/helper/schema"
)

func wafSizeConstraintSetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"size_constraints": &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"field_to_match": {
						Type:     schema.TypeSet,
						Required: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"data": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"type": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					"comparison_operator": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
					"size": &schema.Schema{
						Type:     schema.TypeInt,
						Required: true,
					},
					"text_transformation": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
	}
}

func diffWafSizeConstraints(oldS, newS []interface{}) []*waf.SizeConstraintSetUpdate {
	updates := make([]*waf.SizeConstraintSetUpdate, 0)

	for _, os := range oldS {
		constraint := os.(map[string]interface{})

		if idx, contains := sliceContainsMap(newS, constraint); contains {
			newS = append(newS[:idx], newS[idx+1:]...)
			continue
		}

		updates = append(updates, &waf.SizeConstraintSetUpdate{
			Action: aws.String(waf.ChangeActionDelete),
			SizeConstraint: &waf.SizeConstraint{
				FieldToMatch:       expandFieldToMatch(constraint["field_to_match"].(*schema.Set).List()[0].(map[string]interface{})),
				ComparisonOperator: aws.String(constraint["comparison_operator"].(string)),
				Size:               aws.Int64(int64(constraint["size"].(int))),
				TextTransformation: aws.String(constraint["text_transformation"].(string)),
			},
		})
	}

	for _, ns := range newS {
		constraint := ns.(map[string]interface{})

		updates = append(updates, &waf.SizeConstraintSetUpdate{
			Action: aws.String(waf.ChangeActionInsert),
			SizeConstraint: &waf.SizeConstraint{
				FieldToMatch:       expandFieldToMatch(constraint["field_to_match"].(*schema.Set).List()[0].(map[string]interface{})),
				ComparisonOperator: aws.String(constraint["comparison_operator"].(string)),
				Size:               aws.Int64(int64(constraint["size"].(int))),
				TextTransformation: aws.String(constraint["text_transformation"].(string)),
			},
		})
	}
	return updates
}

func flattenWafSizeConstraints(sc []*waf.SizeConstraint) []interface{} {
	out := make([]interface{}, len(sc), len(sc))
	for i, c := range sc {
		m := make(map[string]interface{})
		m["comparison_operator"] = *c.ComparisonOperator
		if c.FieldToMatch != nil {
			m["field_to_match"] = flattenFieldToMatch(c.FieldToMatch)
		}
		m["size"] = *c.Size
		m["text_transformation"] = *c.TextTransformation
		out[i] = m
	}
	return out
}
