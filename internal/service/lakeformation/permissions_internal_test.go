package lakeformation

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
)

func TestNormalizeGrantOptionPermissions_FiltersSuperUser(t *testing.T) {
	t.Parallel()

	in := []awstypes.Permission{
		awstypes.PermissionSuperUser,
		awstypes.PermissionAll,
		awstypes.PermissionDescribe,
	}

	out := normalizeGrantOptionPermissions(in)

	if len(out) != 2 {
		t.Fatalf("expected 2 permissions after filtering, got %d: %#v", len(out), out)
	}
	for _, p := range out {
		if p == awstypes.PermissionSuperUser {
			t.Fatalf("did not expect %q in output: %#v", awstypes.PermissionSuperUser, out)
		}
	}
}

func TestNormalizeGrantOptionPermissions_EmptyInput(t *testing.T) {
	t.Parallel()

	if got := normalizeGrantOptionPermissions(nil); got != nil {
		t.Fatalf("expected nil output for nil input, got %#v", got)
	}

	if got := normalizeGrantOptionPermissions([]awstypes.Permission{}); got != nil {
		t.Fatalf("expected nil output for empty input, got %#v", got)
	}
}

func TestFlattenGrantPermissions_FiltersSuperUser(t *testing.T) {
	t.Parallel()

	apiObjects := []awstypes.PrincipalResourcePermissions{
		{
			PermissionsWithGrantOption: []awstypes.Permission{
				awstypes.PermissionDescribe,
				awstypes.PermissionSuperUser,
				awstypes.PermissionAll,
			},
		},
	}

	got := flattenGrantPermissions(apiObjects)

	for _, p := range got {
		if p == string(awstypes.PermissionSuperUser) {
			t.Fatalf("did not expect %q in output: %#v", awstypes.PermissionSuperUser, got)
		}
	}
}
