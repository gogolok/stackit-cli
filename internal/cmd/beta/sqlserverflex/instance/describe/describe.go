package describe

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/stackitcloud/stackit-cli/internal/pkg/args"
	"github.com/stackitcloud/stackit-cli/internal/pkg/errors"
	"github.com/stackitcloud/stackit-cli/internal/pkg/examples"
	"github.com/stackitcloud/stackit-cli/internal/pkg/globalflags"
	"github.com/stackitcloud/stackit-cli/internal/pkg/print"
	"github.com/stackitcloud/stackit-cli/internal/pkg/services/sqlserverflex/client"
	"github.com/stackitcloud/stackit-cli/internal/pkg/tables"
	"github.com/stackitcloud/stackit-cli/internal/pkg/utils"

	"github.com/spf13/cobra"
	"github.com/stackitcloud/stackit-sdk-go/services/sqlserverflex"
)

const (
	instanceIdArg = "INSTANCE_ID"
)

type inputModel struct {
	*globalflags.GlobalFlagModel
	InstanceId string
}

func NewCmd(p *print.Printer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("describe %s", instanceIdArg),
		Short: "Shows details  of an SQLServer Flex instance",
		Long:  "Shows details  of an SQLServer Flex instance.",
		Args:  args.SingleArg(instanceIdArg, utils.ValidateUUID),
		Example: examples.Build(
			examples.NewExample(
				`Get details of an SQLServer Flex instance with ID "xxx"`,
				"$ stackit beta sqlserverflex instance describe xxx"),
			examples.NewExample(
				`Get details of an SQLServer Flex instance with ID "xxx" in JSON format`,
				"$ stackit beta sqlserverflex instance describe xxx --output-format json"),
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
				return fmt.Errorf("read SQLServer Flex instance: %w", err)
			}

			return outputResult(p, model.OutputFormat, resp.Item)
		},
	}
	return cmd
}

func parseInput(p *print.Printer, cmd *cobra.Command, inputArgs []string) (*inputModel, error) {
	instanceId := inputArgs[0]

	globalFlags := globalflags.Parse(p, cmd)
	if globalFlags.ProjectId == "" {
		return nil, &errors.ProjectIdError{}
	}

	model := inputModel{
		GlobalFlagModel: globalFlags,
		InstanceId:      instanceId,
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

func buildRequest(ctx context.Context, model *inputModel, apiClient *sqlserverflex.APIClient) sqlserverflex.ApiGetInstanceRequest {
	req := apiClient.GetInstance(ctx, model.ProjectId, model.InstanceId)
	return req
}

func outputResult(p *print.Printer, outputFormat string, instance *sqlserverflex.Instance) error {
	switch outputFormat {
	case print.JSONOutputFormat:
		details, err := json.MarshalIndent(instance, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal SQLServer Flex instance: %w", err)
		}
		p.Outputln(string(details))

		return nil
	case print.YAMLOutputFormat:
		details, err := yaml.MarshalWithOptions(instance, yaml.IndentSequence(true))
		if err != nil {
			return fmt.Errorf("marshal SQLServer Flex instance: %w", err)
		}
		p.Outputln(string(details))

		return nil
	default:
		aclsArray := *instance.Acl.Items
		acls := strings.Join(aclsArray, ",")

		table := tables.NewTable()
		table.AddRow("ID", *instance.Id)
		table.AddSeparator()
		table.AddRow("NAME", *instance.Name)
		table.AddSeparator()
		table.AddRow("STATUS", *instance.Status)
		table.AddSeparator()
		table.AddRow("STORAGE SIZE (GB)", *instance.Storage.Size)
		table.AddSeparator()
		table.AddRow("VERSION", *instance.Version)
		table.AddSeparator()
		table.AddRow("BACKUP SCHEDULE (UTC)", *instance.BackupSchedule)
		table.AddSeparator()
		table.AddRow("ACL", acls)
		table.AddSeparator()
		table.AddRow("FLAVOR DESCRIPTION", *instance.Flavor.Description)
		table.AddSeparator()
		table.AddRow("CPU", *instance.Flavor.Cpu)
		table.AddSeparator()
		table.AddRow("RAM (GB)", *instance.Flavor.Memory)
		table.AddSeparator()

		err := table.Display(p)
		if err != nil {
			return fmt.Errorf("render table: %w", err)
		}

		return nil
	}
}
