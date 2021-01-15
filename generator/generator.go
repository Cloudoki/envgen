package generator

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

const DefaultEnvFileName = ".env"

type branchConfig struct {
	Name   string `yaml:"name"`
	Suffix string `yaml:"suffix"`
}

type packageConfig struct {
	Package   string   `yaml:"package"`
	EnvFile   string   `yaml:"envFile"`
	Variables []string `yaml:"variables"`
}

type config struct {
	BranchVarName    string          `yaml:"branchVarName"`
	BranchVarDefault string          `yaml:"branchVarDefault"`
	Branches         []branchConfig  `yaml:"branches"`
	Packages         []packageConfig `yaml:"packages"`
	Globals          []string        `yaml:"globals"`
}

type Generator struct {
	conf         config
	branchSuffix string
	globals      []string
	logger       *log.Logger
	verbose      bool
}

// New creates a new env file generator, with loaded configurations.
func New(filepath string) (*Generator, error) {
	gen := &Generator{
		// TODO: accept external logger?
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}
	err := gen.LoadConfig(filepath)
	if err != nil {
		return nil, err
	}

	return gen, nil
}

// LoadConfig loads the configuration from the provided yaml file.
// It also determines the branch suffix property.
func (g *Generator) LoadConfig(filepath string) error {
	config := &config{}
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return err
	}

	g.conf = *config
	g.loadBranchSuffix()

	return nil
}

func (g *Generator) LoadGlobals() {
	globals, missing := getVariablesValues(g.conf.Globals, "")
	if len(missing) > 0 {
		g.logger.Printf("[globals] missing env vars: %s\n", strings.Join(missing, ", "))
	}

	g.globals = globals
}

// GenerateFiles creates all the env files as specified in the loaded configuration.
func (g *Generator) GenerateFiles() {
	g.LoadGlobals()

	var wg sync.WaitGroup
	for _, p := range g.conf.Packages {
		wg.Add(1)
		go g.GeneratePackageEnvFile(p, &wg)
	}

	wg.Wait()
}

// GeneratePackageEnvFile generates and writes the env file for a package
func (g *Generator) GeneratePackageEnvFile(pckg packageConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	packageVars, missing := getVariablesValues(pckg.Variables, g.branchSuffix)
	if len(missing) > 0 {
		// TODO create flag to break execution as error ?
		g.logger.Printf("[%s] missing env vars: %s\n", pckg.Package, strings.Join(missing, ", "))
	}

	envFile := pckg.EnvFile
	if envFile == "" {
		envFile = DefaultEnvFileName
	}

	genEnvFilePath := fmt.Sprintf("%s/%s", pckg.Package, envFile)
	err := writeFile(genEnvFilePath, append(packageVars, g.globals...))
	if err != nil {
		g.logger.Println(err)
	}

	g.logger.Printf("[%s] generated env file\n", pckg.Package)
}

// loadBranchSuffix finds and loads the branch suffix to use, depending on the current CI branch
func (g *Generator) loadBranchSuffix() {
	branch := getEnv(g.conf.BranchVarName, "", g.conf.BranchVarDefault)

	for _, b := range g.conf.Branches {
		if b.Name == branch {
			g.branchSuffix = b.Suffix
			return
		}
	}
}

// returns a string slice loaded with the env var declarations
// and an array with those not found in the environment
func getVariablesValues(envVars []string, suffix string) ([]string, []string) {
	vars := []string{}
	missing := []string{}
	for _, v := range envVars {
		val, ok := os.LookupEnv(v + suffix)
		if !ok || len(val) == 0 {
			missing = append(missing, v)
			continue
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