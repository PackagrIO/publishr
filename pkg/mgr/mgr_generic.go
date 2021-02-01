package mgr

import (
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/publishr/pkg/config"
	"net/http"
)

func DetectGeneric(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) bool {
	return false
}

type mgrGeneric struct {
	Config       config.Interface
	PipelineData *pipeline.Data
	Client       *http.Client
}

func (m *mgrGeneric) Init(pipelineData *pipeline.Data, myconfig config.Interface, client *http.Client) error {
	m.PipelineData = pipelineData
	m.Config = myconfig

	if client != nil {
		//primarily used for testing.
		m.Client = client
	}

	return nil
}

func (m *mgrGeneric) MgrValidateTools() error {
	return nil
}

func (m *mgrGeneric) MgrDistStep(currentMetadata interface{}, nextMetadata interface{}) error {
	return nil
}
