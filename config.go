package main

import "os"

// Config represents the complete configuration information for the service.
type Config struct {
	ServerPort  string
	DBHost      string
	DBPort      string
	DBName      string
	VisitsTable string
	CitiesTable string
}

// ConfigFromEnv sources configuration from environment variables.
func ConfigFromEnv() Config {
	return Config{
		ServerPort:  getEnvOrElse("SERVER_PORT", "8080"),
		DBPort:      getEnvOrElse("DB_PORT", "28015"),
		DBHost:      getEnvOrElse("DB_HOST", "localhost"),
		DBName:      getEnvOrElse("DB_NAME", "been_there"),
		VisitsTable: getEnvOrElse("VISITS_TABLE", "visits"),
		CitiesTable: getEnvOrElse("CITIES_TABLE", "cities"),
	}
}

// getEnvOrElse looks up an environment variable and if it does not exist,
// the function returns a default value.
func getEnvOrElse(name string, other string) string {
	env, ok := os.LookupEnv(name)
	if ok {
		log.WithField(name, env).Info("using env variable")
	} else {
		env = other
		log.WithField(name, env).Info("using default setting")
	}
	return env
}

// mustGetEnv looks up a given environment variable and fatally logs an
// error if it does not exist.
func mustGetEnv(name string) string {
	env, ok := os.LookupEnv(name)
	if !ok {
		log.WithField("env", name).Fatalf("missing required environment variable for configuration")
	}
	log.WithField(name, env).Info("using env variable")
	return env
}
