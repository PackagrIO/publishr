package git

import git2go "gopkg.in/libgit2/git2go.v25"

//Add all modified files to index, and commit.
func GitCommit(repoPath string, message string, signature *git2go.Signature) error {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return oerr
	}

	//get repo index.
	idx, ierr := repo.Index()
	if ierr != nil {
		return ierr
	}
	aerr := idx.AddAll([]string{}, git2go.IndexAddDefault, nil)
	if aerr != nil {
		return aerr
	}
	treeId, wterr := idx.WriteTree()
	if wterr != nil {
		return wterr
	}
	werr := idx.Write()
	if werr != nil {
		return werr
	}

	tree, lerr := repo.LookupTree(treeId)
	if lerr != nil {
		return lerr
	}

	currentBranch, berr := repo.Head()
	if berr != nil {
		return berr
	}

	commitTarget, terr := repo.LookupCommit(currentBranch.Target())
	if terr != nil {
		return terr
	}

	_, cerr := repo.CreateCommit("HEAD", signature, signature, message, tree, commitTarget)
	//if(cerr != nil){return cerr}

	return cerr
}

func GitGetHeadCommit(repoPath string) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}
	commitHead, herr := repo.Head()
	if herr != nil {
		return "", herr
	}

	commit, lerr := repo.LookupCommit(commitHead.Target())
	if lerr != nil {
		return "", lerr
	}
	return commit.Id().String(), nil
}
