package api

import (
	"math/rand"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/stretchr/testify/require"
	z "go.uber.org/zap"
	"gorm.io/gorm"

	"transcode-demo/config"
	"transcode-demo/internal/middleware"
	"transcode-demo/internal/presentation/api/healthcheck"
	"transcode-demo/internal/presentation/api/transcoderequest"
	"transcode-demo/internal/services/transcoderequestsvc"
	"transcode-demo/pkg/json"
	"transcode-demo/pkg/logger"
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
			StructValidator:              &StructValidator{validate: validator.New()},
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
	healthHandler := healthcheck.NewHandler()
	s.r.Get("/health", healthHandler.HealthCheck)

	// transcode request
	transReqHandler := transcoderequest.NewHandler(l, transcoderequestsvc.NewService(s.cfg, s.db))
	transReqGroup := s.r.Group("/transcode-request")
	transReqGroup.Post("", middleware.GormTransaction(l, s.db), transReqHandler.CreateTranscodeRequest)
}

// test utils

// NewTestServer creates a test server of Server
func NewTestServer(t testing.TB) string {
	cfg := config.MustNewDevConfig(t)
	l := logger.New(cfg.Log)
	db := config.MustNewGorm(t)
	s := NewServer(cfg, l, db)
	cfg.HttpConfig.Port = rand.Intn(10000) + 9000
	go func() {
		err := s.Listen(cfg.HttpConfig.GetAddr())
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		s.Shutdown()
	})
	time.Sleep(50 * time.Millisecond)
	return "http://" + cfg.HttpConfig.GetAddr()
}
