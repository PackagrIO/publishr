// +build node

package mgr_test

import (
	"github.com/golang/mock/gomock"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/publishr/pkg/config"
	"github.com/packagrio/publishr/pkg/config/mock"
	"github.com/packagrio/publishr/pkg/mgr"
	"github.com/packagrio/publishr/pkg/mgr/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type MgrNodeNpmTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Mgr          *mock_mgr.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *MgrNodeNpmTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Mgr = mock_mgr.NewMockInterface(suite.MockCtrl)

}

func (suite *MgrNodeNpmTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMgrNodeNpm_TestSuite(t *testing.T) {
	suite.Run(t, new(MgrNodeNpmTestSuite))
}

func (suite *MgrNodeNpmTestSuite) TestMgrNodeNpmTestSuite_MgrDistStep_WithoutCredentials() {
	//setup
	suite.Config.EXPECT().IsSet(config.PACKAGR_NPM_AUTH_TOKEN).Return(false).MinTimes(1)

	//suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	mgrNodeNpm, err := mgr.Create("npm", suite.PipelineData, suite.Config, nil)
	require.NoError(suite.T(), err)
	currentVersion := new(metadata.NodeMetadata)
	nextVersion := new(metadata.NodeMetadata)

	//test
	berr := mgrNodeNpm.MgrDistStep(currentVersion, nextVersion)

	//assert
	require.Error(suite.T(), berr)
}
