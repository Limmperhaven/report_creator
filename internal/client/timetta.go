package client

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strings"
)

type TimettaClient struct {
	Client      *http.Client
	AccessToken string
}

type WorkDay struct {
	Date      string
	Hours     int64
	Comment   string
	User      Employee
	ProjectId string
}

type Employee struct {
	Id  string
	Fio string
}

type TimeSheet struct {
	Id   string
	User Employee
}

func NewTimettaClient() *TimettaClient {
	return &TimettaClient{&http.Client{}, ""}
}

func (c *TimettaClient) Authorize(email, password string) error {
	url := "https://auth.timetta.com/connect/token"
	method := "POST"
	payload := strings.NewReader("client_id=external&scope=all offline_access&grant_type=password&username=" + email +
		"&password=" + password)

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return errors.New(fmt.Sprintf("error creating request: %s", err.Error()))
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.Client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("error processing request: %s", err.Error()))
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("error reading response: %s", err.Error()))
	}

	c.AccessToken = fmt.Sprintf("Bearer %s", gjson.Get(string(body), "access_token").String())
	return nil
}

func (c *TimettaClient) GetMembersByProjects(projects map[string]interface{}) ([]Employee, error) {
	result := make([]Employee, 0)

	for k := range projects {
		url := fmt.Sprintf("https://api.timetta.com/odata/Projects(%s)/ProjectTeamMembers?$expand=resource($select=name)", k)
		method := "GET"

		req, err := http.NewRequest(method, url, nil)

		if err != nil {
			return nil, errors.New(fmt.Sprintf("error creating request: %s", err.Error()))
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", c.AccessToken)

		res, err := c.Client.Do(req)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error processing request: %s", err.Error()))
		}

		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			return nil, errors.New(fmt.Sprintf("error reading response: %s", err.Error()))
		}

		members := gjson.Get(string(body), "value").Array()

		for _, mem := range members {
			obj := mem.Map()
			result = append(result, Employee{Id: obj["resourceId"].String(), Fio: mem.Get("resource.name").String()})
		}

		res.Body.Close()
	}

	return result, nil
}

func (c *TimettaClient) GetUserTimeSheets(user Employee, dateFrom, dateTo string) ([]TimeSheet, error) {
	url := "https://api.timetta.com/odata/TimeSheets?$filter=dateFrom%20ge%20" + dateFrom + "%20and%20dateTo%20le%20" + dateTo + "%20and%20userId%20eq%20" + user.Id
	method := "GET"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("error creating request: %s", err.Error()))
	}

	req.Header.Add("Authorization", c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error processing request: %s", err.Error()))
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("error reading response: %s", err.Error()))
	}

	result := make([]TimeSheet, 0)

	timeSheets := gjson.Get(string(body), "value").Array()

	for _, ts := range timeSheets {
		result = append(result, TimeSheet{Id: ts.Get("id").String(), User: user})
	}

	return result, nil
}

func (c *TimettaClient) GetWorkDays(ts TimeSheet, projects map[string]interface{}) ([]WorkDay, error) {
	result := make([]WorkDay, 0)

	url := fmt.Sprintf("https://api.timetta.com/odata/TimeSheets(%s)?$expand=timeSheetLines($orderBy=orderNumber;$select=id,orderNumber,projectId,projectTaskId,activityId,roleId;$expand=timeAllocations($select=id,date,duration,comments)),approvalStatus($select=id,name),user($select=id,name)&$select=id,dateFrom,dateTo,rowVersion", ts.Id)
	method := "GET"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("error creating request: %s", err.Error()))
	}

	req.Header.Add("Authorization", c.AccessToken)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error processing request: %s", err.Error()))
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("error reading response: %s", err.Error()))
	}

	for k := range projects {
		timeAllocations := gjson.Get(string(body), fmt.Sprintf("timeSheetLines.#(projectId=%s).timeAllocations",
			k)).Array()

		for _, ta := range timeAllocations {
			result = append(result, WorkDay{
				Date:      ta.Get("date").String(),
				Hours:     ta.Get("duration").Int(),
				Comment:   ta.Get("comments").String(),
				User:      ts.User,
				ProjectId: k,
			})
		}
	}

	return result, nil
}
