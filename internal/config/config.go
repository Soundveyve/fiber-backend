package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config структура содержит все настройки приложения
// Мы группируем настройки по категориям для лучшей организации
type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

// AppConfig содержит основные настройки приложения
type AppConfig struct {
	Name string // Имя приложения
	Port string // Порт на котором будет слушать HTTP сервер
	Env  string // Окружение (development, production)
}

// DatabaseConfig содержит настройки подключения к базе данных
// Эта структура универсальна и подходит для разных типов БД
type DatabaseConfig struct {
	Driver          string        // Тип БД: postgres, mysql
	Host            string        // Хост БД
	Port            string        // Порт БД
	User            string        // Имя пользователя
	Password        string        // Пароль
	Name            string        // Имя базы данных
	SSLMode         string        // Режим SSL (для PostgreSQL)
	MaxOpenConns    int           // Максимум открытых соединений
	MaxIdleConns    int           // Максимум простаивающих соединений
	ConnMaxLifetime time.Duration // Время жизни соединения
}

// LoadConfig загружает конфигурацию из переменных окружения
// Она сначала пытается загрузить .env файл, затем читает переменные
func LoadConfig() (*Config, error) {
	// Загружаем .env файл если он существует
	// В продакшене .env может не быть, и это нормально
	// В Docker контейнерах переменные будут переданы напрямую
	_ = godotenv.Load()

	// Создаем конфигурацию со значениями по умолчанию
	config := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "fiber-backend"),
			Port: getEnv("APP_PORT", "3000"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "postgres"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "fiber_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			// Парсим числовые значения с дефолтными значениями
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 5)) * time.Minute,
		},
	}

	// Валидируем обязательные параметры
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate проверяет что все критичные параметры заданы
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST не может быть пустым")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER не может быть пустым")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME не может быть пустым")
	}
	return nil
}

// GetDSN возвращает строку подключения к БД в зависимости от драйвера
// DSN (Data Source Name) - это строка с параметрами подключения
func (c *DatabaseConfig) GetDSN() string {
	switch c.Driver {
	case "postgres":
		// Формат для PostgreSQL
		return fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
		)
	case "mysql":
		// Формат для MySQL
		// parseTime=true позволяет автоматически парсить DATE/DATETIME в time.Time
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true",
			c.User, c.Password, c.Host, c.Port, c.Name,
		)
	default:
		return ""
	}
}

// getEnv получает переменную окружения или возвращает дефолтное значение
// Это удобная функция-хелпер для работы с переменными окружения
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt получает переменную окружения как число
// Если не удается распарсить или переменная не задана - возвращает дефолт
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
