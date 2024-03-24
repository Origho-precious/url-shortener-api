package configs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type config struct {
	PORT                             string
	BASE_URL                         string
	GIN_MODE                         string
	MONGO_URI                        string
	CLIENT_URL                       string
	JWT_SECRET                       string
	MAILTRAP_AUTH                    string
	IMAGEKIT_PUBLIC_KEY              string
	URL_REDIRECT_PREFIX              string
	IMAGEKIT_PRIVATE_KEY             string
	MAILTRAP_SENDER_EMAIL            string
	IMAGEKIT_URL_ENDPOINT            string
	RESET_PASSWORD_TEMPLATE_UUID     string
	EMAIL_VERIFICATION_TEMPLATE_UUID string
}

func LoadEnvs() (config, error) {
	var cfg config

	ginMode, exists := os.LookupEnv("GIN_MODE")
	if !exists {
		ginMode = "debug"
	}

	fmt.Println(ginMode)

	err := godotenv.Load(filepath.Join("./", ".env"))
	if err != nil {
		return cfg, err
	}

	cfg.PORT = os.Getenv("PORT")
	cfg.BASE_URL = os.Getenv("BASE_URL")
	cfg.GIN_MODE = os.Getenv("GIN_MODE")
	cfg.MONGO_URI = os.Getenv("MONGO_URI")
	cfg.CLIENT_URL = os.Getenv("CLIENT_URL")
	cfg.JWT_SECRET = os.Getenv("JWT_SECRET")
	cfg.MAILTRAP_AUTH = os.Getenv("MAILTRAP_AUTH")
	cfg.URL_REDIRECT_PREFIX = os.Getenv("URL_REDIRECT_PREFIX")
	cfg.IMAGEKIT_PUBLIC_KEY = os.Getenv("IMAGEKIT_PUBLIC_KEY")
	cfg.IMAGEKIT_PRIVATE_KEY = os.Getenv("IMAGEKIT_PRIVATE_KEY")
	cfg.IMAGEKIT_URL_ENDPOINT = os.Getenv("IMAGEKIT_URL_ENDPOINT")
	cfg.MAILTRAP_SENDER_EMAIL = os.Getenv("MAILTRAP_SENDER_EMAIL")
	cfg.RESET_PASSWORD_TEMPLATE_UUID = os.Getenv("RESET_PASSWORD_TEMPLATE_UUID")
	cfg.EMAIL_VERIFICATION_TEMPLATE_UUID =
		os.Getenv("EMAIL_VERIFICATION_TEMPLATE_UUID")

	return cfg, nil
}
