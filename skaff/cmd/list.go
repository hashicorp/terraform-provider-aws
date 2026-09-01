// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cmd

import (
	"github.com/hashicorp/terraform-provider-aws/skaff/list"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Create scaffolding for a list resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return list.Create(name, snakeName, !clearComments, framework, force)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&snakeName, "snakename", "s", "", "if skaff doesn't get it right, explicitly give name in snake case (e.g., db_vpc_instance)")
	listCmd.Flags().BoolVarP(&clearComments, "clear-comments", "c", false, "do not include instructional comments in source")
	listCmd.Flags().StringVarP(&name, "name", "n", "", "name of the entity")
	listCmd.Flags().BoolVarP(&force, "force", "f", false, "force creation, overwriting existing files")
	listCmd.Flags().BoolVarP(&framework, "framework", "p", false, "use scaffolding for resources written using framework")
}
