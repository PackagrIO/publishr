package pkg

import (
	"errors"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/packagrio/publishr/pkg/config"
	"github.com/packagrio/publishr/pkg/engine"
	"github.com/packagrio/publishr/pkg/mgr"
	"log"
	"os"
	"path"
	"path/filepath"
)

type Pipeline struct {
	Data           *pipeline.Data
	Config         config.Interface
	Scm            scm.Interface
	Engine 			engine.Interface
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

	if err := p.ParseRepoConfig(); err != nil {
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
	//by default the current working directory is the local directory to execute in
	cwdPath, _ := os.Getwd()
	p.Data.GitLocalPath = cwdPath
	p.Data.GitParentPath = filepath.Dir(cwdPath)

	// start the source, and whatever work needs to be done there.
	// MUST set options.GitParentPath
	log.Println("pipeline_init_step")
	scmImpl, serr := scm.Create(p.Config.GetString(config.PACKAGR_SCM), p.Data)
	if serr != nil {
		return serr
	}
	p.Scm = scmImpl

	return nil
}

func (p *Pipeline) ScmRetrievePayloadStep() (*scm.Payload, error) {

	log.Println("scm_retrieve_payload_step")
	payload, perr := p.Scm.RetrievePayload()
	if perr != nil {
		return nil, perr
	}

	return payload, nil
}

//func (p *Pipeline) ScmCheckoutPullRequestStep(payload *scm.Payload) error {
//
//	// PRE HOOK
//	if err := p.RunHook("scm_checkout_pull_request_step.pre"); err != nil {
//		return err
//	}
//
//	if p.Config.IsSet("scm_checkout_pull_request_step.override") {
//		if err := p.RunHook("scm_checkout_pull_request_step.override"); err != nil {
//			return err
//		}
//	} else {
//		log.Println("scm_checkout_pull_request_step")
//		if err := p.Scm.CheckoutPullRequestPayload(payload); err != nil {
//			return err
//		}
//	}
//
//	// POST HOOK
//	if err := p.RunHook("scm_checkout_pull_request_step.post"); err != nil {
//		return err
//	}
//	return nil
//}

//func (p *Pipeline) ScmCheckoutPushPayloadStep(payload *scm.Payload) error {
//
//	// PRE HOOK
//	if err := p.RunHook("scm_checkout_push_payload_step.pre"); err != nil {
//		return err
//	}
//
//	if p.Config.IsSet("scm_checkout_push_payload_step.override") {
//		if err := p.RunHook("scm_checkout_push_payload_step.override"); err != nil {
//			return err
//		}
//	} else {
//		log.Println("scm_checkout_push_payload_step")
//		if err := p.Scm.CheckoutPushPayload(payload); err != nil {
//			return err
//		}
//	}
//
//	// POST HOOK
//	if err := p.RunHook("scm_checkout_push_payload_step.post"); err != nil {
//		return err
//	}
//	return nil
//}

func (p *Pipeline) ParseRepoConfig() error {
	log.Println("parse_repo_config")
	// update the config with repo config file options
	repoConfig := path.Join(p.Data.GitLocalPath, p.Config.GetString(config.PACKAGR_ENGINE_REPO_CONFIG_PATH))
	if utils.FileExists(repoConfig) {
		if err := p.Config.ReadConfig(repoConfig); err != nil {
			return errors.New("An error occured while parsing repository capsule.yml file")
		}
	} else {
		log.Println("No repo capsule.yml file found, using existing config.")
	}

	if p.Config.IsSet("scm_release_assets") {
		//unmarshall config data.
		parsedAssets := new([]pipeline.ScmReleaseAsset)
		if err := p.Config.UnmarshalKey("scm_release_assets", parsedAssets); err != nil {
			return err
		}

		//append the parsed Assets to the current ReleaseAssets storage (incase assets were defined in system yml)
		p.Data.ReleaseAssets = append(p.Data.ReleaseAssets, (*parsedAssets)...)
	}
	return nil
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
		return err
	}

	return nil
}
