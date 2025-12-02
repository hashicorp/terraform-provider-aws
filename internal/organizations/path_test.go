package organizations

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

// mockOrganizationsClient implements organizationsClient for testing
type mockOrganizationsClient struct {
	ParentsMap     map[string][]types.Parent
	OrganizationID string
}

func (m *mockOrganizationsClient) ListParents(ctx context.Context, input *organizations.ListParentsInput, opts ...func(*organizations.Options)) (*organizations.ListParentsOutput, error) {
	if parents, ok := m.ParentsMap[*input.ChildId]; ok {
		return &organizations.ListParentsOutput{Parents: parents}, nil
	}
	return &organizations.ListParentsOutput{Parents: []types.Parent{}}, nil
}

func (m *mockOrganizationsClient) DescribeOrganization(ctx context.Context, input *organizations.DescribeOrganizationInput, opts ...func(*organizations.Options)) (*organizations.DescribeOrganizationOutput, error) {
	if m.OrganizationID == "" {
		return nil, errors.New("no org")
	}
	return &organizations.DescribeOrganizationOutput{
		Organization: &types.Organization{Id: aws.String(m.OrganizationID)},
	}, nil
}

func TestBuildPrincipalOrgPath_OU(t *testing.T) {
	ctx := context.Background()
	client := &mockOrganizationsClient{
		OrganizationID: "o-org456",
		ParentsMap: map[string][]types.Parent{
			"ou-aaa111": {{Id: aws.String("r-root123"), Type: types.ParentTypeRoot}},
		},
	}

	path, err := BuildPrincipalOrgPath(ctx, client, "ou-aaa111")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "o-org456/r-root123/ou-aaa111/"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestBuildPrincipalOrgPath_Account(t *testing.T) {
	ctx := context.Background()
	client := &mockOrganizationsClient{
		OrganizationID: "o-org456",
		ParentsMap: map[string][]types.Parent{
			"123456789012": {{Id: aws.String("ou-aaa111"), Type: types.ParentTypeOrganizationalUnit}},
			"ou-aaa111":    {{Id: aws.String("r-root123"), Type: types.ParentTypeRoot}},
		},
	}

	path, err := BuildPrincipalOrgPath(ctx, client, "123456789012")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "o-org456/r-root123/ou-aaa111/123456789012/"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestBuildPrincipalOrgPath_NestedOUs(t *testing.T) {
	ctx := context.Background()
	client := &mockOrganizationsClient{
		OrganizationID: "o-org456",
		ParentsMap: map[string][]types.Parent{
			"ou-child":       {{Id: aws.String("ou-parent"), Type: types.ParentTypeOrganizationalUnit}},
			"ou-parent":      {{Id: aws.String("ou-grandparent"), Type: types.ParentTypeOrganizationalUnit}},
			"ou-grandparent": {{Id: aws.String("r-root123"), Type: types.ParentTypeRoot}},
		},
	}

	path, err := BuildPrincipalOrgPath(ctx, client, "ou-child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "o-org456/r-root123/ou-grandparent/ou-parent/ou-child/"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestBuildPrincipalOrgPath_NoParents(t *testing.T) {
	ctx := context.Background()
	client := &mockOrganizationsClient{
		OrganizationID: "o-org456",
		ParentsMap:     map[string][]types.Parent{}, // No parents at all
	}

	_, err := BuildPrincipalOrgPath(ctx, client, "ou-aaa111")
	if err == nil {
		t.Fatal("expected error when no parents found")
	}
}

func TestBuildPrincipalOrgPath_AccountNestedOUs(t *testing.T) {
	ctx := context.Background()
	client := &mockOrganizationsClient{
		OrganizationID: "o-org456",
		ParentsMap: map[string][]types.Parent{
			"123456789012":   {{Id: aws.String("ou-child"), Type: types.ParentTypeOrganizationalUnit}},
			"ou-child":       {{Id: aws.String("ou-parent"), Type: types.ParentTypeOrganizationalUnit}},
			"ou-parent":      {{Id: aws.String("ou-grandparent"), Type: types.ParentTypeOrganizationalUnit}},
			"ou-grandparent": {{Id: aws.String("r-root123"), Type: types.ParentTypeRoot}},
		},
	}

	path, err := BuildPrincipalOrgPath(ctx, client, "123456789012")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "o-org456/r-root123/ou-grandparent/ou-parent/ou-child/123456789012/"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestBuildPrincipalOrgPath_RootAsChild(t *testing.T) {
	ctx := context.Background()
	client := &mockOrganizationsClient{
		OrganizationID: "o-org456",
		ParentsMap: map[string][]types.Parent{
			"r-root123": {}, // root has no parent
		},
	}

	_, err := BuildPrincipalOrgPath(ctx, client, "r-root123")
	if err == nil {
		t.Fatal("expected error when root has no parent")
	}
}
