// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package docs

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/tools/injectr/model"
)

func TestResourcePermissions(t *testing.T) {
	t.Parallel()

	ref := func(perms ...string) *model.FuncRef { return &model.FuncRef{IAMPermissions: perms} }
	r := &model.Resource{
		Create: ref("svc:Create", "svc:Describe"),
		Read:   ref("svc:Describe"),
		Update: ref("svc:Update"),
		Delete: ref("svc:Delete"),
		CustomizeDiff: &model.FuncRef{Functions: []*model.FuncRef{
			ref("svc:Validate"),
		}},
		Import: ref("svc:Import"),
		Tags:   &model.ResourceTags{List: ref("svc:ListTags"), Update: ref("svc:TagResource")},
	}

	got := resourcePermissions(r)

	wantRead := []string{"svc:Describe", "svc:Import", "svc:ListTags", "svc:Validate"}
	if !slices.Equal(got.Read, wantRead) {
		t.Errorf("Read = %v, want %v", got.Read, wantRead)
	}
	wantWrite := []string{"svc:Create", "svc:Delete", "svc:Describe", "svc:Import", "svc:ListTags", "svc:TagResource", "svc:Update", "svc:Validate"}
	if !slices.Equal(got.Write, wantWrite) {
		t.Errorf("Write = %v, want %v", got.Write, wantWrite)
	}
}

func TestResourcePermissionsEmptyRead(t *testing.T) {
	t.Parallel()

	got := resourcePermissions(&model.Resource{Create: &model.FuncRef{IAMPermissions: []string{"svc:Create"}}})

	if len(got.Read) != 0 {
		t.Errorf("Read = %v, want none", got.Read)
	}
	if want := []string{"svc:Create"}; !slices.Equal(got.Write, want) {
		t.Errorf("Write = %v, want %v", got.Write, want)
	}
}

func TestDataSourcePermissions(t *testing.T) {
	t.Parallel()

	d := &model.DataSource{
		Read: &model.FuncRef{IAMPermissions: []string{"svc:Get", "svc:Describe"}},
		Tags: &model.ResourceTags{List: &model.FuncRef{IAMPermissions: []string{"svc:Describe", "svc:ListTags"}}},
	}

	got := dataSourcePermissions(d)

	if want := []string{"svc:Describe", "svc:Get", "svc:ListTags"}; !slices.Equal(got.Read, want) {
		t.Errorf("Read = %v, want %v", got.Read, want)
	}
	if got.Write != nil {
		t.Errorf("Write = %v, want nil", got.Write)
	}
}
