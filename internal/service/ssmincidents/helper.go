package ssmincidents

// contains misc functions used by multiple files

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
)

func getReplicationSetARN(context context.Context, client *ssmincidents.Client) (string, error) {
	replicationSets, err := client.ListReplicationSets(context, &ssmincidents.ListReplicationSetsInput{})

	if err != nil {
		return "", err
	}

	if len(replicationSets.ReplicationSetArns) == 0 {
		return "", fmt.Errorf("replication set could not be found")
	}

	// currently only one replication set is supported
	return replicationSets.ReplicationSetArns[0], nil
}
