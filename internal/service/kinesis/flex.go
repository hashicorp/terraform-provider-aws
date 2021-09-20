package kinesis

import (
	"github.com/aws/aws-sdk-go/service/kinesis"
)

func flattenShardLevelMetrics(list []*kinesis.EnhancedMetrics) []string {
	if len(list) == 0 {
		return []string{}
	}
	strs := make([]string, 0, len(list[0].ShardLevelMetrics))
	for _, s := range list[0].ShardLevelMetrics {
		strs = append(strs, *s)
	}
	return strs
}
