package create

import (
	"fmt"

	"github.com/spf13/cobra"

	scaffoldcmd "github.com/kyma-project/modulectl/cmd/modulectl/create/scaffold"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/filegenerator"
	"github.com/kyma-project/modulectl/internal/service/filegenerator/reusefilegenerator"
	"github.com/kyma-project/modulectl/internal/service/moduleconfig"
	"github.com/kyma-project/modulectl/internal/service/scaffold"
	"github.com/kyma-project/modulectl/tools/filesystem"
	"github.com/kyma-project/modulectl/tools/yaml"

	_ "embed"
)

const (
	moduleConfigKind   = "module-config"
	manifestKind       = "manifest"
	defaultCRKind      = "defaultcr"
	securityConfigKind = "security-config"
)

//go:embed use.txt
var use string

//go:embed short.txt
var short string

//go:embed long.txt
var long string

func NewCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
	}

	svc, err := buildScaffoldService()
	if err != nil {
		return nil, fmt.Errorf("failed to build scaffold service: %w", err)
	}

	cmd, err := scaffoldcmd.NewCmd(svc)
	if err != nil {
		return nil, fmt.Errorf("failed to build scaffold command: %w", err)
	}

	rootCmd.AddCommand(cmd)
	return rootCmd, nil
}

func buildScaffoldService() (*scaffold.Service, error) {
	fileSystemUtil := &filesystem.Util{}
	yamlConverter := &yaml.ObjectToYAMLConverter{}

	moduleConfigContentProvider, err := contentprovider.NewModuleConfig(yamlConverter)
	if err != nil {
		return nil, fmt.Errorf("failed to create module config content provider: %w", err)
	}

	moduleConfigFileGenerator, err := filegenerator.NewService(moduleConfigKind, fileSystemUtil, moduleConfigContentProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create module config file generator: %w", err)
	}

	moduleConfigService, err := moduleconfig.NewService(fileSystemUtil, moduleConfigFileGenerator)
	if err != nil {
		return nil, fmt.Errorf("failed to create module config service: %w", err)
	}

	manifestFileGenerator, err := filegenerator.NewService(manifestKind, fileSystemUtil, contentprovider.NewManifest())
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest file generator: %w", err)
	}

	manifestReuseFileGenerator, err := reusefilegenerator.NewService(manifestKind, fileSystemUtil, manifestFileGenerator)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest reuse file generator: %w", err)
	}

	defaultCRFileGenerator, err := filegenerator.NewService(defaultCRKind, fileSystemUtil, contentprovider.NewDefaultCR())
	if err != nil {
		return nil, fmt.Errorf("failed to create default CR file generator: %w", err)
	}

	defaultCRReuseFileGenerator, err := reusefilegenerator.NewService(defaultCRKind, fileSystemUtil, defaultCRFileGenerator)
	if err != nil {
		return nil, fmt.Errorf("failed to create default CR reuse file generator: %w", err)
	}

	securitConfigContentProvider, err := contentprovider.NewSecurityConfig(yamlConverter)
	if err != nil {
		return nil, fmt.Errorf("failed to create security config content provider: %w", err)
	}

	securityConfigFileGenerator, err := filegenerator.NewService(securityConfigKind, fileSystemUtil, securitConfigContentProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create security config file generator: %w", err)
	}

	securityConfigReuseFileGenerator, err := reusefilegenerator.NewService(securityConfigKind, fileSystemUtil, securityConfigFileGenerator)
	if err != nil {
		return nil, fmt.Errorf("failed to create security config reuse file generator: %w", err)
	}

	scaffoldService, err := scaffold.NewService(
		moduleConfigService,
		manifestReuseFileGenerator,
		defaultCRReuseFileGenerator,
		securityConfigReuseFileGenerator,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scaffold service: %w", err)
	}

	return scaffoldService, nil
}
