package servers

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RandySteven/go-kopi/apps"
	"github.com/RandySteven/go-kopi/configs"
	"github.com/RandySteven/go-kopi/routes"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type (
	Server struct {
		Host string
		Port string

		ServerTimeout time.Duration
		ReadTimeout   time.Duration
		WriteTimeout  time.Duration
		IdleTimeout   time.Duration
	}
)

func init() {
	err := godotenv.Load("./files/env/.env")
	if err != nil {
		log.Fatalln(`failed to load .env `, err)
		return
	}
}

func getConfig() *configs.Config {
	configPath, err := configs.ParseFlags()
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	config, err := configs.NewConfig(configPath)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return config
}

func initServer(config *configs.Config) *Server {

	serverConfig := config.Configs.Server

	server := &Server{
		Host:          serverConfig.Host,
		Port:          serverConfig.Port,
		ServerTimeout: serverConfig.Timeout.Server,
		ReadTimeout:   serverConfig.Timeout.Read,
		WriteTimeout:  serverConfig.Timeout.Write,
		IdleTimeout:   serverConfig.Timeout.Idle,
	}

	return server
}

// Run starts the HTTP server with graceful shutdown support.
// It listens for SIGINT and SIGTERM signals and performs a graceful shutdown
// with a 5-second timeout when a termination signal is received.
// The server configuration (host, port, timeouts) is read from the Config struct.
func (s *Server) run(r *mux.Router) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:         s.Host + ":" + s.Port,
		Handler:      r,
		ReadTimeout:  s.ReadTimeout * time.Second,
		WriteTimeout: s.WriteTimeout * time.Second,
		IdleTimeout:  s.IdleTimeout * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func Serve() {
	server := initServer(getConfig())
	ctx := context.TODO()

	app, err := apps.NewApp(getConfig())
	if err != nil {
		log.Fatalln(`Error starting app `, err)
		return
	}

	apis := app.PrepareHttpHandler(ctx)
	r := mux.NewRouter()
	if apis == nil {
		log.Fatalln("api is not valid")
		return
	}
	router := routes.NewEndpointRouters(apis)
	routes.InitRouter(router, r)

	if err = app.Temporal.Start(); err != nil {
		log.Fatalln("Failed to start Temporal worker:", err)
		return
	}
	defer app.Temporal.Stop()

	go server.run(r)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = app.RefreshRedis(ctx); err != nil {
		log.Fatal(err)
		return
	}
}
