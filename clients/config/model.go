package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the Clients service.
type Config struct {
	LogLevel      string
	ListenPort    int
	DB            DBConfig
	MigrationPath string
	RunMigrations bool
	HighLevel     HighLevelConfig
}

// DBConfig holds database configuration.
type DBConfig struct {
	DBHost      string
	DBPort      int
	DBName      string
	DBUser      string
	DBPassword  string
	TLSDisabled bool
}

// HighLevelConfig holds HighLevel provider configuration.
type HighLevelConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	// WebhookPublicKey is the PEM-encoded Ed25519 public key used to verify
	// HighLevel webhook signatures (HIGHLEVEL_WEBHOOK_PUBLIC_KEY). It is
	// public cryptographic material, not a private credential.
	WebhookPublicKey string
	// PaymentURL is the frontend checkout URL supplied to HighLevel as the
	// payment provider's paymentsUrl (HIGHLEVEL_PAYMENT_URL). It is
	// configuration, never hard-coded.
	PaymentURL string
	// QueryURL is the backend query URL supplied to HighLevel as the payment
	// provider's queryUrl (HIGHLEVEL_QUERY_URL). It is configuration, never
	// hard-coded.
	QueryURL string
	// ProviderName is the display name of the payment provider
	// (HIGHLEVEL_PROVIDER_NAME).
	ProviderName string
	// ProviderDescription is the description of the payment provider
	// (HIGHLEVEL_PROVIDER_DESCRIPTION).
	ProviderDescription string
	// ProviderImageURL is the image URL of the payment provider
	// (HIGHLEVEL_PROVIDER_IMAGE_URL).
	ProviderImageURL string
	// APIBaseURL is the HighLevel API base URL used for outbound Custom
	// Payment Provider calls (HIGHLEVEL_API_BASE_URL). It is configuration,
	// never hard-coded.
	APIBaseURL string
	// PublicBaseURL is the publicly reachable base URL of the deployed
	// backend (PUBLIC_BASE_URL). It is used to derive the backend query URL
	// for the payment provider configuration. It is configuration, never
	// hard-coded.
	PublicBaseURL string
}

// LoadConfig loads configuration from environment variables.
func (c *Config) LoadConfig() error {
	c.LogLevel = getEnv("LOG_LEVEL", "info")
	c.ListenPort = getEnvAsInt("LISTEN_PORT", 50051)
	c.MigrationPath = getEnv("MIGRATION_PATH", "db/migrations")
	c.RunMigrations = getEnvAsBool("RUN_MIGRATIONS", true)

	c.DB.DBHost = getEnv("DB_HOST", "localhost")
	c.DB.DBPort = getEnvAsInt("DB_PORT", 5432)
	c.DB.DBName = getEnv("DB_NAME", "rvpay")
	c.DB.DBUser = getEnv("DB_USER", "postgres")
	c.DB.DBPassword = getEnv("DB_PASSWORD", "postgres")
	c.DB.TLSDisabled = getEnvAsBool("DB_TLS_DISABLED", true)

	c.HighLevel.ClientID = getEnv("HIGHLEVEL_CLIENT_ID", "")
	c.HighLevel.ClientSecret = getEnv("HIGHLEVEL_CLIENT_SECRET", "")
	c.HighLevel.RedirectURI = getEnv("HIGHLEVEL_REDIRECT_URI", "")
	c.HighLevel.WebhookPublicKey = getEnv("HIGHLEVEL_WEBHOOK_PUBLIC_KEY", "")
	c.HighLevel.PaymentURL = getEnv("HIGHLEVEL_PAYMENT_URL", "")
	c.HighLevel.QueryURL = getEnv("HIGHLEVEL_QUERY_URL", "")
	c.HighLevel.ProviderName = getEnv("HIGHLEVEL_PROVIDER_NAME", "RVPay")
	c.HighLevel.ProviderDescription = getEnv("HIGHLEVEL_PROVIDER_DESCRIPTION", "RVPay payment provider")
	c.HighLevel.ProviderImageURL = getEnv("HIGHLEVEL_PROVIDER_IMAGE_URL", "")
	c.HighLevel.APIBaseURL = getEnv("HIGHLEVEL_API_BASE_URL", "https://services.leadconnectorhq.com")
	c.HighLevel.PublicBaseURL = getEnv("PUBLIC_BASE_URL", "")

	return nil
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as int or returns a default value.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsBool gets an environment variable as bool or returns a default value.
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
