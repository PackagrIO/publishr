package config

import (
	"github.com/spf13/viper"
)

// Create mock using:
// mockgen -source=pkg/config/interface.go -destination=pkg/config/mock/mock_config.go
type Interface interface {
	Init() error
	ReadConfig(configFilePath string) error
	Set(key string, value interface{})
	SetDefault(key string, value interface{})
	AllSettings() map[string]interface{}
	IsSet(key string) bool
	Get(key string) interface{}
	GetBool(key string) bool
	GetInt(key string) int
	GetString(key string) string
	GetStringSlice(key string) []string
	UnmarshalKey(key string, rawVal interface{}, decoder ...viper.DecoderConfigOption) error
	GetBase64Decoded(key string) (string, error)
}

const PACKAGR_PACKAGE_TYPE = "package_type"
const PACKAGR_SCM = "scm"
const PACKAGR_SCM_LOCAL_BRANCH = "scm_local_branch"
const PACKAGR_SCM_REMOTE_BRANCH = "scm_remote_branch"
const PACKAGR_SCM_RELEASE_ASSETS = "scm_release_assets"
const PACKAGR_ENGINE_REPO_CONFIG_PATH = "engine_repo_config_path"
const PACKAGR_VERSION_METADATA_PATH = "version_metadata_path"
const PACKAGR_GENERIC_VERSION_TEMPLATE = "generic_version_template"
const PACKAGR_GENERIC_MERGE_VERSION_FILE = "generic_merge_version_file"
const PACKAGR_MGR_TYPE = "mgr_type"
const PACKAGR_NPM_REGISTRY = "npm_registry"
const PACKAGR_NPM_AUTH_TOKEN = "npm_auth_token"
const PACKAGR_CHEF_SUPERMARKET_USERNAME = "chef_supermarket_username"
const PACKAGR_CHEF_SUPERMARKET_TYPE = "chef_supermarket_type"
const PACKAGR_PYPI_REPOSITORY = "pypi_repository"
const PACKAGR_PYPI_USERNAME = "pypi_username"
const PACKAGR_PYPI_PASSWORD = "pypi_password"
const PACKAGR_RUBYGEMS_API_KEY = "rubygems_api_key"
