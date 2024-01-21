package config

type Config struct {
	PgString       string `env:"POSTGRES" env-default:"postgres://go:12345678@127.0.0.1:5432/orders"`
	StanClusterId  string `env:"STAN_CLUSTER_ID" env-default:"test-cluster"`
	StanClientId   string `env:"STAN_CLIENT_ID" env-default:"client-"`
	StanSubject    string `env:"STAN_SUBJECT" env-default:"order"`
	StanQueue      string `env:"STAN_QUEUE" env-default:"queue"`
}