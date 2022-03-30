package main

import (
	"sync"

	"github.com/kingcobra2468/cot/internal/config"
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/kingcobra2468/cot/internal/text"
	"github.com/kingcobra2468/cot/internal/text/gvoice"
	"github.com/spf13/viper"
)

// init initialized the parsing of the config file and associated
// environment variables.
func init() {
	viper.SetConfigName("cot_sm")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("COT")
	viper.BindEnv("text_encryption", "gvms_hostname", "gvms_port")
	viper.AutomaticEnv()
}

// parseServices retrieves all of the services that have been registered
// in the config file.
func parseServices() (*config.Services, error) {
	var c config.Services
	err := viper.Unmarshal(&c)

	return &c, err
}

// parseGVMSC retrieves GVMS connection configuration.
func parseGVMS() (*config.GVMSConfig, error) {
	var c config.GVMSConfig
	err := viper.Sub("gvms").Unmarshal(&c)

	return &c, err
}

func main() {
	// read config
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	// read in service config and check integrity
	services, err := parseServices()
	if err != nil {
		panic(err)
	}
	// read in gvms config and check integrity
	gvms, err := parseGVMS()
	if err != nil {
		panic(err)
	}
	// register gvms connection config with gvms client
	gvoice.Setup(gvms.Hostname, gvms.Port)

	// create cache and register all services with it
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
