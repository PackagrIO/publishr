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

func DetectGolangGlide(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	glideyamlPath := path.Join(pipelineData.GitLocalPath, "glide.yaml")
	return utils.FileExists(glideyamlPath)
}

type mgrGolangGlide struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrGolangGlide) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrGolangGlide) MgrValidateTools() error {
	if _, kerr := exec.LookPath("glide"); kerr != nil {
		return errors.EngineValidateToolError("glide binary is missing")
	}
	return nil
}

func (m *mgrGolangGlide) MgrDistStep(nextMetadata interface{}) error {
	// no real packaging for golang.
	// libraries are stored in version control.
	return nil
}
