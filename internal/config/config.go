package config

type Services struct {
	Services []Service `mapstructure:"services"`
}

type Service struct {
	Name          string `mapstructure:"name"`
	Domain        string `mapstructure:"domain"`
	ClientNumbers []int  `mapstructure:"client_numbers"`
}
