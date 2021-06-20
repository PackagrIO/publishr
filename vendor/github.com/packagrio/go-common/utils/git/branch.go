package git

import (
	git2go "gopkg.in/libgit2/git2go.v25"
	"log"
)

func GitGetBranch(repoPath string) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}

	currentBranch, berr := repo.Head()
	if berr != nil {
		return "", berr
	}

	return currentBranch.Branch().Name()
}

func GitCreateBranchFromHead(repoPath string, localBranchName string) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}

	// Lookup head commit
	commitHead, herr := repo.Head()
	if herr != nil {
		return "", herr
	}

	commit, lerr := repo.LookupCommit(commitHead.Target())
	if lerr != nil {
		return "", lerr
	}
	newLocalBranch, err := repo.CreateBranch(localBranchName, commit, false)
	if err != nil {
		log.Print("Failed to create local branch: " + localBranchName)
		return "", err
	}
	return newLocalBranch.Name()
}
