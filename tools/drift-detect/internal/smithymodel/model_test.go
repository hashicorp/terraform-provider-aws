// Copyright IBM Corp. 2026, 2026
// SPDX-License-Identifier: MPL-2.0

package smithymodel_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/smithymodel"
)

const (
	fixturesRoot = "../../testdata/smithy"
	ampModel     = fixturesRoot + "/models/amp/service/2020-08-01/amp-2020-08-01.json"
	sqsModel     = fixturesRoot + "/models/sqs/service/2012-11-05/sqs-2012-11-05.json"
	snsModel     = fixturesRoot + "/models/sns/service/2010-03-31/sns-2010-03-31.json"
)

func fixtureBaseURL(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.FileServer(http.Dir(fixturesRoot)))
	t.Cleanup(server.Close)

	return server.URL
}

// ---------------------------------------------------------------------------
// LoadFile basics
// ---------------------------------------------------------------------------

func TestLoadFile_AMP_NoError(t *testing.T) {
	t.Parallel()
	m, err := smithymodel.LoadFile(ampModel)
	if err != nil {
		t.Fatalf("LoadFile(amp): %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil Model")
	}
}

func TestLoadFile_SQS_NoError(t *testing.T) {
	t.Parallel()
	_, err := smithymodel.LoadFile(sqsModel)
	if err != nil {
		t.Fatalf("LoadFile(sqs): %v", err)
	}
}

func TestLoadFile_SNS_NoError(t *testing.T) {
	t.Parallel()
	_, err := smithymodel.LoadFile(snsModel)
	if err != nil {
		t.Fatalf("LoadFile(sns): %v", err)
	}
}

func TestLoadURL_AMP_NoError(t *testing.T) {
	t.Parallel()

	_, err := smithymodel.LoadURL(fixtureBaseURL(t) + "/models/amp/service/2020-08-01/amp-2020-08-01.json")
	if err != nil {
		t.Fatalf("LoadURL(amp): %v", err)
	}
}

// ---------------------------------------------------------------------------
// AMP — operation → typed structure members
// ---------------------------------------------------------------------------

func TestAMP_CreateWorkspaceOperation(t *testing.T) {
	t.Parallel()
	m, _ := smithymodel.LoadFile(ampModel)

	op := m.Shape("com.amazonaws.amp#CreateWorkspace")
	if op == nil {
		t.Fatal("CreateWorkspace operation not found")
	}
	if op.Kind != smithymodel.KindOperation {
		t.Errorf("Kind = %q, want operation", op.Kind)
	}
	if op.InputTarget != "com.amazonaws.amp#CreateWorkspaceRequest" {
		t.Errorf("InputTarget = %q", op.InputTarget)
	}
}

func TestAMP_CreateWorkspaceRequest_Members(t *testing.T) {
	t.Parallel()
	m, _ := smithymodel.LoadFile(ampModel)

	req := m.Shape("com.amazonaws.amp#CreateWorkspaceRequest")
	if req == nil {
		t.Fatal("CreateWorkspaceRequest not found")
	}
	if req.Kind != smithymodel.KindStructure {
		t.Errorf("Kind = %q, want structure", req.Kind)
	}

	// alias, clientToken, tags, kmsKeyArn — all optional
	for _, name := range []string{"alias", "clientToken", "tags", "kmsKeyArn"} {
		if _, ok := req.Members[name]; !ok {
			t.Errorf("member %q not found", name)
		}
	}
}

func TestAMP_ClientToken_HasIdempotencyTokenTrait(t *testing.T) {
	t.Parallel()
	m, _ := smithymodel.LoadFile(ampModel)

	req := m.Shape("com.amazonaws.amp#CreateWorkspaceRequest")
	if req == nil {
		t.Fatal("CreateWorkspaceRequest not found")
	}
	ct, ok := req.Members["clientToken"]
	if !ok {
		t.Fatal("clientToken member not found")
	}
	if !ct.Traits.IdempotencyToken {
		t.Error("clientToken: IdempotencyToken trait not detected")
	}
	if !ct.Traits.IsSuppressible() {
		t.Error("clientToken: IsSuppressible() = false, want true")
	}
}

func TestAMP_WorkspaceId_HasHTTPLabelTrait(t *testing.T) {
	t.Parallel()
	m, _ := smithymodel.LoadFile(ampModel)

	// DescribeWorkspaceRequest has workspaceId as an httpLabel
	req := m.Shape("com.amazonaws.amp#DescribeWorkspaceRequest")
	if req == nil {
		t.Fatal("DescribeWorkspaceRequest not found")
	}
	wid, ok := req.Members["workspaceId"]
	if !ok {
		t.Fatal("workspaceId member not found")
	}
	if !wid.Traits.HTTPLabel {
		t.Error("workspaceId: HTTPLabel trait not detected")
	}
	if !wid.Traits.IsSuppressible() {
		t.Error("workspaceId: IsSuppressible() = false, want true")
	}
}

func TestAMP_ResolveToKind_WorkspaceAlias(t *testing.T) {
	t.Parallel()
	m, _ := smithymodel.LoadFile(ampModel)

	// WorkspaceAlias is a string typedef
	got := m.ResolveToKind("com.amazonaws.amp#WorkspaceAlias")
	if got != smithymodel.KindString {
		t.Errorf("ResolveToKind(WorkspaceAlias) = %q, want string", got)
	}
}

func TestAMP_ResolveToKind_BuiltinString(t *testing.T) {
	t.Parallel()
	m, _ := smithymodel.LoadFile(ampModel)

	got := m.ResolveToKind("smithy.api#String")
	if got != smithymodel.KindString {
		t.Errorf("ResolveToKind(smithy.api#String) = %q, want string", got)
	}
}

func TestAMP_WorkspaceResourceShapeTargets(t *testing.T) {
	t.Parallel()

	m, _ := smithymodel.LoadFile(ampModel)

	r := m.Shape("com.amazonaws.amp#WorkspaceResourcePolicy")
	if r == nil {
		t.Fatal("WorkspaceResourcePolicy resource not found")
	}
	if r.Kind != smithymodel.KindResource {
		t.Errorf("Kind = %q, want resource", r.Kind)
	}
	if r.PutTarget != "com.amazonaws.amp#PutResourcePolicy" {
		t.Errorf("PutTarget = %q", r.PutTarget)
	}
	if r.ReadTarget != "com.amazonaws.amp#DescribeResourcePolicy" {
		t.Errorf("ReadTarget = %q", r.ReadTarget)
	}
	if r.DeleteTarget != "com.amazonaws.amp#DeleteResourcePolicy" {
		t.Errorf("DeleteTarget = %q", r.DeleteTarget)
	}
	if got := r.Identifiers["workspaceId"]; got != "com.amazonaws.amp#WorkspaceId" {
		t.Errorf("Identifiers[workspaceId] = %q", got)
	}
	if len(r.Operations) != 1 {
		t.Fatalf("Operations len = %d, want 1", len(r.Operations))
	}
	if r.Operations[0] != "com.amazonaws.amp#CreateWorkspace" {
		t.Errorf("Operations[0] = %q, want %q", r.Operations[0], "com.amazonaws.amp#CreateWorkspace")
	}
}

// ---------------------------------------------------------------------------
// SQS — QueueAttributeName enum
// ---------------------------------------------------------------------------

func TestSQS_QueueAttributeNameEnum(t *testing.T) {
	t.Parallel()
	m, _ := smithymodel.LoadFile(sqsModel)

	shape := m.Shape("com.amazonaws.sqs#QueueAttributeName")
	if shape == nil {
		t.Fatal("QueueAttributeName not found")
	}
	if shape.Kind != smithymodel.KindEnum {
		t.Errorf("Kind = %q, want enum", shape.Kind)
	}

	// Spot-check a few expected enum member values
	wantValues := map[string]string{
		"DelaySeconds":         "DelaySeconds",
		"VisibilityTimeout":    "VisibilityTimeout",
		"FifoQueue":            "FifoQueue",
		"KmsMasterKeyId":       "KmsMasterKeyId",
		"RedrivePolicy":        "RedrivePolicy",
		"SqsManagedSseEnabled": "SqsManagedSseEnabled",
	}
	for memberName, wantEnumVal := range wantValues {
		mem, ok := shape.Members[memberName]
		if !ok {
			t.Errorf("enum member %q not found", memberName)
			continue
		}
		if mem.EnumValue != wantEnumVal {
			t.Errorf("member %q: EnumValue = %q, want %q",
				memberName, mem.EnumValue, wantEnumVal)
		}
	}
}

func TestSQS_CreateQueueRequest(t *testing.T) {
	t.Parallel()
	m, _ := smithymodel.LoadFile(sqsModel)

	req := m.Shape("com.amazonaws.sqs#CreateQueueRequest")
	if req == nil {
		t.Fatal("CreateQueueRequest not found")
	}
	// QueueName and Attributes should be present
	for _, name := range []string{"QueueName", "Attributes"} {
		if _, ok := req.Members[name]; !ok {
			t.Errorf("member %q not found in CreateQueueRequest", name)
		}
	}
}

// ---------------------------------------------------------------------------
// SNS — CreateTopicInput with untyped Attributes map
// ---------------------------------------------------------------------------

func TestSNS_CreateTopicInput(t *testing.T) {
	t.Parallel()
	m, _ := smithymodel.LoadFile(snsModel)

	req := m.Shape("com.amazonaws.sns#CreateTopicInput")
	if req == nil {
		t.Fatal("CreateTopicInput not found")
	}

	// Name is required; Attributes and Tags are optional
	nameMember, ok := req.Members["Name"]
	if !ok {
		t.Fatal("Name member not found")
	}
	if !nameMember.Traits.Required {
		t.Error("Name member should be required")
	}

	attrMember, ok := req.Members["Attributes"]
	if !ok {
		t.Fatal("Attributes member not found")
	}
	if attrMember.Traits.Required {
		t.Error("Attributes member should NOT be required")
	}
}

func TestSNS_TopicAttributesMap_IsUntyped(t *testing.T) {
	t.Parallel()
	m, _ := smithymodel.LoadFile(snsModel)

	attrMap := m.Shape("com.amazonaws.sns#TopicAttributesMap")
	if attrMap == nil {
		t.Fatal("TopicAttributesMap not found")
	}
	if attrMap.Kind != smithymodel.KindMap {
		t.Errorf("Kind = %q, want map", attrMap.Kind)
	}
	// Both key and value resolve to string — confirming "untyped" pattern
	keyKind := m.ResolveToKind(attrMap.KeyTarget)
	valKind := m.ResolveToKind(attrMap.ValueTarget)
	if keyKind != smithymodel.KindString {
		t.Errorf("TopicAttributesMap key kind = %q, want string", keyKind)
	}
	if valKind != smithymodel.KindString {
		t.Errorf("TopicAttributesMap value kind = %q, want string", valKind)
	}
}
