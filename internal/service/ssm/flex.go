package ssm

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandTargets(in []interface{}) []*ssm.Target {
	targets := make([]*ssm.Target, 0)

	for _, tConfig := range in {
		config := tConfig.(map[string]interface{})

		target := &ssm.Target{
			Key:    aws.String(config["key"].(string)),
			Values: flex.ExpandStringList(config["values"].([]interface{})),
		}

		targets = append(targets, target)
	}

	return targets
}

func flattenParameters(parameters map[string][]*string) map[string]string {
	result := make(map[string]string)
	for p, values := range parameters {
		var vs []string
		for _, vPtr := range values {
			if v := aws.StringValue(vPtr); v != "" {
				vs = append(vs, v)
			}
		}
		result[p] = strings.Join(vs, ",")
	}
	return result
}

func flattenTargets(targets []*ssm.Target) []map[string]interface{} {
	if len(targets) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(targets))
	for _, target := range targets {
		t := make(map[string]interface{}, 1)
		t["key"] = aws.StringValue(target.Key)
		t["values"] = flex.FlattenStringList(target.Values)

		result = append(result, t)
	}

	return result
}
