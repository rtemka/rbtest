package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rtemka/rbtest/domain"
	"github.com/rtemka/rbtest/pkg/api"
	"github.com/rtemka/rbtest/pkg/cache"
	"github.com/rtemka/rbtest/pkg/repo/mongo"
)

// переменная окружения.
const (
	portEnv = "APP_PORT"
	dbEnv   = "DB_URL"
)

const cacheUpdInterval = 5 * time.Minute

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load() // загружаем переменные окружения
	em, err := envs(portEnv, dbEnv)
	if err != nil {
		return err
	}

	db, err := mongo.New(em[dbEnv], "rbtest", "items")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// создание контекста для регулирования
	// закрытие всех подсистем
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cl := log.New(os.Stdout, "Cache:", log.Lmsgprefix|log.LstdFlags)
	cache := cache.New(ctx, db, cl, cacheUpdInterval)

	var wg sync.WaitGroup
	wg.Add(1)

	al := log.New(os.Stdout, "API:", log.Lmsgprefix|log.LstdFlags)

	servers := []*http.Server{
		startRestServer(cache, al, em, &wg),
	}

	// логика закрытия сервера
	cancelation(cancel, servers)

	wg.Wait()

	return nil
}

// cancellation отслеживает сигналы прерывания и,
// если они получены, "мягко" отменяет контекст приложения и
// гасит серверы.
func cancelation(cancel context.CancelFunc, servers []*http.Server) {
	// ловим сигналов прерывания, типа CTRL-C
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-stop // получили сигнал
		log.Printf("got signal %q", sig)

		// закрываем серверы
		for i := range servers {
			if err := servers[i].Shutdown(context.Background()); err != nil {
				log.Fatal(err)
			}
		}

		cancel() // закрываем контекст приложения
	}()
}

// envs собирает ожидаемые переменные окружения,
// возвращает ошибку, если какая-либо из переменных env не задана.
func envs(envs ...string) (map[string]string, error) {
	em := make(map[string]string, len(envs))
	var ok bool
	for _, env := range envs {
		if em[env], ok = os.LookupEnv(env); !ok {
			return nil, fmt.Errorf("environment variable %q must be set", env)
		}
	}
	return em, nil
}

// startRestServer запускает сервер REST API.
func startRestServer(db domain.Repository, logger *log.Logger, env map[string]string, wg *sync.WaitGroup) *http.Server {
	// REST API
	api := api.New(db, logger, cacheUpdInterval)

	// конфигурируем сервер
	srv := &http.Server{
		Addr:              env[portEnv],
		Handler:           api.Router(),
		IdleTimeout:       3 * time.Minute,
		ReadHeaderTimeout: time.Minute,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
		logger.Println("server is shut down")
		wg.Done()
	}()
	return srv
}
