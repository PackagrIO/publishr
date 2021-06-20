package git

import (
	"fmt"
	"github.com/packagrio/go-common/errors"
	git2go "gopkg.in/libgit2/git2go.v25"
	"log"
	"time"
)

// https://stackoverflow.com/questions/13638235/git-checkout-remote-reference
// https://gist.github.com/danielfbm/ba4ae91efa96bb4771351bdbd2c8b06f
// https://github.com/libgit2/git2go/issues/126
// https://www.atlassian.com/git/articles/pull-request-proficiency-fetching-abilities-unlocked
// https://www.atlassian.com/blog/archives/how-to-fetch-pull-requests
// https://stackoverflow.com/questions/48806891/bitbucket-does-not-update-refspec-for-pr-causing-jenkins-to-build-old-commits
func GitFetchPullRequest(repoPath string, pullRequestNumber string, localBranchName string, srcPatternTmpl string, destPatternTmpl string) error {

	//defaults for Templates if they are not specified.
	if len(srcPatternTmpl) == 0 {
		srcPatternTmpl = "refs/pull/%s/merge" //this default template is for Github
	}

	if len(destPatternTmpl) == 0 {
		destPatternTmpl = "refs/remotes/origin/pr/%s/merge"
	}

	//populate the templates
	srcPattern := fmt.Sprintf(srcPatternTmpl, pullRequestNumber)
	destPattern := fmt.Sprintf(destPatternTmpl, pullRequestNumber)
	refspec := fmt.Sprintf("+%s:%s", srcPattern, destPattern)

	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return oerr
	}

	checkoutOpts := &git2go.CheckoutOpts{
		Strategy: git2go.CheckoutSafe | git2go.CheckoutRecreateMissing | git2go.CheckoutAllowConflicts | git2go.CheckoutUseTheirs,
	}

	remote, lerr := repo.Remotes.Lookup("origin")
	if lerr != nil {
		log.Print("Failed to lookup origin remote")
		return lerr
	}
	time.Sleep(time.Second)

	// fetch the pull request merge and head references into this repo.
	ferr := remote.Fetch([]string{refspec}, new(git2go.FetchOptions), "")
	if ferr != nil {
		log.Print("Failed to fetch PR reference from remote")
		return ferr
	}

	// Get a reference to the PR merge branch in this repo
	prRef, err := repo.References.Lookup(destPattern)
	if err != nil {
		log.Print("Failed to find PR reference locally: " + destPattern)
		return err
	}

	// Lookup commmit for PR branch
	prCommit, err := repo.LookupCommit(prRef.Target())
	if err != nil {
		log.Print(fmt.Sprintf("Failed to find PR head commit: %s", prRef.Target()))
		return err
	}
	defer prCommit.Free()

	prLocalBranch, err := repo.LookupBranch(localBranchName, git2go.BranchLocal)
	// No local branch, lets create one
	if prLocalBranch == nil || err != nil {
		// Creating local branch
		prLocalBranch, err = repo.CreateBranch(localBranchName, prCommit, false)
		if err != nil {
			log.Print("Failed to create local branch: " + localBranchName)
			return err
		}
	}
	if prLocalBranch == nil {
		return errors.ScmFilesystemError("Error while locating/creating local branch")
	}
	defer prLocalBranch.Free()

	// Getting the tree for the branch
	localCommit, err := repo.LookupCommit(prLocalBranch.Target())
	if err != nil {
		log.Print("Failed to lookup for commit in local branch " + localBranchName)
		return err
	}
	//defer localCommit.Free()

	tree, err := repo.LookupTree(localCommit.TreeId())
	if err != nil {
		log.Print("Failed to lookup for tree " + localBranchName)
		return err
	}
	//defer tree.Free()

	// Checkout the tree
	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		log.Print("Failed to checkout tree " + localBranchName)
		return err
	}
	// Setting the Head to point to our branch
	return repo.SetHead("refs/heads/" + localBranchName)
}
