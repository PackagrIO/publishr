package mgr

import (
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/publishr/pkg/config"
	"net/http"
	"path"
)

func DetectGolangMod(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	gomodPath := path.Join(pipelineData.GitLocalPath, "go.mod")
	return utils.FileExists(gomodPath)
}

type mgrGolangMod struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrGolangMod) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrGolangMod) MgrValidateTools() error {
	//if _, kerr := exec.LookPath("dep"); kerr != nil {
	//	return errors.EngineValidateToolError("dep binary is missing")
	//}
	return nil
}

func (m *mgrGolangMod) MgrDistStep(nextMetadata interface{}) error {
	// no real packaging for golang.
	// libraries are stored in version control.
	return nil
}
