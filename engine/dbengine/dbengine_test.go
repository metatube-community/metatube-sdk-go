package dbengine

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"log"
	"net"
	"net/url"
	"slices"
	"sort"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/datatypes"
	"gorm.io/gorm/logger"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/database"
	"github.com/metatube-community/metatube-sdk-go/engine/providerid"
	"github.com/metatube-community/metatube-sdk-go/model"
)

var (
	//go:embed testdata/actor_metadata.json
	actorMetadata string

	//go:embed testdata/movie_metadata.json
	movieMetadata string

	//go:embed testdata/movie_reviews.json
	movieReviews string
)

type DBEngineTestSuite struct {
	suite.Suite

	dsn string
	eng DBEngine
}

func TestDBEngineTestSuite(t *testing.T) {
	suites := []struct {
		name string
		dsn  string
	}{
		{
			// SQLite (memory mode)
			name: database.Sqlite,
			dsn:  ":memory:",
		},
		{
			// Postgres
			name: database.Postgres,
			dsn:  postgresDSN,
		},
	}

	for _, s := range suites {
		t.Run(s.name, func(t *testing.T) {
			suite.Run(t, &DBEngineTestSuite{dsn: s.dsn})
		})
	}
}

func (s *DBEngineTestSuite) SetupSuite() {
	db, err := database.Open(&database.Config{
		DSN:                  s.dsn,
		DisableAutomaticPing: true,
		LogLevel:             logger.Warn,
	})
	s.Require().NoError(err)

	s.eng = New(db)
	err = s.eng.AutoMigrate()
	s.Require().NoError(err)
}

func (s *DBEngineTestSuite) TestVersion() {
	version, err := s.eng.Version()
	s.Require().NoError(err)
	s.Require().NotEmpty(version)
}

func (s *DBEngineTestSuite) TestActor() {
	var jsonData []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Provider string `json:"provider"`
		Homepage string `json:"homepage"`
	}
	if err := json.
		NewDecoder(strings.NewReader(actorMetadata)).
		Decode(&jsonData); err != nil {
		s.Require().NoError(err)
	}
	for _, data := range jsonData {
		err := s.eng.SaveActorInfo(&model.ActorInfo{
			ID:       data.ID,
			Name:     data.Name,
			Provider: data.Provider,
			Homepage: data.Homepage,
		})
		s.Require().NoError(err)
	}

	s.T().Run("get actor info", func(t *testing.T) {
		got, err := s.eng.GetActorInfo(providerid.MustParse("AV-LEAGUE:1607"))
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "上原亜衣", got.Name)
	})

	s.T().Run("get actor info (case-insensitive)", func(t *testing.T) {
		got, err := s.eng.GetActorInfo(providerid.MustParse("av-league:1981"))
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "大場ゆい", got.Name)
	})

	s.T().Run("search actor by name (no case)", func(t *testing.T) {
		actors, err := s.eng.SearchActor("hitOmi", SearchOptions{})
		require.NoError(t, err)
		require.Len(t, actors, 1)
		assert.Equal(t, "7106", actors[0].ID)
	})

	s.T().Run("search actor by name (match)", func(t *testing.T) {
		actors, err := s.eng.SearchActor("加藤あやの", SearchOptions{})
		require.NoError(t, err)
		require.Len(t, actors, 1)
		assert.Equal(t, "2477", actors[0].ID)
	})

	s.T().Run("search actor by name (fuzz)", func(t *testing.T) {
		actors, err := s.eng.SearchActor("加藤", SearchOptions{})
		require.NoError(t, err)
		require.Len(t, actors, 2)
		sort.Slice(actors, func(i, j int) bool { return actors[i].ID < actors[j].ID })
		assert.Equal(t, "2477", actors[0].ID)
		assert.Equal(t, "25352", actors[1].ID)
	})

	s.T().Run("search actor by name (fuzzer)", func(t *testing.T) {
		actors, err := s.eng.SearchActor("三", SearchOptions{Threshold: 0.1})
		require.NoError(t, err)
		require.Len(t, actors, 3)
	})

	s.T().Run("search actor by name (limit)", func(t *testing.T) {
		actors, err := s.eng.SearchActor("夏", SearchOptions{Threshold: 0.1, Limit: 2})
		require.NoError(t, err)
		require.Len(t, actors, 2)
	})

	s.T().Run("search actor by name (not found)", func(t *testing.T) {
		actors, err := s.eng.SearchActor("无名氏", SearchOptions{})
		require.NoError(t, err)
		require.Empty(t, actors)
	})
}

func (s *DBEngineTestSuite) TestMovie() {
	_ = movieMetadata
}

func (s *DBEngineTestSuite) TestMovie_Reviews() {
	var jsonData []struct {
		ID         string          `json:"id"`
		Provider   string          `json:"provider"`
		RawReviews json.RawMessage `json:"reviews"`
		Reviews    []struct {
			Title   string  `json:"title"`
			Author  string  `json:"author"`
			Comment string  `json:"comment"`
			Score   float64 `json:"score"`
			Date    string  `json:"date"`
		} `json:"-"`
	}
	if err := json.
		NewDecoder(strings.NewReader(movieReviews)).
		Decode(&jsonData); err != nil {
		s.Require().NoError(err)
	}
	for _, data := range jsonData {
		var reviewJSON string
		err := json.Unmarshal(data.RawReviews, &reviewJSON)
		s.Require().NoError(err)
		err = json.Unmarshal([]byte(reviewJSON), &data.Reviews)
		s.Require().NoError(err)
		err = s.eng.SaveMovieReviewInfo(&model.MovieReviewInfo{
			ID:       data.ID,
			Provider: data.Provider,
			Reviews: datatypes.NewJSONType(
				slices.Collect(func(yield func(detail *model.MovieReviewDetail) bool) {
					for _, review := range data.Reviews {
						if !yield(&model.MovieReviewDetail{
							Title:   review.Title,
							Author:  review.Author,
							Comment: review.Comment,
							Score:   review.Score,
							Date:    parser.ParseDate(review.Date),
						}) {
							return
						}
					}
				})),
		})
		s.Require().NoError(err)
	}

	s.T().Run("get review info", func(t *testing.T) {
		got, err := s.eng.GetMovieReviewInfo(providerid.MustParse("FANZA:ebwh00024"))
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.NotEmpty(t, got.Reviews)
	})

	s.T().Run("get review info (case-insensitive)", func(t *testing.T) {
		got, err := s.eng.GetMovieReviewInfo(providerid.MustParse("fanZA:DAsS00465"))
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.NotEmpty(t, got.Reviews)
	})
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	if err := pool.Client.Ping(); err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "15-alpine",
			Env: []string{
				"POSTGRES_DB=" + postgresDB,
				"POSTGRES_USER=" + postgresUser,
				"POSTGRES_PASSWORD=" + postgresPass,
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	postgresDSN = (&url.URL{
		Scheme:   postgresDriver,
		User:     url.UserPassword(postgresUser, postgresPass),
		Host:     net.JoinHostPort("localhost", resource.GetPort("5432/tcp")),
		Path:     postgresDB,
		RawQuery: url.Values{"sslmode": []string{"disable"}}.Encode(),
	}).String()

	if err := pool.Retry(func() error {
		db, err := sql.Open(postgresDriver, postgresDSN)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}()

	m.Run()
}

var postgresDSN string

const (
	postgresDriver = "postgres"
	postgresDB
	postgresUser
	postgresPass
)
