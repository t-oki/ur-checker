package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handler)
}

func handler() error {
	slackURL := os.Getenv("SLACK_URL")
	if slackURL == "" {
		log.Fatal("ENV[SLACK_URL] not specified")
	}
	floorType := os.Getenv("FLOOR_TYPE")
	if floorType == "" {
		log.Fatal("ENV[FLOOR_TYPE]not specified")
	}
	floorTypes := strings.Split(floorType, ",")
	upperPrice := os.Getenv("UPPER_PRICE")
	if upperPrice == "" {
		log.Fatal("ENV[UPPER_PRICE]not specified")
	}

	v := url.Values{}
	v.Set("rent_high", upperPrice)
	v.Set("shisya", "20")
	v.Set("danchi", "597")
	v.Set("shikibetu", "0")
	v.Set("orderByField", "0")
	v.Set("orderBySort", "0")
	v.Set("pageIndex", "0")
	resp, err := http.PostForm(
		"https://chintai.sumai.ur-net.go.jp/chintai/api/bukken/detail/detail_bukken_room/",
		v,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	urResList := &[]urRes{}
	if err := json.Unmarshal(body, urResList); err != nil {
		log.Fatal(err)
	}
	for _, r := range *urResList {
		for _, t := range floorTypes {
			if r.Type == t {
				if err := postSlack(r, slackURL); err != nil {
					log.Fatal(err)
				}
				break
			}
		}
	}
	return nil
}

type urRes struct {
	PageIndex        string `json:"pageIndex"`
	RowMax           string `json:"rowMax"`
	RowMaxSp         string `json:"rowMaxSp"`
	RowMaxNext       string `json:"rowMaxNext"`
	PageMax          string `json:"pageMax"`
	AllCount         string `json:"allCount"`
	Block            string `json:"block"`
	Tdfk             string `json:"tdfk"`
	Shisya           string `json:"shisya"`
	Danchi           string `json:"danchi"`
	Shikibetu        string `json:"shikibetu"`
	FloorAll         string `json:"floorAll"`
	RoomDetailLink   string `json:"roomDetailLink"`
	RoomDetailLinkSp string `json:"roomDetailLinkSp"`
	System           []struct {
		IMG           string `json:"制度_IMG"`
		NAMING_FAILED string `json:"制度名"`
		HTML          string `json:"制度HTML"`
	} `json:"system"`
	Parking       interface{}   `json:"parking"`
	Design        []interface{} `json:"design"`
	FeatureParam  []interface{} `json:"featureParam"`
	Traffic       interface{}   `json:"traffic"`
	Place         interface{}   `json:"place"`
	Kanris        interface{}   `json:"kanris"`
	Kouzou        interface{}   `json:"kouzou"`
	Soukosu       interface{}   `json:"soukosu"`
	ID            string        `json:"id"`
	Year          interface{}   `json:"year"`
	Name          string        `json:"name"`
	Shikikin      string        `json:"shikikin"`
	Requirement   string        `json:"requirement"`
	Madori        string        `json:"madori"`
	Rent          string        `json:"rent"`
	RentNormal    string        `json:"rent_normal"`
	RentNormalCSS string        `json:"rent_normal_css"`
	Commonfee     string        `json:"commonfee"`
	CommonfeeSp   interface{}   `json:"commonfee_sp"`
	Status        interface{}   `json:"status"`
	Type          string        `json:"type"`
	Floorspace    string        `json:"floorspace"`
	Floor         string        `json:"floor"`
	URLDetail     interface{}   `json:"urlDetail"`
	URLDetailSp   interface{}   `json:"urlDetail_sp"`
	Feature       interface{}   `json:"feature"`
}

type slackReq struct {
	Text string `json:"text"`
}

func postSlack(r urRes, slackURL string) error {
	v := slackReq{Text: fmt.Sprintf("物件: %s, 価格: %s", r.Name, r.Rent)}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	resp, err := http.Post(
		slackURL,
		"application/json",
		strings.NewReader(string(b)),
	)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		log.Println(fmt.Sprintf("posted to slack: %s", v))
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	log.Println(string(body), err)
	return fmt.Errorf("error sending to slack")
}
