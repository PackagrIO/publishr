package github

import (
	"github.com/google/go-github/v32/github"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm/models"
)

func PayloadFromGithubPushEvent(pushEvent github.PushEvent) *models.Payload {
	return &models.Payload{
		Head: &pipeline.ScmCommitInfo{
			Sha: pushEvent.GetAfter(),
			Ref: pushEvent.GetRef(),
			Repo: &pipeline.ScmRepoInfo{
				CloneUrl: pushEvent.GetRepo().GetCloneURL(),
				Name:     pushEvent.GetRepo().GetName(),
				FullName: pushEvent.GetRepo().GetFullName(),
			}},
	}
}
