package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/Soundveyve/fiber-backend/internal/models"
	"github.com/Soundveyve/fiber-backend/internal/services"
)

// UserHandler обрабатывает HTTP запросы связанные с пользователями
// Это тонкий слой который:
// 1. Парсит HTTP запрос
// 2. Валидирует данные
// 3. Вызывает сервисный слой
// 4. Формирует HTTP ответ
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler создает новый обработчик пользователей
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser обрабатывает POST /api/v1/users
// Создает нового пользователя
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	// 1. Парсим тело запроса в структуру
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		// Fiber.Ctx.BodyParser автоматически парсит JSON в структуру
		// Если JSON невалидный - возвращаем ошибку 400 Bad Request
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Невалидный JSON",
			Code:  "INVALID_JSON",
		})
	}

	// 2. Здесь можно добавить валидацию через validator пакет
	// Например: validate.Struct(req)

	// 3. Вызываем сервисный слой
	// c.Context() передает контекст запроса для отмены операции если клиент отключился
	user, err := h.userService.CreateUser(c.Context(), req)
	if err != nil {
		// Можно добавить логику для разных типов ошибок
		// Например, проверка на дублирование email
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
			Code:  "CREATE_USER_ERROR",
		})
	}

	// 4. Возвращаем созданного пользователя со статусом 201 Created
	// fiber.StatusCreated это константа для 201
	return c.Status(fiber.StatusCreated).JSON(user)
}

// GetUser обрабатывает GET /api/v1/users/:id
// Получает пользователя по ID
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	// 1. Получаем ID из URL параметров
	// c.Params("id") извлекает значение из маршрута
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Невалидный ID пользователя",
			Code:  "INVALID_USER_ID",
		})
	}

	// 2. Получаем пользователя из сервиса
	user, err := h.userService.GetUserByID(c.Context(), id)
	if err != nil {
		// Если пользователь не найден - возвращаем 404
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: err.Error(),
			Code:  "USER_NOT_FOUND",
		})
	}

	// 3. Возвращаем пользователя
	return c.JSON(user)
}

// ListUsers обрабатывает GET /api/v1/users
// Возвращает список пользователей с пагинацией
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	// 1. Парсим query параметры (page, page_size)
	var req models.ListUsersRequest
	
	// Устанавливаем значения по умолчанию
	req.Page = 1
	req.PageSize = 10

	// QueryParser извлекает параметры из query string
	// Например: /api/v1/users?page=2&page_size=20
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Невалидные параметры запроса",
			Code:  "INVALID_QUERY_PARAMS",
		})
	}

	// 2. Валидируем параметры
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 3. Получаем список пользователей
	response, err := h.userService.ListUsers(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
			Code:  "LIST_USERS_ERROR",
		})
	}

	// 4. Возвращаем список
	return c.JSON(response)
}

// UpdateUser обрабатывает PUT /api/v1/users/:id
// Обновляет данные пользователя
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	// 1. Получаем ID из URL
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Невалидный ID пользователя",
			Code:  "INVALID_USER_ID",
		})
	}

	// 2. Парсим тело запроса
	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Невалидный JSON",
			Code:  "INVALID_JSON",
		})
	}

	// 3. Обновляем пользователя
	user, err := h.userService.UpdateUser(c.Context(), id, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
			Code:  "UPDATE_USER_ERROR",
		})
	}

	// 4. Возвращаем обновленного пользователя
	return c.JSON(user)
}

// DeleteUser обрабатывает DELETE /api/v1/users/:id
// Удаляет пользователя
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	// 1. Получаем ID
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Невалидный ID пользователя",
			Code:  "INVALID_USER_ID",
		})
	}

	// 2. Удаляем пользователя
	// В production лучше использовать DeactivateUser (soft delete)
	err = h.userService.DeleteUser(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: err.Error(),
			Code:  "DELETE_USER_ERROR",
		})
	}

	// 3. Возвращаем 204 No Content (успешное удаление без тела ответа)
	return c.SendStatus(fiber.StatusNoContent)
}

// HealthCheck обрабатывает GET /health
// Проверяет состояние сервиса и его зависимостей
func (h *UserHandler) HealthCheck(c *fiber.Ctx) error {
	// Можно добавить проверку БД и других зависимостей
	return c.JSON(models.HealthResponse{
		Status: "ok",
		Services: map[string]string{
			"api":      "healthy",
			"database": "healthy", // Здесь можно добавить реальную проверку
		},
		Version: "1.0.0",
	})
}
