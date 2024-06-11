package kubeconfig

import (
	"github.com/stackitcloud/stackit-cli/internal/cmd/skcf/kubeconfig/create"
	"github.com/stackitcloud/stackit-cli/internal/cmd/skcf/kubeconfig/login"
	"github.com/stackitcloud/stackit-cli/internal/pkg/args"
	"github.com/stackitcloud/stackit-cli/internal/pkg/print"
	"github.com/stackitcloud/stackit-cli/internal/pkg/utils"

	"github.com/spf13/cobra"
)

func NewCmd(p *print.Printer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubeconfig",
		Short: "Provides functionality for SKCF kubeconfig",
		Long:  "Provides functionality for STACKIT Korifi Cloud Foundry (SKCF) kubeconfig.",
		Args:  args.NoArgs,
		Run:   utils.CmdHelp,
	}
	addSubcommands(cmd, p)
	return cmd
}

func addSubcommands(cmd *cobra.Command, p *print.Printer) {
	cmd.AddCommand(create.NewCmd(p))
	cmd.AddCommand(login.NewCmd(p))
}
