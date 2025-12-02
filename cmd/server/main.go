package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	httpAdapter "github.com/vlone310/cfguardian/internal/adapters/inbound/http"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/handlers"
	"github.com/vlone310/cfguardian/internal/adapters/inbound/http/middleware"
	"github.com/vlone310/cfguardian/internal/adapters/outbound/postgres"
	"github.com/vlone310/cfguardian/internal/adapters/outbound/raft"
	"github.com/vlone310/cfguardian/internal/domain/services"
	"github.com/vlone310/cfguardian/internal/infrastructure/config"
	"github.com/vlone310/cfguardian/internal/infrastructure/telemetry"
	"github.com/vlone310/cfguardian/internal/usecases/auth"
	configUseCase "github.com/vlone310/cfguardian/internal/usecases/config"
	"github.com/vlone310/cfguardian/internal/usecases/project"
	"github.com/vlone310/cfguardian/internal/usecases/role"
	"github.com/vlone310/cfguardian/internal/usecases/schema"
	"github.com/vlone310/cfguardian/internal/usecases/user"
)

const (
	appName    = "cfguardian"
	appVersion = "0.1.0"
)

func main() {
	// Display banner
	displayBanner()
	
	// Setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting GoConfig Guardian",
		"app", appName,
		"version", appVersion,
	)

	// Create context that listens for termination signals
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	
	slog.Info("Configuration loaded",
		"database_host", cfg.Database.Host,
		"server_port", cfg.Server.Port,
	)

	// Initialize database connection
	dbPool, err := initDatabase(ctx, cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	
	slog.Info("Database connection established")

	// Initialize repositories
	userRepo := postgres.NewUserRepositoryAdapter(dbPool)
	projectRepo := postgres.NewProjectRepositoryAdapter(dbPool)
	roleRepo := postgres.NewRoleRepositoryAdapter(dbPool)
	configSchemaRepo := postgres.NewConfigSchemaRepositoryAdapter(dbPool)
	configRevisionRepo := postgres.NewConfigRevisionRepositoryAdapter(dbPool)
	
	// Initialize Raft consensus for config repository
	raftStore, err := initRaft(cfg)
	if err != nil {
		slog.Error("Failed to initialize Raft", "error", err)
		os.Exit(1)
	}
	defer raftStore.Shutdown()
	
	configRepo := raft.NewConfigRepository(raftStore)
	
	slog.Info("Raft consensus initialized")

	// Initialize domain services
	passwordHasher := services.NewPasswordHasher(12) // bcrypt cost
	apiKeyGenerator := services.NewAPIKeyGenerator()
	schemaValidator := services.NewSchemaValidator()
	versionManager := services.NewVersionManager()

	// Initialize use cases
	// Auth
	loginUseCase := auth.NewLoginUserUseCase(userRepo, passwordHasher)
	registerUseCase := auth.NewRegisterUserUseCase(userRepo, passwordHasher)

	// User
	createUserUseCase := user.NewCreateUserUseCase(userRepo, passwordHasher)
	listUsersUseCase := user.NewListUsersUseCase(userRepo)
	getUserUseCase := user.NewGetUserUseCase(userRepo)
	deleteUserUseCase := user.NewDeleteUserUseCase(userRepo, roleRepo)

	// Project
	createProjectUseCase := project.NewCreateProjectUseCase(projectRepo, userRepo, roleRepo, apiKeyGenerator)
	listProjectsUseCase := project.NewListProjectsUseCase(projectRepo)
	getProjectUseCase := project.NewGetProjectUseCase(projectRepo)
	deleteProjectUseCase := project.NewDeleteProjectUseCase(projectRepo)

	// Role
	assignRoleUseCase := role.NewAssignRoleUseCase(roleRepo, userRepo, projectRepo)
	revokeRoleUseCase := role.NewRevokeRoleUseCase(roleRepo)
	checkPermissionUseCase := role.NewCheckPermissionUseCase(roleRepo)

	// Schema
	createSchemaUseCase := schema.NewCreateSchemaUseCase(configSchemaRepo, schemaValidator)
	listSchemasUseCase := schema.NewListSchemasUseCase(configSchemaRepo)
	updateSchemaUseCase := schema.NewUpdateSchemaUseCase(configSchemaRepo, schemaValidator)
	deleteSchemaUseCase := schema.NewDeleteSchemaUseCase(configSchemaRepo)

	// Config
	createConfigUseCase := configUseCase.NewCreateConfigUseCase(
		configRepo,
		configRevisionRepo,
		configSchemaRepo,
		projectRepo,
		schemaValidator,
	)
	getConfigUseCase := configUseCase.NewGetConfigUseCase(configRepo)
	updateConfigUseCase := configUseCase.NewUpdateConfigUseCase(
		configRepo,
		configRevisionRepo,
		configSchemaRepo,
		schemaValidator,
		versionManager,
	)
	deleteConfigUseCase := configUseCase.NewDeleteConfigUseCase(configRepo)
	rollbackConfigUseCase := configUseCase.NewRollbackConfigUseCase(
		configRepo,
		configRevisionRepo,
		configSchemaRepo,
		schemaValidator,
		versionManager,
	)
	readConfigByAPIKeyUseCase := configUseCase.NewReadConfigByAPIKeyUseCase(
		projectRepo,
		configRepo,
	)

	// Initialize Prometheus metrics
	prometheusMetrics := telemetry.NewPrometheusMetrics(appName)
	slog.Info("Prometheus metrics initialized")
	
	// Initialize HTTP handlers
	authHandler := handlers.NewAuthHandler(loginUseCase, registerUseCase, cfg.JWT.Secret, cfg.JWT.Expiration)
	userHandler := handlers.NewUserHandler(createUserUseCase, listUsersUseCase, getUserUseCase, deleteUserUseCase)
	projectHandler := handlers.NewProjectHandler(createProjectUseCase, listProjectsUseCase, getProjectUseCase, deleteProjectUseCase)
	roleHandler := handlers.NewRoleHandler(assignRoleUseCase, revokeRoleUseCase)
	schemaHandler := handlers.NewSchemaHandler(createSchemaUseCase, listSchemasUseCase, updateSchemaUseCase, deleteSchemaUseCase)
	configHandler := handlers.NewConfigHandler(createConfigUseCase, getConfigUseCase, updateConfigUseCase, deleteConfigUseCase, rollbackConfigUseCase)
	readHandler := handlers.NewReadHandler(readConfigByAPIKeyUseCase)
	healthHandler := handlers.NewHealthHandler(dbPool)
	metricsHandler := handlers.NewMetricsHandler()

	// Initialize router
	router := httpAdapter.NewRouter(httpAdapter.RouterConfig{
		JWTSecret:      cfg.JWT.Secret,
		RateLimitRPS:   100,
		RateLimitBurst: 200,
		AuthHandler:    authHandler,
		UserHandler:    userHandler,
		ProjectHandler: projectHandler,
		RoleHandler:    roleHandler,
		SchemaHandler:  schemaHandler,
		ConfigHandler:  configHandler,
		ReadHandler:    readHandler,
		HealthHandler:  healthHandler,
		MetricsHandler: metricsHandler,
		PrometheusMetrics: prometheusMetrics,
		AuthorizationConfig: middleware.AuthorizationConfig{
			CheckPermission: checkPermissionUseCase,
		},
	})

	// Initialize HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("HTTP server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Application initialized successfully")
	slog.Info("GoConfig Guardian is ready to accept requests", "address", server.Addr)

	// Wait for interrupt signal
	<-ctx.Done()

	// Graceful shutdown
	slog.Info("Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown failed", "error", err)
	}

	// Close Raft node
	if err := raftStore.Shutdown(); err != nil {
		slog.Error("Raft shutdown failed", "error", err)
	}

	// Close database connections
	dbPool.Close()

	slog.Info("Shutdown complete")
}

// initDatabase initializes the PostgreSQL connection pool
func initDatabase(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Connection pool settings
	poolConfig.MaxConns = int32(cfg.Database.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.Database.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.Database.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.Database.ConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// initRaft initializes the Raft consensus layer
func initRaft(cfg *config.Config) (*raft.Store, error) {
	storeConfig := raft.StoreConfig{
		NodeID:            cfg.Raft.NodeID,
		BindAddr:          cfg.Raft.BindAddr,
		DataDir:           cfg.Raft.DataDir,
		Bootstrap:         cfg.Raft.Bootstrap,
		HeartbeatTimeout:  cfg.Raft.HeartbeatTimeout,
		ElectionTimeout:   cfg.Raft.ElectionTimeout,
		SnapshotInterval:  cfg.Raft.SnapshotInterval,
		SnapshotThreshold: cfg.Raft.SnapshotThreshold,
	}

	store, err := raft.NewStore(storeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Raft store: %w", err)
	}

	// Wait for leader election
	if err := store.WaitForLeader(10 * time.Second); err != nil {
		store.Shutdown()
		return nil, fmt.Errorf("failed to elect leader: %w", err)
	}

	return store, nil
}

func displayBanner() {
	banner := `
   ____      ____                    _ _             
  / ___|    / ___|_   _  __ _ _ __ __| (_) __ _ _ __    
 | |   _   | |  _| | | |/ _' | '__/ _' | |/ _' | '_ \   
 | |__| |  | |_| | |_| | (_| | | | (_| | | (_| | | | |  
  \____|   \____|\__,_|\__,_|_|  \__,_|_|\__,_|_| |_|  
                                                         
  GoConfig Guardian - Distributed Configuration Management
  Version: %s
	`
	fmt.Printf(banner, appVersion)
	fmt.Println()
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
