// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// AccountRole is a provider-level assume_role_in_account block:
// the IAM role the provider assumes, from its own credentials, for resources in other AWS accounts.
type AccountRole struct {
	awsbase.AssumeRole          // RoleARN is not used: the ARN is built from the resource's account and RoleName.
	AccountIDs         []string // The accounts the block applies to. Empty for the block that applies to every other account.
	RoleName           string
}

// accountRoles indexes assume_role_in_account blocks by account ID.
// The block that names no accounts is indexed under the empty string.
func accountRoles(blocks []AccountRole) (map[string]*AccountRole, error) {
	m := make(map[string]*AccountRole, len(blocks))
	for i := range blocks {
		accountIDs := blocks[i].AccountIDs
		if len(accountIDs) == 0 {
			accountIDs = []string{""}
		}
		for _, accountID := range accountIDs {
			if _, ok := m[accountID]; ok {
				if accountID == "" {
					return nil, errors.New("assume_role_in_account: only one block may omit account_ids")
				}
				return nil, fmt.Errorf("assume_role_in_account: account (%s) is in more than one block", accountID)
			}
			m[accountID] = &blocks[i]
		}
	}
	return m, nil
}

// WithOverrideAccount records a per-resource AWS account on the resource's context.
// It must be called after [NewResourceContext].
func WithOverrideAccount(ctx context.Context, accountID string) context.Context {
	inContext, ok := FromContext(ctx)
	if !ok || accountID == inContext.overrideAccountID {
		return ctx
	}
	v := *inContext
	v.overrideAccountID = accountID
	return context.WithValue(ctx, contextKey, &v)
}

// overrideAccountID returns the per-resource account in ctx, or "" if there is none or it is the provider's own account:
// the provider is already there.
func (c *AWSClient) overrideAccountID(ctx context.Context) string {
	if inContext, ok := FromContext(ctx); ok && inContext.overrideAccountID != c.accountID {
		return inContext.overrideAccountID
	}
	return ""
}

// ValidateInContextAccount verifies any per-resource account in ctx: one that the provider configuration allows
// and that one of its assume_role_in_account blocks applies to.
func (c *AWSClient) ValidateInContextAccount(ctx context.Context) error {
	accountID := c.overrideAccountID(ctx)
	if accountID == "" {
		return nil
	}

	if c.awsbaseConfig != nil {
		if err := c.awsbaseConfig.VerifyAccountIDAllowed(accountID); err != nil {
			return fmt.Errorf("account (%s): %w", accountID, err)
		}
	}
	if c.accountRole(accountID) == nil {
		return fmt.Errorf("account (%s) is not the provider's account and no assume_role_in_account block in the provider configuration applies to it", accountID)
	}

	return nil
}

func (c *AWSClient) accountRole(accountID string) *AccountRole {
	if v, ok := c.accountRoles[accountID]; ok {
		return v
	}
	return c.accountRoles[""]
}

// effectiveConfig returns the AWS configuration for the operation in ctx and the key its API clients are cached under.
// The key is empty when the operation is in the provider's own account.
func (c *AWSClient) effectiveConfig(ctx context.Context) (*aws.Config, string, error) {
	accountID := c.overrideAccountID(ctx)
	if accountID == "" {
		return c.awsConfig, "", nil
	}

	c.accountLock.Lock()
	defer c.accountLock.Unlock()

	if cfg, ok := c.accountConfigs[accountID]; ok {
		return cfg, accountID, nil
	}

	role := c.accountRole(accountID)
	if role == nil {
		return nil, "", fmt.Errorf("account (%s): no assume_role_in_account block in the provider configuration applies to it", accountID)
	}

	// The STS client is built by the STS service package so that the provider's
	// endpoint, sts_region and retry configuration apply here too.
	// It uses the provider's own account and Region: that is where the role is assumed from.
	// Passing the configuration explicitly makes the client a non-default one, which is created without AWSClient.lock.
	inContext, _ := FromContext(ctx)
	providerContext := *inContext
	providerContext.overrideAccountID, providerContext.overrideRegion = "", ""
	stsClient, err := c.client[*sts.Client](context.WithValue(ctx, contextKey, &providerContext), names.STS, map[string]any{
		"aws_sdkv2_config": c.awsConfig,
	})
	if err != nil {
		return nil, "", fmt.Errorf("account (%s): creating STS client: %w", accountID, err)
	}

	roleARN := arn.ARN{
		Partition: c.Partition(ctx),
		Service:   "iam",
		AccountID: accountID,
		Resource:  "role/" + strings.TrimPrefix(role.RoleName, "/"),
	}.String()
	cfg := c.awsConfig.Copy()
	cfg.Credentials = aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(stsClient, roleARN, func(opts *stscreds.AssumeRoleOptions) {
		if role.SessionName != "" {
			opts.RoleSessionName = role.SessionName
		}
		if role.Duration != 0 {
			opts.Duration = role.Duration
		}
		if role.ExternalID != "" {
			opts.ExternalID = aws.String(role.ExternalID)
		}
		if role.Policy != "" {
			opts.Policy = aws.String(role.Policy)
		}
		for _, v := range role.PolicyARNs {
			opts.PolicyARNs = append(opts.PolicyARNs, ststypes.PolicyDescriptorType{Arn: aws.String(v)})
		}
		for k, v := range role.Tags {
			opts.Tags = append(opts.Tags, ststypes.Tag{Key: aws.String(k), Value: aws.String(v)})
		}
		opts.TransitiveTagKeys = role.TransitiveTagKeys
		if role.SourceIdentity != "" {
			opts.SourceIdentity = aws.String(role.SourceIdentity)
		}
	}))

	if c.accountConfigs == nil {
		c.accountConfigs = make(map[string]*aws.Config)
	}
	c.accountConfigs[accountID] = &cfg

	return &cfg, accountID, nil
}
