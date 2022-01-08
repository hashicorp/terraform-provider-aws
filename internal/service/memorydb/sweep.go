//go:build sweep
// +build sweep

package memorydb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aws_memorydb_acl", &resource.Sweeper{
		Name: "aws_memorydb_acl",
		F:    sweepACLs,
		Dependencies: []string{
			"aws_memorydb_cluster",
		},
	})

	resource.AddTestSweepers("aws_memorydb_cluster", &resource.Sweeper{
		Name: "aws_memorydb_cluster",
		F:    sweepClusters,
	})

	resource.AddTestSweepers("aws_memorydb_parameter_group", &resource.Sweeper{
		Name: "aws_memorydb_parameter_group",
		F:    sweepParameterGroups,
		Dependencies: []string{
			"aws_memorydb_cluster",
		},
	})

	resource.AddTestSweepers("aws_memorydb_snapshot", &resource.Sweeper{
		Name: "aws_memorydb_snapshot",
		F:    sweepSnapshots,
		Dependencies: []string{
			"aws_memorydb_cluster",
		},
	})

	resource.AddTestSweepers("aws_memorydb_subnet_group", &resource.Sweeper{
		Name: "aws_memorydb_subnet_group",
		F:    sweepSubnetGroups,
		Dependencies: []string{
			"aws_memorydb_cluster",
		},
	})

	resource.AddTestSweepers("aws_memorydb_user", &resource.Sweeper{
		Name: "aws_memorydb_user",
		F:    sweepUsers,
		Dependencies: []string{
			"aws_memorydb_acl",
		},
	})
}

func sweepACLs(region string) error {
	return nil
}

func sweepClusters(region string) error {
	return nil
}

func sweepParameterGroups(region string) error {
	return nil
}

func sweepSnapshots(region string) error {
	return nil
}

func sweepSubnetGroups(region string) error {
	return nil
}

func sweepUsers(region string) error {
	return nil
}
