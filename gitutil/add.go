package gitutil

import git "github.com/libgit2/git2go"

// gitAdd returns the tree object created by adding files to the index.
func gitAdd(repo *git.Repository, paths ...string) (*git.Tree, error) {
	indx, err := repo.Index()
	if err != nil {
		return nil, err
	}
	defer indx.Free()

	for _, file := range paths {
		err = indx.AddByPath(file)
		if err != nil {
			return nil, err
		}
	}

	err = indx.Write()
	if err != nil {
		return nil, err
	}

	treeID, err := indx.WriteTree()
	if err != nil {
		return nil, err
	}

	tree, err := repo.LookupTree(treeID)
	if err != nil {
		return nil, err
	}
	return tree, nil
}
