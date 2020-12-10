package cmd

import (
	"context"
	"github.com/njucz/terraform-provider-azurerm-analysis/internal/common"
	"os/exec"
)

type AzurermProviderUrlExtractor struct {
	ExtractCmdPath string
	ProviderDir    string
}

func (extractor *AzurermProviderUrlExtractor) ExtractUrl(ctx context.Context) (string, []common.Endpoint, error) {
	cmd := exec.CommandContext(ctx, extractor.ExtractCmdPath, "./...")
	cmd.Dir = extractor.ProviderDir

	outputBytes, err := cmd.Output()
	if err != nil {
		return "", nil, err
	}
	raw := string(outputBytes)
	endpoints, err := common.ConstructEndpointList(raw)
	return raw, endpoints, err
}
