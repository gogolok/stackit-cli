package describe

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/stackitcloud/stackit-cli/internal/pkg/args"
	"github.com/stackitcloud/stackit-cli/internal/pkg/errors"
	"github.com/stackitcloud/stackit-cli/internal/pkg/examples"
	"github.com/stackitcloud/stackit-cli/internal/pkg/globalflags"
	"github.com/stackitcloud/stackit-cli/internal/pkg/print"
	"github.com/stackitcloud/stackit-cli/internal/pkg/services/skcf/client"
	"github.com/stackitcloud/stackit-cli/internal/pkg/tables"

	"github.com/spf13/cobra"
	"github.com/stackitcloud/stackit-sdk-go/services/skcf"
)

const (
	clusterNameArg = "CLUSTER_NAME"
)

type inputModel struct {
	*globalflags.GlobalFlagModel
	ClusterName string
}

func NewCmd(p *print.Printer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("describe %s", clusterNameArg),
		Short: "Shows details  of a SKCF cluster",
		Long:  "Shows details  of a STACKIT Korifi Cloud Foundry (SKCF) cluster.",
		Args:  args.SingleArg(clusterNameArg, nil),
		Example: examples.Build(
			examples.NewExample(
				`Get details of an SKCF cluster with name "my-cluster"`,
				"$ stackit skcf cluster describe my-cluster"),
			examples.NewExample(
				`Get details of an SKCF cluster with name "my-cluster" in JSON format`,
				"$ stackit skcf cluster describe my-cluster --output-format json"),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			model, err := parseInput(p, cmd, args)
			if err != nil {
				return err
			}
			// Configure API client
			apiClient, err := client.ConfigureClient(p)
			if err != nil {
				return err
			}

			// Call API
			req := buildRequest(ctx, model, apiClient)
			resp, err := req.Execute()
			if err != nil {
				return fmt.Errorf("read SKCF cluster: %w", err)
			}

			return outputResult(p, model.OutputFormat, resp)
		},
	}
	return cmd
}

func parseInput(p *print.Printer, cmd *cobra.Command, inputArgs []string) (*inputModel, error) {
	clusterName := inputArgs[0]

	globalFlags := globalflags.Parse(p, cmd)
	if globalFlags.ProjectId == "" {
		return nil, &errors.ProjectIdError{}
	}

	model := inputModel{
		GlobalFlagModel: globalFlags,
		ClusterName:     clusterName,
	}

	if p.IsVerbosityDebug() {
		modelStr, err := print.BuildDebugStrFromInputModel(model)
		if err != nil {
			p.Debug(print.ErrorLevel, "convert model to string for debugging: %v", err)
		} else {
			p.Debug(print.DebugLevel, "parsed input values: %s", modelStr)
		}
	}

	return &model, nil
}

func buildRequest(ctx context.Context, model *inputModel, apiClient *skcf.APIClient) skcf.ApiGetClusterRequest {
	req := apiClient.GetCluster(ctx, model.ProjectId, model.ClusterName)
	return req
}

func outputResult(p *print.Printer, outputFormat string, cluster *skcf.Cluster) error {
	switch outputFormat {
	case print.JSONOutputFormat:
		details, err := json.MarshalIndent(cluster, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal SKCF cluster: %w", err)
		}
		p.Outputln(string(details))

		return nil
	case print.YAMLOutputFormat:
		details, err := yaml.MarshalWithOptions(cluster, yaml.IndentSequence(true))
		if err != nil {
			return fmt.Errorf("marshal SKCF cluster: %w", err)
		}
		p.Outputln(string(details))

		return nil
	default:
		table := tables.NewTable()
		table.AddRow("NAME", *cluster.Name)
		table.AddSeparator()
		table.AddRow("STATE", *cluster.Status.Aggregated)
		err := table.Display(p)
		if err != nil {
			return fmt.Errorf("render table: %w", err)
		}

		return nil
	}
}
