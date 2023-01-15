package ssmincidents

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
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

func getReplicationSetArn(ctx context.Context, conn *ssmincidents.Client) (string, error) {
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

func expandRegions(regions []interface{}) map[string]types.RegionMapInputValue {
	if len(regions) == 0 {
		return nil
	}

	ret := make(map[string]types.RegionMapInputValue)
	for _, region := range regions {
		regionData := region.(map[string]interface{})

		input := types.RegionMapInputValue{}

		if kmsKey := regionData["kms_key_arn"].(string); kmsKey != "DefaultKey" {
			input.SseKmsKeyId = aws.String(kmsKey)
		}

		ret[regionData["name"].(string)] = input
	}

	return ret
}

func flattenRegions(regions map[string]types.RegionInfo) []interface{} {
	if len(regions) == 0 {
		return nil
	}

	ret := make([]interface{}, 0)
	for name, v := range regions {
		region := make(map[string]interface{})

		region["name"] = name
		region["status"] = v.Status
		region["status_update_time"] = aws.ToTime(v.StatusUpdateDateTime).String()
		region["kms_key_arn"] = aws.ToString(v.SseKmsKeyId)
		region["status_message"] = aws.ToString(v.StatusMessage)

		ret = append(ret, region)
	}

	return ret
}
