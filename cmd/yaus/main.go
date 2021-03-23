package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/pippolo84/yaus/internal/short"
	"github.com/pippolo84/yaus/internal/storage"
	"github.com/pippolo84/yaus/internal/yaus"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const cooldown time.Duration = 5 * time.Second

type config struct {
	storagePath     string
	srvAddress      string
	srvTimeoutWrite time.Duration
	srvTimeoutRead  time.Duration
	srvTimeoutIdle  time.Duration
}

func readConfig() (config, error) {
	viper.SetDefault("storage.path", "./")
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.timeout.write", time.Duration(15*time.Second))
	viper.SetDefault("server.timeout.read", time.Duration(15*time.Second))
	viper.SetDefault("server.timeout.idle", time.Duration(60*time.Second))

	// viper.SetEnvPrefix("yaus") // will be uppercased automatically
	// if err := viper.BindEnv("secret"); err != nil {
	// 	return config{}, err
	// }

	// development config directory and path
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join("configs", "devel"))

	if err := viper.ReadInConfig(); err != nil {
		return config{}, err
	}

	return config{
		storagePath:     viper.GetString("storage.path"),
		srvAddress:      viper.GetString("server.address"),
		srvTimeoutWrite: viper.GetDuration("server.timeout.write"),
		srvTimeoutRead:  viper.GetDuration("server.timeout.read"),
		srvTimeoutIdle:  viper.GetDuration("server.timeout.idle"),
	}, nil
}

func main() {
	conf, err := readConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	sugar := logger.Sugar()
	sugar.Infow("configuration",
		zap.String("storage path", conf.storagePath),
		zap.Duration("server write timeout", conf.srvTimeoutWrite),
		zap.Duration("server read timeout", conf.srvTimeoutRead),
		zap.Duration("server idle timeout", conf.srvTimeoutIdle),
	)

	cache, err := storage.NewBadgerBackend(conf.storagePath)
	if err != nil {
		sugar.Fatal(err)
	}
	defer cache.Close()

	shortener := short.NewMD5()

	srv := yaus.NewServer(
		cache,
		shortener,
		sugar,
		yaus.Address(conf.srvAddress),
		yaus.WriteTimeout(conf.srvTimeoutWrite),
		yaus.ReadTimeout(conf.srvTimeoutRead),
		yaus.IdleTimeout(conf.srvTimeoutIdle),
	)

	errs := srv.Run()

	// trap incoming SIGINT and SIGTERM
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// block until a signal or an error from the server is received
	select {
	case err := <-errs:
		sugar.Errorw("yaus", zap.Error(err))
	case sig := <-signalChan:
		sugar.Infow("signal shutdown", zap.String("signal", sig.String()))
	}

	// graceful shutdown the server
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cooldown)
	defer cancelShutdown()

	var wg sync.WaitGroup
	wg.Add(1)

	if err := srv.Shutdown(shutdownCtx, &wg); err != nil {
		sugar.Errorw("shutdown", zap.Error(err))
	}

	wg.Wait()
}
