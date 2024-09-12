package status_pages

import (
	"github.com/metoro-io/statusphere/common/api"
)

var PageAres = api.StatusPage{
	URL:              "https://ares.gov.cz/ekonomicke-subjekty-v-be/rest/ekonomicke-subjekty/25073958",
	Name:             "ARES - ekonomicke subjekty",
	PreferredScraper: "REST",
	Method:           api.MethodHead,
}

var PageInstacover = api.StatusPage{
	URL:  "https://api.instacover.ai/instacar/v2.0/status",
	Name: "Instacover - Focení HAV",
	ValidationRules: map[string]interface{}{
		"status": "OK",
	},
	PreferredScraper: "REST",
	Method:           api.MethodGet,
}

var PageSmartform = api.StatusPage{
	URL:  "https://api.smartform.cz/oracleAddress/v11",
	Name: "Smartform",
	Headers: map[string]string{
		"Authorization": "Basic {env.SMARTFORM_AUTH_HEADER}",
		"Content-Type":  "application/json",
		"test":          "true",
	},
	RequestPayload: map[string]interface{}{
		"fieldType": "STREET_AND_NUMBER",
		"values": map[string]string{
			"STREET_AND_NUMBER": "nové sady",
			"CITY":              "Brno",
			"ZIP":               "",
		},
		"country": "CZ",
		"limit":   1,
	},
	Method: api.MethodPost,
	ValidationRules: map[string]interface{}{
		"resultCode": "TEST",
	},
	PreferredScraper: "REST",
}

var PagePayU = api.StatusPage{
	URL:              "https://status.secure.payu.com/",
	Name:             "PayU - Platební brána",
	PreferredScraper: "Atlassian",
}

var PageCKP = api.StatusPage{
	URL:              "https://www.supin.cz/Aplikace/Support/RSS/",
	Name:             "ČKP",
	PreferredScraper: "CKP_RSS",
}

var PageIPEX = api.StatusPage{
	URL:  "https://direct.servicedesk.net/api/Tickets/GetTickets",
	Name: "IPEX",
	Headers: map[string]string{
		"Authorization": "Basic {env.IPEX_AUTH_HEADER}",
		"Content-Type":  "application/json",
		"test":          "true",
	},
	PreferredScraper: "REST",
	Method:           api.MethodGet,
}

// seznam vsech ktere chceme overovat
var StatusPages = []api.StatusPage{
	PageAres,
	PageCKP,
	PagePayU,
	PageSmartform,
	PageInstacover,
	PageIPEX,
}
