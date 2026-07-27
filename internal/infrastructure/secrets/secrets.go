package secrets

type Postgres struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
}

type Secrets struct {
	Postgres Postgres
}
