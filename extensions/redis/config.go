package redis

type Config struct {
	Host            string `envconfig:"REDIS_HOST" required:"true"`
	Port            int    `envconfig:"REDIS_PORT" required:"true"`
	Password        string `envconfig:"REDIS_PASSWORD" required:"true"`
	DB              int    `envconfig:"REDIS_DB" required:"true"`
	CounterKey      string `envconfig:"COUNTER_KEY" required:"true"`
	CounterStartVal int    `envconfig:"COUNTER_START_VAL" required:"true"`
}
