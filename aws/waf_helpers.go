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

func flattenWafGeoMatchConstraint(ts []*waf.GeoMatchConstraint) []interface{} {
	out := make([]interface{}, len(ts), len(ts))
	for i, t := range ts {
		m := make(map[string]interface{})
		m["type"] = *t.Type
		m["value"] = *t.Value
		out[i] = m
	}
	return out
}

func diffWafGeoMatchSetConstraints(oldT, newT []interface{}) []*waf.GeoMatchSetUpdate {
	updates := make([]*waf.GeoMatchSetUpdate, 0)

	for _, od := range oldT {
		constraint := od.(map[string]interface{})

		if idx, contains := sliceContainsMap(newT, constraint); contains {
			newT = append(newT[:idx], newT[idx+1:]...)
			continue
		}

		updates = append(updates, &waf.GeoMatchSetUpdate{
			Action: aws.String(waf.ChangeActionDelete),
			GeoMatchConstraint: &waf.GeoMatchConstraint{
				Type:  aws.String(constraint["type"].(string)),
				Value: aws.String(constraint["value"].(string)),
			},
		})
	}

	for _, nd := range newT {
		constraint := nd.(map[string]interface{})

		updates = append(updates, &waf.GeoMatchSetUpdate{
			Action: aws.String(waf.ChangeActionInsert),
			GeoMatchConstraint: &waf.GeoMatchConstraint{
				Type:  aws.String(constraint["type"].(string)),
				Value: aws.String(constraint["value"].(string)),
			},
		})
	}
	return updates
}

func diffWafRegexPatternSetPatternStrings(oldPatterns, newPatterns []interface{}) []*waf.RegexPatternSetUpdate {
	updates := make([]*waf.RegexPatternSetUpdate, 0)

	for _, op := range oldPatterns {
		if idx, contains := sliceContainsString(newPatterns, op.(string)); contains {
			newPatterns = append(newPatterns[:idx], newPatterns[idx+1:]...)
			continue
		}

		updates = append(updates, &waf.RegexPatternSetUpdate{
			Action:             aws.String(waf.ChangeActionDelete),
			RegexPatternString: aws.String(op.(string)),
		})
	}

	for _, np := range newPatterns {
		updates = append(updates, &waf.RegexPatternSetUpdate{
			Action:             aws.String(waf.ChangeActionInsert),
			RegexPatternString: aws.String(np.(string)),
		})
	}
	return updates
}

func sliceContainsString(slice []interface{}, s string) (int, bool) {
	for idx, value := range slice {
		v := value.(string)
		if v == s {
			return idx, true
		}
	}
	return -1, false
}
