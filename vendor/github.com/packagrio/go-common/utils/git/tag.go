package git

import (
	"fmt"
	"github.com/packagrio/go-common/pipeline"
	git2go "gopkg.in/libgit2/git2go.v25"
	"log"
)

func GitTag(repoPath string, version string, message string, signature *git2go.Signature) (string, error) {
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

	//tagId, terr := repo.Tags.CreateLightweight(version, commit, false)
	tagId, terr := repo.Tags.Create(version, commit, signature, fmt.Sprintf("(%s) %s", version, message))
	if terr != nil {
		return "", terr
	}

	tagObj, terr := repo.LookupTag(tagId)
	if terr != nil {
		return "", terr
	}
	return tagObj.TargetId().String(), terr
}

func GitGetTagDetails(repoPath string, tagName string) (*pipeline.GitTagDetails, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return nil, oerr
	}

	id, aerr := repo.References.Dwim(tagName)
	if aerr != nil {
		return nil, aerr
	}
	tag, lerr := repo.LookupTag(id.Target()) //assume its an annotated tag.

	var currentTag *pipeline.GitTagDetails
	if lerr != nil {
		//this is a lightweight tag, not an annotated tag.
		commitRef, rerr := repo.LookupCommit(id.Target())
		if rerr != nil {
			return nil, rerr
		}

		author := commitRef.Author()

		log.Printf("Light-weight tag (%s) Commit ID: %s, DATE: %s", tagName, commitRef.Id().String(), author.When.String())

		currentTag = &pipeline.GitTagDetails{
			TagShortName: tagName,
			CommitSha:    commitRef.Id().String(),
			CommitDate:   author.When,
		}

	} else {

		log.Printf("Annotated tag (%s) Tag ID: %s, Commit ID: %s, DATE: %s", tagName, tag.Id().String(), tag.TargetId().String(), tag.Tagger().When.String())

		currentTag = &pipeline.GitTagDetails{
			TagShortName: tagName,
			CommitSha:    tag.TargetId().String(),
			CommitDate:   tag.Tagger().When,
		}
	}
	return currentTag, nil

}

// Get the nearest tag on branch.
// tag must be nearest, ie. sorted by their distance from the HEAD of the branch, not the date or tagname.
// basically `git describe --tags --abbrev=0`
func GitFindNearestTagName(repoPath string) (string, error) {
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return "", oerr
	}

	//get the previous commit
	ref, lerr := repo.References.Lookup("HEAD")
	if lerr != nil {
		return "", lerr
	}
	resRef, err := ref.Resolve()
	if err != nil {
		return "", err
	}
	headCommit, cerr := repo.LookupCommit(resRef.Target())
	if cerr != nil {
		return "", cerr
	}

	parentComit := headCommit.Parent(0)
	defer parentComit.Free()

	parentCommit, err := parentComit.AsCommit()
	if err != nil {
		return "", err
	}

	descOptions, derr := git2go.DefaultDescribeOptions()
	if derr != nil {
		return "", derr
	}
	descOptions.Strategy = git2go.DescribeTags
	//descOptions.Pattern = "HEAD^"

	formatOptions, ferr := git2go.DefaultDescribeFormatOptions()
	if ferr != nil {
		return "", ferr
	}
	formatOptions.AbbreviatedSize = 0

	descr, derr := parentCommit.Describe(&descOptions)
	if derr != nil {
		return "", derr
	}

	nearestTag, ferr := descr.Format(&formatOptions)
	if ferr != nil {
		return "", ferr
	}

	return nearestTag, nil
}
