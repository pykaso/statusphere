package status_pages

import "github.com/metoro-io/statusphere/common/api"

var StatusPages = []api.StatusPage{
	{
		URL:  "https://status.secure.payu.com/",
		Name: "PayU - Platební brána",
	}, {
		URL:  "https://api.instacover.ai/instacar/v2.0/status",
		Name: "Instacover - Focení HAV",
	}, {
		URL:  "https://www.supin.cz/Aplikace/Support/RSS/",
		Name: "ČKP",
	}, {
		URL:  "http://10.1.10.102:5050/rss",
		Name: "CKP - TEST",
	}, {
		URL:  "http://10.1.10.102:5050/status",
		Name: "Instacover - Test",
	},
}
