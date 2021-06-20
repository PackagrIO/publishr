package git

import (
	"fmt"
	"github.com/packagrio/go-common/errors"
	git2go "gopkg.in/libgit2/git2go.v25"
	"log"
)

// https://github.com/welaw/welaw/blob/100be9cf9a4c6d26f8126678c05072ff725202dd/pkg/easyrepo/merge.go#L11
// https://gist.github.com/danielfbm/37b0ca88b745503557b2b3f16865d8c3
// https://gist.github.com/danielfbm/ba4ae91efa96bb4771351bdbd2c8b06f
// https://github.com/Devying/git2go-example/blob/master/fetch1.go
//https://github.com/jandre/passward/blob/e37bce388cf6417d7123c802add1937574c2b30e/passward/git.go#L186-L206
// https://github.com/electricbookworks/electric-book-gui/blob/4d9ad588dbdf7a94345ef10a1bb6944bc2a2f69a/src/go/src/ebw/git/RepoConflict.go
func GitMergeRemoteBranch(repoPath string, localBranchName string, baseBranchName string, remoteUrl string, remoteBranchName string, signature *git2go.Signature) error {

	checkoutOpts := &git2go.CheckoutOpts{
		Strategy: git2go.CheckoutSafe | git2go.CheckoutRecreateMissing | git2go.CheckoutAllowConflicts | git2go.CheckoutUseTheirs,
	}

	//get current checked out repository.
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return oerr
	}

	// Lookup commmit for base branch
	baseBranch, err := repo.LookupBranch(baseBranchName, git2go.BranchLocal)
	if err != nil {
		log.Print("Failed to find local base branch: " + baseBranchName)
		return err
	}

	baseCommit, err := repo.LookupCommit(baseBranch.Target())
	if err != nil {
		log.Print(fmt.Sprintf("Failed to find head commit for base branch: %s", baseBranchName))
		return err
	}
	defer baseCommit.Free()

	// Check if there's a local branch with the pr_* name already.
	prLocalBranch, err := repo.LookupBranch(localBranchName, git2go.BranchLocal)
	// No local branch, lets create one
	if prLocalBranch == nil || err != nil {
		// Creating local pr branch from the base branch commit.
		prLocalBranch, err = repo.CreateBranch(localBranchName, baseCommit, false)
		if err != nil {
			log.Print("Failed to create local branch: " + localBranchName)
			return err
		}
	}

	// Getting the tree for the branch
	prLocalCommit, err := repo.LookupCommit(prLocalBranch.Target())
	if err != nil {
		log.Print("Failed to lookup for commit in local branch " + localBranchName)
		return err
	}
	//defer localCommit.Free()

	tree, err := repo.LookupTree(prLocalCommit.TreeId())
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
	herr := repo.SetHead("refs/heads/" + localBranchName)
	if herr != nil {
		return herr
	}

	//add a new remote for the PR head.
	prRemoteAlias := "pr_origin"
	prRemote, rerr := repo.Remotes.Create(prRemoteAlias, remoteUrl)
	if rerr != nil {
		return rerr
	}

	//fetch the commits for the remoteBranchName
	rferr := prRemote.Fetch([]string{"refs/heads/" + remoteBranchName}, new(git2go.FetchOptions), "")
	if rferr != nil {
		return rferr
	}

	remoteBranch, errRef := repo.References.Lookup(fmt.Sprintf("refs/remotes/%s/%s", prRemoteAlias, remoteBranchName))
	if errRef != nil {
		return errRef
	}
	remoteBranchID := remoteBranch.Target()

	//Assuming we are already checkout as the destination branch
	remotePrAnnCommit, err := repo.AnnotatedCommitFromRef(remoteBranch)
	if err != nil {
		log.Print("Failed get annotated commit from remote ")
		return err
	}
	defer remotePrAnnCommit.Free()

	//Getting repo HEAD
	head, err := repo.Head()
	if err != nil {
		log.Print("Failed get head ")
		return err
	}

	// Do merge analysis
	mergeHeads := make([]*git2go.AnnotatedCommit, 1)
	mergeHeads[0] = remotePrAnnCommit
	analysis, _, err := repo.MergeAnalysis(mergeHeads)

	if analysis&git2go.MergeAnalysisNone != 0 || analysis&git2go.MergeAnalysisUpToDate != 0 {
		log.Print("Found nothing to merge. This should not happen for valid PR's")
		return errors.ScmMergeNothingToMergeError("Found nothing to merge. This should not happen for valid PR's")
	} else if analysis&git2go.MergeAnalysisNormal != 0 {
		// Just merge changes

		//Options for merge
		mergeOpts, err := git2go.DefaultMergeOptions()
		if err != nil {
			return err
		}
		mergeOpts.FileFavor = git2go.MergeFileFavorNormal
		mergeOpts.TreeFlags = git2go.MergeTreeFailOnConflict

		//Options for checkout
		mergeCheckoutOpts := &git2go.CheckoutOpts{
			Strategy: git2go.CheckoutSafe | git2go.CheckoutRecreateMissing | git2go.CheckoutUseTheirs,
		}

		//Merge action
		if err = repo.Merge(mergeHeads, &mergeOpts, mergeCheckoutOpts); err != nil {
			log.Print("Failed to merge heads")
			// Check for conflicts
			index, err := repo.Index()
			if err != nil {
				log.Print("Failed to get repo index")
				return err
			}
			if index.HasConflicts() {
				log.Printf("Conflicts encountered. Please resolve them. %v", err)
				return errors.ScmMergeConflictError("Merge resulted in conflicts. Please solve the conflicts before merging.")
			}
			return err
		}

		//Getting repo Index
		index, err := repo.Index()
		if err != nil {
			log.Print("Failed to get repo index")
			return err
		}
		defer index.Free()

		//Checking for conflicts
		if index.HasConflicts() {
			return errors.ScmMergeConflictError("Merge resulted in conflicts. Please solve the conflicts before merging.")
		}

		// Make the merge commit

		// Get Write Tree
		treeId, err := index.WriteTree()
		if err != nil {
			return err
		}

		tree, err := repo.LookupTree(treeId)
		if err != nil {
			return err
		}

		localCommit, err := repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}

		remoteCommit, err := repo.LookupCommit(remoteBranchID)
		if err != nil {
			return err
		}

		repo.CreateCommit("HEAD", signature, signature, "", tree, localCommit, remoteCommit)
		// Clean up
		repo.StateCleanup()
	} else if analysis&git2go.MergeAnalysisFastForward != 0 {
		// Fast-forward changes
		// Get remote tree

		remoteTree, err := repo.LookupTree(remoteBranchID)
		if err != nil {
			return err
		}

		// Checkout
		if err := repo.CheckoutTree(remoteTree, nil); err != nil {
			return err
		}

		// Point branch to the object
		prLocalBranch.SetTarget(remoteBranchID, "")
		if _, err := head.SetTarget(remoteBranchID, ""); err != nil {
			return err
		}

	} else {
		log.Printf("Unexpected merge analysis result %d", analysis)
		return errors.ScmMergeAnalysisUnknownError(fmt.Sprintf("Unexpected merge analysis result: %d", analysis))
	}
	return nil

}
