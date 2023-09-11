package csv

import (
	"encoding/csv"
	"fmt"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/client"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/config"
	"os"
)

func GenerateCSV(wds []client.WorkDay, cfg *config.Config) error {
	projectsCosts := make(map[string]float64)

	for _, wd := range wds {
		projectsCosts[wd.ProjectId] += wd.Hours * float64(cfg.EmployeeById[wd.User.Id].Salary)
	}

	data := make([][]string, 0, len(cfg.Projects))

	for _, project := range cfg.Projects {
		data = append(data, []string{
			project.Name,
			cfg.Timetta.Settings.DateFrom + " - " + cfg.Timetta.Settings.DateTo,
			fmt.Sprintf("%.2f", projectsCosts[project.Id]),
		})
	}

	var path string

	if cfg.Output == "" {
		path = "Costs.csv"
	} else {
		path = cfg.Output + "/Costs.csv"
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	writer := csv.NewWriter(file)

	err = writer.WriteAll(data)
	if err != nil {
		return err
	}

	return nil
}
