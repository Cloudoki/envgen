package cmd

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

const DefaultEnvFileName = ".env"

type ConfBranch struct {
	Name   string `yaml:"name"`
	Suffix string `yaml:"suffix"`
}

type ConfPackage struct {
	Package   string   `yaml:"package"`
	EnvFile   string   `yaml:"envFile"`
	Variables []string `yaml:"variables"`
}

type GeneratorConfig struct {
	BranchVarName    string        `yaml:"branchVarName"`
	BranchVarDefault string        `yaml:"branchVarDefault"`
	Branches         []ConfBranch  `yaml:"branches"`
	Packages         []ConfPackage `yaml:"packages"`
	Globals          []string      `yaml:"globals"`
}

type Generator struct {
	conf         GeneratorConfig
	branchSuffix string
	globals      []string
}

var rootCmd = &cobra.Command{
	Version:       "2.1.0",
	SilenceErrors: true,
	Use:           "envgen <configFilePath> [envFile1] ... [envFileN]",
	Short:         "envgen generates env files for sub packages",
	Long:          "envgen is CLI tool that generates .env files for subpackages in your project based on a configuration file",
	Args:          cobra.MinimumNArgs(1),
	RunE:          runGenerator,
}

func Execute() {
	rootCmd.SetErr(errorWriter{})

	envFiles := os.Args[2:]
	if len(envFiles) > 0 {
		err := godotenv.Load(envFiles...)
		if err != nil {
			rootCmd.PrintErr("Error loading .env file")
		}
	}

	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErr(err)
		os.Exit(1)
	}
}

func runGenerator(cmd *cobra.Command, args []string) error {
	gen, err := newGenerator(args[0])
	if err != nil {
		return err
	}

	logInfo("Starting env files generation\n")
	gen.loadGlobals()
	gen.generateEnvFiles()
	logInfo("Finished env files generation!\n")

	return nil
}

func newGenerator(filepath string) (*Generator, error) {
	gen := &Generator{}
	err := gen.loadConfig(filepath)
	if err != nil {
		return nil, err
	}

	return gen, nil
}

func (g *Generator) generateEnvFiles() {
	var wg sync.WaitGroup
	for _, p := range g.conf.Packages {
		wg.Add(1)
		go g.generatePackageEnvFile(p, &wg)
	}

	wg.Wait()
}

func (g *Generator) generatePackageEnvFile(pckg ConfPackage, wg *sync.WaitGroup) {
	defer wg.Done()
	logInfo(fmt.Sprintf("[%s] loading variables", pckg.Package))

	packageVars, missing := getVariablesValues(pckg.Variables, g.branchSuffix)
	if len(missing) > 0 {
		// TODO create flag to break execution as error
		logWarn(fmt.Sprintf("[%s] missing env vars: %s\n", pckg.Package, strings.Join(missing, ", ")))
	}

	logInfo(fmt.Sprintf("[%s] writing env file", pckg.Package))

	envFile := pckg.EnvFile
	if envFile == "" {
		envFile = DefaultEnvFileName
	}

	genEnvFilePath := fmt.Sprintf("%s/%s", pckg.Package, envFile)
	err := writeFile(genEnvFilePath, append(packageVars, g.globals...))
	if err != nil {
		logError(err)
	}

	logInfo(fmt.Sprintf("[%s] generated env file\n", pckg.Package))
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

// loads global env variables (to be added on all packages)
func (g *Generator) loadGlobals() {
	globals, missing := getVariablesValues(g.conf.Globals, "")
	if len(missing) > 0 {
		logWarn(fmt.Sprintf("[globals] missing env vars: %s\n", strings.Join(missing, ", ")))
	}

	g.globals = globals
}

// returns a string slice loaded with the env var declarations
// and an array with those not found in the environment
func getVariablesValues(envVars []string, suffix string) ([]string, []string) {
	vars := []string{}
	missing := []string{}
	for _, v := range envVars {
		val, ok := os.LookupEnv(v + suffix)
		if !ok {
			missing = append(missing, v)
		}

		vars = append(vars, fmt.Sprintf("%s=%s", v, val))
	}

	return vars, missing
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
