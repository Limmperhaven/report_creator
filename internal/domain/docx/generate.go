package docx

import (
	"errors"
	docxt "github.com/qida/go-docx-templates"
	"github.com/spf13/viper"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/client"
	"strconv"
)

type TemplateStruct struct {
	Date       string
	TotalHours string
	Rows       []TemplateRow
}

type TemplateRow struct {
	Date      string
	Comment   string
	Hours     string
	Category  string
	Initiator string
}

func NewTemplateStruct(wds []client.WorkDay, categories, projects map[string]interface{}) (TemplateStruct, error) {
	ts := TemplateStruct{Date: viper.GetString("timetta.settings.documentDate")}

	var totalHrs int64 = 0

	for _, wd := range wds {
		totalHrs += wd.Hours

		category, exists := categories[wd.User.Fio]
		if !exists {
			category = "Разработка"
		}

		cat, ok := category.(string)
		if !ok {
			return *new(TemplateStruct), errors.New("invalid category detected")
		}

		init, ok := projects[wd.ProjectId].(string)
		if !ok {
			return *new(TemplateStruct), errors.New("invalid initiator detected")
		}

		ts.Rows = append(ts.Rows, TemplateRow{
			Date:      wd.Date,
			Comment:   wd.User.Fio + ": " + wd.Comment,
			Hours:     strconv.Itoa(int(wd.Hours)),
			Category:  cat,
			Initiator: init,
		})
	}

	ts.TotalHours = strconv.Itoa(int(totalHrs))

	return ts, nil
}

func (t *TemplateStruct) Generate(path string) error {
	template, err := docxt.OpenTemplate("etc/template.docx")
	if err != nil {
		return err
	}

	if err := template.RenderTemplate(t); err != nil {
		return err
	}

	if err := template.Save(path + "result.docx"); err != nil {
		return err
	}

	return nil
}
