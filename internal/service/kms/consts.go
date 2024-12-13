// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	aliasNamePrefix = "alias/"
	cmkAliasPrefix  = aliasNamePrefix + "aws/"
)

const (
	policyNameDefault = "default"
)

func customKeyStoreType_Values() []string {
	return enum.Values[awstypes.CustomKeyStoreType]()
}

func proxyConnectivityType_Values() []string {
	return enum.Values[awstypes.XksProxyConnectivityType]()
}
