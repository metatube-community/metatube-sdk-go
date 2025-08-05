package testkit

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

type internalTestSuite struct {
	t *testing.T
}

func (s *internalTestSuite) T() *testing.T {
	return s.t
}

func (s *internalTestSuite) testItems(items []string, call func(*testing.T, string)) {
	for i, item := range items {
		s.T().Run(item, func(t *testing.T) {
			call(t, item)
		})
		if i < len(items)-1 {
			// Add random delay milliseconds.
			time.Sleep(time.Duration(rand.Intn(500)+500) * time.Millisecond)
		}
	}
}

func (s *internalTestSuite) testGetInfo(f any, items []string, vfs ...ValidateFunc) {
	ff := func(item string) (info interface{ IsValid() bool }, err error) {
		switch v := f.(type) {
		case func(string) (*model.ActorInfo, error):
			info, err = v(item)
		case func(string) (*model.MovieInfo, error):
			info, err = v(item)
		default:
			return nil, fmt.Errorf("invalid function type: %T", v)
		}
		return
	}
	s.testItems(items, func(t *testing.T, item string) {
		info, err := ff(item)
		require.NoError(t, err)
		require.NotNil(t, info)
		for _, vf := range append([]ValidateFunc{
			logJSONContent(),
			assertIsValid(),
		}, vfs...) {
			vf(t, info)
		}
	})
}

func (s *internalTestSuite) TestGetActorInfoByID(p mt.ActorProvider, items []string, vfs ...ValidateFunc) {
	s.testGetInfo(p.GetActorInfoByID, items, vfs...)
}

func (s *internalTestSuite) TestGetActorInfoByURL(p mt.ActorProvider, items []string, vfs ...ValidateFunc) {
	s.testGetInfo(p.GetActorInfoByURL, items, vfs...)
}

func (s *internalTestSuite) TestGetMovieInfoByID(p mt.MovieProvider, items []string, vfs ...ValidateFunc) {
	s.testGetInfo(p.GetMovieInfoByID, items, vfs...)
}

func (s *internalTestSuite) TestGetMovieInfoByURL(p mt.MovieProvider, items []string, vfs ...ValidateFunc) {
	s.testGetInfo(p.GetMovieInfoByURL, items, vfs...)
}

func (s *internalTestSuite) testGetMovieReviews(f func(string) ([]*model.MovieReviewDetail, error), items []string, vfs ...ValidateFunc) {
	s.testItems(items, func(t *testing.T, item string) {
		reviews, err := f(item)
		require.NoError(t, err)
		require.NotEmpty(t, reviews)
		for _, vf := range append([]ValidateFunc{
			logJSONContent(),
			assertIsValid(),
		}, vfs...) {
			vf(t, reviews)
		}
	})
}

func (s *internalTestSuite) TestGetMovieReviewsByID(p mt.MovieReviewer, items []string, vfs ...ValidateFunc) {
	s.testGetMovieReviews(p.GetMovieReviewsByID, items, vfs...)
}

func (s *internalTestSuite) TestGetMovieReviewsByURL(p mt.MovieReviewer, items []string, vfs ...ValidateFunc) {
	s.testGetMovieReviews(p.GetMovieReviewsByURL, items, vfs...)
}

func (s *internalTestSuite) TestSearchActor(p mt.ActorSearcher, items []string, vfs ...ValidateFunc) {
	s.testItems(items, func(t *testing.T, item string) {
		results, err := p.SearchActor(item)
		require.NoError(t, err)
		require.NotEmpty(t, results)
		for _, vf := range append([]ValidateFunc{
			logJSONContent(),
			assertIsValid(),
		}, vfs...) {
			vf(t, results)
		}
	})
}

func (s *internalTestSuite) TestSearchMovie(p mt.MovieSearcher, items []string, vfs ...ValidateFunc) {
	s.testItems(items, func(t *testing.T, item string) {
		results, err := p.SearchMovie(p.NormalizeMovieKeyword(item))
		require.NoError(t, err)
		require.NotEmpty(t, results)
		for _, vf := range append([]ValidateFunc{
			logJSONContent(),
			assertIsValid(),
		}, vfs...) {
			vf(t, results)
		}
	})
}

func (s *internalTestSuite) TestFetch(p mt.Fetcher, items []string, vfs ...ValidateFunc) {
	s.testItems(items, func(t *testing.T, item string) {
		resp, err := p.Fetch(item)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		assert.NotEmpty(t, data)
		for _, vf := range append([]ValidateFunc{
			logJSONContent(),
		}, vfs...) {
			vf(t, data)
		}
	})
}

func Test[T mt.Provider](t *testing.T, new func() T, items []string, vfs ...ValidateFunc) {
	if ci, _ := strconv.ParseBool(os.Getenv("GITHUB_ACTIONS")); ci {
		t.SkipNow() // Skip in GitHub Actions
	}

	functionName := getFrame(1).Function
	providerName, testMethod, err := parseTestFunction(functionName)
	require.NoError(t, err)

	provider := new()
	structName := reflect.TypeOf(provider).Elem().Name()
	require.Equal(t, providerName, structName)

	s := &internalTestSuite{t}
	m := reflect.ValueOf(s).MethodByName("Test" + testMethod)
	require.Truef(t, m.IsValid(), "invalid test method: %s", testMethod)

	// Build test function args.
	args := []reflect.Value{
		reflect.ValueOf(provider),
		reflect.ValueOf(items),
	}
	for _, vf := range vfs {
		args = append(args, reflect.ValueOf(vf))
	}

	// Make the test function call.
	results := m.Call(args)
	require.Empty(t, results)
}
