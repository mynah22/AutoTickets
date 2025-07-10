package api

import (
	"AutoTickets/tickets"
	"fmt"
	"io"
	"net/http"

	"github.com/tidwall/gjson"
)

const openTicketsQueryUrl = `https://webservices14.autotask.net/atservicesrest/v1.0/tickets/query?search={"filter":[{"op":"noteq","field":"Status","value":5}]}&pagesize=200`

// polls API and returns open tickets
func GetOpenTickets(apiIntegrationCode, apiSecret, apiUsername string) ([]tickets.AutotaskTicket, error) {
	req, err := http.NewRequest("GET", openTicketsQueryUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("ApiIntegrationCode", apiIntegrationCode)
	req.Header.Set("Secret", apiSecret)
	req.Header.Set("UserName", apiUsername)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response from API: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var openTickets []tickets.AutotaskTicket
	jsonStr := string(body)
	gjson.Get(jsonStr, "items").
		ForEach(func(_, t gjson.Result) bool {
			openTickets = append(openTickets, tickets.AutotaskTicket{
				ID:                 t.Get("id").Int(),
				AssignedResourceID: t.Get("assignedResourceID").String(),
				CreateDate:         t.Get("createDate").String(),
				Description:        t.Get("description").String(),
				Title:              t.Get("title").String(),
			})
			return true
		})

	return openTickets, nil
}
