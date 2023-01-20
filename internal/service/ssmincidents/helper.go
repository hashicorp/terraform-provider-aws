package ssmincidents

// contains misc functions used by multiple files

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
)

func ConvertInterfaceMapToStringMap(in map[string]interface{}) map[string]string {
	result := map[string]string{}

	for k, v := range in {
		result[k] = v.(string)
	}
	return result
}

func GetReplicationSetARN(ctx context.Context, conn *ssmincidents.Client) (string, error) {
	replicationSets, err := conn.ListReplicationSets(ctx, &ssmincidents.ListReplicationSetsInput{})

	if err != nil {
		return "", err
	}

	if len(replicationSets.ReplicationSetArns) == 0 {
		return "", fmt.Errorf("replication set could not be found")
	}

	// currently only one replication set is supported
	return replicationSets.ReplicationSetArns[0], nil
}
