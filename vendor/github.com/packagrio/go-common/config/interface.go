package config

import (
	"github.com/spf13/viper"
)

type BaseInterface interface {
	Init() error
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
}

const PACKAGR_SCM_PULL_REQUEST = "scm_pull_request"

const PACKAGR_SCM_GITHUB_ACCESS_TOKEN_TYPE = "scm_github_access_token_type"
const PACKAGR_SCM_GITHUB_ACCESS_TOKEN = "scm_github_access_token"
const PACKAGR_SCM_GITHUB_API_ENDPOINT = "scm_github_api_endpoint"
const PACKAGR_SCM_REPO_FULL_NAME = "scm_repo_full_name"
const PACKAGR_SCM_DISABLE_NEAREST_TAG_CHANGELOG = "scm_disable_nearest_tag_changelog"
const PACKAGR_SCM_ENABLE_BRANCH_CLEANUP = "scm_enable_branch_cleanup"
const PACKAGR_SCM_RELEASE_ASSETS = "scm_release_assets"
