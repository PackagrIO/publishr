package scm

import (
	"github.com/packagrio/go-common/config"
	"github.com/packagrio/go-common/pipeline"
	"net/http"
)

//Generic SCM makes no assumptions about a cloud SCM (like Bitbucket, Github, Gitlab). Instead it'll just use native git commands
type scmGeneric struct {
	scmBase
}

// configure method will generate an authenticated client that can be used to comunicate with Github
// MUST set @git_parent_path
// MUST set @client field
func (s *scmGeneric) Init(pipelineData *pipeline.Data, myConfig config.BaseInterface, httpClient *http.Client) error {
	return s.scmBase.Init(pipelineData, myConfig, httpClient)
}

// comes from scmBase
//func (s *scmGeneric) RetrievePayload() (*Payload, error) {
//	return s.scmBase.RetrievePayload()
//}

// use scmBase
//func (s *scmGeneric) Publish() error {
//	return s.scmBase.Publish()
//}

func (s *scmGeneric) PublishAssets(releaseData interface{}) error {
	return nil
}

func (s *scmGeneric) Cleanup() error {
	return nil
}

func (s *scmGeneric) SetEnvironmentalVariable(name string, value string) error {
	return nil
}

// To prepend a string to PATH
func (s *scmGeneric) AddPath(path string) error {
	return nil
}

// To set an output for the step
func (s *scmGeneric) SetOutput(name string, value string) error {
	return nil
}

// To mask a value in the logs
func (s *scmGeneric) MaskSecret(secret string) error {
	return nil
}
