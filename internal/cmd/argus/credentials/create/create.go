package create

import (
	"context"
	"fmt"

	"github.com/stackitcloud/stackit-cli/internal/pkg/args"
	cliErr "github.com/stackitcloud/stackit-cli/internal/pkg/errors"
	"github.com/stackitcloud/stackit-cli/internal/pkg/examples"
	"github.com/stackitcloud/stackit-cli/internal/pkg/flags"
	"github.com/stackitcloud/stackit-cli/internal/pkg/globalflags"
	"github.com/stackitcloud/stackit-cli/internal/pkg/print"
	"github.com/stackitcloud/stackit-cli/internal/pkg/services/argus/client"
	argusUtils "github.com/stackitcloud/stackit-cli/internal/pkg/services/argus/utils"

	"github.com/spf13/cobra"
	"github.com/stackitcloud/stackit-sdk-go/services/argus"
)

const (
	instanceIdFlag = "instance-id"
)

type inputModel struct {
	*globalflags.GlobalFlagModel

	InstanceId string
}

func NewCmd(p *print.Printer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates credentials for an Argus instance.",
		Long: fmt.Sprintf("%s\n%s",
			"Creates credentials (username and password) for an Argus instance.",
			"The credentials will be generated and included in the response. You won't be able to retrieve the password later."),
		Args: args.NoArgs,
		Example: examples.Build(
			examples.NewExample(
				`Create credentials for Argus instance with ID "xxx"`,
				"$ stackit argus credentials create --instance-id xxx"),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			model, err := parseInput(p, cmd)
			if err != nil {
				return err
			}

			// Configure API client
			apiClient, err := client.ConfigureClient(p)
			if err != nil {
				return err
			}

			instanceLabel, err := argusUtils.GetInstanceName(ctx, apiClient, model.InstanceId, model.ProjectId)
			if err != nil {
				p.Debug(print.ErrorLevel, "get instance name: %v", err)
				instanceLabel = model.InstanceId
			}

			if !model.AssumeYes {
				prompt := fmt.Sprintf("Are you sure you want to create credentials for instance %q?", instanceLabel)
				err = p.PromptForConfirmation(prompt)
				if err != nil {
					return err
				}
			}

			// Call API
			req := buildRequest(ctx, model, apiClient)
			if err != nil {
				return err
			}
			resp, err := req.Execute()
			if err != nil {
				return fmt.Errorf("create credentials for Argus instance: %w", err)
			}

			p.Outputf("Created credentials for instance %q.\n\n", instanceLabel)
			// The username field cannot be set by the user so we only display it if it's not returned empty
			username := *resp.Credentials.Username
			if username != "" {
				p.Outputf("Username: %s\n", username)
			}

			p.Outputf("Password: %s\n", *resp.Credentials.Password)
			return nil
		},
	}
	configureFlags(cmd)
	return cmd
}

func configureFlags(cmd *cobra.Command) {
	cmd.Flags().Var(flags.UUIDFlag(), instanceIdFlag, "Instance ID")

	err := flags.MarkFlagsRequired(cmd, instanceIdFlag)
	cobra.CheckErr(err)
}

func parseInput(p *print.Printer, cmd *cobra.Command) (*inputModel, error) {
	globalFlags := globalflags.Parse(p, cmd)
	if globalFlags.ProjectId == "" {
		return nil, &cliErr.ProjectIdError{}
	}

	return &inputModel{
		GlobalFlagModel: globalFlags,
		InstanceId:      flags.FlagToStringValue(p, cmd, instanceIdFlag),
	}, nil
}

func buildRequest(ctx context.Context, model *inputModel, apiClient *argus.APIClient) argus.ApiCreateCredentialsRequest {
	req := apiClient.CreateCredentials(ctx, model.InstanceId, model.ProjectId)
	return req
}
