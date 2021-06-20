package github

import (
	"encoding/json"
	"github.com/analogj/go-util/utils"
	"github.com/google/go-github/v32/github"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm/models"
)

// TAKEN from: https://github.com/google/go-github/blob/master/github/event_types.go
// TODO: this is not yet available in master, once it is, we should remove this Struct.
//
// WorkflowDispatchEvent is triggered when someone triggers a workflow run on GitHub or
// sends a POST request to the create a workflow dispatch event endpoint.
//
// GitHub API docs: https://docs.github.com/en/developers/webhooks-and-events/webhook-events-and-payloads#workflow_dispatch
type WorkflowDispatchEvent struct {
	Inputs   json.RawMessage `json:"inputs,omitempty"`
	Ref      *string         `json:"ref,omitempty"`
	Workflow *string         `json:"workflow,omitempty"`

	// The following fields are only populated by Webhook events.
	Repo   *github.Repository   `json:"repository,omitempty"`
	Org    *github.Organization `json:"organization,omitempty"`
	Sender *github.User         `json:"sender,omitempty"`
}

func PayloadFromGithubWorkflowDispatchEvent(wfDispatchEvent WorkflowDispatchEvent) *models.Payload {
	return &models.Payload{
		Head: &pipeline.ScmCommitInfo{
			Sha: utils.GetEnv("GITHUB_SHA", ""),
			Ref: *wfDispatchEvent.Ref,
			Repo: &pipeline.ScmRepoInfo{
				CloneUrl: wfDispatchEvent.Repo.GetCloneURL(),
				Name:     wfDispatchEvent.Repo.GetName(),
				FullName: wfDispatchEvent.Repo.GetFullName(),
			}},
	}
}
