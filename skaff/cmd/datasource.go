package cmd

import (
	"github.com/hashicorp/terraform-provider-aws/skaff/datasource"
	"github.com/spf13/cobra"
)

var datasourceCmd = &cobra.Command{
	Use:   "datasource",
	Short: "Create scaffolding for a data source",
	RunE: func(cmd *cobra.Command, args []string) error {
		return datasource.Create(name, snakeName, !clearComments, force)
	},
}

func init() {
	rootCmd.AddCommand(datasourceCmd)
	datasourceCmd.Flags().StringVarP(&snakeName, "snakename", "s", "", "If skaff doesn't get it right, explicitly give name in snake case (e.g., db_vpc_instance)")
	datasourceCmd.Flags().BoolVarP(&clearComments, "clear-comments", "c", false, "Do not include instructional comments in source")
	datasourceCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the entity")
	datasourceCmd.Flags().BoolVarP(&force, "force", "f", false, "Force creation, overwriting existing files")
}
