package main

import (
	"sync"

	"github.com/kingcobra2468/cot/internal/config"
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/kingcobra2468/cot/internal/text"
	"github.com/kingcobra2468/cot/internal/text/gvoice"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("cot_sm")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("COT")
	viper.BindEnv("text_encryption")
	viper.AutomaticEnv()
}

func parseConfig() (*config.Services, error) {
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var services config.Services
	err = viper.Unmarshal(&services)

	return &services, err
}

func main() {
	services, err := parseConfig()
	if err != nil {
		panic(err)
	}

	// TODO: fetch env vars when setting config below
	gvoice.Init("0.0.0.0", 50051)

	serviceCache := service.NewCache()
	serviceCache.Add(services.Names()...)

	serviceRouter := service.NewRouter(serviceCache)
	listeners := services.Listeners()
	commandExecutor := text.NewExecutor(5, len(*listeners), serviceRouter)
	commandExecutor.AddRecipient((*listeners)...)
	done := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(1)
	commandExecutor.Start(done)
	wg.Wait()

}
