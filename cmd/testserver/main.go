package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Hari-Kiri/virest/cmd/testserver/auth"
	"github.com/Hari-Kiri/virest/cmd/testserver/handlers"
	"github.com/Hari-Kiri/virest/storagePool"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/Hari-Kiri/virest/cmd/testserver/docs"
)

//	@title			ViRest Storage Pool Testserver
//	@version		1.0
//	@description	Optional HTTP harness for the storagepool Go library. Authenticate with Basic Auth to get a JWT, then call pool APIs with Bearer token and Hypervisor-Uri header. Use Authorize in Swagger UI to set credentials (Postman-like try-it-out).
//	@host			localhost:8000
//	@BasePath		/
//
//	@securityDefinitions.basic	BasicAuth
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT from /storage-pool/authenticate. Prefer value: Bearer {token}
func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
	loadDotEnv(".env")

	appName := envOr("VIREST_STORAGE_POOL_APPLICATION_NAME", "virest-storage-pool-testserver")
	port := mustAtoi(envOr("VIREST_STORAGE_POOL_APPLICATION_PORT", "8000"), "port")
	lifetimeSec := mustAtoi(envOr("VIREST_STORAGE_POOL_APPLICATION_JWT_LIFETIME_SECONDS", "3600"), "jwt lifetime")
	store := mustAuthStore(appName, envOr("VIREST_USERS_FILE", "users.yaml"), lifetimeSec)
	startLibvirtEvents()

	deps := &handlers.Deps{Auth: store}
	runServer(port, appName, registerRoutes(deps))
}

func mustAtoi(raw, label string) int {
	n, err := strconv.Atoi(raw)
	if err != nil {
		slog.Error("invalid "+label, "err", err)
		os.Exit(1)
	}
	return n
}

func mustAuthStore(appName, usersPath string, lifetimeSec int) *auth.Store {
	cfg, err := auth.LoadUsersFile(usersPath)
	if err != nil {
		slog.Error("load users file", "path", usersPath, "err", err)
		os.Exit(1)
	}
	store, err := auth.NewStore(
		cfg,
		appName,
		[]byte(os.Getenv("VIREST_STORAGE_POOL_APPLICATION_JWT_SIGNATURE_KEY")),
		time.Duration(lifetimeSec)*time.Second,
	)
	if err != nil {
		slog.Error("auth store", "err", err)
		os.Exit(1)
	}
	return store
}

func startLibvirtEvents() {
	storagePool.RegisterEventLoop()
}

func registerRoutes(deps *handlers.Deps) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/storage-pool/authenticate", deps.Authenticate)
	mux.HandleFunc("/storage-pool/find-storage-pool-sources", deps.FindSources)
	mux.HandleFunc("/storage-pool/find-sources-as", deps.FindSourcesAs)
	mux.HandleFunc("/storage-pool/list", deps.List)
	mux.HandleFunc("/storage-pool/info", deps.Info)
	mux.HandleFunc("/storage-pool/detail", deps.Detail)
	mux.HandleFunc("/storage-pool/dump", deps.Dump)
	mux.HandleFunc("/storage-pool/name", deps.Name)
	mux.HandleFunc("/storage-pool/uuid-by-name", deps.UUIDByName)
	mux.HandleFunc("/storage-pool/get-uid", deps.ProcessUID)
	mux.HandleFunc("/storage-pool/get-gid", deps.ProcessGID)
	mux.HandleFunc("/storage-pool/define", deps.Define)
	mux.HandleFunc("/storage-pool/define-as", deps.DefineAs)
	mux.HandleFunc("/storage-pool/create-transient", deps.CreateTransient)
	mux.HandleFunc("/storage-pool/create-as", deps.CreateAs)
	mux.HandleFunc("/storage-pool/build", deps.Build)
	mux.HandleFunc("/storage-pool/create", deps.Create)
	mux.HandleFunc("/storage-pool/start", deps.Start)
	mux.HandleFunc("/storage-pool/autostart", deps.Autostart)
	mux.HandleFunc("/storage-pool/destroy", deps.Destroy)
	mux.HandleFunc("/storage-pool/undefine", deps.Undefine)
	mux.HandleFunc("/storage-pool/delete", deps.Delete)
	mux.HandleFunc("/storage-pool/refresh", deps.Refresh)
	mux.HandleFunc("/storage-pool/capabilities", deps.Capabilities)
	mux.HandleFunc("/storage-pool/event", deps.Event)
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	return mux
}

func runServer(port int, appName string, mux *http.ServeMux) {
	addr := fmt.Sprintf(":%d", port)
	cert := os.Getenv("VIREST_TLS_CERT_FILE")
	key := os.Getenv("VIREST_TLS_KEY_FILE")
	slog.Info("testserver listening", "addr", addr, "app", appName, "swagger", fmt.Sprintf("http://localhost:%d/swagger/index.html", port))
	if cert != "" && key != "" {
		slog.Error("listen tls", "err", http.ListenAndServeTLS(addr, cert, key, mux))
		os.Exit(1)
	}
	slog.Info("TLS cert/key not set; serving plaintext (lab only)")
	slog.Error("listen", "err", http.ListenAndServe(addr, mux))
	os.Exit(1)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, row := range strings.Split(string(data), "\n") {
		row = strings.TrimSpace(row)
		if row == "" || strings.HasPrefix(row, "#") {
			continue
		}
		parts := strings.SplitN(row, "=", 2)
		if len(parts) != 2 {
			continue
		}
		_ = os.Setenv(parts[0], parts[1])
	}
}
