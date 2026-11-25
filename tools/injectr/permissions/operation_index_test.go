// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package permissions

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/tools/injectr/model"
	"github.com/hashicorp/terraform-provider-aws/tools/injectr/serviceref"
)

func testOperationIndex() *OperationIndex {
	boto3 := func(names ...string) []serviceref.SDKMethod {
		var methods []serviceref.SDKMethod
		for _, name := range names {
			methods = append(methods, serviceref.SDKMethod{Name: name, Package: "Boto3"})
		}
		return methods
	}
	actions := func(service string, names ...string) []serviceref.AuthorizedAction {
		var aas []serviceref.AuthorizedAction
		for _, name := range names {
			aas = append(aas, serviceref.AuthorizedAction{Name: name, Service: service})
		}
		return aas
	}

	return NewOperationIndex(serviceref.ServiceRefs{
		"s3": {Name: "s3", Operations: []serviceref.Operation{
			{
				Name:              "CopyObject",
				AuthorizedActions: append(actions("s3", "PutObject", "GetObject", "PutObject"), actions("s3-object-lambda", "PutObject")...),
				SDK:               boto3("s3"),
			},
			{Name: "ListDirectoryBuckets", SDK: boto3("s3")},
		}},
		"elasticloadbalancing": {Name: "elasticloadbalancing", Operations: []serviceref.Operation{
			{Name: "CreateLoadBalancer", AuthorizedActions: actions("elasticloadbalancing", "CreateLoadBalancer"), SDK: boto3("elb", "elbv2")},
		}},
		"sso": {Name: "sso", Operations: []serviceref.Operation{
			{Name: "CreatePermissionSet", AuthorizedActions: actions("sso", "CreatePermissionSet"), SDK: boto3("sso-admin")},
		}},
		"cognito-idp": {Name: "cognito-idp", Operations: []serviceref.Operation{
			{Name: "CreateUserPool", AuthorizedActions: actions("cognito-idp", "CreateUserPool"), SDK: boto3("cognito-idp")},
		}},
		"other": {Name: "other", Operations: []serviceref.Operation{
			{Name: "DoThing", AuthorizedActions: actions("other", "DoThing"), SDK: []serviceref.SDKMethod{{Name: "other", Package: "NotBoto3"}}},
		}},
		"qldb": {Name: "qldb"},
	})
}

func TestOperationIndexResolve(t *testing.T) {
	t.Parallel()

	idx := testOperationIndex()

	testCases := map[string]struct {
		sdkCall    string
		want       []string
		wantReason model.UnresolvedReason
	}{
		"same package and client name": {
			sdkCall: "s3.CopyObject",
			want:    []string{"s3-object-lambda:PutObject", "s3:GetObject", "s3:PutObject"},
		},
		"aliased package": {
			sdkCall: "elasticloadbalancingv2.CreateLoadBalancer",
			want:    []string{"elasticloadbalancing:CreateLoadBalancer"},
		},
		"second client for the same operation": {
			sdkCall: "elasticloadbalancing.CreateLoadBalancer",
			want:    []string{"elasticloadbalancing:CreateLoadBalancer"},
		},
		"dashed client name": {
			sdkCall: "ssoadmin.CreatePermissionSet",
			want:    []string{"sso:CreatePermissionSet"},
		},
		"aliased package to dashed client name": {
			sdkCall: "cognitoidentityprovider.CreateUserPool",
			want:    []string{"cognito-idp:CreateUserPool"},
		},
		"no IAM actions": {
			sdkCall:    "s3.ListDirectoryBuckets",
			wantReason: reasonNoIAMActions,
		},
		"operation not found": {
			sdkCall:    "s3.NoSuchOperation",
			wantReason: reasonOperationNotFound,
		},
		"service without operations": {
			sdkCall:    "qldb.CreateLedger",
			wantReason: reasonServiceHasNoOperations,
		},
		"unknown service": {
			sdkCall:    "evidently.GetProject",
			wantReason: reasonServiceHasNoOperations,
		},
		"non-Boto3 SDK ignored": {
			sdkCall:    "other.DoThing",
			wantReason: reasonServiceHasNoOperations,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, reason := idx.Resolve(tc.sdkCall)
			if reason != tc.wantReason {
				t.Errorf("reason = %q, want %q", reason, tc.wantReason)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("permissions = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestBoto3ClientAliasesExist catches aliases made stale by service reference
// data changes.
func TestBoto3ClientAliasesExist(t *testing.T) {
	t.Parallel()
	skipWithoutRealData(t)

	refs, err := serviceref.LoadServiceRefs(realServiceDataDir)
	if err != nil {
		t.Fatal(err)
	}
	idx := NewOperationIndex(refs)

	for goSDKPackage, alias := range boto3ClientAliases {
		if _, ok := idx.clients[normalizeClient(alias)]; !ok {
			t.Errorf("alias %s -> %s: no operations for client %q", goSDKPackage, alias, alias)
		}
	}
}

func BenchmarkNewOperationIndex(b *testing.B) {
	skipWithoutRealData(b)

	refs, err := serviceref.LoadServiceRefs(realServiceDataDir)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		NewOperationIndex(refs)
	}
}
