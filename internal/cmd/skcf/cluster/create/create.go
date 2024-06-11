package create

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/stackitcloud/stackit-cli/internal/pkg/args"
	"github.com/stackitcloud/stackit-cli/internal/pkg/errors"
	"github.com/stackitcloud/stackit-cli/internal/pkg/examples"
	"github.com/stackitcloud/stackit-cli/internal/pkg/flags"
	"github.com/stackitcloud/stackit-cli/internal/pkg/globalflags"
	"github.com/stackitcloud/stackit-cli/internal/pkg/print"
	"github.com/stackitcloud/stackit-cli/internal/pkg/projectname"
	"github.com/stackitcloud/stackit-cli/internal/pkg/services/skcf/client"
	skcfUtils "github.com/stackitcloud/stackit-cli/internal/pkg/services/skcf/utils"
	"github.com/stackitcloud/stackit-cli/internal/pkg/spinner"

	"github.com/spf13/cobra"
	"github.com/stackitcloud/stackit-sdk-go/services/skcf"
	"github.com/stackitcloud/stackit-sdk-go/services/skcf/wait"
)

const (
	clusterNameArg = "CLUSTER_NAME"

	payloadFlag = "payload"
)

type inputModel struct {
	*globalflags.GlobalFlagModel
	ClusterName string
	Payload     *skcf.CreateOrUpdateClusterPayload
}

func NewCmd(p *print.Printer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("create %s", clusterNameArg),
		Short: "Creates an SKCF cluster",
		Long: fmt.Sprintf("%s\n%s\n%s",
			"Creates a STACKIT Kubernetes Engine (SKCF) cluster.",
			"The payload can be provided as a JSON string or a file path prefixed with \"@\".",
			"See https://docs.api.stackit.cloud/documentation/skcf/version/v1alpha1#tag/Cluster/operation/SkcfService_CreateOrUpdateCluster for information regarding the payload structure.",
		),
		Args: args.SingleArg(clusterNameArg, nil),
		Example: examples.Build(
			examples.NewExample(
				`Create an SKCF cluster using default configuration`,
				"$ stackit skcf cluster create my-cluster"),
			examples.NewExample(
				`Create an SKCF cluster using an API payload sourced from the file "./payload.json"`,
				"$ stackit skcf cluster create my-cluster --payload @./payload.json"),
			examples.NewExample(
				`Create an SKCF cluster using an API payload provided as a JSON string`,
				`$ stackit skcf cluster create my-cluster --payload "{...}"`),
			examples.NewExample(
				`Generate a payload with default values, and adapt it with custom values for the different configuration options`,
				`$ stackit skcf cluster generate-payload > ./payload.json`,
				`<Modify payload in file, if needed>`,
				`$ stackit skcf cluster create my-cluster --payload @./payload.json`),
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

			projectLabel, err := projectname.GetProjectName(ctx, p, cmd)
			if err != nil {
				p.Debug(print.ErrorLevel, "get project name: %v", err)
				projectLabel = model.ProjectId
			}

			if !model.AssumeYes {
				prompt := fmt.Sprintf("Are you sure you want to create a cluster for project %q?", projectLabel)
				err = p.PromptForConfirmation(prompt)
				if err != nil {
					return err
				}
			}

			// Check if SKCF is enabled for this project
			// FIXME
			//enabled, err := skcfUtils.ProjectEnabled(ctx, apiClient, model.ProjectId)
			//if err != nil {
			//	return err
			//}
			//if !enabled {
			//	return fmt.Errorf("SKCF isn't enabled for this project, please run 'stackit skcf enable'")
			//}

			// Check if cluster exists
			exists, err := skcfUtils.ClusterExists(ctx, apiClient, model.ProjectId, model.ClusterName)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("cluster with name %s already exists", model.ClusterName)
			}

			if model.Payload == nil {
				//	defaultPayload, err := skcfUtils.GetDefaultPayload(ctx, apiClient)
				//	if err != nil {
				//		return fmt.Errorf("get default payload: %w", err)
				//	}
				// FIXME just a hack
				model.Payload = &skcf.CreateOrUpdateClusterPayload{}
			}

			// Call API
			req := buildRequest(ctx, model, apiClient)
			resp, err := req.Execute()
			if err != nil {
				return fmt.Errorf("create SKCF cluster: %w", err)
			}
			name := *resp.Name

			// Wait for async operation, if async mode not enabled
			if !model.Async {
				s := spinner.New(p)
				s.Start("Creating cluster")
				_, err = wait.CreateOrUpdateClusterWaitHandler(ctx, apiClient, model.ProjectId, name).WaitWithContext(ctx)
				if err != nil {
					return fmt.Errorf("wait for SKCF cluster creation: %w", err)
				}
				s.Stop()
			}

			return outputResult(p, model, projectLabel, resp)
		},
	}
	configureFlags(cmd)
	return cmd
}

func configureFlags(cmd *cobra.Command) {
	cmd.Flags().Var(flags.ReadFromFileFlag(), payloadFlag, `Request payload (JSON). Can be a string or a file path, if prefixed with "@" (example: @./payload.json). If unset, will use a default payload (you can check it by running "stackit skcf cluster generate-payload")`)
}

func parseInput(p *print.Printer, cmd *cobra.Command, inputArgs []string) (*inputModel, error) {
	clusterName := inputArgs[0]

	globalFlags := globalflags.Parse(p, cmd)
	if globalFlags.ProjectId == "" {
		return nil, &errors.ProjectIdError{}
	}

	payloadValue := flags.FlagToStringPointer(p, cmd, payloadFlag)
	var payload *skcf.CreateOrUpdateClusterPayload
	if payloadValue != nil {
		payload = &skcf.CreateOrUpdateClusterPayload{}
		err := json.Unmarshal([]byte(*payloadValue), payload)
		if err != nil {
			return nil, fmt.Errorf("encode payload: %w", err)
		}
	}

	model := inputModel{
		GlobalFlagModel: globalFlags,
		ClusterName:     clusterName,
		Payload:         payload,
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

func buildRequest(ctx context.Context, model *inputModel, apiClient *skcf.APIClient) skcf.ApiCreateOrUpdateClusterRequest {
	req := apiClient.CreateOrUpdateCluster(ctx, model.ProjectId, model.ClusterName)

	req = req.CreateOrUpdateClusterPayload(*model.Payload)
	return req
}

func outputResult(p *print.Printer, model *inputModel, projectLabel string, resp *skcf.Cluster) error {
	switch model.OutputFormat {
	case print.JSONOutputFormat:
		details, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal SKCF cluster: %w", err)
		}
		p.Outputln(string(details))

		return nil
	case print.YAMLOutputFormat:
		details, err := yaml.MarshalWithOptions(resp, yaml.IndentSequence(true))
		if err != nil {
			return fmt.Errorf("marshal SKCF cluster: %w", err)
		}
		p.Outputln(string(details))

		return nil
	default:
		operationState := "Created"
		if model.Async {
			operationState = "Triggered creation of"
		}
		p.Outputf("%s cluster for project %q. Cluster name: %s\n", operationState, projectLabel, *resp.Name)
		return nil
	}
}
