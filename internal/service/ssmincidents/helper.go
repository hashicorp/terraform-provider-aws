package ssmincidents

// contains misc functions used by multiple files

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
)

func GetReplicationSetARN(context context.Context, client *ssmincidents.Client) (string, error) {
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

func GenerateMapFromList[T, U any](list []T, keyFunction func(T) string, valueFunction func(T) U) map[string]U {
	output := make(map[string]U)

	for _, itr := range list {
		output[keyFunction(itr)] = valueFunction(itr)
	}

	return output
}

func GenerateListFromMap[T, U any](m map[string]T, f func(string, T) U) []U {
	output := make([]U, 0)

	for k, v := range m {
		output = append(output, f(k, v))
	}

	return output
}
