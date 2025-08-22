// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfrontkeyvaluestore

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore/types"
)

// resourceKeyValuePairEqual determines whether two key value pairs are semantically equal
func resourceKeyValuePairEqual(s1, s2 awstypes.ListKeysResponseListItem) bool {
	return aws.ToString(s1.Key) == aws.ToString(s2.Key) && aws.ToString(s1.Value) == aws.ToString(s2.Value)
}

// resourceKeyValuePairKeyEqual determines whether two key value pairs have an equal key
func resourceKeyValuePairKeyEqual(s1, s2 awstypes.ListKeysResponseListItem) bool {
	return aws.ToString(s1.Key) == aws.ToString(s2.Key)
}
