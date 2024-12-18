package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"ncc/go-mezon-bot/pkg/responses"
	"net/http"
	"strings"
	"time"
)

func CheckinApi(imageBase64 string) (*responses.CheckinRes, error) {

	url := "https://checkin.nccsoft.vn/v1/employees/facial-recognition/ims-verify"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
    "currentDateTime": "%s",
    "employeeFacialSetupDTO": {
        "timeVerify": "",
        "secondsTime": "",
        "imgs": [
            "%s"
        ]
    }
}`, time.Now().Format("2006-01-02T15:04:05"), imageBase64))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json, text/plain, */*")
	req.Header.Add("Accept-Language", "vi,en-US;q=0.9,en;q=0.8")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Account-Id", "dd0f2097-ad1a-4575-be15-a8bba7b559f2")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data *responses.CheckinRes
	err = json.Unmarshal(body, &data)
	return data, err
}
