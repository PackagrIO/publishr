package config_test

import (
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/releasr/pkg/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfiguration_init_ShouldCorrectlyInitializeConfiguration(t *testing.T) {
	t.Parallel()

	//setup
	defer utils.UnsetEnv("PACKAGR_")()

	//test
	testConfig, _ := config.Create()

	//assert
	require.Equal(t, "generic", testConfig.GetString(config.PACKAGR_PACKAGE_TYPE), "should populate package_type with generic default")
	require.Equal(t, "default", testConfig.GetString(config.PACKAGR_SCM), "should populate scm with default")
}

//func TestConfiguration_init_EnvVariablesShouldLoadProperly(t *testing.T) {
//	//setup
//	os.Setenv("PACKAGR_VERSION_BUMP_TYPE", "major")
//
//	//test
//	testConfig, _ := config.Create()
//
//	//assert
//	require.Equal(t, "major", testConfig.GetString(config.PACKAGR_VERSION_BUMP_TYPE), "should populate Engine Version Bump Type from environmental variable")
//
//	//teardown
//	os.Unsetenv("PACKAGR_VERSION_BUMP_TYPE")
//}
