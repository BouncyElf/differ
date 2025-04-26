package main

import (
	"differ/config"
	"differ/handlers"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: differ <config_file>")
	}
	err := config.InitConfig(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid Config file: %v", err)
	}
	app := handlers.InitApp()
	app.Run(fmt.Sprintf(":%d", config.Conf.ProxyConfig.Port))

}
