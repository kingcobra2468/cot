package main

import (
	"sync"

	"github.com/kingcobra2468/cot/internal/config"
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/kingcobra2468/cot/internal/text"
	"github.com/kingcobra2468/cot/internal/text/crypto"
	"github.com/kingcobra2468/cot/internal/text/gvoice"
	"github.com/spf13/viper"
)

// init initialized the parsing of the config file and associated
// environment variables.
func init() {
	viper.SetConfigName("cot_sm")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("cot")
	viper.AutomaticEnv()

	viper.BindEnv("text_encryption")
	viper.BindEnv("public_key_file")
	viper.BindEnv("private_key_file")
	viper.BindEnv("passphrase")
	viper.BindEnv("cn_public_key_dir")
	viper.BindEnv("sig_verification")
	viper.BindEnv("base64_encoding")
}

// parseServices retrieves all of the services that have been registered
// in the config file.
func parseServices() (*config.Services, error) {
	var c config.Services
	err := viper.Unmarshal(&c)

	return &c, err
}

// parseGVMS retrieves GVMS connection configuration.
func parseGVMS() (*config.GVMS, error) {
	var c config.GVMS
	err := viper.Sub("gvms").Unmarshal(&c)

	return &c, err
}

// parseGVMSC retrieves GVMS connection configuration.
func parseEncryption() (*config.Encryption, error) {
	var c config.Encryption
	err := viper.Unmarshal(&c)

	return &c, err
}

func main() {
	// read config
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	// read in service config and check integrity
	sc, err := parseServices()
	if err != nil {
		panic(err)
	}
	// read in gvms config and check integrity
	gvms, err := parseGVMS()
	if err != nil {
		panic(err)
	}
	// register gvms connection config with gvms client
	gvoice.Setup(gvms)

	// read in gvms config and check integrity
	encryption, err := parseEncryption()
	if err != nil {
		panic(err)
	}

	if encryption.TextEncryption {
		err := crypto.SetConfig(encryption)
		if err != nil {
			panic(err)
		}

		crypto.LoadClientNumberKeys(encryption.ClientNumberPublicKeyDir)
	}

	done := make(chan struct{})

	// create cache and register all services with it
	services := service.GenerateServices(sc)
	serviceCache := service.NewCache()
	serviceCache.Add(services...)

	listeners := text.GenerateListeners(sc)
	commandExecutor := text.NewExecutor(5, len(*listeners), serviceCache)
	commandExecutor.AddRecipient((*listeners)...)

	wg := sync.WaitGroup{}
	wg.Add(1)
	commandExecutor.Start(done)
	wg.Wait()
}
