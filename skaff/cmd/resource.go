package cmd

import (
	"github.com/hashicorp/terraform-provider-aws/skaff/resource"
	"github.com/spf13/cobra"
)

var (
	snakeName     string
	clearComments bool
	name          string
	force         bool
)

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Create scaffolding for a resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return resource.Create(name, snakeName, !clearComments, force)
	},
}

func init() {
	rootCmd.AddCommand(resourceCmd)
	resourceCmd.Flags().StringVarP(&snakeName, "snakename", "s", "", "If skaff doesn't get it right, explicitly give name in snake case (e.g., db_vpc_instance)")
	resourceCmd.Flags().BoolVarP(&clearComments, "clear-comments", "c", false, "Do not include instructional comments in source")
	resourceCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the entity")
	resourceCmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation, overwriting existing files")
}
