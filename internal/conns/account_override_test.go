// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
)

const (
	testProviderAccountID = "111111111111"
	testSpokeAccountID    = "222222222222"
	testOtherAccountID    = "333333333333"
)

func testAccountClient(t *testing.T, blocks ...AccountRole) *AWSClient {
	t.Helper()

	roles, err := accountRoles(blocks)
	if err != nil {
		t.Fatal(err)
	}
	c := &AWSClient{
		accountID:     testProviderAccountID,
		accountRoles:  roles,
		awsConfig:     &aws.Config{Region: "eu-central-1"}, //lintignore:AWSAT003
		awsbaseConfig: &awsbase.Config{},
		clients:       make(map[string]map[string]any),
	}
	for _, p := range endpoints.DefaultPartitions() {
		if p.ID() == endpoints.AwsPartitionID {
			c.partition = p
		}
	}

	return c
}

func testAccountContext(accountID string) context.Context {
	return WithOverrideAccount(NewResourceContext(context.Background(), "S3", "Bucket", "aws_s3_bucket", ""), accountID)
}

func TestIndexAccountRoles(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		blocks  []AccountRole
		wantErr bool
	}{
		"none":                 {},
		"default":              {blocks: []AccountRole{{RoleName: "terraform"}}},
		"default and override": {blocks: []AccountRole{{RoleName: "terraform"}, {RoleName: "legacy", AccountIDs: []string{testSpokeAccountID}}}},
		"two defaults":         {blocks: []AccountRole{{RoleName: "a"}, {RoleName: "b"}}, wantErr: true},
		"account in two blocks": {blocks: []AccountRole{
			{RoleName: "a", AccountIDs: []string{testSpokeAccountID}},
			{RoleName: "b", AccountIDs: []string{testOtherAccountID, testSpokeAccountID}},
		}, wantErr: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := accountRoles(tc.blocks); (err != nil) != tc.wantErr {
				t.Errorf("err = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}

func TestClientAccountRole(t *testing.T) {
	t.Parallel()

	c := testAccountClient(t, AccountRole{RoleName: "terraform"}, AccountRole{RoleName: "legacy", AccountIDs: []string{testOtherAccountID}})

	if got, want := c.accountRole(testSpokeAccountID).RoleName, "terraform"; got != want {
		t.Errorf("role for an account without its own block = %s, want %s", got, want)
	}
	if got, want := c.accountRole(testOtherAccountID).RoleName, "legacy"; got != want {
		t.Errorf("role for an account with its own block = %s, want %s", got, want)
	}
	if got := testAccountClient(t, AccountRole{RoleName: "legacy", AccountIDs: []string{testOtherAccountID}}).accountRole(testSpokeAccountID); got != nil {
		t.Errorf("role without a matching block = %v, want none", got)
	}
}

func TestClientValidateInContextAccount(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		accountID           string
		blocks              []AccountRole
		forbiddenAccountIDs []string
		wantErr             bool
	}{
		"no account":                   {},
		"provider's account, no block": {accountID: testProviderAccountID},
		"other account":                {accountID: testSpokeAccountID, blocks: []AccountRole{{RoleName: "terraform"}}},
		"other account, no block":      {accountID: testSpokeAccountID, wantErr: true},
		"other account, other block":   {accountID: testSpokeAccountID, blocks: []AccountRole{{RoleName: "x", AccountIDs: []string{testOtherAccountID}}}, wantErr: true},
		"forbidden account":            {accountID: testSpokeAccountID, blocks: []AccountRole{{RoleName: "terraform"}}, forbiddenAccountIDs: []string{testSpokeAccountID}, wantErr: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c := testAccountClient(t, tc.blocks...)
			c.awsbaseConfig.ForbiddenAccountIds = tc.forbiddenAccountIDs

			if err := c.ValidateInContextAccount(testAccountContext(tc.accountID)); (err != nil) != tc.wantErr {
				t.Errorf("err = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}

func TestClientAccountIDWithOverride(t *testing.T) {
	t.Parallel()

	c := testAccountClient(t, AccountRole{RoleName: "terraform"})

	if got, want := c.AccountID(testAccountContext("")), testProviderAccountID; got != want {
		t.Errorf("AccountID without an account = %s, want %s", got, want)
	}
	ctx := testAccountContext(testSpokeAccountID)
	if got, want := c.AccountID(ctx), testSpokeAccountID; got != want {
		t.Errorf("AccountID = %s, want %s", got, want)
	}
	if got, want := c.RegionalARN(ctx, "sqs", "q"), "arn:aws:sqs:eu-central-1:222222222222:q"; got != want { //lintignore:AWSAT003,AWSAT005
		t.Errorf("RegionalARN = %q, want %q", got, want)
	}

	// The provider's own account needs no role.
	if cfg, key, err := c.effectiveConfig(testAccountContext(testProviderAccountID)); err != nil || key != "" || cfg != c.awsConfig {
		t.Errorf("effectiveConfig for the provider's account = %v, %q, %v, want the provider's configuration", cfg, key, err)
	}
}
