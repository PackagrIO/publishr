package scm

import (
	"github.com/packagrio/go-common/config"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm/models"
	"net/http"
)

type scmBitbucket struct {
	scmBase
}

// configure method will generate an authenticated client that can be used to comunicate with Github
// MUST set @git_parent_path
// MUST set @client field
func (b *scmBitbucket) Init(pipelineData *pipeline.Data, myConfig config.BaseInterface, httpClient *http.Client) error {
	return b.scmBase.Init(pipelineData, myConfig, httpClient)
}

func (b *scmBitbucket) RetrievePayload() (*models.Payload, error) {
	return nil, nil
}

func (b *scmBitbucket) Publish() error {
	return nil
}

func (g *scmBitbucket) PublishAssets(releaseData interface{}) error {
	return nil
}

func (g *scmBitbucket) Cleanup() error {
	return nil
}

func (b *scmBitbucket) SetEnvironmentalVariable(name string, value string) error {
	return nil
}

// To prepend a string to PATH
func (b *scmBitbucket) AddPath(path string) error {
	return nil
}

// To set an output for the step
func (b *scmBitbucket) SetOutput(name string, value string) error {
	return nil
}

// To mask a value in the logs
func (b *scmBitbucket) MaskSecret(secret string) error {
	return nil
}
