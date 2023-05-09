package locator

type Locator struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	TimeZone string `json:"timeZone"`
}
