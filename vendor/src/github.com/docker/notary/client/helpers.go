package client

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/docker/notary/client/changelist"
	"github.com/endophage/gotuf"
	"github.com/endophage/gotuf/data"
	"github.com/endophage/gotuf/keys"
	"github.com/endophage/gotuf/store"
)

// Use this to initialize remote HTTPStores from the config settings
func getRemoteStore(baseURL, gun string, rt http.RoundTripper) (store.RemoteStore, error) {
	return store.NewHTTPStore(
		baseURL+"/v2/"+gun+"/_trust/tuf/",
		"",
		"json",
		"",
		"key",
		rt,
	)
}

func applyChangelist(repo *tuf.TufRepo, cl changelist.Changelist) error {
	changes := cl.List()
	var err error
	for _, c := range changes {
		if c.Scope() == "targets" {
			applyTargetsChange(repo, c)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func applyTargetsChange(repo *tuf.TufRepo, c changelist.Change) error {
	var err error
	meta := &data.FileMeta{}
	err = json.Unmarshal(c.Content(), meta)
	if err != nil {
		return nil
	}
	if c.Action() == changelist.ActionCreate {
		files := data.Files{c.Path(): *meta}
		_, err = repo.AddTargets("targets", files)
	} else if c.Action() == changelist.ActionDelete {
		err = repo.RemoveTargets("targets", c.Path())
	}
	if err != nil {
		return err
	}
	return nil
}

func nearExpiry(r *data.SignedRoot) bool {
	plus6mo := time.Now().AddDate(0, 6, 0)
	return r.Signed.Expires.Before(plus6mo)
}

func initRoles(kdb *keys.KeyDB, rootKey, targetsKey, snapshotKey, timestampKey data.PublicKey) error {
	rootRole, err := data.NewRole("root", 1, []string{rootKey.ID()}, nil, nil)
	if err != nil {
		return err
	}
	targetsRole, err := data.NewRole("targets", 1, []string{targetsKey.ID()}, nil, nil)
	if err != nil {
		return err
	}
	snapshotRole, err := data.NewRole("snapshot", 1, []string{snapshotKey.ID()}, nil, nil)
	if err != nil {
		return err
	}
	timestampRole, err := data.NewRole("timestamp", 1, []string{timestampKey.ID()}, nil, nil)
	if err != nil {
		return err
	}

	if err := kdb.AddRole(rootRole); err != nil {
		return err
	}
	if err := kdb.AddRole(targetsRole); err != nil {
		return err
	}
	if err := kdb.AddRole(snapshotRole); err != nil {
		return err
	}
	if err := kdb.AddRole(timestampRole); err != nil {
		return err
	}
	return nil
}
