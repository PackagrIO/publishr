package engine

import (
	"encoding/json"
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/publishr/pkg/config"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

type engineChef struct {
	engineBase
	NextMetadata *metadata.ChefMetadata
	Scm          scm.Interface
}

func (g *engineChef) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Config = configData
	g.Scm = sourceScm
	g.PipelineData = pipelineData
	g.NextMetadata = new(metadata.ChefMetadata)
	return nil
}

func (g *engineChef) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineChef) ValidateTools() error {
	if _, kerr := exec.LookPath("knife"); kerr != nil {
		return errors.EngineValidateToolError("knife binary is missing")
	}
	// TODO, check for knife spork
	return nil
}

func (g *engineChef) PopulateNextMetadata() error {
	return g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath)
}

//private Helpers

func (g *engineChef) retrieveCurrentMetadata(gitLocalPath string) error {
	//dat, err := ioutil.ReadFile(path.Join(gitLocalPath, "metadata.rb"))
	//knife cookbook metadata -o ../ chef-mycookbook -- will generate a metadata.json file.
	if cerr := utils.BashCmdExec(fmt.Sprintf("knife cookbook metadata -o ../ %s", path.Base(gitLocalPath)), gitLocalPath, nil, ""); cerr != nil {
		return cerr
	}
	defer os.Remove(path.Join(gitLocalPath, "metadata.json"))

	//read metadata.json file.
	metadataContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, "metadata.json"))
	if rerr != nil {
		return rerr
	}

	if uerr := json.Unmarshal(metadataContent, g.NextMetadata); uerr != nil {
		return uerr
	}

	return nil
}
