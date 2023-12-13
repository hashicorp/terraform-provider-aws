// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindReplicationGroupByID retrieves an ElastiCache Replication Group by id.
func FindReplicationGroupByID(ctx context.Context, conn *elasticache.ElastiCache, id string) (*elasticache.ReplicationGroup, error) {
	input := &elasticache.DescribeReplicationGroupsInput{
		ReplicationGroupId: aws.String(id),
	}
	output, err := conn.DescribeReplicationGroupsWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeReplicationGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ReplicationGroups) == 0 || output.ReplicationGroups[0] == nil {
		return nil, &retry.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return output.ReplicationGroups[0], nil
}

// FindReplicationGroupMemberClustersByID retrieves all of an ElastiCache Replication Group's MemberClusters by the id of the Replication Group.
func FindReplicationGroupMemberClustersByID(ctx context.Context, conn *elasticache.ElastiCache, id string) ([]*elasticache.CacheCluster, error) {
	rg, err := FindReplicationGroupByID(ctx, conn, id)
	if err != nil {
		return nil, err
	}

	clusters, err := FindCacheClustersByID(ctx, conn, aws.StringValueSlice(rg.MemberClusters))
	if err != nil {
		return clusters, err
	}
	if len(clusters) == 0 {
		return clusters, &retry.NotFoundError{
			Message: fmt.Sprintf("No Member Clusters found in Replication Group (%s)", id),
		}
	}

	return clusters, nil
}

// FindCacheClusterByID retrieves an ElastiCache Cache Cluster by id.
func FindCacheClusterByID(ctx context.Context, conn *elasticache.ElastiCache, id string) (*elasticache.CacheCluster, error) {
	input := &elasticache.DescribeCacheClustersInput{
		CacheClusterId: aws.String(id),
	}
	return FindCacheCluster(ctx, conn, input)
}

// FindCacheClusterWithNodeInfoByID retrieves an ElastiCache Cache Cluster with Node Info by id.
func FindCacheClusterWithNodeInfoByID(ctx context.Context, conn *elasticache.ElastiCache, id string) (*elasticache.CacheCluster, error) {
	input := &elasticache.DescribeCacheClustersInput{
		CacheClusterId:    aws.String(id),
		ShowCacheNodeInfo: aws.Bool(true),
	}
	return FindCacheCluster(ctx, conn, input)
}

// FindCacheCluster retrieves an ElastiCache Cache Cluster using DescribeCacheClustersInput.
func FindCacheCluster(ctx context.Context, conn *elasticache.ElastiCache, input *elasticache.DescribeCacheClustersInput) (*elasticache.CacheCluster, error) {
	result, err := conn.DescribeCacheClustersWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheClusterNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil || len(result.CacheClusters) == 0 || result.CacheClusters[0] == nil {
		return nil, &retry.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return result.CacheClusters[0], nil
}

// FindCacheClustersByID retrieves a list of ElastiCache Cache Clusters by id.
// Order of the clusters is not guaranteed.
func FindCacheClustersByID(ctx context.Context, conn *elasticache.ElastiCache, idList []string) ([]*elasticache.CacheCluster, error) {
	var results []*elasticache.CacheCluster
	ids := make(map[string]bool)
	for _, v := range idList {
		ids[v] = true
	}

	input := &elasticache.DescribeCacheClustersInput{}
	err := conn.DescribeCacheClustersPagesWithContext(ctx, input, func(page *elasticache.DescribeCacheClustersOutput, _ bool) bool {
		if page == nil || page.CacheClusters == nil || len(page.CacheClusters) == 0 {
			return true
		}

		for _, v := range page.CacheClusters {
			if ids[aws.StringValue(v.CacheClusterId)] {
				results = append(results, v)
				delete(ids, aws.StringValue(v.CacheClusterId))
				if len(ids) == 0 {
					break
				}
			}
		}

		return len(ids) != 0
	})

	return results, err
}

// FindGlobalReplicationGroupByID retrieves an ElastiCache Global Replication Group by id.
func FindGlobalReplicationGroupByID(ctx context.Context, conn *elasticache.ElastiCache, id string) (*elasticache.GlobalReplicationGroup, error) {
	input := &elasticache.DescribeGlobalReplicationGroupsInput{
		GlobalReplicationGroupId: aws.String(id),
		ShowMemberInfo:           aws.Bool(true),
	}
	output, err := conn.DescribeGlobalReplicationGroupsWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.GlobalReplicationGroups) == 0 || output.GlobalReplicationGroups[0] == nil {
		return nil, &retry.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return output.GlobalReplicationGroups[0], nil
}

// FindGlobalReplicationGroupMemberByID retrieves a member Replication Group by id from a Global Replication Group.
func FindGlobalReplicationGroupMemberByID(ctx context.Context, conn *elasticache.ElastiCache, globalReplicationGroupID string, id string) (*elasticache.GlobalReplicationGroupMember, error) {
	globalReplicationGroup, err := FindGlobalReplicationGroupByID(ctx, conn, globalReplicationGroupID)
	if err != nil {
		return nil, &retry.NotFoundError{
			Message:   "unable to retrieve enclosing Global Replication Group",
			LastError: err,
		}
	}

	if globalReplicationGroup == nil || len(globalReplicationGroup.Members) == 0 {
		return nil, &retry.NotFoundError{
			Message: "empty result",
		}
	}

	for _, member := range globalReplicationGroup.Members {
		if aws.StringValue(member.ReplicationGroupId) == id {
			return member, nil
		}
	}

	return nil, &retry.NotFoundError{
		Message: fmt.Sprintf("Replication Group (%s) not found in Global Replication Group (%s)", id, globalReplicationGroupID),
	}
}

func FindParameterGroupByName(ctx context.Context, conn *elasticache.ElastiCache, name string) (*elasticache.CacheParameterGroup, error) {
	input := elasticache.DescribeCacheParameterGroupsInput{
		CacheParameterGroupName: aws.String(name),
	}

	output, err := conn.DescribeCacheParameterGroupsWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheParameterGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSinglePtrResult(output.CacheParameterGroups)
}

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

func FindCacheSubnetGroupByName(ctx context.Context, conn *elasticache.ElastiCache, name string) (*elasticache.CacheSubnetGroup, error) {
	input := elasticache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: aws.String(name),
	}

	output, err := conn.DescribeCacheSubnetGroupsWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheSubnetGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CacheSubnetGroups) == 0 || output.CacheSubnetGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.CacheSubnetGroups); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.CacheSubnetGroups[0], nil
}
