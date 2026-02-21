package kaggle

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Service struct {
	// CLI relies on KAGGLE_USERNAME and KAGGLE_KEY env vars
	WorkDir string
}

func NewService(workDir string) *Service {
	return &Service{WorkDir: workDir}
}

// DatasetMetadata matches dataset-metadata.json
type DatasetMetadata struct {
	Title     string    `json:"title"`
	Id        string    `json:"id"`
	Licenses  []License `json:"licenses"`
	IsPrivate bool      `json:"isPrivate"`
}

type License struct {
	Name string `json:"name"`
}

// KernelMetadata matches kernel-metadata.json
type KernelMetadata struct {
	Id                 string   `json:"id"`
	Title              string   `json:"title"`
	CodeFile           string   `json:"code_file"`
	Language           string   `json:"language"`
	KernelType         string   `json:"kernel_type"`
	IsPrivate          bool     `json:"is_private"`
	EnableGpu          bool     `json:"enable_gpu"`
	EnableInternet     bool     `json:"enable_internet"`
	DatasetSources     []string `json:"dataset_sources"`
	KernelSources      []string `json:"kernel_sources"`
	CompetitionSources []string `json:"competition_sources"`
}

// CreateDataset creates a new dataset on Kaggle from a local file
func (s *Service) CreateDataset(name string, filePath string) (string, error) {
	// 1. Prepare Staging Directory
	stagingDir := filepath.Join(s.WorkDir, "staging_datasets", sanitize(name))
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return "", err
	}
	defer os.RemoveAll(stagingDir)

	// 2. Copy File
	destFile := filepath.Join(stagingDir, filepath.Base(filePath))
	// Simulating copy (or actually copy/link)
	// For now, assuming input is content byte array or we copy from existing path
	// In the worker, we download from MinIO to a temp path, so we can just move/copy it.
	// For this implementation, I'll assume filePath is accessible.
	input, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read input file: %v", err)
	}
	if err := os.WriteFile(destFile, input, 0644); err != nil {
		return "", fmt.Errorf("failed to copy file to staging: %v", err)
	}

	// 3. Create Metadata
	slug := sanitize(name)
	username := os.Getenv("KAGGLE_USERNAME") // Assuming env var is set
	if username == "" {
		return "", fmt.Errorf("KAGGLE_USERNAME not set")
	}
	ref := fmt.Sprintf("%s/%s", username, slug)

	meta := DatasetMetadata{
		Title:     name,
		Id:        ref,
		Licenses:  []License{{Name: "CC0-1.0"}},
		IsPrivate: true,
	}

	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(stagingDir, "dataset-metadata.json"), metaBytes, 0644); err != nil {
		return "", err
	}

	// 4. Run CLI: kaggle datasets create -p .
	cmd := exec.Command("kaggle", "datasets", "create", "-p", stagingDir, "--dir-mode", "zip")
	cmd.Env = os.Environ() // Pass KAGGLE_USERNAME/KEY
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kaggle cli error: %v, output: %s", err, string(out))
	}

	return ref, nil
}

// PushKernel deploys a notebook to Kaggle
func (s *Service) PushKernel(slug string, notebookContent []byte, datasetRefs []string) (string, error) {
	stagingDir := filepath.Join(s.WorkDir, "staging_kernels", slug)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return "", err
	}
	defer os.RemoveAll(stagingDir)

	// 1. Write Notebook
	if err := os.WriteFile(filepath.Join(stagingDir, "notebook.ipynb"), notebookContent, 0644); err != nil {
		return "", err
	}

	// 2. Write Metadata
	username := os.Getenv("KAGGLE_USERNAME")
	ref := fmt.Sprintf("%s/%s", username, slug)

	meta := KernelMetadata{
		Id:                 ref,
		Title:              slug, // Use slug as title for simplicity
		CodeFile:           "notebook.ipynb",
		Language:           "python",
		KernelType:         "notebook",
		IsPrivate:          true,
		EnableGpu:          true,
		EnableInternet:     true,
		DatasetSources:     datasetRefs,
		KernelSources:      []string{},
		CompetitionSources: []string{},
	}
	if meta.DatasetSources == nil {
		meta.DatasetSources = []string{}
	}

	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(stagingDir, "kernel-metadata.json"), metaBytes, 0644); err != nil {
		return "", err
	}

	// 3. Run CLI: kaggle kernels push -p .
	cmd := exec.Command("kaggle", "kernels", "push", "-p", stagingDir)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kaggle cli error: %v, output: %s", err, string(out))
	}

	return ref, nil
}

func (s *Service) GetKernelStatus(ref string) (string, error) {
	// kaggle kernels status [ref]
	cmd := exec.Command("kaggle", "kernels", "status", ref)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %v, out: %s", err, string(out))
	}

	output := strings.ToLower(string(out))
	if strings.Contains(output, "complete") {
		return "completed", nil
	} else if strings.Contains(output, "running") || strings.Contains(output, "queued") {
		return "running", nil
	} else if strings.Contains(output, "error") || strings.Contains(output, "failed") {
		return "failed", nil
	} else if strings.Contains(output, "cancel") {
		return "cancelled", nil
	}

	return "unknown", nil
}

// Helper
func sanitize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric except dash
	// Simplified...
	return s
}
