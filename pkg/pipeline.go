package pkg

import (
	"errors"
	packagrErrors "github.com/packagrio/go-common/errors"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/go-common/scm/models"
	"github.com/packagrio/publishr/pkg/config"
	"github.com/packagrio/publishr/pkg/engine"
	"github.com/packagrio/publishr/pkg/mgr"
	"log"
	"path"
	"os"
	"path/filepath"
)

type Pipeline struct {
	Data           *pipeline.Data
	Config         config.Interface
	Scm            scm.Interface
	Engine         engine.Interface
	PackageManager mgr.Interface
}

func (p *Pipeline) Start(config config.Interface) error {
	// Initialize Pipeline.
	p.Config = config
	p.Data = new(pipeline.Data)

	if err := p.PipelineInitStep(); err != nil {
		return err
	}

	payload, err := p.ScmRetrievePayloadStep()
	if err != nil {
		return err
	}

	perr := p.ScmPopulatePipelineData(payload)
	if perr != nil {
		return perr
	}

	if err := p.ValidateTools(); err != nil {
		return err
	}

	if err := p.Engine.PopulateReleaseVersion(); err != nil {
		return err
	}

	if err := p.MgrInitStep(); err != nil {
		return err
	}

	if err := p.MgrValidateTools(); err != nil {
		return err
	}

	if err := p.MgrDistStep(); err != nil {
		return err
	}

	if err := p.ScmPublishStep(); err != nil {
		return err
	}

	if err := p.ScmCleanupStep(); err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) PipelineInitStep() error {
	// start the source, and whatever work needs to be done there.
	// MUST set options.GitParentPath
	cwdPath, _ := os.Getwd()
	p.Data.GitLocalPath = cwdPath
	p.Data.GitParentPath = filepath.Dir(cwdPath)

	if err := p.ParseRepoConfig(); err != nil {
		return err
	}


	log.Println("pipeline_init_step")
	scmImpl, serr := scm.Create(p.Config.GetString(config.PACKAGR_SCM), p.Data, p.Config, nil)
	if serr != nil {
		return serr
	}
	p.Scm = scmImpl

	//Generate a new instance of the engine
	engineImpl, eerr := engine.Create(p.Config.GetString(config.PACKAGR_PACKAGE_TYPE), p.Data, p.Config, p.Scm)
	if eerr != nil {
		return eerr
	}
	p.Engine = engineImpl

	return nil
}

func (p *Pipeline) ScmRetrievePayloadStep() (*models.Payload, error) {

	log.Println("scm_retrieve_payload_step")
	payload, perr := p.Scm.RetrievePayload()
	if perr != nil {
		return nil, perr
	}

	return payload, nil
}
func (p *Pipeline) ScmPopulatePipelineData(payload *models.Payload) error {

	log.Println("scm_populate_pipeline_data_step")
	return p.Scm.PopulatePipelineData(payload)
}

func (p *Pipeline) ParseRepoConfig() error {
	log.Println("parse_repo_config")
	// update the config with repo config file options
	repoConfig := path.Join(p.Data.GitLocalPath, p.Config.GetString(config.PACKAGR_ENGINE_REPO_CONFIG_PATH))
	if utils.FileExists(repoConfig) {
		if err := p.Config.ReadConfig(repoConfig); err != nil {
			return errors.New("An error occured while parsing repository packagr.yml file")
		}
	} else {
		log.Println("No repo capsule.yml file found, using existing config.")
	}

	return nil
}

func (p *Pipeline) ValidateTools() error {
	log.Println("validate_tools")
	return p.Engine.ValidateTools()
}

func (p *Pipeline) MgrInitStep() error {
	log.Println("mgr_init_step")
	if p.Config.IsSet("mgr_type") {
		manager, merr := mgr.Create(p.Config.GetString(config.PACKAGR_MGR_TYPE), p.Data, p.Config, nil)
		if merr != nil {
			return merr
		}
		p.PackageManager = manager
	} else {
		manager, merr := mgr.Detect(p.Config.GetString(config.PACKAGR_PACKAGE_TYPE), p.Data, p.Config, nil)
		if merr != nil {
			return merr
		}
		p.PackageManager = manager
	}
	return nil
}

func (p *Pipeline) MgrValidateTools() error {
	log.Println("mgr_validate_tools")
	return p.PackageManager.MgrValidateTools()
}

// this step should push the release to the package repository (ie. npm, chef supermarket, rubygems)
func (p *Pipeline) MgrDistStep() error {
	if p.Config.GetBool("mgr_disable_dist") {
		log.Println("skipping mgr_dist_step.pre, mgr_dist_step, mgr_dist_step.post")
		return nil
	}

	log.Println("mgr_dist_step")
	if err := p.PackageManager.MgrDistStep(p.Engine.GetNextMetadata()); err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) ScmPublishStep() error {
	if p.Config.GetBool("scm_disable_publish") {
		log.Println("skipping scm_publish_step.pre, scm_publish_step, scm_publish_step.post")
		return nil
	}

	log.Println("scm_publish_step")
	if err := p.Scm.Publish(); err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) ScmCleanupStep() error {
	if p.Config.GetBool("scm_disable_cleanup") {
		log.Println("skipping scm_cleanup_step.pre, scm_cleanup_step, scm_cleanup_step.post")
		return nil
	}

	log.Println("scm_cleanup_step")
	if err := p.Scm.Cleanup(); err != nil {
		if _, typeMatch := err.(packagrErrors.ScmCleanupFailed); !typeMatch {
			//not a cleanup failure message (which we can ignore), so expose this.
			return err
		}
	}

	return nil
}
