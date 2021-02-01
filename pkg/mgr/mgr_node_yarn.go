package mgr

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/publishr/pkg/config"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func DetectNodeYarn(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	//theres no way to automatically determine if a project was created via Yarn (vs NPM)
	return false
}

type mgrNodeYarn struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrNodeYarn) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrNodeYarn) MgrValidateTools() error {
	if _, kerr := exec.LookPath("yarn"); kerr != nil {
		return errors.EngineValidateToolError("yarn binary is missing")
	}
	return nil
}

func (m *mgrNodeYarn) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
	if !m.Config.IsSet(config.PACKAGR_NPM_AUTH_TOKEN) {
		return errors.MgrDistCredentialsMissing("cannot deploy page to npm, credentials missing")
	}

	npmrcFile, _ := ioutil.TempFile("", ".npmrc")
	defer os.Remove(npmrcFile.Name())

	// write the .npmrc config jfile.
	npmrcContent := fmt.Sprintf(
		"//%s/:_authToken=%s",
		m.Config.GetString(config.PACKAGR_NPM_REGISTRY),
		m.Config.GetString(config.PACKAGR_NPM_AUTH_TOKEN),
	)

	if _, werr := npmrcFile.Write([]byte(npmrcContent)); werr != nil {
		return werr
	}

	//TODO: is it worth using the Yarn publish command as well?
	npmPublishCmd := fmt.Sprintf("npm --userconfig %s publish .", npmrcFile.Name())
	derr := utils.BashCmdExec(npmPublishCmd, m.PipelineData.GitLocalPath, nil, "")
	if derr != nil {
		return errors.MgrDistPackageError("npm publish failed. Check log for exact error")
	}
	return nil
}
