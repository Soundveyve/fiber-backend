package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Soundveyve/fiber-backend/internal/models"
	"github.com/Soundveyve/fiber-backend/internal/repository"
	
	"golang.org/x/crypto/bcrypt"
)

// UserService содержит бизнес-логику для работы с пользователями
// Это промежуточный слой между HTTP handlers и repository (БД)
type UserService struct {
	queries *repository.Queries // Сгенерированные sqlc запросы
	db      *sql.DB             // Прямой доступ к БД для транзакций
}

// NewUserService создает новый экземпляр сервиса пользователей
func NewUserService(queries *repository.Queries, db *sql.DB) *UserService {
	return &UserService{
		queries: queries,
		db:      db,
	}
}

// CreateUser создает нового пользователя
// Хеширует пароль перед сохранением в БД
func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (*models.UserResponse, error) {
	// 1. Хешируем пароль с помощью bcrypt
	// bcrypt автоматически добавляет соль и использует безопасный алгоритм
	// DefaultCost (10) это хороший баланс между безопасностью и производительностью
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	// 2. Создаем пользователя в БД через сгенерированный sqlc метод
	user, err := s.queries.CreateUser(ctx, repository.CreateUserParams{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		FirstName:    sql.NullString{String: req.FirstName, Valid: req.FirstName != ""},
		LastName:     sql.NullString{String: req.LastName, Valid: req.LastName != ""},
	})
	if err != nil {
		// Здесь можно добавить проверку на дублирование email/username
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	// 3. Конвертируем модель БД в модель ответа API
	return s.toUserResponse(&user), nil
}

// GetUserByID получает пользователя по ID
func (s *UserService) GetUserByID(ctx context.Context, id int) (*models.UserResponse, error) {
	user, err := s.queries.GetUserByID(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	return s.toUserResponse(&user), nil
}

// GetUserByEmail получает пользователя по email
// Полезно для аутентификации
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.UserResponse, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	return s.toUserResponse(&user), nil
}

// ListUsers возвращает список пользователей с пагинацией
func (s *UserService) ListUsers(ctx context.Context, req models.ListUsersRequest) (*models.ListUsersResponse, error) {
	// 1. Рассчитываем offset для SQL запроса
	// Например: страница 2, размер 10 -> offset = (2-1) * 10 = 10
	offset := (req.Page - 1) * req.PageSize

	// 2. Получаем пользователей из БД
	users, err := s.queries.ListUsers(ctx, repository.ListUsersParams{
		Limit:  int32(req.PageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка пользователей: %w", err)
	}

	// 3. Получаем общее количество пользователей для пагинации
	totalCount, err := s.queries.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка подсчета пользователей: %w", err)
	}

	// 4. Конвертируем в формат ответа
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *s.toUserResponse(&user)
	}

	// 5. Рассчитываем общее количество страниц
	totalPages := int(totalCount) / req.PageSize
	if int(totalCount)%req.PageSize != 0 {
		totalPages++
	}

	return &models.ListUsersResponse{
		Users:      userResponses,
		TotalCount: int(totalCount),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(ctx context.Context, id int, req models.UpdateUserRequest) (*models.UserResponse, error) {
	// Конвертируем указатели в sql.Null* типы
	// Это позволяет различать "не передано" (nil) и "установить пусто" ("")
	params := repository.UpdateUserParams{
		ID: int32(id),
	}

	if req.Email != nil {
		params.Email = sql.NullString{String: *req.Email, Valid: true}
	}
	if req.Username != nil {
		params.Username = sql.NullString{String: *req.Username, Valid: true}
	}
	if req.FirstName != nil {
		params.FirstName = sql.NullString{String: *req.FirstName, Valid: true}
	}
	if req.LastName != nil {
		params.LastName = sql.NullString{String: *req.LastName, Valid: true}
	}
	if req.IsActive != nil {
		params.IsActive = sql.NullBool{Bool: *req.IsActive, Valid: true}
	}

	user, err := s.queries.UpdateUser(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	return s.toUserResponse(&user), nil
}

// DeleteUser удаляет пользователя (физически)
func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	err := s.queries.DeleteUser(ctx, int32(id))
	if err != nil {
		return fmt.Errorf("ошибка удаления пользователя: %w", err)
	}
	return nil
}

// DeactivateUser деактивирует пользователя (soft delete)
// Предпочтительный способ в production
func (s *UserService) DeactivateUser(ctx context.Context, id int) error {
	err := s.queries.DeactivateUser(ctx, int32(id))
	if err != nil {
		return fmt.Errorf("ошибка деактивации пользователя: %w", err)
	}
	return nil
}

// VerifyPassword проверяет пароль пользователя
// Используется при аутентификации
func (s *UserService) VerifyPassword(ctx context.Context, email, password string) (*models.UserResponse, error) {
	// Получаем пользователя с хешем пароля
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("неверный email или пароль")
		}
		return nil, fmt.Errorf("ошибка проверки пароля: %w", err)
	}

	// Сравниваем хеш с введенным паролем
	// bcrypt.CompareHashAndPassword безопасно сравнивает пароли
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("неверный email или пароль")
	}

	return s.toUserResponse(&user), nil
}

// toUserResponse конвертирует модель БД в модель API ответа
// Убирает sensitive данные (пароль) и преобразует типы
func (s *UserService) toUserResponse(user *repository.User) *models.UserResponse {
	resp := &models.UserResponse{
		ID:        int(user.ID),
		Email:     user.Email,
		Username:  user.Username,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Преобразуем sql.NullString в *string
	if user.FirstName.Valid {
		resp.FirstName = &user.FirstName.String
	}
	if user.LastName.Valid {
		resp.LastName = &user.LastName.String
	}

	return resp
}
