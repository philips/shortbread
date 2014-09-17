package gitutil

import (
	"errors"

	git "github.com/coreos/shortbread/Godeps/_workspace/src/github.com/libgit2/git2go"
)

// Push has the same effect as 'git push origin master'
func Push(repo *git.Repository) error {
	remoteRepo, err := repo.LoadRemote("origin")
	if err != nil {
		return err
	}
	defer remoteRepo.Free()

	err = remoteRepo.SetPushRefspecs([]string{MasterRef + ":" + MasterRef})
	if err != nil {
		return err
	}

	credCallBack := &git.RemoteCallbacks{}
	credCallBack.CredentialsCallback = getCredentials
	err = remoteRepo.SetCallbacks(credCallBack)
	if err != nil {
		return err
	}

	err = remoteRepo.Save()
	if err != nil {
		return err
	}

	pushObj, err := remoteRepo.NewPush()
	if err != nil {
		return err
	}
	defer pushObj.Free()

	err = pushObj.AddRefspec(MasterRef + ":" + MasterRef)
	if err != nil {
		return err
	}

	err = pushObj.Finish()
	if err != nil {
		return err
	}

	ok := pushObj.UnpackOk()
	if !ok {
		return errors.New("objects from push not unpacked properly")

	}

	StatusForEachCallback := func(ref string, msg string) int {
		if msg != "" {
			return -1
		}
		return 0
	}

	err = pushObj.StatusForeach(StatusForEachCallback)
	if err != nil {
		return err
	}

	sig, err := getSignatureFromLastCommit(repo)
	if err != nil {
		return err
	}

	err = pushObj.UpdateTips(sig, "update by push")
	return nil
}

func getSignatureFromLastCommit(repo *git.Repository) (*git.Signature, error) {
	master, err := repo.LookupReference(MasterRef)
	if err != nil {
		return nil, err
	}
	defer master.Free()

	objId := master.Target()
	c, err := repo.LookupCommit(objId)
	if err != nil {
		return nil, err
	}
	defer c.Free()

	return c.Committer(), nil
}
