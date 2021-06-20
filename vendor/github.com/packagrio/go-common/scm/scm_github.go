package scm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/google/go-github/v32/github"
	"github.com/packagrio/go-common/config"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/pipeline"
	githubHelper "github.com/packagrio/go-common/scm/github"
	"github.com/packagrio/go-common/scm/models"
	"github.com/packagrio/go-common/utils/git"
	gitUrl "github.com/whilp/git-urls"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type scmGithub struct {
	scmBase
	Client *github.Client
	Config config.BaseInterface

	isGithubActionEnv bool
}

func (g *scmGithub) Init(pipelineData *pipeline.Data, myConfig config.BaseInterface, httpClient *http.Client) error {
	err := g.scmBase.Init(pipelineData, myConfig, httpClient)
	if err != nil {
		return err
	}

	g.Config = myConfig
	g.Config.SetDefault(config.PACKAGR_SCM_GITHUB_ACCESS_TOKEN_TYPE, "user")

	ctx := context.Background()

	//TODO: autopaginate turned on.
	if httpClient != nil {
		//primarily used for testing.
		g.Client = github.NewClient(httpClient)
	} else if githubToken, present := os.LookupEnv("GITHUB_TOKEN"); present && len(githubToken) > 0 {
		log.Printf("found GITHUB_TOKEN")
		g.Config.Set(config.PACKAGR_SCM_GITHUB_ACCESS_TOKEN, githubToken)
		if action, isAction := os.LookupEnv("GITHUB_ACTION"); isAction && len(action) > 0 {
			log.Printf("Running in a Github Action")
			//running as a github action.
			g.Config.Set(config.PACKAGR_SCM_GITHUB_ACCESS_TOKEN_TYPE, "app")
			g.isGithubActionEnv = true
		}
	} else if g.Config.IsSet(config.PACKAGR_SCM_GITHUB_ACCESS_TOKEN) {
		log.Printf("found PACKAGR_SCM_GITHUB_ACCESS_TOKEN")

		//already set, do nothing.
	} else {
		//no access token present
		return fmt.Errorf("github SCM requires an access token")
	}

	//create an authenticated client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.Config.GetString(config.PACKAGR_SCM_GITHUB_ACCESS_TOKEN)},
	)
	tc := oauth2.NewClient(ctx, ts)

	if g.Config.IsSet(config.PACKAGR_SCM_GITHUB_API_ENDPOINT) {
		gheClient, err := github.NewEnterpriseClient(
			g.Config.GetString(config.PACKAGR_SCM_GITHUB_API_ENDPOINT),
			g.Config.GetString(config.PACKAGR_SCM_GITHUB_API_ENDPOINT),
			tc,
		)
		if err != nil {
			return err
		}
		g.Client = gheClient
	} else {
		g.Client = github.NewClient(tc)
	}

	return nil
}

func (g *scmGithub) RetrievePayload() (*models.Payload, error) {

	if !g.isGithubActionEnv {
		// this is not a github action.
		// check if the user has provided information for us to determine if this is a push or a pull request

		if !g.Config.IsSet(config.PACKAGR_SCM_PULL_REQUEST) {
			log.Print("This is not a pull request.")

			payload, err := g.scmBase.RetrievePayload()
			if err != nil {
				return nil, err
			}

			if !g.Config.IsSet(config.PACKAGR_SCM_REPO_FULL_NAME) {
				g.Config.Set(config.PACKAGR_SCM_REPO_FULL_NAME, payload.Head.Repo.FullName)
			}
			return payload, nil
			//make this as similar to a pull request as possible
		} else {
			log.Print("This is a pull request")
			if !g.Config.IsSet(config.PACKAGR_SCM_REPO_FULL_NAME) {
				return nil, errors.ScmPayloadFormatError("repository full name must be provided for Github pull request processing")
			}
			g.PipelineData.IsPullRequest = true
			ctx := context.Background()
			parts := strings.Split(g.Config.GetString(config.PACKAGR_SCM_REPO_FULL_NAME), "/")
			pr, _, err := g.Client.PullRequests.Get(ctx, parts[0], parts[1], g.Config.GetInt(config.PACKAGR_SCM_PULL_REQUEST))

			if err != nil {
				return nil, errors.ScmAuthenticationFailed(fmt.Sprintf("Could not retrieve pull request from Github: %s", err))
			}

			//validate pullrequest
			if pr.GetState() != "open" {
				return nil, errors.ScmPayloadUnsupported("Pull request has an invalid action")
			}
			if pr.Base.Repo.GetDefaultBranch() != pr.Base.GetRef() {
				return nil, errors.ScmPayloadUnsupported(fmt.Sprintf("Pull request is not being created against the default branch of this repository (%s vs %s)", pr.Base.Repo.GetDefaultBranch(), pr.Base.GetRef()))
			}
			// check the payload push user.

			//TODO: figure out how to do optional authenication. possible options, Source USER, token based auth, no auth when used with capsulecd.com.
			// unless @source_client.collaborator?(payload['base']['repo']['full_name'], payload['user']['login'])
			//
			//   @source_client.add_comment(payload['base']['repo']['full_name'], payload['number'], CapsuleCD::BotUtils.pull_request_comment)
			//   fail CapsuleCD::Error::SourceUnauthorizedUser, 'Pull request was opened by an unauthorized user'
			// end

			return githubHelper.PayloadFromGithubPullRequest(*pr), nil
		}

	} else {
		//this is a github action, retrieve event data
		eventType := utils.GetEnv("GITHUB_EVENT_NAME", "push")
		eventPayloadPath, present := os.LookupEnv("GITHUB_EVENT_PATH")
		if !present {
			return nil, errors.ScmPayloadFormatError("Event Payload not present")
		}

		//set the fullname
		if !g.Config.IsSet(config.PACKAGR_SCM_REPO_FULL_NAME) {
			g.Config.Set(config.PACKAGR_SCM_REPO_FULL_NAME, utils.GetEnv("GITHUB_REPOSITORY", ""))
		}

		//open & parse JSON File
		jsonBytes, err := ioutil.ReadFile(eventPayloadPath)
		if err != nil {
			return nil, errors.ScmFilesystemError("Event Payload Path does not exist")
		}

		if eventType == "push" {
			var pushEvent github.PushEvent
			err := json.Unmarshal(jsonBytes, &pushEvent)
			if err != nil {
				return nil, errors.ScmPayloadFormatError(err.Error())
			}

			g.PipelineData.IsPullRequest = false
			return githubHelper.PayloadFromGithubPushEvent(pushEvent), nil
			//make this as similar to a pull request as possible
		} else if eventType == "pull_request" {

			//parse Pull Request event payload
			var pullRequestEvent github.PullRequestEvent
			err := json.Unmarshal(jsonBytes, &pullRequestEvent)
			if err != nil {
				return nil, errors.ScmPayloadFormatError(err.Error())
			}

			g.PipelineData.IsPullRequest = true
			return githubHelper.PayloadFromGithubPullRequest(*pullRequestEvent.GetPullRequest()), nil
		} else if eventType == "workflow_dispatch" {
			//parse Workflow Dispatch (manual) event payload
			var wfDispatchEvent githubHelper.WorkflowDispatchEvent
			err := json.Unmarshal(jsonBytes, &wfDispatchEvent)
			if err != nil {
				return nil, errors.ScmPayloadFormatError(err.Error())
			}

			g.PipelineData.IsPullRequest = false
			return githubHelper.PayloadFromGithubWorkflowDispatchEvent(wfDispatchEvent), nil

		} else {
			return nil, errors.ScmPayloadUnsupported("Unknown Event Type. Exiting.")
		}
	}
}

//
//func (g *scmGithub) PopulatePipelineData(payload *Payload) error {
//	err := g.scmBase.PopulatePipelineData(payload)
//	if err != nil {
//		return err
//	}
//	//this is a Github repo, so we can assume that the Clone URL is a https url
//
//	//var gitRemoteUsername string
//	//var gitRemotePassword string
//	//
//	//if g.Config.GetString("scm_github_access_token_type") == "app" {
//	//	// see https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/
//	//	gitRemoteUsername = "x-access-token"
//	//	gitRemotePassword = g.Config.GetString("scm_github_access_token")
//	//} else {
//	//	gitRemoteUsername = g.Config.GetString("scm_github_access_token")
//	//	gitRemotePassword = ""
//	//}
//	//
//	//authRemote, aerr := authGitRemote(g.PipelineData.GitBaseInfo.Repo.CloneUrl, gitRemoteUsername, gitRemotePassword)
//	//if aerr != nil {
//	//	return aerr
//	//}
//	//g.PipelineData.GitRemote = authRemote
//	//
//	//
//	//g.PipelineData.GitLocalBranch = g.PipelineData.GitHeadInfo.Ref
//}

func (g *scmGithub) Publish() error {

	//convert the remote url to an authenticated HTTP Remote
	httpGitRemote, err := g.authGitRemoteUrl(g.PipelineData.GitRemote)
	if err != nil {
		return err
	}
	g.PipelineData.GitRemote = httpGitRemote

	//do a basic publish (git push).
	err = g.scmBase.Publish()
	if err != nil {
		return err
	}

	//sleep because github needs time to process the new tag.
	time.Sleep(5 * time.Second)

	//the repo has already been pushed at this point, now we need to create a Github release.
	if g.Client == nil {
		log.Println("Skipping scm publish, no client credentials found")
		return nil
	}

	// try to find the nearest tag for this repo
	err = g.populateNearestTag()
	if err != nil {
		//ignore errors, we will just have an empty changelog.
	}

	// calculate the release sha
	releaseSha := utils.LeftPad2Len(g.PipelineData.ReleaseCommit, "0", 40)

	//get the release changelog
	// logic is complicated.
	// If this is a push we can only do a tag-tag Changelog
	// If this is a pull request we can do either
	// if disable_nearest_tag_changelog is true, we must attempt
	var releaseBody string = ""
	if g.PipelineData.GitNearestTag != nil && !g.Config.GetBool(config.PACKAGR_SCM_DISABLE_NEAREST_TAG_CHANGELOG) {
		releaseBody, _ = git.GitGenerateChangelog(
			g.PipelineData.GitLocalPath,
			g.PipelineData.GitNearestTag.TagShortName,
			g.PipelineData.GitLocalBranch,
		)
	}
	//fallback to using diff if pullrequest.
	if g.PipelineData.IsPullRequest && releaseBody == "" {
		releaseBody, _ = git.GitGenerateChangelog(
			g.PipelineData.GitLocalPath,
			g.PipelineData.GitBaseInfo.Sha,
			g.PipelineData.GitHeadInfo.Sha,
		)
	}

	//create release.
	ctx := context.Background()
	parts := strings.Split(g.Config.GetString(config.PACKAGR_SCM_REPO_FULL_NAME), "/")
	version := fmt.Sprintf("v%s", g.PipelineData.ReleaseVersion)

	log.Printf("Creating new release for `%s/%s` with version: `%s` on commit: `%s`. Commit message: `%s`", parts[0], parts[1], version, releaseSha, releaseBody)

	releaseData, _, rerr := g.Client.Repositories.CreateRelease(
		ctx,
		parts[0],
		parts[1],
		&github.RepositoryRelease{
			TargetCommitish: &releaseSha,
			Body:            &releaseBody,
			TagName:         &version,
			Name:            &version,
		},
	)
	if rerr != nil {
		return rerr
	}

	if perr := g.PublishAssets(releaseData.GetID()); perr != nil {
		log.Print("An error occured while publishing assets:")
		log.Print(perr)
		log.Print("Continuing...")
	}

	return nil
}

func (g *scmGithub) PublishAssets(releaseData interface{}) error {
	//releaseData should be an ID (int)
	releaseId, ok := releaseData.(int64)
	if !ok {
		return fmt.Errorf("Invalid releaseID, cannot upload assets")
	}

	ctx := context.Background()
	parts := strings.Split(g.Config.GetString(config.PACKAGR_SCM_REPO_FULL_NAME), "/")

	//populate the release assets
	releaseAssets, err := g.scmBase.parseReleaseAssetNames(g.Config.GetStringSlice(config.PACKAGR_SCM_RELEASE_ASSETS))
	if err != nil {
		return err
	}
	g.PipelineData.ReleaseAssets = releaseAssets
	for _, assetData := range g.PipelineData.ReleaseAssets {
		// handle templated destination artifact names
		artifactNamePopulated, aerr := utils.PopulateTemplate(assetData.ArtifactName, g.PipelineData)
		if aerr != nil {
			return aerr
		}

		localPathPopulated, lerr := utils.PopulateTemplate(assetData.LocalPath, g.PipelineData)
		if lerr != nil {
			return lerr
		}

		g.publishGithubAsset(
			g.Client,
			ctx,
			parts[0],
			parts[1],
			artifactNamePopulated,
			path.Join(g.PipelineData.GitLocalPath, localPathPopulated),
			releaseId,
			5)
	}
	return nil
}

func (g *scmGithub) Cleanup() error {

	if !g.Config.GetBool(config.PACKAGR_SCM_ENABLE_BRANCH_CLEANUP) { //Default is false, so this will just return without doing anything.
		// - exit if "scm_enable_branch_cleanup" is not true
		return errors.ScmCleanupFailed("scm_enable_branch_cleanup is false. Skipping cleanup")
	} else if !g.PipelineData.IsPullRequest {
		return errors.ScmCleanupFailed("scm cleanup unnecessary for push's. Skipping cleanup")
	} else if g.PipelineData.GitHeadInfo.Repo.FullName != g.PipelineData.GitBaseInfo.Repo.FullName {
		// exit if the HEAD PR branch is not in the same organization and repository as the BASE
		return errors.ScmCleanupFailed("HEAD PR branch is not in the same organization & repo as the BASE. Skipping cleanup")
	} else if g.Client == nil {
		return nil
	}

	ctx := context.Background()
	parts := strings.Split(g.PipelineData.GitBaseInfo.Repo.FullName, "/")

	repoData, _, err := g.Client.Repositories.Get(ctx, parts[0], parts[1])
	if err != nil {
		return err
	}

	if g.PipelineData.GitHeadInfo.Ref == repoData.GetDefaultBranch() || g.PipelineData.GitHeadInfo.Ref == "master" {
		//exit if the HEAD branch is the repo default branch
		//exit if the HEAD branch is master
		return errors.ScmCleanupFailed("HEAD PR branch is default repo branch, or master. Skipping cleanup")
	}

	_, drerr := g.Client.Git.DeleteRef(ctx, parts[0], parts[1], fmt.Sprintf("heads/%s", g.PipelineData.GitHeadInfo.Ref))
	if drerr != nil {
		return drerr
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////
// Github Specific Functionality (need to figure out a proper location for this code).
///////////////////////////////////////////////////////////////////////////

// see https://github.com/actions/toolkit/blob/main/docs/commands.md
func (g *scmGithub) SetEnvironmentalVariable(name string, value string) error {
	fmt.Printf("\n::set-env name=%s::%s\n", name, value)
	return nil
}

// To prepend a string to PATH
func (g *scmGithub) AddPath(path string) error {
	fmt.Printf("\n::add-path::%s\n", path)
	return nil
}

// To set an output for the step
func (g *scmGithub) SetOutput(name string, value string) error {
	fmt.Printf("\n::set-output name=%s::%s\n", name, value)
	return nil
}

// To mask a value in the logs
func (g *scmGithub) MaskSecret(secret string) error {
	fmt.Printf("\n::add-mask::%s\n", secret)
	return nil
}

//private

// since this is a GithubSCM, we're going to assume that we have a token of some sort.
// We're going to attempt to create a new authenticated https remote URL and use that for authentication, rather than
// the existing origin, incase its setup for ssh (public/private key) auth.
// git@github.com:AnalogJ/golang_analogj_test.git
// https://github.com/AnalogJ/golang_analogj_test.git
func (g *scmGithub) authGitRemoteUrl(originRemote string) (string, error) {
	originRemoteUrl, err := gitUrl.Parse(originRemote)
	if err != nil {
		return "", err
	}
	// now we need to override theURL parts, and put them together as a HTTP remote (with authentication).
	originRemoteUrl.Scheme = "https"

	var gitRemoteUsername string
	var gitRemotePassword string
	if g.Config.GetString(config.PACKAGR_SCM_GITHUB_ACCESS_TOKEN_TYPE) == "app" {
		// see https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/
		gitRemoteUsername = "x-access-token"
		gitRemotePassword = g.Config.GetString(config.PACKAGR_SCM_GITHUB_ACCESS_TOKEN)
	} else {
		gitRemoteUsername = g.Config.GetString(config.PACKAGR_SCM_GITHUB_ACCESS_TOKEN)
		gitRemotePassword = ""
	}

	if gitRemoteUsername != "" || gitRemotePassword != "" {
		originRemoteUrl.User = url.UserPassword(gitRemoteUsername, gitRemotePassword)
	}
	return originRemoteUrl.String(), nil
}

// the current commit should be a tag, we need to find the previous tag
func (g *scmGithub) populateNearestTag() error {
	//retrieve and store the nearestTag to this commit.
	nearestTag, err := git.GitFindNearestTagName(g.PipelineData.GitLocalPath)
	if err != nil {
		return nil // we dont care about failures finding the nearest tag, we'll just have an empty changelog.
	}

	tagDetails, err := git.GitGetTagDetails(g.PipelineData.GitLocalPath, nearestTag)
	if err != nil {
		return nil // we dont care about failures finding the nearest tag, we'll just have an empty changelog.
	}
	g.PipelineData.GitNearestTag = tagDetails
	return nil
}

func (g *scmGithub) publishGithubAsset(client *github.Client, ctx context.Context, repoOwner string, repoName string, assetName, filePath string, releaseID int64, retries int) error {

	log.Printf("Attempt (%d) to upload release asset %s from %s", retries, assetName, filePath)
	f, err := os.Open(filePath)
	if err != nil {
		log.Print(err)
		return err
	}

	_, _, err = client.Repositories.UploadReleaseAsset(ctx, repoOwner, repoName, releaseID, &github.UploadOptions{
		Name: assetName,
	}, f)

	if err != nil && retries > 0 {
		fmt.Println("artifact upload errored out, retrying in one second. Err:", err)
		time.Sleep(time.Second)
		err = g.publishGithubAsset(client, ctx, repoOwner, repoName, assetName, filePath, releaseID, retries-1)
	}

	return err
}
