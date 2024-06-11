package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/stackitcloud/stackit-cli/internal/pkg/utils"

	"github.com/stackitcloud/stackit-sdk-go/services/skcf"
)

const (
	supportedState = "supported"
)

type SKCFClient interface {
	//GetServiceStatusExecute(ctx context.Context, projectId string) (*skcf.ProjectResponse, error)
	ListClustersExecute(ctx context.Context, projectId string) (*skcf.ListClustersResponse, error)
}

//func ProjectEnabled(ctx context.Context, apiClient SKCFClient, projectId string) (bool, error) {
//	project, err := apiClient.GetServiceStatusExecute(ctx, projectId)
//	if err != nil {
//		return false, fmt.Errorf("get SKCF status: %w", err)
//	}
//	return *project.State == skcf.PROJECTSTATE_CREATED, nil
//}

func ClusterExists(ctx context.Context, apiClient SKCFClient, projectId, clusterName string) (bool, error) {
	clusters, err := apiClient.ListClustersExecute(ctx, projectId)
	if err != nil {
		return false, fmt.Errorf("list SKCF clusters: %w", err)
	}
	for _, cl := range *clusters.Items {
		if cl.Name != nil && *cl.Name == clusterName {
			return true, nil
		}
	}
	return false, nil
}

// The time string must be in the format of <value><unit>, where unit is one of s, m, h, d, M.
func ConvertToSeconds(timeStr string) (*string, error) {
	if len(timeStr) < 2 {
		return nil, fmt.Errorf("invalid time: %s", timeStr)
	}

	unit := timeStr[len(timeStr)-1:]

	valueStr := timeStr[:len(timeStr)-1]
	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid time value: %s", valueStr)
	}

	var multiplier uint64
	switch unit {
	// second
	case "s":
		multiplier = 1
	// minute
	case "m":
		multiplier = 60
	// hour
	case "h":
		multiplier = 60 * 60
	// day
	case "d":
		multiplier = 60 * 60 * 24
	// month, assume 30 days
	case "M":
		multiplier = 60 * 60 * 24 * 30
	default:
		return nil, fmt.Errorf("invalid time unit: %s", unit)
	}

	result := uint64(value) * multiplier
	return utils.Ptr(strconv.FormatUint(result, 10)), nil
}

// WriteConfigFile writes the given data to the given path.
// The directory is created if it does not exist.
func WriteConfigFile(configPath, data string) error {
	if data == "" {
		return fmt.Errorf("no data to write")
	}

	dir := filepath.Dir(configPath)

	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	err = os.WriteFile(configPath, []byte(data), 0o600)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// GetDefaultKubeconfigPath returns the default location for the kubeconfig file.
func GetDefaultKubeconfigPath() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home directory: %w", err)
	}

	return filepath.Join(userHome, ".kube", "config"), nil
}
