package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

var AppConfig ConfigPath

type Timeout struct{}

type FileConfig struct {
	Extentsion     string
	FileName       string
	ConfigPath     string
	ConfigFilePath string
}

type AppConfigField struct {
	Fields map[string]any
	Keys   []string
}

type ConfigPath struct {
	RootPath       string
	CmdPath        string
	AppConfigPath  string
	ConfigsPath    string
	PkgPath        string
	InternalPath   string
	ServicePath    string
	DomainPath     string
	RepositoryPath string
	AdapterPath    string
	HandlerPath    string
	SwaggerPath    string
	InfraPath      string
	ExternalPath   string
}

func init() {
	_, filename, _, _ := runtime.Caller(0)

	// Root directory
	AppConfig.RootPath = filepath.Dir(filepath.Dir(filename))

	// System paths
	AppConfig.ConfigsPath = (filepath.Join(AppConfig.RootPath, "configs"))
	AppConfig.CmdPath = (filepath.Join(AppConfig.RootPath, "cmd"))
	AppConfig.InternalPath = (filepath.Join(AppConfig.RootPath, "internal"))
	AppConfig.PkgPath = (filepath.Join(AppConfig.RootPath, "pkg"))
	AppConfig.InfraPath = (filepath.Join(AppConfig.RootPath, "infra"))
	AppConfig.SwaggerPath = (filepath.Join(AppConfig.RootPath, "docs"))

	// System paths inside internal
	AppConfig.ServicePath = (filepath.Join(AppConfig.InternalPath, "service"))
	AppConfig.DomainPath = (filepath.Join(AppConfig.InternalPath, "entity"))
	AppConfig.RepositoryPath = (filepath.Join(AppConfig.InternalPath, "repository"))
	AppConfig.AdapterPath = (filepath.Join(AppConfig.InternalPath, "adapter"))
	AppConfig.HandlerPath = (filepath.Join(AppConfig.InternalPath, "handler"))
	AppConfig.ExternalPath = filepath.Join(AppConfig.RootPath, "..")
}

type Connections struct {
	PathConfigFile string `mapstructure:"path_config_file"`
	Paths          *ConfigPath
	FileConfig     *FileConfig
	Fields map[string]interface{} 
	*AppConfigField
}

func LoadConfig() (*Connections, error) {

	if os.Getenv("PATH_CONFIG") != "" {
		AppConfig.ExternalPath = os.Getenv("PATH_CONFIG")
	}

	fc := FileConfig{
		ConfigPath:     AppConfig.ExternalPath,
		Extentsion:     "yaml",
		FileName:       "app-config",
		ConfigFilePath: AppConfig.ExternalPath,
	}

	var cfg *Connections

	viper.AddConfigPath(fc.ConfigPath)
	viper.SetConfigName(fc.FileName)
	viper.SetConfigType(fc.Extentsion)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err.(viper.ConfigFileNotFoundError)
		}
		return nil, err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	err = os.Setenv("JSON_CONFIG_PATH", fc.ConfigFilePath)
	if err != nil {
		return nil, err
	}

	log.Println("Config file loaded successfully")
	log.Println(viper.AllSettings())

	return &Connections{
		PathConfigFile: fc.ConfigFilePath,
		Paths:          &AppConfig,
		FileConfig:     &fc,
		Fields: viper.AllSettings(),
		AppConfigField: &AppConfigField{
			Fields: viper.AllSettings(),
			Keys:   viper.AllKeys(),
		},
	}, err
}