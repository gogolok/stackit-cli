package skcf

import (
	"github.com/stackitcloud/stackit-cli/internal/cmd/skcf/cluster"
	"github.com/stackitcloud/stackit-cli/internal/cmd/skcf/kubeconfig"
	"github.com/stackitcloud/stackit-cli/internal/pkg/args"
	"github.com/stackitcloud/stackit-cli/internal/pkg/print"
	"github.com/stackitcloud/stackit-cli/internal/pkg/utils"

	"github.com/spf13/cobra"
)

func NewCmd(p *print.Printer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skcf",
		Short: "Provides functionality for SKCF",
		Long:  "Provides functionality for STACKIT Korifi Cloud Foundry (SKCF).",
		Args:  args.NoArgs,
		Run:   utils.CmdHelp,
	}
	addSubcommands(cmd, p)
	return cmd
}

func addSubcommands(cmd *cobra.Command, p *print.Printer) {
	cmd.AddCommand(kubeconfig.NewCmd(p))
	cmd.AddCommand(cluster.NewCmd(p))
}
