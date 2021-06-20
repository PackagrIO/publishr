package git

import (
	"fmt"
	goUtils "github.com/analogj/go-util/utils"
	"github.com/packagrio/go-common/errors"
	git2go "gopkg.in/libgit2/git2go.v25"
	"os"
	"path"
	"path/filepath"
)

// Clone a git repo into a local directory.
// Credentials need to be specified by embedding in gitRemote url.
// TODO: this pattern may not work on Bitbucket/GitLab
func GitClone(parentPath string, repositoryName string, gitRemote string) (string, error) {
	absPath, _ := filepath.Abs(path.Join(parentPath, repositoryName))

	if !goUtils.FileExists(absPath) {
		os.MkdirAll(absPath, os.ModePerm)
	} else {
		return "", errors.ScmFilesystemError(fmt.Sprintf("The local repository path already exists, this should never happen. %s", absPath))
	}

	_, err := git2go.Clone(gitRemote, absPath, new(git2go.CloneOptions))
	return absPath, err
}
