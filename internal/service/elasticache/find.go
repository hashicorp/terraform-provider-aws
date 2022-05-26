package elasticache

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindReplicationGroupByID retrieves an ElastiCache Replication Group by id.
func FindReplicationGroupByID(conn *elasticache.ElastiCache, id string) (*elasticache.ReplicationGroup, error) {
	input := &elasticache.DescribeReplicationGroupsInput{
		ReplicationGroupId: aws.String(id),
	}
	output, err := conn.DescribeReplicationGroups(input)
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeReplicationGroupNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ReplicationGroups) == 0 || output.ReplicationGroups[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return output.ReplicationGroups[0], nil
}

// FindReplicationGroupMemberClustersByID retrieves all of an ElastiCache Replication Group's MemberClusters by the id of the Replication Group.
func FindReplicationGroupMemberClustersByID(conn *elasticache.ElastiCache, id string) ([]*elasticache.CacheCluster, error) {
	rg, err := FindReplicationGroupByID(conn, id)
	if err != nil {
		return nil, err
	}

	clusters, err := FindCacheClustersByID(conn, aws.StringValueSlice(rg.MemberClusters))
	if err != nil {
		return clusters, err
	}
	if len(clusters) == 0 {
		return clusters, &resource.NotFoundError{
			Message: fmt.Sprintf("No Member Clusters found in Replication Group (%s)", id),
		}
	}

	return clusters, nil
}

// FindCacheClusterByID retrieves an ElastiCache Cache Cluster by id.
func FindCacheClusterByID(conn *elasticache.ElastiCache, id string) (*elasticache.CacheCluster, error) {
	input := &elasticache.DescribeCacheClustersInput{
		CacheClusterId: aws.String(id),
	}
	return FindCacheCluster(conn, input)
}

// FindCacheClusterWithNodeInfoByID retrieves an ElastiCache Cache Cluster with Node Info by id.
func FindCacheClusterWithNodeInfoByID(conn *elasticache.ElastiCache, id string) (*elasticache.CacheCluster, error) {
	input := &elasticache.DescribeCacheClustersInput{
		CacheClusterId:    aws.String(id),
		ShowCacheNodeInfo: aws.Bool(true),
	}
	return FindCacheCluster(conn, input)
}

// FindCacheCluster retrieves an ElastiCache Cache Cluster using DescribeCacheClustersInput.
func FindCacheCluster(conn *elasticache.ElastiCache, input *elasticache.DescribeCacheClustersInput) (*elasticache.CacheCluster, error) {
	result, err := conn.DescribeCacheClusters(input)
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheClusterNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil || len(result.CacheClusters) == 0 || result.CacheClusters[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return result.CacheClusters[0], nil
}

// FindCacheClustersByID retrieves a list of ElastiCache Cache Clusters by id.
// Order of the clusters is not guaranteed.
func FindCacheClustersByID(conn *elasticache.ElastiCache, idList []string) ([]*elasticache.CacheCluster, error) {
	var results []*elasticache.CacheCluster
	ids := make(map[string]bool)
	for _, v := range idList {
		ids[v] = true
	}

	input := &elasticache.DescribeCacheClustersInput{}
	err := conn.DescribeCacheClustersPages(input, func(page *elasticache.DescribeCacheClustersOutput, _ bool) bool {
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

// FindGlobalReplicationGroupByID() retrieves an ElastiCache Global Replication Group by id.
func FindGlobalReplicationGroupByID(conn *elasticache.ElastiCache, id string) (*elasticache.GlobalReplicationGroup, error) {
	input := &elasticache.DescribeGlobalReplicationGroupsInput{
		GlobalReplicationGroupId: aws.String(id),
		ShowMemberInfo:           aws.Bool(true),
	}
	output, err := conn.DescribeGlobalReplicationGroups(input)
	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeGlobalReplicationGroupNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.GlobalReplicationGroups) == 0 || output.GlobalReplicationGroups[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return output.GlobalReplicationGroups[0], nil
}

// FindGlobalReplicationGroupMemberByID retrieves a member Replication Group by id from a Global Replication Group.
func FindGlobalReplicationGroupMemberByID(conn *elasticache.ElastiCache, globalReplicationGroupID string, id string) (*elasticache.GlobalReplicationGroupMember, error) {
	globalReplicationGroup, err := FindGlobalReplicationGroupByID(conn, globalReplicationGroupID)
	if err != nil {
		return nil, &resource.NotFoundError{
			Message:   "unable to retrieve enclosing Global Replication Group",
			LastError: err,
		}
	}

	if globalReplicationGroup == nil || len(globalReplicationGroup.Members) == 0 {
		return nil, &resource.NotFoundError{
			Message: "empty result",
		}
	}

	for _, member := range globalReplicationGroup.Members {
		if aws.StringValue(member.ReplicationGroupId) == id {
			return member, nil
		}
	}

	return nil, &resource.NotFoundError{
		Message: fmt.Sprintf("Replication Group (%s) not found in Global Replication Group (%s)", id, globalReplicationGroupID),
	}
}

func FindUserByID(conn *elasticache.ElastiCache, userID string) (*elasticache.User, error) {
	input := &elasticache.DescribeUsersInput{
		UserId: aws.String(userID),
	}
	out, err := conn.DescribeUsers(input)

	if err != nil {
		return nil, err
	}

	switch len(out.Users) {
	case 0:
		return nil, &resource.NotFoundError{
			Message: "empty result",
		}
	case 1:
		return out.Users[0], nil
	default:
		return nil, &resource.NotFoundError{
			Message: "too many results",
		}
	}
}

func FindUserGroupByID(conn *elasticache.ElastiCache, groupID string) (*elasticache.UserGroup, error) {
	input := &elasticache.DescribeUserGroupsInput{
		UserGroupId: aws.String(groupID),
	}
	out, err := conn.DescribeUserGroups(input)
	if err != nil {
		return nil, err
	}

	switch count := len(out.UserGroups); count {
	case 0:
		return nil, tfresource.NewEmptyResultError(input)
	case 1:
		return out.UserGroups[0], nil
	default:
		return nil, tfresource.NewTooManyResultsError(count, input)
	}
}

func FindParameterGroupByName(conn *elasticache.ElastiCache, name string) (*elasticache.CacheParameterGroup, error) {
	input := elasticache.DescribeCacheParameterGroupsInput{
		CacheParameterGroupName: aws.String(name),
	}
	out, err := conn.DescribeCacheParameterGroups(&input)

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheParameterGroupNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	switch count := len(out.CacheParameterGroups); count {
	case 0:
		return nil, tfresource.NewEmptyResultError(input)
	case 1:
		return out.CacheParameterGroups[0], nil
	default:
		return nil, tfresource.NewTooManyResultsError(count, input)
	}
}
