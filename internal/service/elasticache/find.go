// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type redisParameterGroupFilter func(group *elasticache.CacheParameterGroup) bool

func FindParameterGroupByFilter(ctx context.Context, conn *elasticache.ElastiCache, filters ...redisParameterGroupFilter) (*elasticache.CacheParameterGroup, error) {
	parameterGroups, err := ListParameterGroups(ctx, conn, filters...)
	if err != nil {
		return nil, err
	}

	switch count := len(parameterGroups); count {
	case 0:
		return nil, tfresource.NewEmptyResultError(nil)
	case 1:
		return parameterGroups[0], nil
	default:
		return nil, tfresource.NewTooManyResultsError(count, nil)
	}
}

func ListParameterGroups(ctx context.Context, conn *elasticache.ElastiCache, filters ...redisParameterGroupFilter) ([]*elasticache.CacheParameterGroup, error) {
	var parameterGroups []*elasticache.CacheParameterGroup
	err := conn.DescribeCacheParameterGroupsPagesWithContext(ctx, &elasticache.DescribeCacheParameterGroupsInput{}, func(page *elasticache.DescribeCacheParameterGroupsOutput, lastPage bool) bool {
	PARAM_GROUPS:
		for _, parameterGroup := range page.CacheParameterGroups {
			for _, filter := range filters {
				if !filter(parameterGroup) {
					continue PARAM_GROUPS
				}
			}
			parameterGroups = append(parameterGroups, parameterGroup)
		}
		return !lastPage
	})
	return parameterGroups, err
}

func FilterRedisParameterGroupFamily(familyName string) redisParameterGroupFilter {
	return func(group *elasticache.CacheParameterGroup) bool {
		return aws.StringValue(group.CacheParameterGroupFamily) == familyName
	}
}

func FilterRedisParameterGroupNameDefault(group *elasticache.CacheParameterGroup) bool {
	name := aws.StringValue(group.CacheParameterGroupName)
	if strings.HasPrefix(name, "default.") && !strings.HasSuffix(name, ".cluster.on") {
		return true
	}
	return false
}

func FilterRedisParameterGroupNameClusterEnabledDefault(group *elasticache.CacheParameterGroup) bool {
	name := aws.StringValue(group.CacheParameterGroupName)
	if strings.HasPrefix(name, "default.") && strings.HasSuffix(name, ".cluster.on") {
		return true
	}
	return false
}
