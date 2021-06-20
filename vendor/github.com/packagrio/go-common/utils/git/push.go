package git

import (
	"fmt"
	git2go "gopkg.in/libgit2/git2go.v25"
	"strings"
)

func GitPush(repoPath string, localBranch string, remoteUrl string, remoteBranch string, tagName string) error {
	//- https://gist.github.com/danielfbm/37b0ca88b745503557b2b3f16865d8c3
	//- https://stackoverflow.com/questions/37026399/git2go-after-createcommit-all-files-appear-like-being-added-for-deletion
	repo, oerr := git2go.OpenRepository(repoPath)
	if oerr != nil {
		return oerr
	}

	// Push
	remote, rerr := repo.Remotes.CreateAnonymous(remoteUrl)
	if rerr != nil {
		return rerr
	}

	remoteCallbacks := git2go.RemoteCallbacks{
		//TODO: check if this is necessary.
		CertificateCheckCallback: func(cert *git2go.Certificate, valid bool, hostname string) git2go.ErrorCode {
			return 0
		},
		//CredentialsCallback: func(remoteUrl string, usernameFromUrl string, allowedTypes git2go.CredType) (git2go.ErrorCode, *git2go.Cred){
		//	log.Printf("Authenticating to git remote: %s (%s) [type:%d]", remoteUrl, usernameFromUrl, allowedTypes)
		//
		//
		//	i, cred := git2go.NewCredDefault()
		//
		//	if allowedTypes&git2go.CredTypeUserpassPlaintext != 0 {
		//		log.Printf("using user-paass")
		//		parsed, err := url.Parse(remoteUrl)
		//		if err != nil {
		//			return git2go.ErrorCode(-1), nil
		//		}
		//		password, _ := parsed.User.Password()
		//		i, cred = git2go.NewCredUserpassPlaintext(parsed.User.Username(),password)
		//		return git2go.ErrorCode(i), &cred
		//	}
		//	//if allowedTypes&git2go.CredTypeSshCustom != 0 {
		//	//	log.Printf("using ssh")
		//	//	i, cred = git2go.NewCredSshKey(usernameFromUrl,"/root/.ssh/id_rsa.pub","/root/.ssh/id_rsa","")
		//	//	return  git2go.ErrorCode(i), &cred
		//	//}
		//	//if allowedTypes&git2go.CredTypeSshKey != 0 {
		//	//	log.Printf("not implemented-sending agaent")
		//	//	//i, cred = git2go.NewCredSshKeyFromAgent("analogj")
		//	//	return git2go.ErrorCode(-1), nil
		//	//}
		//	//
		//	//if allowedTypes&git2go.CredTypeDefault == 0 {
		//	//	log.Printf("invalid-cred-type")
		//	//	return  git2go.ErrorCode(-1), nil
		//	//}
		//
		//	return git2go.ErrorCode(i), &cred
		//},

		SidebandProgressCallback: func(str string) git2go.ErrorCode {
			fmt.Printf("\rremote: %v", str)
			return 0
		},
	}

	//strip the fully qualified branch ref if present.
	localBranch = strings.TrimPrefix(localBranch, "refs/heads/")
	remoteBranch = strings.TrimPrefix(remoteBranch, "refs/heads/")
	pushSpecs := []string{
		fmt.Sprintf("refs/heads/%s:refs/heads/%s", localBranch, remoteBranch),
		fmt.Sprintf("refs/tags/%s:refs/tags/%s", tagName, tagName),
	}
	return remote.Push(pushSpecs, &git2go.PushOptions{RemoteCallbacks: remoteCallbacks})
}
