// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cmd

import (
	"github.com/hashicorp/terraform-provider-aws/skaff/function"
	"github.com/spf13/cobra"
)

var description string

var functionCmd = &cobra.Command{
	Use:   "function",
	Short: "Create scaffolding for a function",
	RunE: func(cmd *cobra.Command, args []string) error {
		return function.Create(name, snakeName, description, !clearComments, force)
	},
}

func init() {
	rootCmd.AddCommand(functionCmd)
	functionCmd.Flags().BoolVarP(&clearComments, "clear-comments", "c", false, "do not include instructional comments in source")
	functionCmd.Flags().BoolVarP(&force, "force", "f", false, "force creation, overwriting existing files")
	functionCmd.Flags().StringVarP(&name, "name", "n", "", "name of the function")
	functionCmd.Flags().StringVarP(&snakeName, "snakename", "s", "", "if skaff doesn't get it right, explicitly give name in snake case (e.g., arn_build)")
	functionCmd.Flags().StringVarP(&description, "description", "d", "", "description of the function")
	functionCmd.MarkFlagRequired("name")
	functionCmd.MarkFlagRequired("description")
}
