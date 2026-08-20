package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	z "go.uber.org/zap"
	"gorm.io/gorm"

	"transcode-demo/config"
	"transcode-demo/internal/middleware"
	"transcode-demo/internal/presentation/api/healthcheck"
	"transcode-demo/pkg/json"
)

type Server struct {
	r   *fiber.App
	cfg *config.Config
	db  *gorm.DB
}

func NewServer(cfg *config.Config, l *z.Logger, db *gorm.DB) *Server {
	s := &Server{
		r: fiber.New(fiber.Config{
			ErrorHandler:                 ErrorHandler(cfg.Log.Env, nil),
			WriteBufferSize:              64 * 1024,
			StreamRequestBody:            true,
			DisablePreParseMultipartForm: true,
			JSONDecoder:                  json.Unmarshal,
			JSONEncoder:                  json.Marshal,
		}),
		cfg: cfg,
		db:  db,
	}
	s.middleware(l)

	s.initHandlers(l)

	return s
}

func (s *Server) Listen(add string) error {
	return s.r.Listen(add)
}

func (s *Server) Shutdown() error {
	return s.r.Shutdown()
}

func (s *Server) middleware(l *z.Logger) {
	s.r.Use(cors.New())
	s.r.Use(pprof.New())
	s.r.Use(compress.New())
	s.r.Use(requestid.New())
	s.r.Use(middleware.RequestIDContext())
	s.r.Use(middleware.Logging(l))
	s.r.Use(middleware.CustomRecover(l))
}

func (s *Server) initHandlers(l *z.Logger) {
	// health check
	health := healthcheck.NewHandler()
	s.r.Get("/health", health.HealthCheck)
}
