package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	
	"github.com/Soundveyve/fiber-backend/internal/config"
	"github.com/Soundveyve/fiber-backend/internal/database"
	"github.com/Soundveyve/fiber-backend/internal/handlers"
	"github.com/Soundveyve/fiber-backend/internal/repository"
	"github.com/Soundveyve/fiber-backend/internal/services"
)

func main() {
	// 1. Загружаем конфигурацию из .env и переменных окружения
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	log.Printf("🚀 Запуск приложения: %s (окружение: %s)", cfg.App.Name, cfg.App.Env)

	// 2. Подключаемся к базе данных
	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	// Выводим статистику пула соединений
	db.LogStats()

	// 3. Создаем слой репозитория (sqlc сгенерированный код)
	queries := repository.New(db.DB)

	// 4. Создаем сервисный слой (бизнес-логика)
	userService := services.NewUserService(queries, db.DB)

	// 5. Создаем HTTP обработчики
	userHandler := handlers.NewUserHandler(userService)

	// 6. Настраиваем Fiber приложение
	app := setupFiberApp(cfg)

	// 7. Регистрируем роуты
	setupRoutes(app, userHandler)

	// 8. Запускаем HTTP сервер в отдельной горутине
	go func() {
		addr := fmt.Sprintf(":%s", cfg.App.Port)
		log.Printf("🌐 HTTP сервер запущен на http://localhost%s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("❌ Ошибка HTTP сервера: %v", err)
		}
	}()

	// 9. Graceful shutdown - ждем сигнал завершения
	quit := make(chan os.Signal, 1)
	// Перехватываем SIGINT (Ctrl+C) и SIGTERM (kill)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Получен сигнал завершения, начинаем graceful shutdown...")

	// Создаем контекст с таймаутом для завершения
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Останавливаем HTTP сервер
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("❌ Ошибка при остановке HTTP сервера: %v", err)
	}

	log.Println("✅ Приложение успешно завершено")
}

// setupFiberApp настраивает Fiber приложение с middleware
func setupFiberApp(cfg *config.Config) *fiber.App {
	// Создаем новое Fiber приложение с настройками
	app := fiber.New(fiber.Config{
		// AppName отображается в заголовках ответов
		AppName: cfg.App.Name,
		
		// ServerHeader добавляет кастомный Server заголовок
		ServerHeader: cfg.App.Name,
		
		// ErrorHandler - кастомный обработчик ошибок
		// Все panic и ошибки будут обработаны здесь
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			
			// Если это Fiber ошибка, используем её код
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware для восстановления после паник
	// Если где-то произойдет panic, приложение не упадет
	app.Use(recover.New())

	// Middleware для логирования запросов
	// Логирует каждый HTTP запрос с информацией о методе, пути, статусе и времени
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "02-Jan-2006 15:04:05",
		TimeZone:   "Local",
	}))

	// CORS middleware для разрешения кросс-доменных запросов
	// Настройте в production для конкретных доменов
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // В production укажите конкретные домены
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	return app
}

// setupRoutes регистрирует все HTTP роуты приложения
func setupRoutes(app *fiber.App, userHandler *handlers.UserHandler) {
	// Health check эндпоинт
	// Используется для проверки доступности сервиса (Kubernetes, Docker)
	app.Get("/health", userHandler.HealthCheck)

	// API группа с префиксом /api/v1
	// Группировка позволяет применять middleware к группе роутов
	api := app.Group("/api/v1")

	// Роуты для пользователей
	users := api.Group("/users")
	{
		// POST /api/v1/users - создание пользователя
		users.Post("/", userHandler.CreateUser)
		
		// GET /api/v1/users - список пользователей
		users.Get("/", userHandler.ListUsers)
		
		// GET /api/v1/users/:id - получение пользователя
		users.Get("/:id", userHandler.GetUser)
		
		// PUT /api/v1/users/:id - обновление пользователя
		users.Put("/:id", userHandler.UpdateUser)
		
		// DELETE /api/v1/users/:id - удаление пользователя
		users.Delete("/:id", userHandler.DeleteUser)
	}

	// 404 обработчик для неизвестных роутов
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Маршрут не найден",
			"path":  c.Path(),
		})
	})
}
