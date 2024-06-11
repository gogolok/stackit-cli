package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stackitcloud/stackit-cli/internal/pkg/utils"

	"github.com/google/uuid"
	"github.com/stackitcloud/stackit-sdk-go/services/skcf"
)

var (
	testProjectId = uuid.NewString()
)

const (
	testClusterName = "test-cluster"
)

type skcfClientMocked struct {
	getServiceStatusFails    bool
	getServiceStatusResp     *skcf.ProjectResponse
	listClustersFails        bool
	listClustersResp         *skcf.ListClustersResponse
	listProviderOptionsFails bool
}

func (m *skcfClientMocked) GetServiceStatusExecute(_ context.Context, _ string) (*skcf.ProjectResponse, error) {
	if m.getServiceStatusFails {
		return nil, fmt.Errorf("could not get service status")
	}
	return m.getServiceStatusResp, nil
}

func (m *skcfClientMocked) ListClustersExecute(_ context.Context, _ string) (*skcf.ListClustersResponse, error) {
	if m.listClustersFails {
		return nil, fmt.Errorf("could not list clusters")
	}
	return m.listClustersResp, nil
}

func TestClusterExists(t *testing.T) {
	tests := []struct {
		description      string
		getClustersFails bool
		getClustersResp  *skcf.ListClustersResponse
		isValid          bool
		expectedExists   bool
	}{
		{
			description:     "cluster exists",
			getClustersResp: &skcf.ListClustersResponse{Items: &[]skcf.Cluster{{Name: utils.Ptr(testClusterName)}}},
			isValid:         true,
			expectedExists:  true,
		},
		{
			description:     "cluster exists 2",
			getClustersResp: &skcf.ListClustersResponse{Items: &[]skcf.Cluster{{Name: utils.Ptr("some-cluster")}, {Name: utils.Ptr("some-other-cluster")}, {Name: utils.Ptr(testClusterName)}}},
			isValid:         true,
			expectedExists:  true,
		},
		{
			description:     "cluster does not exist",
			getClustersResp: &skcf.ListClustersResponse{Items: &[]skcf.Cluster{{Name: utils.Ptr("some-cluster")}, {Name: utils.Ptr("some-other-cluster")}}},
			isValid:         true,
			expectedExists:  false,
		},
		{
			description:      "get clusters fails",
			getClustersFails: true,
			isValid:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			client := &skcfClientMocked{
				listClustersFails: tt.getClustersFails,
				listClustersResp:  tt.getClustersResp,
			}

			exists, err := ClusterExists(context.Background(), client, testProjectId, testClusterName)

			if tt.isValid && err != nil {
				t.Errorf("failed on valid input")
			}
			if !tt.isValid && err == nil {
				t.Errorf("did not fail on invalid input")
			}
			if !tt.isValid {
				return
			}
			if exists != tt.expectedExists {
				t.Errorf("expected exists to be %t, got %t", tt.expectedExists, exists)
			}
		})
	}
}

func TestConvertToSeconds(t *testing.T) {
	tests := []struct {
		description    string
		expirationTime string
		isValid        bool
		expectedOutput string
	}{
		{
			description:    "seconds",
			expirationTime: "30s",
			isValid:        true,
			expectedOutput: "30",
		},
		{
			description:    "minutes",
			expirationTime: "30m",
			isValid:        true,
			expectedOutput: "1800",
		},
		{
			description:    "hours",
			expirationTime: "30h",
			isValid:        true,
			expectedOutput: "108000",
		},
		{
			description:    "days",
			expirationTime: "30d",
			isValid:        true,
			expectedOutput: "2592000",
		},
		{
			description:    "months",
			expirationTime: "30M",
			isValid:        true,
			expectedOutput: "77760000",
		},
		{
			description:    "leading zero",
			expirationTime: "0030M",
			isValid:        true,
			expectedOutput: "77760000",
		},
		{
			description:    "invalid unit",
			expirationTime: "30x",
			isValid:        false,
		},
		{
			description:    "invalid unit 2",
			expirationTime: "3000abcdef",
			isValid:        false,
		},
		{
			description:    "invalid unit 3",
			expirationTime: "3000abcdef000",
			isValid:        false,
		},
		{
			description:    "invalid time",
			expirationTime: "x",
			isValid:        false,
		},
		{
			description:    "empty",
			expirationTime: "",
			isValid:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			output, err := ConvertToSeconds(tt.expirationTime)

			if tt.isValid && err != nil {
				t.Errorf("failed on valid input")
			}
			if !tt.isValid && err == nil {
				t.Errorf("did not fail on invalid input")
			}
			if !tt.isValid {
				return
			}
			if *output != tt.expectedOutput {
				t.Errorf("expected output to be %s, got %s", tt.expectedOutput, *output)
			}
		})
	}
}

func TestWriteConfigFile(t *testing.T) {
	tests := []struct {
		description     string
		location        string
		kubeconfig      string
		isValid         bool
		isLocationDir   bool
		isLocationEmpty bool
		expectedErr     string
	}{
		{
			description: "base",
			location:    filepath.Join("base", "config"),
			kubeconfig:  "kubeconfig",
			isValid:     true,
		},
		{
			description:     "empty location",
			location:        "",
			kubeconfig:      "kubeconfig",
			isValid:         false,
			isLocationEmpty: true,
		},
		{
			description:   "path is only dir",
			location:      "only_dir",
			kubeconfig:    "kubeconfig",
			isValid:       false,
			isLocationDir: true,
		},
		{
			description: "empty kubeconfig",
			location:    filepath.Join("empty", "config"),
			kubeconfig:  "",
			isValid:     false,
		},
	}

	baseTestDir := "test_data/"
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			testLocation := filepath.Join(baseTestDir, tt.location)
			// make sure empty case still works
			if tt.isLocationEmpty {
				testLocation = ""
			}
			// filepath Join cleans trailing separators
			if tt.isLocationDir {
				testLocation += string(filepath.Separator)
			}
			err := WriteConfigFile(testLocation, tt.kubeconfig)

			if tt.isValid && err != nil {
				t.Errorf("failed on valid input")
			}
			if !tt.isValid && err == nil {
				t.Errorf("did not fail on invalid input")
			}

			if tt.isValid {
				data, err := os.ReadFile(testLocation)
				if err != nil {
					t.Errorf("could not read file: %s", tt.location)
				}
				if string(data) != tt.kubeconfig {
					t.Errorf("expected file content to be %s, got %s", tt.kubeconfig, string(data))
				}
			}
		})
	}
	// Cleanup
	err := os.RemoveAll(baseTestDir)
	if err != nil {
		t.Errorf("failed cleaning test data")
	}
}

func TestGetDefaultKubeconfigPath(t *testing.T) {
	tests := []struct {
		description string
	}{
		{
			description: "base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			output, err := GetDefaultKubeconfigPath()

			if err != nil {
				t.Errorf("failed on valid input")
			}
			userHome, err := os.UserHomeDir()
			if err != nil {
				t.Errorf("could not get user home directory")
			}
			if output != filepath.Join(userHome, ".kube", "config") {
				t.Errorf("expected output to be %s, got %s", filepath.Join(userHome, ".kube", "config"), output)
			}
		})
	}
}
