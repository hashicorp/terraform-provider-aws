package ssmincidents

// contains misc functions used by multiple files

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// generates a pseudo-random unique temporary client token
func GenerateClientToken() string {
	rand.Seed(time.Now().UnixNano())

	n := 30
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func CastInterfaceMapToStringMap(in map[string]interface{}) map[string]string {
	result := map[string]string{}

	for k, v := range in {
		result[k] = v.(string)
	}
	return result
}

func Trim(s string) string {
	return strings.Trim(s, "\"")
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
