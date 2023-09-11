package docx

import (
	docxt "github.com/qida/go-docx-templates"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/utils"

	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/client"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/config"
)

type TemplateStruct struct {
	Date        string
	Initiator   string
	ProjectCode string
	TotalHours  string
	Rows        []TemplateRow
}

type TemplateRow struct {
	Date     string
	Comment  string
	Hours    string
	Category string
}

type UserComment struct {
	UserId  string
	Comment string
}

func NewTemplateStruct(wds []client.WorkDay, cfg *config.Config, project config.Project) (TemplateStruct, error) {
	ts := TemplateStruct{
		Date:        cfg.Timetta.Settings.DocumentDate,
		Initiator:   project.Initiator,
		ProjectCode: project.Code,
	}

	var totalHrs float64 = 0
	var wdsMap = make(map[UserComment]client.WorkDay)

	for _, wd := range wds {
		totalHrs += wd.Hours
		mwd, ok := wdsMap[UserComment{wd.User.Id, wd.Comment}]
		if ok {
			wdsMap[UserComment{wd.User.Id, wd.Comment}] = client.WorkDay{
				Date:      mwd.Date,
				Hours:     wd.Hours + mwd.Hours,
				Comment:   mwd.Comment,
				User:      mwd.User,
				ProjectId: mwd.ProjectId,
			}
		} else {
			wdsMap[UserComment{wd.User.Id, wd.Comment}] = wd
		}
	}

	for _, wd := range wdsMap {
		emp, exists := cfg.EmployeeById[wd.User.Id]
		if !exists {
			emp.Category = "Разработка"

		}

		ts.Rows = append(ts.Rows, TemplateRow{
			Date:     wd.Date,
			Comment:  wd.User.Fio + ": " + wd.Comment,
			Hours:    utils.FormatFloat64(wd.Hours),
			Category: emp.Category,
		})
	}

	ts.TotalHours = utils.FormatFloat64(totalHrs)

	return ts, nil
}

func (t *TemplateStruct) Generate(path, name string) error {
	template, err := docxt.OpenTemplate("etc/template1.docx")
	if err != nil {
		return err
	}

	if err := template.RenderTemplate(t); err != nil {
		return err
	}

	var savePath string

	if path == "" {
		savePath = name + ".docx"
	} else {
		savePath = path + "/" + name + ".docx"
	}

	if err := template.Save(savePath); err != nil {
		return err
	}

	return nil
}
