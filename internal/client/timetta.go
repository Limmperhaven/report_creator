package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"

	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/config"
)

type TimettaClient struct {
	Client      *http.Client
	AccessToken string
}

type WorkDay struct {
	Date      string
	Hours     float64
	Comment   string
	User      Person
	ProjectId string
}

type Person struct {
	Id        string
	Fio       string
	ProjectID string
}

type TimeSheet struct {
	Id   string
	User Person
}

var timetta *TimettaClient

func InitTimettaClient(email, password string) error {
	timetta = &TimettaClient{&http.Client{}, ""}

	url := "https://auth.timetta.com/connect/token"
	method := "POST"
	payload := strings.NewReader("client_id=external&scope=all offline_access&grant_type=password&username=" + email +
		"&password=" + password)

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return fmt.Errorf("error creating request: %s", err.Error())
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := timetta.Client.Do(req)
	if err != nil {
		return fmt.Errorf("error processing request: %s", err.Error())
	}

	if res.StatusCode == 400 {
		return errors.New("invalid email or password")
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %s", err.Error())
	}

	timetta.AccessToken = fmt.Sprintf("Bearer %s", gjson.Get(string(body), "access_token").String())
	return nil
}

func GistTimettaClient() *TimettaClient {
	return timetta
}

func (c *TimettaClient) GetMembersByProjects(projects []config.Project) ([]Person, error) {
	result := make([]Person, 0)
	membersMap := make(map[string]bool)

	for _, project := range projects {
		url := fmt.Sprintf("https://api.timetta.com/odata/Projects(%s)/ProjectTeamMembers?$expand=resource($select=name)", project.Id)
		method := "GET"

		req, err := http.NewRequest(method, url, nil)

		if err != nil {
			return nil, fmt.Errorf("error creating request: %s", err.Error())
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", c.AccessToken)

		res, err := c.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error processing request: %s", err.Error())
		}

		body, err := io.ReadAll(res.Body)

		if err != nil {
			return nil, fmt.Errorf("error reading response: %s", err.Error())
		}

		members := gjson.Get(string(body), "value").Array()

		for _, mem := range members {
			obj := mem.Map()

			_, exists := membersMap[obj["resourceId"].String()]

			if !exists {
				result = append(result, Person{
					Id:        obj["resourceId"].String(),
					Fio:       mem.Get("resource.name").String(),
					ProjectID: project.Id,
				})
				membersMap[obj["resourceId"].String()] = true
			}
		}

		res.Body.Close()
	}

	return result, nil
}

func (c *TimettaClient) GetUserTimeSheets(user Person, dateFrom, dateTo string) ([]TimeSheet, error) {
	url := "https://api.timetta.com/odata/TimeSheets?$filter=dateFrom%20ge%20" + dateFrom + "%20and%20dateTo%20le%20" + dateTo + "%20and%20userId%20eq%20" + user.Id
	method := "GET"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err.Error())
	}

	req.Header.Add("Authorization", c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error processing request: %s", err.Error())
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading response: %s", err.Error())
	}

	result := make([]TimeSheet, 0)

	timeSheets := gjson.Get(string(body), "value").Array()

	for _, ts := range timeSheets {
		result = append(result, TimeSheet{Id: ts.Get("id").String(), User: user})
	}

	return result, nil
}

func (c *TimettaClient) GetWorkDays(ts TimeSheet, projects []config.Project) ([]WorkDay, error) {
	result := make([]WorkDay, 0)

	url := fmt.Sprintf("https://api.timetta.com/odata/TimeSheets(%s)?$expand=timeSheetLines($orderBy=orderNumber;$select=id,orderNumber,projectId,projectTaskId,activityId,roleId;$expand=timeAllocations($select=id,date,duration,comments)),state($select=id,name),user($select=id,name)&$select=id,dateFrom,dateTo,rowVersion", ts.Id)
	method := "GET"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err.Error())
	}

	req.Header.Add("Authorization", c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error processing request: %s", err.Error())
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading response: %s", err.Error())
	}

	state := gjson.Get(string(body), "state.name").String()
	if state != "Согласовано" && state != "На согласовании" && state != "Черновик" {
		return result, nil
	}

	for _, project := range projects {
		timeSheetLines := gjson.Get(string(body), fmt.Sprintf("timeSheetLines.#(projectId=%s)#",
			project.Id)).Array()

		for _, tsl := range timeSheetLines {
			timeAllocations := tsl.Get("timeAllocations").Array()
			for _, ta := range timeAllocations {
				result = append(result, WorkDay{
					Date:      ta.Get("date").String(),
					Hours:     ta.Get("duration").Float(),
					Comment:   ta.Get("comments").String(),
					User:      ts.User,
					ProjectId: project.Id,
				})
			}
		}
	}

	return result, nil
}
