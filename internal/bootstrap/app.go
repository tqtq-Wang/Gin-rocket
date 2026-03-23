package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"gin-rocket/internal/handler"
	"gin-rocket/internal/model"
	"gin-rocket/internal/repository"
	"gin-rocket/internal/service"
	"gin-rocket/pkg/cache"
	"gin-rocket/pkg/configx"
	"gin-rocket/pkg/database"
	"gin-rocket/pkg/logger"
	"gin-rocket/router"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Application struct {
	Config     *configx.Config
	Logger     *zap.Logger
	DB         *gorm.DB
	Redis      *redis.Client
	HTTPServer *http.Server
}

func NewApplication() (*Application, error) {
	cfg, err := configx.Load()
	if err != nil {
		return nil, err
	}

	appLogger, err := logger.New(cfg.Log, cfg.App.Env)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	db, err := database.NewMySQL(cfg.MySQL, appLogger)
	if err != nil {
		_ = appLogger.Sync()
		return nil, fmt.Errorf("init mysql: %w", err)
	}

	if err := model.AutoMigrate(db); err != nil {
		closeSQLDB(db)
		_ = appLogger.Sync()
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	redisClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		closeSQLDB(db)
		_ = appLogger.Sync()
		return nil, fmt.Errorf("init redis: %w", err)
	}

	healthHandler, err := handler.NewHealthHandler(cfg.App.Name, db, redisClient)
	if err != nil {
		_ = redisClient.Close()
		closeSQLDB(db)
		_ = appLogger.Sync()
		return nil, fmt.Errorf("init health handler: %w", err)
	}

	userRepo := repository.NewGormUserRepository(db)
	userCacheRepo := repository.NewRedisUserCacheRepository(redisClient)
	userService := service.NewUserService(userRepo, userCacheRepo, appLogger, cfg.Cache.UserTTL)
	userHandler := handler.NewUserHandler(userService)

	engine := router.New(cfg, appLogger, healthHandler, userHandler)
	httpServer := &http.Server{
		Addr:         net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port)),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Application{
		Config:     cfg,
		Logger:     appLogger,
		DB:         db,
		Redis:      redisClient,
		HTTPServer: httpServer,
	}, nil
}

func (a *Application) Start() error {
	a.Logger.Info("http server starting", zap.String("addr", a.HTTPServer.Addr))
	return a.HTTPServer.ListenAndServe()
}

func (a *Application) Shutdown(ctx context.Context) error {
	var shutdownErr error

	if err := a.HTTPServer.Shutdown(ctx); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown http server: %w", err))
	}

	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("close redis: %w", err))
		}
	}

	if err := closeSQLDB(a.DB); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("close mysql: %w", err))
	}

	if a.Logger != nil {
		_ = a.Logger.Sync()
	}

	return shutdownErr
}

func closeSQLDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
