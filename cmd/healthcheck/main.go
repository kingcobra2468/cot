package main

import (
	"errors"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/kingcobra2468/cot/internal/config"
	"github.com/kingcobra2468/cot/internal/router/worker"
	"github.com/kingcobra2468/cot/internal/router/worker/gvoice"
	"github.com/spf13/viper"
)

// init initializes the parsing of the config file and associated
// environment variables.
func init() {
	viper.SetConfigName("cot_sm")
	viper.SetConfigType("yaml")

	if path, exists := os.LookupEnv("COT_CONF_DIR"); exists {
		viper.AddConfigPath(path)
	}
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("cot")
	viper.AutomaticEnv()
}

// parseGVMS retrieves GVMS connection configuration.
func parseGVMS() (*config.GVMS, error) {
	var c config.GVMS
	err := viper.Sub("gvms").Unmarshal(&c)

	return &c, err
}

// ping checks if COT is online.
func ping(number string) error {
	l := worker.NewGVoiceWorker(gvoice.Link{GVoiceNumber: "14159422253", ClientNumber: "14159422253"}, false, 0)
	if err := l.Send("ping"); err != nil {
		glog.Error(err)
		return nil
	}
	time.Sleep(time.Second * 20)
	x := *l.Fetch()
	for _, ui := range x {
		if strings.EqualFold(ui.Name, "pong") {
			return nil
		}
	}

	return errors.New("unable to ping cot")
}

func main() {
	flag.Parse()

	// read config
	err := viper.ReadInConfig()
	if err != nil {
		glog.Error(err)
	}

	// read in gvms config and check integrity
	gvms, err := parseGVMS()
	if err != nil {
		glog.Error(err)
	}
	// register gvms connection config with gvms client
	gvoice.Setup(gvms)

	gvoiceNumber := viper.GetString("gvoice_number")
	if len(gvoiceNumber) == 0 {
		glog.Error("unable to parse gvoice number")
	}

	glog.Infof("send \"ping\" to %s", gvoiceNumber)

	if err := ping(gvoiceNumber); err != nil {
		os.Exit(1)
	}
}
