package dbengine

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"log"
	"net"
	"net/url"
	"sort"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm/logger"

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

	// get actor info.
	got, err := s.eng.GetActorInfo(providerid.MustParse("AV-LEAGUE:1607"))
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Assert().Equal("上原亜衣", got.Name)

	// get actor info (case-insensitive).
	got, err = s.eng.GetActorInfo(providerid.MustParse("av-league:1981"))
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Assert().Equal("大場ゆい", got.Name)

	// search actors by ID.
	actors, err := s.eng.SearchActor("2213", SearchOptions{})
	s.Require().NoError(err)
	s.Require().Len(actors, 1)
	s.Assert().Equal("音羽レオン", actors[0].Name)

	// search actors by name (no case).
	actors, err = s.eng.SearchActor("hitOmi", SearchOptions{})
	s.Require().NoError(err)
	s.Require().Len(actors, 1)
	s.Assert().Equal("7106", actors[0].ID)

	// search actors by name (match).
	actors, err = s.eng.SearchActor("加藤あやの", SearchOptions{})
	s.Require().NoError(err)
	s.Require().Len(actors, 1)
	s.Assert().Equal("2477", actors[0].ID)

	// search actors by name (fuzz).
	actors, err = s.eng.SearchActor("加藤", SearchOptions{})
	s.Require().NoError(err)
	s.Require().Len(actors, 2)
	sort.Slice(actors, func(i, j int) bool { return actors[i].ID < actors[j].ID })
	s.Assert().Equal("2477", actors[0].ID)
	s.Assert().Equal("25352", actors[1].ID)

	// search actors by name (fuzzer).
	actors, err = s.eng.SearchActor("三", SearchOptions{Threshold: 0.1})
	s.Require().NoError(err)
	s.Require().Len(actors, 3)

	// search actors by name (limit).
	actors, err = s.eng.SearchActor("夏", SearchOptions{Threshold: 0.1, Limit: 2})
	s.Require().NoError(err)
	s.Require().Len(actors, 2)

	// search actors by name (not found).
	actors, err = s.eng.SearchActor("无名氏", SearchOptions{})
	s.Require().NoError(err)
	s.Require().Len(actors, 0)
}

func (s *DBEngineTestSuite) TestMovie() {
	_ = movieMetadata
	_ = movieReviews
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
