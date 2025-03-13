// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cmd

import (
	"github.com/hashicorp/terraform-provider-aws/skaff/ephemeral"
	"github.com/spf13/cobra"
)

var ephemeralCmd = &cobra.Command{
	Use:   "ephemeral",
	Short: "Create scaffolding for an ephemeral resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ephemeral.Create(name, snakeName, !clearComments, force)
	},
}

func init() {
	rootCmd.AddCommand(ephemeralCmd)
	ephemeralCmd.Flags().StringVarP(&snakeName, "snakename", "s", "", "if skaff doesn't get it right, explicitly give name in snake case (e.g., db_vpc_instance)")
	ephemeralCmd.Flags().BoolVarP(&clearComments, "clear-comments", "c", false, "do not include instructional comments in source")
	ephemeralCmd.Flags().StringVarP(&name, "name", "n", "", "name of the entity")
	ephemeralCmd.Flags().BoolVarP(&force, "force", "f", false, "force creation, overwriting existing files")
}
