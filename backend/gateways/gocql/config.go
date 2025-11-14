package gocql

type Config struct {
	Host        string `envconfig:"CASSANDRA_HOST" required:"true"`
	Username    string `envconfig:"CASSANDRA_USERNAME" required:"true"`
	Password    string `envconfig:"CASSANDRA_PASSWORD" required:"true"`
	Keyspace    string `envconfig:"CASSANDRA_KEYSPACE" required:"true"`
	Port        int    `envconfig:"CASSANDRA_PORT" required:"true"`
	AutoMigrate bool   `envconfig:"CASSANDRA_AUTO_MIGRATE" default:"false"`
}
