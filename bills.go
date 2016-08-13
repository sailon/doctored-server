package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// OCRResponse is the JSON payload returned
// by the Microsoft Cognitive Services OCR API
type OCRResponse struct {
	Language    string  `json:"language"`
	TextAngle   float64 `json:"textAngle"`
	Orientation string  `json:"orientation"`
	Regions     []struct {
		BoundingBox string `json:"boundingBox"`
		Lines       []struct {
			BoundingBox string `json:"boundingBox"`
			Words       []struct {
				BoundingBox string `json:"boundingBox"`
				Text        string `json:"text"`
			} `json:"words"`
		} `json:"lines"`
	} `json:"regions"`
}

// BillSubmission is the user-submitted payload that contains a URL
// to the uploaded image of the bill.
type BillSubmission struct {
	URL string `json:"url"`
}

// PostBill accepts an S3 URL and calls the MSFT OCR API. Route returns successful
// with a parsed invoice object.
func PostBill(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (*httpResponse, *httpError) {

	// Grab the S3 URL from the request payload
	var bill BillSubmission
	err := decodeRequestPayload(r, &bill)
	if err != nil {
		if err.Error() == parseError {
			return nil, &httpError{Message: parseError, StatusCode: http.StatusBadRequest}
		}

		log.Printf("There was an issue decoding the post bill request: %s", err.Error())
		return nil, &httpError{Message: http.StatusText(http.StatusInternalServerError), StatusCode: http.StatusInternalServerError}
	}

	if bill.URL == "" {
		return nil, &httpError{Message: "Missing URL", StatusCode: http.StatusBadRequest}
	}

	// Create a new request so we can add custom headers
	jsonStr := []byte("{\"url\":\"" + bill.URL + "\"}")
	req, err := http.NewRequest("POST", "https://api.projectoxford.ai/vision/v1.0/ocr?language=unk&detectOrientation=true", bytes.NewBuffer(jsonStr))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Ocp-Apim-Subscription-Key", apiKey)

	// Create http client and send the actual request
	client := &http.Client{}
	msftResp, err := client.Do(req)

	// Parse the analyzed image
	var ocrResponse OCRResponse
	err = json.NewDecoder(msftResp.Body).Decode(&ocrResponse)
	if err != nil {
		return nil, &httpError{Message: http.StatusText(http.StatusInternalServerError), StatusCode: http.StatusInternalServerError}
	}

	return &httpResponse{ContentType: "application/json", Payload: ocrResponse}, nil
}
