package domain

import (
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/client"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/config"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/domain/docx"
	"log"
)

func GenerateReport(cl *client.TimettaClient, path string, config *config.Config) {
	log.Println("Authorizing your account...")
	if err := cl.Authorize(config.Timetta.Credentials.Email, config.Timetta.Credentials.Password); err != nil {
		log.Fatalf("authorization error: %s", err.Error())
	}

	log.Println("Getting project members...")
	members, err := cl.GetMembersByProjects(config.Projects)
	if err != nil {
		log.Fatalf("error getting project members: %s", err.Error())
	}

	allTimeSheets := make([]client.TimeSheet, 0)

	log.Println("Getting TimeSheets...")
	for _, mem := range members {
		userTimeSheets, err := cl.GetUserTimeSheets(mem, config.Timetta.Settings.DateFrom, config.Timetta.Settings.DateTo)
		if err != nil {
			log.Fatalf("error getting user timesheets: %s", err.Error())
		}

		allTimeSheets = append(allTimeSheets, userTimeSheets...)
	}

	allWorkDays := make([]client.WorkDay, 0)

	log.Println("Getting working hours...")
	for _, ts := range allTimeSheets {
		wd, err := cl.GetWorkDays(ts, config.Projects)
		if err != nil {
			log.Fatalf("error getting workdays: %s", err.Error())
		}

		allWorkDays = append(allWorkDays, wd...)
	}

	log.Println("Processing collected info...")
	ts, err := docx.NewTemplateStruct(allWorkDays, config.Categories, config.Projects)
	if err != nil {
		log.Fatalf("error creating template struct: %s", err.Error())
	}

	log.Println("Generating document...")
	if err := ts.Generate(path); err != nil {
		log.Fatalf("error generating document: %s", err.Error())
	}

	log.Println("Completed")
}
