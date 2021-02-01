package engine

import (
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/publishr/pkg/config"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
)

type enginePython struct {
	engineBase

	Scm          scm.Interface //Interface
	NextMetadata *metadata.PythonMetadata
}

func (g *enginePython) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = configData
	g.PipelineData = pipelineData
	g.NextMetadata = new(metadata.PythonMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault(config.PACKAGR_VERSION_METADATA_PATH, "VERSION")
	g.Config.SetDefault(config.PACKAGR_PYPI_REPOSITORY, "https://upload.pypi.org/legacy/")

	return nil
}

func (g *enginePython) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *enginePython) ValidateTools() error {
	if _, berr := exec.LookPath("python"); berr != nil {
		return errors.EngineValidateToolError("python binary is missing")
	}

	return nil
}

func (g *enginePython) PopulateNextMetadata() error {
	return g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath)
}

//private Helpers

func (g *enginePython) retrieveCurrentMetadata(gitLocalPath string) error {
	//read metadata.json file.
	versionContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
	if rerr != nil {
		return rerr
	}
	g.NextMetadata.Version = strings.TrimSpace(string(versionContent))
	return nil
}
