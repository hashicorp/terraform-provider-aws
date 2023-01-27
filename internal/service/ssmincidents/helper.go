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

func generateMapFromList[T, U any](list []T, keyFunction func(T) string, valueFunction func(T) U) map[string]U {
	output := make(map[string]U)

	for _, item := range list {
		output[keyFunction(item)] = valueFunction(item)
	}

	return output
}

func generateListFromMap[T, U any](m map[string]T, f func(string, T) U) []U {
	output := make([]U, len(m))

	i := 0
	for k, v := range m {
		output[i] = f(k, v)
		i++
	}

	return output
}
