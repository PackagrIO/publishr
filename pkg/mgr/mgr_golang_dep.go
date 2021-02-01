package mgr

import (
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/publishr/pkg/config"
	"net/http"
	"os/exec"
	"path"
)

func DetectGolangDep(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	gopkgPath := path.Join(pipelineData.GitLocalPath, "Gopkg.toml")
	return utils.FileExists(gopkgPath)
}

type mgrGolangDep struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrGolangDep) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrGolangDep) MgrValidateTools() error {
	if _, kerr := exec.LookPath("dep"); kerr != nil {
		return errors.EngineValidateToolError("dep binary is missing")
	}
	return nil
}

func (m *mgrGolangDep) MgrDistStep(nextMetadata interface{}) error {
	// no real packaging for golang.
	// libraries are stored in version control.
	return nil
}
