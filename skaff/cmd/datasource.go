// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cmd

import (
	"github.com/hashicorp/terraform-provider-aws/skaff/datasource"
	"github.com/spf13/cobra"
)

var datasourceCmd = &cobra.Command{
	Use:   "datasource",
	Short: "Create scaffolding for a data source",
	RunE: func(cmd *cobra.Command, args []string) error {
		return datasource.Create(name, snakeName, !clearComments, force, !pluginSDKV2, includeTags)
	},
}

func init() {
	rootCmd.AddCommand(datasourceCmd)
	datasourceCmd.Flags().StringVarP(&snakeName, "snakename", "s", "", "if skaff doesn't get it right, explicitly give name in snake case (e.g., db_vpc_instance)")
	datasourceCmd.Flags().BoolVarP(&clearComments, "clear-comments", "c", false, "do not include instructional comments in source")
	datasourceCmd.Flags().StringVarP(&name, "name", "n", "", "name of the entity")
	datasourceCmd.Flags().BoolVarP(&force, "force", "f", false, "force creation, overwriting existing files")
	datasourceCmd.Flags().BoolVarP(&pluginSDKV2, "plugin-sdkv2", "p", false, "generate for Terraform Plugin SDK V2")
	datasourceCmd.Flags().BoolVarP(&includeTags, "include-tags", "t", false, "Indicate that this resource has tags and the code for tagging should be generated")
}
