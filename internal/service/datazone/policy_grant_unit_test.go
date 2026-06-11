// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/google/go-cmp/cmp"
)

func TestFlattenPolicyGrantDetail_CreateDomainUnit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	boolVal := true
	input := &awstypes.PolicyGrantDetailMemberCreateDomainUnit{
		Value: awstypes.CreateDomainUnitPolicyGrantDetail{
			IncludeChildDomainUnits: &boolVal,
		},
	}

	got, diags := flattenPolicyGrantDetail(ctx, input)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	elements, d := got.CreateDomainUnit.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics reading CreateDomainUnit: %s", d)
	}
	if len(elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elements))
	}
	if !elements[0].IncludeChildDomainUnits.ValueBool() {
		t.Error("expected IncludeChildDomainUnits to be true")
	}

	nullElements, d := got.CreateProject.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics reading CreateProject: %s", d)
	}
	if len(nullElements) != 0 {
		t.Errorf("expected CreateProject to be null/empty, got %d elements", len(nullElements))
	}
}

func TestFlattenPolicyGrantDetail_CreateEnvironment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	input := &awstypes.PolicyGrantDetailMemberCreateEnvironment{
		Value: awstypes.Unit{},
	}

	got, diags := flattenPolicyGrantDetail(ctx, input)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	elements, d := got.CreateEnvironment.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %s", d)
	}
	if len(elements) != 1 {
		t.Fatalf("expected 1 element for Unit detail, got %d", len(elements))
	}
}

func TestFlattenPolicyGrantDetail_CreateProjectFromProjectProfile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	boolVal := false
	input := &awstypes.PolicyGrantDetailMemberCreateProjectFromProjectProfile{
		Value: awstypes.CreateProjectFromProjectProfilePolicyGrantDetail{
			IncludeChildDomainUnits: &boolVal,
			ProjectProfiles:         []string{"profile-1", "profile-2"},
		},
	}

	got, diags := flattenPolicyGrantDetail(ctx, input)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	elements, d := got.CreateProjectFromProjectProfile.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %s", d)
	}
	if len(elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elements))
	}
	if elements[0].IncludeChildDomainUnits.ValueBool() {
		t.Error("expected IncludeChildDomainUnits to be false")
	}
	if elements[0].ProjectProfiles.IsNull() {
		t.Error("expected ProjectProfiles to be non-null")
	}

	var profiles []string
	d = elements[0].ProjectProfiles.ElementsAs(ctx, &profiles, false)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics reading ProjectProfiles: %s", d)
	}
	if diff := cmp.Diff([]string{"profile-1", "profile-2"}, profiles); diff != "" {
		t.Errorf("unexpected ProjectProfiles (-want +got):\n%s", diff)
	}
}

func TestFlattenPolicyGrantPrincipal_User_AllUsersGrantFilter(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	input := &awstypes.PolicyGrantPrincipalMemberUser{
		Value: &awstypes.UserPolicyGrantPrincipalMemberAllUsersGrantFilter{
			Value: awstypes.AllUsersGrantFilter{},
		},
	}

	got, diags := flattenPolicyGrantPrincipal(ctx, input)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	elements, d := got.User.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %s", d)
	}
	if len(elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elements))
	}
	if !elements[0].UserIdentifier.IsNull() {
		t.Error("expected UserIdentifier to be null")
	}
	filterElements, fd := elements[0].AllUsersGrantFilter.ToSlice(ctx)
	if fd.HasError() {
		t.Fatalf("unexpected diagnostics: %s", fd)
	}
	if len(filterElements) != 1 {
		t.Errorf("expected AllUsersGrantFilter to have 1 element, got %d", len(filterElements))
	}
}

func TestFlattenPolicyGrantPrincipal_User_UserIdentifier(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	input := &awstypes.PolicyGrantPrincipalMemberUser{
		Value: &awstypes.UserPolicyGrantPrincipalMemberUserIdentifier{
			Value: "user-123",
		},
	}

	got, diags := flattenPolicyGrantPrincipal(ctx, input)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	elements, d := got.User.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %s", d)
	}
	if len(elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elements))
	}
	if got := elements[0].UserIdentifier.ValueString(); got != "user-123" {
		t.Errorf("expected UserIdentifier %q, got %q", "user-123", got)
	}
}

func TestFlattenPolicyGrantPrincipal_Project(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	input := &awstypes.PolicyGrantPrincipalMemberProject{
		Value: awstypes.ProjectPolicyGrantPrincipal{
			ProjectDesignation: awstypes.ProjectDesignationOwner,
			ProjectIdentifier:  aws.String("proj-abc"),
		},
	}

	got, diags := flattenPolicyGrantPrincipal(ctx, input)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	elements, d := got.Project.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %s", d)
	}
	if len(elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elements))
	}
	if got := elements[0].ProjectDesignation.ValueEnum(); got != awstypes.ProjectDesignationOwner {
		t.Errorf("expected ProjectDesignation %q, got %q", awstypes.ProjectDesignationOwner, got)
	}
	if got := elements[0].ProjectIdentifier.ValueString(); got != "proj-abc" {
		t.Errorf("expected ProjectIdentifier %q, got %q", "proj-abc", got)
	}
}

func TestFlattenPolicyGrantPrincipal_DomainUnit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	input := &awstypes.PolicyGrantPrincipalMemberDomainUnit{
		Value: awstypes.DomainUnitPolicyGrantPrincipal{
			DomainUnitDesignation: awstypes.DomainUnitDesignationOwner,
			DomainUnitIdentifier:  aws.String("du-xyz"),
			DomainUnitGrantFilter: &awstypes.DomainUnitGrantFilterMemberAllDomainUnitsGrantFilter{
				Value: awstypes.AllDomainUnitsGrantFilter{},
			},
		},
	}

	got, diags := flattenPolicyGrantPrincipal(ctx, input)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	elements, d := got.DomainUnit.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %s", d)
	}
	if len(elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elements))
	}
	if got := elements[0].DomainUnitDesignation.ValueEnum(); got != awstypes.DomainUnitDesignationOwner {
		t.Errorf("expected DomainUnitDesignation %q, got %q", awstypes.DomainUnitDesignationOwner, got)
	}
	if got := elements[0].DomainUnitIdentifier.ValueString(); got != "du-xyz" {
		t.Errorf("expected DomainUnitIdentifier %q, got %q", "du-xyz", got)
	}
	filterElements, fd := elements[0].AllDomainUnitsGrantFilter.ToSlice(ctx)
	if fd.HasError() {
		t.Fatalf("unexpected diagnostics: %s", fd)
	}
	if len(filterElements) != 1 {
		t.Errorf("expected AllDomainUnitsGrantFilter to have 1 element, got %d", len(filterElements))
	}
}

func TestFlattenPolicyGrantPrincipal_Group(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	input := &awstypes.PolicyGrantPrincipalMemberGroup{
		Value: &awstypes.GroupPolicyGrantPrincipalMemberGroupIdentifier{
			Value: "group-456",
		},
	}

	got, diags := flattenPolicyGrantPrincipal(ctx, input)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags)
	}

	elements, d := got.Group.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("unexpected diagnostics: %s", d)
	}
	if len(elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elements))
	}
	if got := elements[0].GroupIdentifier.ValueString(); got != "group-456" {
		t.Errorf("expected GroupIdentifier %q, got %q", "group-456", got)
	}
}

func TestPolicyGrantImportID_Parse_Valid(t *testing.T) {
	t.Parallel()

	id := "dzd-abc123,DOMAIN_UNIT,du-xyz,CREATE_DOMAIN_UNIT,grantAbcDe1"
	parser := policyGrantImportID{}

	rawID, result, err := parser.Parse(id)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if rawID != id {
		t.Errorf("expected rawID %q, got %q", id, rawID)
	}

	expected := map[string]any{
		"domain_identifier": "dzd-abc123",
		"entity_type":       "DOMAIN_UNIT",
		"entity_identifier": "du-xyz",
		"policy_type":       "CREATE_DOMAIN_UNIT",
		"grant_id":          "grantAbcDe1",
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("unexpected parse result (-want +got):\n%s", diff)
	}
}

func TestPolicyGrantImportID_Parse_TooFewParts(t *testing.T) {
	t.Parallel()

	parser := policyGrantImportID{}
	_, _, err := parser.Parse("dzd-abc,DOMAIN_UNIT,du-xyz")
	if err == nil {
		t.Fatal("expected error for too few parts, got nil")
	}
}

func TestPolicyGrantImportID_Parse_TooManyParts(t *testing.T) {
	t.Parallel()

	parser := policyGrantImportID{}
	_, _, err := parser.Parse("dzd-abc,DOMAIN_UNIT,du-xyz,CREATE_DOMAIN_UNIT,grant123,extra")
	if err == nil {
		t.Fatal("expected error for too many parts, got nil")
	}
}
