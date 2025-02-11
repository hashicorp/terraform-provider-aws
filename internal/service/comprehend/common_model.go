// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package comprehend

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/comprehend/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type safeMutex struct {
	locked bool
	mutex  sync.Mutex
}

func (m *safeMutex) Lock() {
	m.mutex.Lock()
	m.locked = true
}

func (m *safeMutex) Unlock() {
	if m.locked {
		m.locked = false
		m.mutex.Unlock()
	}
}

var modelVPCENILock safeMutex

func findNetworkInterfaces(ctx context.Context, conn *ec2.Client, securityGroups []string, subnets []string) ([]ec2types.NetworkInterface, error) {
	networkInterfaces, err := tfec2.FindNetworkInterfaces(ctx, conn, &ec2.DescribeNetworkInterfacesInput{
		Filters: []ec2types.Filter{
			tfec2.NewFilter("group-id", securityGroups),
			tfec2.NewFilter("subnet-id", subnets),
		},
	})
	if err != nil {
		return []ec2types.NetworkInterface{}, err
	}

	comprehendENIs := make([]ec2types.NetworkInterface, 0, len(networkInterfaces))
	for _, v := range networkInterfaces {
		if strings.HasSuffix(aws.ToString(v.RequesterId), ":Comprehend") {
			comprehendENIs = append(comprehendENIs, v)
		}
	}

	return comprehendENIs, nil
}

func waitNetworkInterfaceCreated(ctx context.Context, conn *ec2.Client, initialENIIds map[string]bool, securityGroups []string, subnets []string, timeout time.Duration) (*ec2types.NetworkInterface, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{},
		Target:     enum.Slice(ec2types.NetworkInterfaceStatusInUse),
		Refresh:    statusNetworkInterfaces(ctx, conn, initialENIIds, securityGroups, subnets),
		Delay:      4 * time.Minute,
		MinTimeout: 10 * time.Second,
		Timeout:    timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(ec2types.NetworkInterface); ok {
		return &output, err
	}

	return nil, err
}

func statusNetworkInterfaces(ctx context.Context, conn *ec2.Client, initialENIs map[string]bool, securityGroups []string, subnets []string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findNetworkInterfaces(ctx, conn, securityGroups, subnets)
		if err != nil {
			return nil, "", err
		}

		var added ec2types.NetworkInterface
		for _, v := range out {
			if _, ok := initialENIs[aws.ToString(v.NetworkInterfaceId)]; !ok {
				added = v
				break
			}
		}

		if added.NetworkInterfaceId == nil {
			return nil, "", nil
		}

		return added, string(added.Status), nil
	}
}

type resourceGetter interface {
	Get(key string) any
}

func flattenVPCConfig(apiObject *types.VpcConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrSecurityGroupIDs: flex.FlattenStringValueSet(apiObject.SecurityGroupIds),
		names.AttrSubnets:          flex.FlattenStringValueSet(apiObject.Subnets),
	}

	return []interface{}{m}
}

func expandVPCConfig(tfList []interface{}) *types.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.VpcConfig{
		SecurityGroupIds: flex.ExpandStringValueSet(tfMap[names.AttrSecurityGroupIDs].(*schema.Set)),
		Subnets:          flex.ExpandStringValueSet(tfMap[names.AttrSubnets].(*schema.Set)),
	}

	return a
}

func flattenAugmentedManifests(apiObjects []types.AugmentedManifestsListItem) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenAugmentedManifestsListItem(&apiObject))
	}

	return l
}

func flattenAugmentedManifestsListItem(apiObject *types.AugmentedManifestsListItem) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"attribute_names": flex.FlattenStringValueList(apiObject.AttributeNames),
		"s3_uri":          aws.ToString(apiObject.S3Uri),
		"document_type":   apiObject.DocumentType,
		"split":           apiObject.Split,
	}

	if v := apiObject.AnnotationDataS3Uri; v != nil {
		m["annotation_data_s3_uri"] = aws.ToString(v)
	}

	if v := apiObject.SourceDocumentsS3Uri; v != nil {
		m["source_documents_s3_uri"] = aws.ToString(v)
	}

	return m
}

func expandAugmentedManifests(tfSet *schema.Set) []types.AugmentedManifestsListItem {
	if tfSet.Len() == 0 {
		return nil
	}

	var s []types.AugmentedManifestsListItem

	for _, r := range tfSet.List() {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandAugmentedManifestsListItem(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandAugmentedManifestsListItem(tfMap map[string]interface{}) *types.AugmentedManifestsListItem {
	if tfMap == nil {
		return nil
	}

	a := &types.AugmentedManifestsListItem{
		AttributeNames: flex.ExpandStringValueList(tfMap["attribute_names"].([]interface{})),
		S3Uri:          aws.String(tfMap["s3_uri"].(string)),
		DocumentType:   types.AugmentedManifestsDocumentTypeFormat(tfMap["document_type"].(string)),
		Split:          types.Split(tfMap["split"].(string)),
	}

	if v, ok := tfMap["annotation_data_s3_uri"].(string); ok && v != "" {
		a.AnnotationDataS3Uri = aws.String(v)
	}

	if v, ok := tfMap["source_documents_s3_uri"].(string); ok && v != "" {
		a.SourceDocumentsS3Uri = aws.String(v)
	}

	return a
}
