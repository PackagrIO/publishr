package scm

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/packagrio/go-common/config"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm/models"
	"github.com/packagrio/go-common/utils/git"
	gitUrl "github.com/whilp/git-urls"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type scmBase struct {
	PipelineData *pipeline.Data
}

// configure method will generate an authenticated client that can be used to comunicate with Github
// MUST set @git_parent_path
// MUST set @client field
func (s *scmBase) Init(pipelineData *pipeline.Data, myConfig config.BaseInterface, httpClient *http.Client) error {
	s.PipelineData = pipelineData

	//by default the current working directory is the local directory to execute in
	cwdPath, _ := os.Getwd()
	s.PipelineData.GitLocalPath = cwdPath
	s.PipelineData.GitParentPath = filepath.Dir(cwdPath)

	return nil
}

//We cant make any assumptions about the SCM or the environment. (No Pull requests or SCM env vars). So lets use native git methods to get
// the current repo status.
func (g *scmBase) RetrievePayload() (*models.Payload, error) {

	g.PipelineData.IsPullRequest = false

	//check the local git repo for relevant info
	remoteUrl, err := git.GitGetRemote(g.PipelineData.GitLocalPath, "origin")
	if err != nil {
		return nil, err
	}

	commit, err := git.GitGetHeadCommit(g.PipelineData.GitLocalPath)
	if err != nil {
		return nil, err
	}

	branch, err := git.GitGetBranch(g.PipelineData.GitLocalPath)
	if err != nil {
		return nil, err
	}

	// try to determine the repo name from the remote
	gitRemoteUrl, err := gitUrl.Parse(remoteUrl)
	gitRemotePath := strings.Trim(gitRemoteUrl.Path, "/")
	gitRemotePath = strings.TrimSuffix(gitRemotePath, ".git")
	gitRemoteParts := strings.Split(gitRemotePath, "/")
	var repoName string
	if len(gitRemoteParts) >= 2 {
		repoName = gitRemoteParts[len(gitRemoteParts)-1]
	} else {
		repoName = "placeholder"
	}
	return &models.Payload{
		Head: &pipeline.ScmCommitInfo{
			Sha: commit,
			Ref: branch,
			Repo: &pipeline.ScmRepoInfo{
				CloneUrl: remoteUrl,
				Name:     repoName,
				FullName: gitRemotePath,
			}},
	}, nil
}

func (g *scmBase) PopulatePipelineData(payload *models.Payload) error {
	//set the processed head info
	g.PipelineData.GitHeadInfo = payload.Head
	if err := g.PipelineData.GitHeadInfo.Validate(); err != nil {
		return err
	}
	if g.PipelineData.IsPullRequest {
		//pull requests need both HEAD and BASE info for processing.
		g.PipelineData.GitBaseInfo = payload.Base
		if err := g.PipelineData.GitBaseInfo.Validate(); err != nil {
			return err
		}
		g.PipelineData.GitRemote = payload.Base.Repo.CloneUrl
	} else {
		//this is not a PR, so use the HEAD for the CloneURL
		g.PipelineData.GitRemote = payload.Head.Repo.CloneUrl
	}
	return nil
}

func (s *scmBase) Publish() error {

	// Create a Version 4 UUID, panicking on error
	branchName := uuid.Must(uuid.NewV4())
	s.PipelineData.GitLocalBranch = branchName.String()

	//create a randomly named local branch based on the head commit.
	_, err := git.GitCreateBranchFromHead(s.PipelineData.GitLocalPath, s.PipelineData.GitLocalBranch)

	var destBranchName string
	if s.PipelineData.IsPullRequest {
		//the branch data is stored in the  "base"
		destBranchName = s.PipelineData.GitBaseInfo.Ref
	} else {
		//the branch info is stored in the "head"
		destBranchName = s.PipelineData.GitHeadInfo.Ref
	}

	perr := git.GitPush(s.PipelineData.GitLocalPath, s.PipelineData.GitLocalBranch, s.PipelineData.GitRemote, destBranchName, fmt.Sprintf("v%s", s.PipelineData.ReleaseVersion))
	if perr != nil {
		return perr
	}

	// calculate the release sha
	releaseCommit, err := git.GitGetHeadCommit(s.PipelineData.GitLocalPath)
	if err != nil {
		return err
	}
	s.PipelineData.ReleaseCommit = releaseCommit

	return nil
}

func (s *scmBase) parseReleaseAssetNames(releaseAssetItems []string) ([]pipeline.ScmReleaseAsset, error) {
	releaseAssets := []pipeline.ScmReleaseAsset{}

	for _, item := range releaseAssetItems {
		if item == "" {
			return nil, fmt.Errorf("invalid release asset path")
		}
		assetParts := strings.SplitN(item, ":", 2)
		itemLocalPath := assetParts[0]
		var itemArtifactName string
		if len(assetParts) == 2 && assetParts[1] != "" {
			itemArtifactName = assetParts[1]
		} else {
			itemArtifactName = filepath.Base(itemLocalPath)
		}

		releaseAsset := pipeline.ScmReleaseAsset{
			LocalPath:    itemLocalPath,
			ArtifactName: itemArtifactName,
			ContentType:  "application/octet-stream",
		}

		releaseAssets = append(releaseAssets, releaseAsset)
	}
	return releaseAssets, nil
}
