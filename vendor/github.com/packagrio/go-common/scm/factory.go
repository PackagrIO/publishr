package scm

import (
	"fmt"
	"github.com/packagrio/go-common/config"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/pipeline"
	"net/http"
)

func Create(scmType string, pipelineData *pipeline.Data, config config.BaseInterface, client *http.Client) (Interface, error) {

	var scm Interface
	switch scmType {
	case "bitbucket":
		scm = new(scmBitbucket)
	case "github":
		scm = new(scmGithub)
	case "generic":
		scm = new(scmGeneric)
	default:
		return nil, errors.ScmUnspecifiedError(fmt.Sprintf("Unknown Scm Type: %s", scmType))
	}

	if err := scm.Init(pipelineData, config, client); err != nil {
		return nil, err
	}
	return scm, nil
}
