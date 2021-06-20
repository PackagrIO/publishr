package git

import (
	git2go "gopkg.in/libgit2/git2go.v25"
	"time"
)

func GitSignature(authorName string, authorEmail string) *git2go.Signature {
	return &git2go.Signature{
		Name:  authorName,
		Email: authorEmail,
		When:  time.Now(),
	}
}
