package git

import (
	"fmt"
	goUtils "github.com/analogj/go-util/utils"
	git2go "gopkg.in/libgit2/git2go.v25"
	"strings"
)

func GitGenerateChangelog(repoPath string, baseSha string, headSha string) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}

	markdown := goUtils.StripIndent(`Timestamp |  SHA | Message | Author
	------------- | ------------- | ------------- | -------------
	`)

	revWalk, werr := repo.Walk()
	if werr != nil {
		return "", werr
	}

	rerr := revWalk.PushRange(fmt.Sprintf("%s..%s", baseSha, headSha))
	if rerr != nil {
		return "", rerr
	}

	revWalk.Iterate(func(commit *git2go.Commit) bool {
		markdown += fmt.Sprintf("%s | %.8s | %s | %s\n", //TODO: this should have a link for the SHA.
			commit.Author().When.UTC().Format("2006-01-02T15:04Z"),
			commit.Id().String(),
			cleanCommitMessage(commit.Message()),
			commit.Author().Name,
		)
		return true
	})
	//for {
	//	err := revWalk.Next()
	//	if err != nil {
	//		break
	//	}
	//
	//	log.Info(gi.String())
	//}

	return markdown, nil
}

// helpers
func cleanCommitMessage(commitMessage string) string {
	commitMessage = strings.TrimSpace(commitMessage)
	if commitMessage == "" {
		return "--"
	}

	commitMessage = strings.Replace(commitMessage, "|", "/", -1)
	commitMessage = strings.Replace(commitMessage, "\n", " ", -1)

	return commitMessage
}
