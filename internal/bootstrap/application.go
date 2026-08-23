package bootstrap

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	assignmentapp "github.com/wyw14/cry-092/internal/application/assignment"
	handlingapp "github.com/wyw14/cry-092/internal/application/handling"
	queryapp "github.com/wyw14/cry-092/internal/application/query"
	responseapp "github.com/wyw14/cry-092/internal/application/response"
	reviewapp "github.com/wyw14/cry-092/internal/application/review"
	submissionapp "github.com/wyw14/cry-092/internal/application/submission"
	supervisionapp "github.com/wyw14/cry-092/internal/application/supervision"
	"github.com/wyw14/cry-092/internal/config"
	"github.com/wyw14/cry-092/internal/domain/handling"
	"github.com/wyw14/cry-092/internal/platform/clock"
	"github.com/wyw14/cry-092/internal/platform/notify"
	"github.com/wyw14/cry-092/internal/platform/outbox"
	"github.com/wyw14/cry-092/internal/repository/postgres"
	httptransport "github.com/wyw14/cry-092/internal/transport/http"
	"go.uber.org/zap"
)

type Application struct {
	Config config.Runtime
	Logger *zap.Logger
	Store  *postgres.Store
	Router http.Handler
	Worker outbox.Worker
	Close  func()
}

func Build(ctx context.Context, cfg config.Runtime, logger *zap.Logger) (*Application, error) {
	store, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	ids := suggestionIDSource{}
	now := clock.Live()
	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("load timezone: %w", err)
	}
	calendar := handling.NewCalendar(location, nil, nil)
	submission := submissionapp.Service{Proposals: store, Audits: store, Tx: store, Clock: now, IDs: ids}
	assignment := assignmentapp.Service{Assignments: store, Proposals: store, Audits: store, Outbox: store, Tx: store, Clock: now, IDs: ids}
	handlingService := handlingapp.Service{Plans: store, Assignments: store, Proposals: store, Audits: store, Tx: store, Clock: now, IDs: ids, Calendar: calendar}
	responseService := responseapp.Service{Replies: store, Plans: store, Assignments: store, Proposals: store, Audits: store, Tx: store, Clock: now, IDs: ids}
	reviewService := reviewapp.Service{Reviews: store, Replies: store, Assignments: store, Proposals: store, Supervision: store, Audits: store, Outbox: store, Tx: store, Clock: now, IDs: ids}
	queryService := queryapp.Service{Reader: store}
	supervisionService := supervisionapp.Service{Plans: store, Assignments: store, Supervision: store, Audits: store, Outbox: store, Tx: store, Clock: now, IDs: ids, Calendar: calendar}
	_ = supervisionService
	localNotifier := notify.NewLocalAdapter(now.Now)
	worker := outbox.Worker{Repo: store, Handler: notificationHandler{Notifier: localNotifier}, Logger: logger, MaxAttempts: 5, Batch: 50}
	router := httptransport.NewRouter(httptransport.Dependencies{
		Submission: submission,
		Assignment: assignment,
		Handling:   handlingService,
		Response:   responseService,
		Review:     reviewService,
		Query:      queryService,
		SigningKey: []byte(cfg.JWTSigningKey),
		Issuer:     cfg.JWTIssuer,
		NewID:      ids.NewID,
		Ready:      store.Pool.Ping,
		Logger:     logger,
	})
	return &Application{Config: cfg, Logger: logger, Store: store, Router: router, Worker: worker, Close: store.Close}, nil
}

type suggestionIDSource struct{}

func (suggestionIDSource) NewID() string {
	var entropy [15]byte
	if _, err := rand.Read(entropy[:]); err != nil {
		panic(fmt.Sprintf("secure suggestion identifier generation failed: %v", err))
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(entropy[:])
	return "suggestion_" + strings.ToLower(encoded)
}

func Command() int {
	logger, err := zap.NewProduction()
	if err != nil {
		return 1
	}
	defer func() { _ = logger.Sync() }()
	runtime, err := config.FromEnvironment()
	if err != nil {
		logger.Error("runtime settings rejected", zap.Error(err))
		return 2
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	application, err := Build(ctx, runtime, logger)
	if err != nil {
		logger.Error("suggestion platform assembly failed", zap.Error(err))
		return 3
	}
	defer application.Close()
	if err := application.Run(ctx); err != nil {
		logger.Error("suggestion platform stopped", zap.Error(err))
		return 4
	}
	return 0
}

type notificationHandler struct {
	Notifier interface {
		Notify(context.Context, string, string, map[string]string) error
	}
}

func (h notificationHandler) Deliver(ctx context.Context, message outbox.Message) error {
	return h.Notifier.Notify(ctx, "local-supervision-desk", message.Topic, map[string]string{"message_id": message.ID})
}
