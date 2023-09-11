package domain

import (
	"fmt"
	"log"

	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/client"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/config"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/domain/csv"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/domain/docx"
)

func GenerateReport(cl *client.TimettaClient, config *config.Config) {

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
	for _, project := range config.Projects {
		projectWorkDays := make([]client.WorkDay, 0)

		for _, wd := range allWorkDays {
			if wd.ProjectId == project.Id {
				projectWorkDays = append(projectWorkDays, wd)
			}
		}

		//for _, pwd := range projectWorkDays {
		//	if pwd.User.Id == "398c3d99-3261-4e54-94fa-5ce59420e3e4" {
		//		fmt.Println(pwd.Comment, pwd.Hours)
		//	}
		//}

		ts, err := docx.NewTemplateStruct(projectWorkDays, config, project)
		if err != nil {
			log.Fatalf("error creating template struct: %s", err.Error())
		}

		log.Printf("Generating document `%s`...\n", project.Name)
		docName := fmt.Sprintf("Рабочий отчет_%s_%s", config.Timetta.Settings.DocumentDate[:7], project.Code)

		if err := ts.Generate(config.Output, docName); err != nil {
			log.Fatalf("error generating document: %s", err.Error())
		}
	}

	log.Println("Generating csv...")
	if err := csv.GenerateCSV(allWorkDays, config); err != nil {
		log.Fatalf("error generating csv: %s", err.Error())
	}

	log.Println("Completed")
}

func PrintMembers(cl *client.TimettaClient, config *config.Config) {
	log.Println("Getting project members...")
	members, err := cl.GetMembersByProjects(config.Projects)
	if err != nil {
		log.Fatalf("error getting project members: %s", err.Error())
	}

	fmt.Print("\n\n\n")

	for _, project := range config.Projects {
		fmt.Printf("%s - %s - %s\n\n", project.Id, project.Name, project.Initiator)
		for _, member := range members {
			if member.ProjectID == project.Id && member.Id != "" && member.Fio != "" {
				fmt.Printf("%s: %s\n", member.Fio, member.Id)
			}
		}

		fmt.Print("\n\n")
	}
}
