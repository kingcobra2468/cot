package main

import (
	"fmt"

	"github.com/kingcobra2468/cot/internal/config"
	"github.com/spf13/viper"
)

func parseConfig() (*config.Services, error) {
	viper.SetConfigName("cot_sm")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

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

	fmt.Println(services)
}
