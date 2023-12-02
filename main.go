/*
Copyright © 2023 Dor Munis <dormunis@gmail.com>
*/
package main

import (
	"github.com/dormunis/punch/cmd/cli"
	"github.com/dormunis/punch/pkg/config"
	"log"
	"os"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Unable to retrieve config: %v", err)
		os.Exit(1)
	}
	err = cli.Execute(cfg)
	if err != nil {
		log.Fatalf("Unable to execute command: %v", err)
		os.Exit(1)
	}
}
