package engine

import (
	"encoding/json"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/publishr/pkg/config"
	"io/ioutil"
	"os/exec"
	"path"
)

type engineNode struct {
	engineBase

	Scm          scm.Interface //Interface
	NextMetadata *metadata.NodeMetadata
}

func (g *engineNode) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = configData
	g.PipelineData = pipelineData
	g.NextMetadata = new(metadata.NodeMetadata)

	g.Config.SetDefault(config.PACKAGR_NPM_REGISTRY, "registry.npmjs.org")

	return nil
}

func (g *engineNode) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineNode) ValidateTools() error {

	if _, kerr := exec.LookPath("node"); kerr != nil {
		return errors.EngineValidateToolError("node binary is missing")
	}

	return nil
}

func (g *engineNode) PopulateNextMetadata() error {
	return g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath)
}

//private Helpers

func (g *engineNode) retrieveCurrentMetadata(gitLocalPath string) error {
	//read package.json file.
	packageContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, "package.json"))
	if rerr != nil {
		return rerr
	}

	if uerr := json.Unmarshal(packageContent, g.NextMetadata); uerr != nil {
		return uerr
	}

	return nil
}
