package conf

import (
	"reflect"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Viper struct{}

type app struct {
	ENV                  string `mapstructure:"APP_ENV"`
	APIKey               string `mapstructure:"API_KEY"`
	MaxRequestsPerMinute int    `mapstructure:"MAX_REQUESTS_PER_MINUTE"`
}

type jobConfig struct {
	JobInQueuedSize int `mapstructure:"JOB_IN_QUEUE_SIZE"`
	ParallelJobs    int `mapstructure:"PARALLEL_JOBS"`
	TickerInterval  int `mapstructure:"TICKER_INTERVAL"`
}

type database struct {
	PgSQLConnection string `mapstructure:"PG_DB_CONNECTION"`
	PgSQLHost       string `mapstructure:"PG_DB_HOST"`
	PgSQLPort       string `mapstructure:"PG_DB_PORT"`
	PgSQLDatabase   string `mapstructure:"PG_DB_DATABASE"`
	PgSQLUser       string `mapstructure:"PG_DB_USERNAME"`
	PgSQLPassword   string `mapstructure:"PG_DB_PASSWORD"`
	PgSQLDebug      bool   `mapstructure:"PG_DB_DEBUG"`
	PgSQLSSL        bool   `mapstructure:"PG_DB_SSL"`
}

type kafka struct {
	Brokers       []string `mapstructure:"KAFKA_BROKERS"`
	Username      string   `mapstructure:"KAFKA_USERNAME"`
	Password      string   `mapstructure:"KAFKA_PASSWORD"`
	SASLMechanism string   `mapstructure:"KAFKA_SASL_MECHANISM"`
	Version       string   `mapstructure:"KAFKA_VERSION"`

	JobsTopic     string `mapstructure:"JOBS_TOPIC"`
	ConsumerGroup string `mapstructure:"CONSUMER_GROUP"`
}

var AppConfig = &app{}
var DatabaseConfig = &database{}
var JobConfig = &jobConfig{}
var KafkaConfig = &kafka{}

func (v *Viper) Init() {
	viper.AddConfigPath("./")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	if err := viper.MergeInConfig(); err != nil {
		log.Fatal().Msgf("Viper merge failed: %v", err)
	}

	v.setDefaults()
	v.unmarshal(&AppConfig)
	v.unmarshal(&DatabaseConfig)
	v.unmarshal(&JobConfig)
	v.unmarshal(&KafkaConfig)
	log.Info().Msgf("Viper initialized successfully")
}

func (v *Viper) setDefaults() {
	defer func() {
		if err := recover(); err != nil {
			log.Info().Msgf("Panic occurred: %v", err)
			return
		}
	}()
	structFields := [][]reflect.StructField{
		reflect.VisibleFields(reflect.TypeOf(struct{ app }{})),
		reflect.VisibleFields(reflect.TypeOf(struct{ database }{})),
		reflect.VisibleFields(reflect.TypeOf(struct{ jobConfig }{})),
		reflect.VisibleFields(reflect.TypeOf(struct{ kafka }{})),
	}
	v.setFields(structFields)
	log.Info().Msgf("Setting defaults for viper, completed")
}

func (v *Viper) setFields(fields [][]reflect.StructField) {
	for _, fs := range fields {
		for _, field := range fs {
			viper.SetDefault(field.Tag.Get("mapstructure"), "")
		}
	}
}

func (v *Viper) unmarshal(conf interface{}) {
	err := viper.Unmarshal(conf)
	if err != nil {
		log.Fatal().Msgf("Viper app unmarshal error: %s", err)
	}
}
