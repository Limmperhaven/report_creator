package main

import (
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/client"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/config"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/domain"
	"log"
)

func main() {
	log.Println("Reading configs...")
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("error reading config: %s", err.Error())
	}

	log.Println("Reading flags...")
	output := config.ParseFlags()

	log.Println("Creating http client...")
	cl := client.NewTimettaClient()

	domain.GenerateReport(cl, output, cfg)
}
