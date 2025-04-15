package engine

import (
	"net/url"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/database"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func setupTestEngine(t *testing.T) *Engine {
	db, err := database.Open(&database.Config{}) // In-memory SQLite
	assert.NoError(t, err)
	err = db.AutoMigrate(&model.ActorInfo{})
	assert.NoError(t, err)
	engine := &Engine{
		db:      db,
		name:    DefaultEngineName,
		timeout: DefaultRequestTimeout,
	}
	engine.initLogger()
	return engine
}

type testActorProvider struct {
	lang language.Tag
}

func (s *testActorProvider) Name() string {
	return ""
}

func (s *testActorProvider) Priority() float64 {
	return 0.0
}

func (s *testActorProvider) SetPriority(v float64) {
}

func (s *testActorProvider) Language() language.Tag {
	return s.lang
}

func (s *testActorProvider) URL() *url.URL {
	return nil
}

func (s *testActorProvider) NormalizeActorID(id string) string {
	return ""
}

func (s *testActorProvider) ParseActorIDFromURL(rawURL string) (string, error) {
	return "", nil
}

func (s *testActorProvider) GetActorInfoByID(id string) (*model.ActorInfo, error) {
	return nil, nil
}

func (s *testActorProvider) GetActorInfoByURL(url string) (*model.ActorInfo, error) {
	return nil, nil
}

type testActorImageProvider struct {
	returnImages []string
}

func (s *testActorImageProvider) Name() string {
	return ""
}

func (s *testActorImageProvider) Priority() float64 {
	return 0.0
}

func (s *testActorImageProvider) SetPriority(v float64) {
}

func (s *testActorImageProvider) Language() language.Tag {
	return language.Japanese
}

func (s *testActorImageProvider) URL() *url.URL {
	return nil
}

func (s *testActorImageProvider) GetActorImagesByName(name string) ([]string, error) {
	return s.returnImages, nil
}

func TestGetActorInfoWithCallback(t *testing.T) {
	const notLazy = false

	engine := setupTestEngine(t)
	actorImageProvider := &testActorImageProvider{}

	engine.actorImageLanguageProviders = make(map[string][]mt.ActorImageProvider)
	engine.actorImageLanguageProviders[language.Japanese.String()] = []mt.ActorImageProvider{
		actorImageProvider,
	}

	actorProviderInJapanese := &testActorProvider{lang: language.Japanese}
	actorProviderInChinese := &testActorProvider{lang: language.Chinese}

	tests := []struct {
		name                           string
		callbackReturnActorInfo        *model.ActorInfo
		actorImageProviderReturnImages []string
		actorProvider                  mt.ActorProvider
		expectedError                  error
		expectedImages                 pq.StringArray
	}{
		{
			name:                           "CallbackCannotFindInfo",
			callbackReturnActorInfo:        nil,
			actorImageProviderReturnImages: nil,
			actorProvider:                  actorProviderInJapanese,
			expectedError:                  mt.ErrInfoNotFound,
			expectedImages:                 nil,
		},
		{
			name: "NoImageProviders",
			callbackReturnActorInfo: &model.ActorInfo{
				ID:       "id",
				Name:     "name",
				Provider: "provider",
				Homepage: "homepage",
				Images:   []string{"image1.jpg"},
			},
			actorImageProviderReturnImages: nil,
			actorProvider:                  actorProviderInChinese,
			expectedError:                  nil,
			expectedImages:                 []string{"image1.jpg"},
		},
		{
			name: "ImageProviderReturnsNothing",
			callbackReturnActorInfo: &model.ActorInfo{
				ID:       "id",
				Name:     "name",
				Provider: "provider",
				Homepage: "homepage",
				Images:   []string{"image1.jpg"},
			},
			actorImageProviderReturnImages: nil,
			actorProvider:                  actorProviderInJapanese,
			expectedError:                  nil,
			expectedImages:                 []string{"image1.jpg"},
		},
		{
			name: "ImageProviderReturnsItems",
			callbackReturnActorInfo: &model.ActorInfo{
				ID:       "id",
				Name:     "name",
				Provider: "provider",
				Homepage: "homepage",
				Images:   []string{"image1.jpg"},
			},
			actorImageProviderReturnImages: []string{"image2.jpg", "image3.jpg"},
			actorProvider:                  actorProviderInJapanese,
			expectedError:                  nil,
			expectedImages:                 []string{"image1.jpg", "image2.jpg", "image3.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorImageProvider.returnImages = tt.actorImageProviderReturnImages

			callback := func() (*model.ActorInfo, error) {
				return tt.callbackReturnActorInfo, nil
			}

			info, err := engine.getActorInfoWithCallback(tt.actorProvider, "id", notLazy, callback)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
				assert.Equal(t, tt.expectedImages, info.Images)
			}
		})
	}
}
