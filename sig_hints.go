package libkb

import (
	"github.com/keybase/go-jsonw"
)

type SigHint struct {
	sigId    *SigId
	remoteId string
	apiUrl   string
	humanUrl string
}

type SigHints struct {
	uid     UID
	version int
	hints   map[SigId]*SigHint
	dirty   bool
}

func NewSigHint(jw *jsonw.Wrapper) (sh *SigHint, err error) {
	sh = &SigHint{}
	sh.sigId, err = GetSigId(jw.AtKey("sig_id"), true)
	sh.remoteId, _ = jw.AtKey("remote_id").GetString()
	sh.apiUrl, _ = jw.AtKey("api_url").GetString()
	sh.humanUrl, _ = jw.AtKey("human_url").GetString()
	return
}

func (sh SigHints) Lookup(i SigId) *SigHint {
	obj, _ := sh.hints[i]
	return obj
}

func NewSigHints(jw *jsonw.Wrapper, uid UID, dirty bool) (sh *SigHints, err error) {
	sh = &SigHints{uid: uid, dirty: dirty, version: 0}
	err = sh.PopulateWith(jw)
	if err != nil {
		sh = nil
	}
	return
}

func (sh *SigHints) PopulateWith(jw *jsonw.Wrapper) (err error) {

	if jw == nil || jw.IsNil() {
		return
	}

	jw.AtKey("version").GetIntVoid(&sh.version, &err)
	if err != nil {
		return
	}

	sh.hints = make(map[SigId]*SigHint)
	var n int
	n, err = jw.AtKey("hints").Len()
	if err != nil {
		return
	}

	for i := 0; i < n; i++ {
		hint, tmpe := NewSigHint(jw.AtKey("hints").AtIndex(i))
		if tmpe != nil {
			G.Log.Warning("Bad SigHint Loaded: %s", tmpe.Error())
		} else {
			sh.hints[*hint.sigId] = hint
		}
	}
	return
}

func (sh SigHint) MarshalToJson() *jsonw.Wrapper {
	ret := jsonw.NewDictionary()
	ret.SetKey("sig_id", jsonw.NewString(sh.sigId.ToString(true)))
	ret.SetKey("remote_id", jsonw.NewString(sh.remoteId))
	ret.SetKey("api_url", jsonw.NewString(sh.apiUrl))
	ret.SetKey("human_url", jsonw.NewString(sh.humanUrl))
	return ret
}

func (sh SigHints) MarshalToJson() *jsonw.Wrapper {
	ret := jsonw.NewDictionary()
	ret.SetKey("version", jsonw.NewInt(sh.version))
	ret.SetKey("hints", jsonw.NewArray(len(sh.hints)))
	i := 0
	for _, v := range sh.hints {
		ret.AtKey("hints").SetIndex(i, v.MarshalToJson())
		i++
	}
	return ret
}

func (sh *SigHints) Store() (err error) {
	G.Log.Debug("+ SigHints.Store() for uid=%s", string(sh.uid))
	if sh.dirty {
		err = G.LocalDb.Put(
			DbKey{Typ: DB_SIG_HINTS, Key: string(sh.uid)},
			[]DbKey{},
			sh.MarshalToJson(),
		)
		sh.dirty = false
	} else {
		G.Log.Debug("| SigHints.Store() skipped; wasn't dirty")
	}
	G.Log.Debug("- SigHints.Store() for uid=%s -> %v", string(sh.uid), ErrToOk(err))
	return err
}

func LoadSigHints(uid UID) (sh *SigHints, err error) {
	G.Log.Debug("+ LoadSigHints(%s)", string(uid))
	var jw *jsonw.Wrapper
	jw, err = G.LocalDb.Get(DbKey{Typ: DB_SIG_HINTS, Key: string(uid)})
	if err != nil {
		return
	}
	sh, err = NewSigHints(jw, uid, false)
	if err == nil {
		G.Log.Debug("| SigHints loaded @v%d", sh.version)
	}
	G.Log.Debug("- LoadSigHints(%s)", string(uid))
	return
}

func (sh *SigHints) Refresh() error {
	G.Log.Debug("+ Refresh SigHints() for uid=%s", string(sh.uid))
	res, err := G.API.Get(ApiArg{
		Endpoint:    "sig/hints",
		NeedSession: false,
		Args: HttpArgs{
			"uid": S{string(sh.uid)},
			"low": I{sh.version},
		},
	})
	if err != nil {
		return err
	}
	var n int
	n, err = res.Body.AtKey("hints").Len()
	if err != nil {
		return err
	}
	if n == 0 {
		G.Log.Debug("| No changes; version %d was up-to-date", sh.version)
	} else if err = sh.PopulateWith(res.Body); err != nil {
		return err
	} else {
		sh.dirty = true
	}
	G.Log.Debug("- Refresh SigHints() for uid=%s", string(sh.uid))
	return nil
}

func LoadAndRefreshSigHints(uid UID) (sh *SigHints, err error) {
	sh, err = LoadSigHints(uid)
	if err == nil {
		err = sh.Refresh()
	}
	return
}
