package cluster

import (
	"github.com/stackitcloud/stackit-cli/internal/cmd/skcf/cluster/create"
	"github.com/stackitcloud/stackit-cli/internal/cmd/skcf/cluster/delete"
	"github.com/stackitcloud/stackit-cli/internal/cmd/skcf/cluster/describe"
	"github.com/stackitcloud/stackit-cli/internal/cmd/skcf/cluster/list"
	"github.com/stackitcloud/stackit-cli/internal/cmd/skcf/cluster/update"
	"github.com/stackitcloud/stackit-cli/internal/pkg/args"
	"github.com/stackitcloud/stackit-cli/internal/pkg/print"
	"github.com/stackitcloud/stackit-cli/internal/pkg/utils"

	"github.com/spf13/cobra"
)

func NewCmd(p *print.Printer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Provides functionality for SKCF cluster",
		Long:  "Provides functionality for STACKIT Korifi Cloud Foundry (SKCF) cluster.",
		Args:  args.NoArgs,
		Run:   utils.CmdHelp,
	}
	addSubcommands(cmd, p)
	return cmd
}

func addSubcommands(cmd *cobra.Command, p *print.Printer) {
	cmd.AddCommand(create.NewCmd(p))
	cmd.AddCommand(delete.NewCmd(p))
	cmd.AddCommand(describe.NewCmd(p))
	cmd.AddCommand(list.NewCmd(p))
	cmd.AddCommand(update.NewCmd(p))
}
