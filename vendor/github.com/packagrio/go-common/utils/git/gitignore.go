package git

import (
	stderrors "errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

func GitGenerateGitIgnore(repoPath string, ignoreType string) error {
	//https://github.com/GlenDC/go-gitignore/blob/master/gitignore/provider/github.go

	gitIgnoreBytes, err := getGitIgnore(ignoreType)
	if err != nil {
		return err
	}

	gitIgnorePath := filepath.Join(repoPath, ".gitignore")
	return ioutil.WriteFile(gitIgnorePath, gitIgnoreBytes, 0644)
}

// helpers

func getGitIgnore(languageName string) ([]byte, error) {
	gitURL := fmt.Sprintf("https://raw.githubusercontent.com/github/gitignore/master/%s.gitignore", languageName)

	resp, err := http.Get(gitURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, stderrors.New(fmt.Sprintf("Could not find .gitignore for '%s'", languageName))
	}

	return ioutil.ReadAll(resp.Body)
}
