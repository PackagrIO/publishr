package scm

import (
	"github.com/packagrio/go-common/config"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm/models"
	"net/http"
)

// Create mock using:
// mockgen -source=scm/interface.go -destination=scm/mock/mock_scm.go
type Interface interface {

	// init method will generate an authenticated client that can be used to comunicate with Scm
	// MUST set pipelineData.GitParentPath
	// MUST set pipelineData.GitLocalPath
	Init(pipelineData *pipeline.Data, config config.BaseInterface, client *http.Client) error

	// Should correctly parse Github Events, and fallback to reading the local repository
	// If Github Event:
	// - Determine if this is a pull request or a push.
	//   - if it's a pull request the scm must retrieve the pull request payload and return it
	//   - if its a push, the scm must retrieve the push payload and return it
	// MUST set pipelineData.IsPullRequest
	RetrievePayload() (*models.Payload, error)

	//Populate the pipeline data using data retrieved from Git repo or APIs
	// if PUSH
	// MUST set pipelienData.GitRemote
	// MUST set pipelineData.GitHeadInfo
	// SHOULD set pipelineData.NearestTagDetails
	//
	// if PULL_REQUEST
	// MUST set pipelineData.GitRemote
	// MUST set pipelineData.GitBaseInfo
	// MUST set pipelineData.GitHeadInfo
	// SHOULD set pipelineData.NearestTagDetails
	PopulatePipelineData(payload *models.Payload) error

	// The local & remote repository should now contain code that has been the merged, tested and version bumped.
	// this step should also do any scm specific releases (github release, asset uploading, etc)
	// CAN override
	// MUST set pipelineData.GitLocalBranch
	// REQUIRES pipelineData.GitRemote
	// REQUIRES config.scm_repo_full_name
	// REQUIRES pipelineData.ScmReleaseCommit
	// REQUIRES pipelineData.GitLocalPath
	// REQUIRES pipelineData.GitBaseInfo
	// REQUIRES pipelineData.GitHeadInfo
	// REQUIRES pipelineData.ReleaseArtifacts
	// REQUIRES pipelineData.ReleaseVersion
	// REQUIRES pipelineData.ReleaseCommit
	// REQUIRES pipelineData.GitParentPath
	// USES set pipelineData.NearestTagDetails
	Publish() error //create release.

	//Upload assets to SCM, and attach to SCM release if possible.
	//Failing to upload Assets to SCM will not fail the publish (we'll retry 5 times)
	//Should not be called directly, will be called via Publish()
	//ReleaseData will be different for each SCM, but is probably a release ID that we can attach files to.
	//REQUIRES config.scm_repo_full_name
	//REQUIRES pipelineData.ReleaseAssets
	//REQUIRES pipelineData.GitLocalPath
	PublishAssets(releaseData interface{}) error

	// optionally delete the PR branch after the code has been merged into master.
	// only do so if:
	// - "scm_enable_branch_cleanup" is true
	// - HEAD PR branch is in the same repository as the BASE
	// - branch is not the default branch or "master" for this repository
	// - branch is not protected (SCM specific feature)
	// CAN override
	// USES scm_enable_branch_cleanup
	// REQUIRES config.scm_repo_full_name
	// REQUIRES pipelineData.GitBaseInfo.Repo.FullName
	// REQUIRES pipelineData.GitHeadInfo.Repo.FullName
	// REQUIRES pipelineData.GitHeadInfo.Ref
	Cleanup() error

	///////////////////////////////////////////////////////////////////////////
	// Github Specific Functionality (need to figure out a proper location for this code).
	///////////////////////////////////////////////////////////////////////////

	// To set an environment variable for future out of process steps
	SetEnvironmentalVariable(name string, value string) error

	// To prepend a string to PATH
	AddPath(path string) error

	// To set an output for the step
	SetOutput(name string, value string) error

	// To mask a value in the logs
	MaskSecret(secret string) error
}
