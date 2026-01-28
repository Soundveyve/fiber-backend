package models

import "time"

// CreateUserRequest представляет данные для создания пользователя
// Эти поля приходят от клиента в JSON формате
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`     // Email обязателен и должен быть валидным
	Username  string `json:"username" validate:"required,min=3"`  // Username минимум 3 символа
	Password  string `json:"password" validate:"required,min=8"`  // Пароль минимум 8 символов
	FirstName string `json:"first_name,omitempty"`                // Опциональное поле
	LastName  string `json:"last_name,omitempty"`                 // Опциональное поле
}

// UpdateUserRequest представляет данные для обновления пользователя
// Все поля опциональны (указатели позволяют различить "не передано" и "пусто")
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// UserResponse представляет пользователя в ответе API
// Не включаем password_hash для безопасности
type UserResponse struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName *string   `json:"first_name,omitempty"` // Указатель чтобы null был null, а не пустой строкой
	LastName  *string   `json:"last_name,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListUsersRequest представляет параметры для получения списка пользователей
type ListUsersRequest struct {
	Page     int `query:"page" validate:"min=1"`               // Номер страницы (начиная с 1)
	PageSize int `query:"page_size" validate:"min=1,max=100"` // Размер страницы (макс 100)
}

// ListUsersResponse представляет ответ со списком пользователей
type ListUsersResponse struct {
	Users      []UserResponse `json:"users"`       // Список пользователей
	TotalCount int            `json:"total_count"` // Общее количество
	Page       int            `json:"page"`        // Текущая страница
	PageSize   int            `json:"page_size"`   // Размер страницы
	TotalPages int            `json:"total_pages"` // Всего страниц
}

// ErrorResponse представляет ошибку в API ответе
// Стандартизированный формат ошибок упрощает обработку на клиенте
type ErrorResponse struct {
	Error   string                 `json:"error"`             // Текст ошибки
	Code    string                 `json:"code,omitempty"`    // Код ошибки (для программной обработки)
	Details map[string]interface{} `json:"details,omitempty"` // Дополнительные детали
}

// SuccessResponse представляет успешный ответ без данных
type SuccessResponse struct {
	Message string `json:"message"` // Сообщение об успехе
}

// HealthResponse представляет статус здоровья сервиса
type HealthResponse struct {
	Status   string            `json:"status"`   // "ok" или "error"
	Services map[string]string `json:"services"` // Статусы подсервисов (БД и т.д.)
	Version  string            `json:"version"`  // Версия приложения
}
