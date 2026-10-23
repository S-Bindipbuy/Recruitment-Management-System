package config

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// TursoReplicaPath enables an embedded replica at the given local path.
// TursoSyncInterval is the background sync period in seconds (0 = initial sync only).
type Config struct {
	ServerPort     string
	TursoURL       string
	TursoAuthToken string
	TursoReplicaPath  string
	TursoSyncInterval int
	JWTSecret         string
	JWTExpiration     int
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	AllowedOrigins    string
	FrontendPath      string
	UploadDir         string
}

var AppConfig *Config

func Load() {
	loadDotEnv()

	required := []string{"TURSO_DATABASE_URL", "JWT_SECRET", "FRONTEND_PATH"}
	if missing := missingEnv(required); len(missing) > 0 {
		log.Fatalf(
			"config: missing required environment variable(s): %s; set them in the environment or .env (see .env-example)",
			strings.Join(missing, ", "),
		)
	}

	AppConfig = &Config{
		ServerPort:        getEnv("SERVER_PORT", "3000"),
		TursoURL:          os.Getenv("TURSO_DATABASE_URL"),
		TursoAuthToken:    getEnv("TURSO_AUTH_TOKEN", ""),
		TursoReplicaPath:  getEnv("TURSO_REPLICA_PATH", ""),
		TursoSyncInterval: getEnvAsInt("TURSO_SYNC_INTERVAL", 0),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTExpiration:     getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           getEnvAsInt("REDIS_DB", 0),
		AllowedOrigins:    getEnv("ALLOWED_ORIGINS", "*"),
		FrontendPath:      os.Getenv("FRONTEND_PATH"),
		UploadDir:         getEnv("UPLOAD_DIR", "storage/uploads/cvs"),
	}
}

func loadDotEnv() {
	if envFile := strings.TrimSpace(os.Getenv("ENV_FILE")); envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			log.Fatalf("config: failed to load ENV_FILE %q: %v", envFile, err)
		}
		return
	}

	if err := godotenv.Load(".env"); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("config: failed to load .env: %v", err)
	}
}

func missingEnv(keys []string) []string {
	var missing []string
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if intVal, err := strconv.Atoi(getEnv(key, "")); err == nil {
		return intVal
	}
	return fallback
}
