package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"gotalk-api/app/service"
	"gotalk-api/helper"
	"gotalk-api/middleware"
)

type Dependencies struct {
	Pool        *pgxpool.Pool
	JWT         *helper.JWTManager
	Permissions *helper.PermissionSet
	AuthService *service.AuthService
	UserService *service.UserService
	ChatService *service.ChatService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool))

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimit(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	users := api.Group("/users", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))

	perms := deps.Permissions

	users.Get("/",
		middleware.RequirePermission(perms, "user:list"),
		deps.UserService.List)
	users.Post("/",
		middleware.RequirePermission(perms, "user:create"),
		deps.UserService.Create)
	users.Patch("/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		deps.UserService.AssignRole)
	users.Delete("/:id",
		middleware.RequirePermission(perms, "user:delete"),
		deps.UserService.Delete)

	users.Get("/search", deps.UserService.SearchPublic)

	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)

	chats := api.Group("/chats", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))

	chats.Post("/", middleware.RequireJSON, deps.ChatService.CreateOrGetRoom)
	chats.Get("/", deps.ChatService.ListRooms)
	chats.Get("/:id/messages", deps.ChatService.ListMessages)
	chats.Post("/:id/messages", middleware.RequireJSON, deps.ChatService.SendMessage)

	messages := api.Group("/messages", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))

	messages.Put("/:id", middleware.RequireJSON, deps.ChatService.UpdateMessage)
	messages.Delete("/:id", deps.ChatService.DeleteMessage)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if pool != nil {
			if err := pool.Ping(ctx); err != nil {
				return helper.ServiceUnavailable("Database tidak dapat dihubungi")
			}
		}

		return helper.Success(c, fiber.StatusOK, "Server dan database berjalan", nil)
	}
}
