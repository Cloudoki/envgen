package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"gopkg.in/yaml.v2"
)

const DefaultEnvFileName = ".env"

type ConfBranches struct {
	Name   string `yaml:"name"`
	Suffix string `yaml:"suffix"`
}

type ConfPackages struct {
	Package   string   `yaml:"package"`
	EnvFile   string   `yaml:"envFile"`
	Variables []string `yaml:"variables"`
}

type GeneratorConfig struct {
	BranchVarName    string         `yaml:"branchVarName"`
	BranchVarDefault string         `yaml:"branchVarDefault"`
	Branches         []ConfBranches `yaml:"branches"`
	Packages         []ConfPackages `yaml:"packages"`
	Globals          []string       `yaml:"globals"`
}

type Generator struct {
	conf         GeneratorConfig
	branchSuffix string
}

var rootCmd = &cobra.Command{
	Version:       "1.0.0",
	SilenceErrors: true,
	Use:           "envgen <configFilePath>",
	Short:         "envgen generates env files for sub packages",
	Long:          "envgen is CLI tool that generates .env files for subpackages in your project based on a configuration file",
	Args:          cobra.ExactArgs(1),
	RunE:          generateEnvFiles,
}

func Execute() {
	rootCmd.SetErr(errorWriter{})
	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErr(err)
		os.Exit(1)
	}
}

func generateEnvFiles(cmd *cobra.Command, args []string) error {
	gen := &Generator{}
	err := gen.loadConfig(args[0])
	if err != nil {
		return err
	}

	logInfo("Starting env files generation")
	fmt.Println()

	globals, err := getVariablesValues(gen.conf.Globals, "")
	if err != nil {
		return err
	}

	for _, p := range gen.conf.Packages {
		logInfo("> Loading variables for " + p.Package)

		packageVars, err := getVariablesValues(p.Variables, gen.branchSuffix)
		if err != nil {
			return err
		}

		packageVars = append(packageVars, globals...)

		logInfo("> Writing env file for " + p.Package)
		envFile := p.EnvFile
		if envFile == "" {
			envFile = DefaultEnvFileName
		}
		genEnvFilePath := fmt.Sprintf("%s/%s", p.Package, envFile)
		err = writeFile(genEnvFilePath, packageVars)
		if err != nil {
			return err
		}

		logInfo("> Done generating env file for " + p.Package)
		fmt.Println()
	}

	logInfo("Finished env files generation!")
	return nil
}

func getVariablesValues(envVars []string, suffix string) ([]string, error) {
	vars := []string{}
	for _, v := range envVars {
		val, ok := os.LookupEnv(v + suffix)
		if !ok {
			err := fmt.Errorf("missing variable %s", v)
			return nil, err
		}

		vars = append(vars, fmt.Sprintf("%s=%s", v, val))
	}

	return vars, nil
}

// loadConfig loads the configuration from the provided yaml file
// into the instance of the Generator. It also determines the branch suffix property.
func (g *Generator) loadConfig(filepath string) error {
	config := &GeneratorConfig{}
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return err
	}

	g.conf = *config

	g.branchSuffix = g.findBranchSuffix()

	return nil
}

// findBranchSuffix determines the branch suffix to use, depending on the current CI branch
func (g *Generator) findBranchSuffix() string {
	branch := getEnv(g.conf.BranchVarName, "", g.conf.BranchVarDefault)

	for _, b := range g.conf.Branches {
		if b.Name == branch {
			return b.Suffix
		}
	}

	return ""
}

// getEnv looks up for a loaded environment variable.
// An optional suffix may be passed, as well as a default value to return if the env var is not loaded.
func getEnv(key string, suffix string, defaultVal string) string {
	if value, exists := os.LookupEnv(key + suffix); exists {
		return value
	}

	return defaultVal
}

// writeFile writes a slice of strings into a file, separated by new lines
func writeFile(path string, vars []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	sep := "\n"
	for _, line := range vars {
		if _, err = file.WriteString(line + sep); err != nil {
			return err
		}
	}

	return nil
}
