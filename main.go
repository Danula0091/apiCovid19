package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type CovidData struct {
	SummaryStats SummaryStats `json:"summaryStats"`
	Cache        Cache        `json:"cache"`
	DataSource   DataSource   `json:"dataSource"`
	RawData      []SumStats   `json:"rawData"`
}

type SummaryStats struct {
	Global   Stats `json:"global"`
	China    Stats `json:"china"`
	NonChina Stats `json:"nonChina"`
}

type Stats struct {
	Confirmed int  `json:"confirmed"`
	Recovered *int `json:"recovered"`
	Deaths    int  `json:"deaths"`
}

type Cache struct {
	LastUpdated          string `json:"lastUpdated"`
	Expires              string `json:"expires"`
	LastUpdatedTimestamp int64  `json:"lastUpdatedTimestamp"`
	ExpiresTimestamp     int64  `json:"expiresTimestamp"`
}

type DataSource struct {
	URL              string `json:"url"`
	LastGithubCommit string `json:"lastGithubCommit"`
	PublishedBy      string `json:"publishedBy"`
	Ref              string `json:"ref"`
}

type SumStats struct {
	FIPS          string `json:"FIPS"`
	Admin2        string `json:"Admin2"`
	ProvinceState string `json:"Province_State,omitempty"`
	CountryRegion string `json:"Country_Region"`
	LastUpdate    string `json:"Last_Update"`
	Lat           string `json:"Lat"`
	Long          string `json:"Long_"`
	Confirmed     string `json:"Confirmed"`
	Deaths        string `json:"Deaths"`
	Recovered     string `json:"Recovered"`
	Active        string `json:"Active"`
	CombKey       string `json:"Combined_Key"`
	IncRate       string `json:"Incident_Rate"`
	CasFatRat     string `json:"Case_Fatality_Ratio"`
}

func (c SumStats) Info() string {
	return fmt.Sprintf("[%s] %s, TotalDeaths: %s, TotalConfirmed: %s, LastUpdate %s\n", c.CountryRegion, c.ProvinceState,
		c.Deaths, c.Confirmed, c.LastUpdate)
}

type loggingRoundTripper struct {
	logger io.Writer
	next   http.RoundTripper
}

func (l loggingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	fmt.Fprintf(l.logger, "[%s] %s %s\n", time.Now().Format(time.ANSIC), r.Method, r.URL)
	return l.next.RoundTrip(r)
}
func main() {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			fmt.Println(req.Response.StatusCode)
			fmt.Println("Redirect!")
			return nil
		},
		Transport: &loggingRoundTripper{
			logger: os.Stdout,
			next:   http.DefaultTransport,
		},
		Timeout: time.Second * 30,
	}
	resp, err := client.Get("https://coronavirus.m.pipedream.net/")
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	fmt.Println("Resp status", resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var r CovidData
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("CovidStats.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	for _, asset := range r.RawData {
		_, err = io.Copy(f, strings.NewReader(asset.Info()))
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(`Copy is completed`)
}
