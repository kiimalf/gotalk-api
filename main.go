package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"gotalk-api/app/repository"
	"gotalk-api/app/service"
	"gotalk-api/config"
	"gotalk-api/database"
	"gotalk-api/helper"
	"gotalk-api/route"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("Gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "gotalk-api"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)
	chatRepository := repository.NewChatRepository(pool)
	messageRepository := repository.NewMessageRepository(pool)

	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("Gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Info("Permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	userService := service.NewUserService(userRepository, tokenRepository, permissions)
	authService := service.NewAuthService(
		userRepository, tokenRepository, jwtManager, permissions,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)
	chatService := service.NewChatService(chatRepository, messageRepository, userRepository)

	app := config.NewApp(logger, route.Dependencies{
		Pool:        pool,
		JWT:         jwtManager,
		Permissions: permissions,
		AuthService: authService,
		UserService: userService,
		ChatService: chatService,
	})

	port := config.GetEnv("APP_PORT", "3000")
	logger.Info("Server berjalan", slog.String("port", port))
	if err := app.Listen(":" + port); err != nil {
		logger.Error("Server berhenti", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
