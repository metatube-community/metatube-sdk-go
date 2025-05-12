package engine

import (
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/engine/providerid"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/gfriends"
)

func (e *Engine) getActorInfoWithCallback(provider mt.ActorProvider, id string, lazy bool, callback func() (*model.ActorInfo, error)) (info *model.ActorInfo, err error) {
	defer func() {
		// metadata validation check.
		if err == nil && (info == nil || !info.IsValid()) {
			err = mt.ErrIncompleteMetadata
		}
	}()
	if provider.Name() == gfriends.Name {
		return provider.GetActorInfoByID(id)
	}
	defer func() {
		// gfriends actor image injection for JAV actor providers.
		if err == nil && info != nil && provider.Language() == language.Japanese {
			if gInfo, gErr := e.MustGetActorProviderByName(gfriends.Name).GetActorInfoByID(info.Name); gErr == nil && len(gInfo.Images) > 0 {
				info.Images = append(gInfo.Images, info.Images...)
			}
		}
	}()
	// Query DB first (by id).
	if lazy {
		if info, err = e.db.GetActorInfo(
			providerid.ProviderID{
				Provider: provider.Name(), ID: id,
			}); err == nil && info.IsValid() {
			return
		}
	}
	// Delayed info auto-save.
	defer func() {
		if err == nil {
			// Make sure we save the original info here.
			_ = e.db.SaveActorInfo(info) // ignore error
		}
	}()
	return callback()
}

func (e *Engine) getActorInfoByProviderID(provider mt.ActorProvider, id string, lazy bool) (*model.ActorInfo, error) {
	if id = provider.NormalizeActorID(id); id == "" {
		return nil, mt.ErrInvalidID
	}
	return e.getActorInfoWithCallback(provider, id, lazy, func() (*model.ActorInfo, error) {
		return provider.GetActorInfoByID(id)
	})
}

func (e *Engine) GetActorInfoByProviderID(pid providerid.ProviderID, lazy bool) (*model.ActorInfo, error) {
	provider, err := e.GetActorProviderByName(pid.Provider)
	if err != nil {
		return nil, err
	}
	return e.getActorInfoByProviderID(provider, pid.ID, lazy)
}

func (e *Engine) getActorInfoByProviderURL(provider mt.ActorProvider, rawURL string, lazy bool) (*model.ActorInfo, error) {
	id, err := provider.ParseActorIDFromURL(rawURL)
	switch {
	case err != nil:
		return nil, err
	case id == "":
		return nil, mt.ErrInvalidURL
	}
	return e.getActorInfoWithCallback(provider, id, lazy, func() (*model.ActorInfo, error) {
		return provider.GetActorInfoByURL(rawURL)
	})
}

func (e *Engine) GetActorInfoByURL(rawURL string, lazy bool) (*model.ActorInfo, error) {
	provider, err := e.GetActorProviderByURL(rawURL)
	if err != nil {
		return nil, err
	}
	return e.getActorInfoByProviderURL(provider, rawURL, lazy)
}
