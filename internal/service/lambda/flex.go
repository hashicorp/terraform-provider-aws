package lambda

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func flattenAliasRoutingConfiguration(arc *lambda.AliasRoutingConfiguration) []interface{} {
	if arc == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"additional_version_weights": aws.Float64ValueMap(arc.AdditionalVersionWeights),
	}

	return []interface{}{m}
}

func flattenLayers(layers []*lambda.Layer) []interface{} {
	arns := make([]*string, len(layers))
	for i, layer := range layers {
		arns[i] = layer.Arn
	}
	return flex.FlattenStringList(arns)
}

func flattenVPCConfigResponse(s *lambda.VpcConfigResponse) []map[string]interface{} {
	settings := make(map[string]interface{})

	if s == nil {
		return nil
	}

	var emptyVpc bool
	if aws.StringValue(s.VpcId) == "" {
		emptyVpc = true
	}
	if len(s.SubnetIds) == 0 && len(s.SecurityGroupIds) == 0 && emptyVpc {
		return nil
	}

	settings["subnet_ids"] = flex.FlattenStringSet(s.SubnetIds)
	settings["security_group_ids"] = flex.FlattenStringSet(s.SecurityGroupIds)
	if s.VpcId != nil {
		settings["vpc_id"] = aws.StringValue(s.VpcId)
	}

	return []map[string]interface{}{settings}
}

// Expands a map of string to interface to a map of string to *float
func expandFloat64Map(m map[string]interface{}) map[string]*float64 {
	float64Map := make(map[string]*float64, len(m))
	for k, v := range m {
		float64Map[k] = aws.Float64(v.(float64))
	}
	return float64Map
}
