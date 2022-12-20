package configx

import (
	"fmt"
	"strings"

	"github.com/jessevdk/go-flags"
)

type Configuration struct {
	ListenAddr string `long:"port" env:"LISTEN_ADDR" description:"Listen to port (format: :8500|127.0.0.1:8500)" required:"false" default:":8500"`
	// PostgresHostname Postgres host address or name
	PostgresHostname string `long:"psql_host" env:"POSTGRESQL_HOST" description:"Postgres hostname" required:"false" default:"0.0.0.0"`
	// PostgresPort Postgres server port
	PostgresPort int64 `long:"psql_port" env:"POSTGRESQL_PORT" description:"Postgres port" required:"false" default:"8601"`
	// PostgresDBName Postgres server models name
	PostgresDBName string `long:"psql_db" env:"POSTGRESQL_DB" description:"Postgres models name" required:"false" default:"simple"`
	// PostgresUsername Postgres server user name
	PostgresUsername string `long:"psql_user" env:"POSTGRESQL_USER" description:"Postgres username" required:"false" default:"simple"`
	// PostgresPassword Postgres server password
	PostgresPassword string `long:"psql_password" env:"POSTGRESQL_PASSWORD" description:"Postgres password" required:"false" default:"simple"`
	PostgresSSLMode  string `long:"psql_ssl_mode" env:"POSTGRESQL_SSL_MODE" description:"Postgres SSL mode" required:"false" default:"disable"`
	// time
	TimeFormat string `long:"time_format" env:"TIME_FORMAT" description:"default time format" required:"false" default:"15:04:05 02/01/2006"`
	// gin package
	GinMode string `long:"gin_mode" env:"GIN_MODE" description:"set to debug mode to get more information" required:"false" default:"debug"`
	// Kakfa env variables
	MBHosts             string `long:"mb_brokers" env:"MB_HOSTS" description:"the list of kafka brokers' hosts" required:"false" default:"0.0.0.0:29092,0.0.0.0:39092"`
	MBRequestsTopic     string `long:"mb_requests_topic" env:"MB_REQUESTS_TOPIC" description:"" required:"false" default:"hometask.requests.external.test.1"`
	MBRequestsPartition int    `long:"mb_requests_host" env:"MB_REQUESTS_PARTITION" description:"" required:"false" default:"0"`
	// env variable
	Environment string `long:"environment" env:"ENVIRONMENT" description:"prod or staging; it affects other parts of the code (e.g. logs)" required:"false" default:"staging"`
	// redis host
	RedisHost         string `long:"redis_host" env:"REDIS_HOST" description:"redis host" required:"false" default:"0.0.0.0"`
	RedisPort         int32  `long:"redis_port" env:"REDIS_PORT" description:"redis port" required:"false" default:"8602"`
	RedisDatabase     int32  `long:"redis_database" env:"REDIS_DATABASE" description:"database number" required:"false" default:"0"`
	RedisTestDatabase int32  `long:"redis_test_database" env:"REDIS_TEST_DATABASE" description:"will be used for running tests" required:"false" default:"1"`
}

// NewConfiguration
// either parses env variables if it exists or parses default values
// Note: the field must not be required (has required tag set to false)
func NewConfiguration() (*Configuration, error) {
	var c Configuration
	p := flags.NewParser(&c, flags.HelpFlag|flags.PrintErrors|flags.PassDoubleDash|flags.IgnoreUnknown)
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			return nil, fmt.Errorf("this err indicates that the built-in help was shown (the error contains the help message). err: %w", err)
		} else {
			return nil, fmt.Errorf("failed to parse conf. err: %w", err)
		}
	}
	return &c, nil
}

func (c Configuration) GetMBHosts() []string {
	return strings.Split(c.MBHosts, ",")
}
