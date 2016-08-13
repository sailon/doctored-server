package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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

type lineItem struct {
	YCoordinate int
	Code        int
	Description string
	Price       string
}

// BillSubmission is the user-submitted payload that contains a URL
// to the uploaded image of the bill.
type BillSubmission struct {
	URL string `json:"url"`
}

// PostBill accepts an image URL and calls the MSFT OCR API. Route returns successful
// with a parsed invoice object.
func PostBill(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (*httpResponse, *httpError) {

	// Grab the image URL from the request payload
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

	// lineItems are stored with their Y coordinate as the map key
	var lineItems []lineItem

	codeRegex := regexp.MustCompile(`^\d{3}$`)
	priceRegex := regexp.MustCompile(`\d+,?\d+\.\d{2}$`)

	// Loop through each region, line, and word looking for codes
	for _, region := range ocrResponse.Regions {
		for _, line := range region.Lines {
			for _, word := range line.Words {

				// Found a code, create a new line item
				if codeRegex.MatchString(word.Text) {
					code, _ := strconv.Atoi(word.Text)
					coordinates := strings.Split(word.BoundingBox, ",")
					yCoordinate, _ := strconv.Atoi(coordinates[1])

					lineItems = append(lineItems, lineItem{
						YCoordinate: yCoordinate,
						Code:        code,
					})

					// fmt.Println("Code: ", word.Text, " Path: ", region.BoundingBox, " ", line.BoundingBox, " ", word.BoundingBox)
				}

			}
		}
	}

	// Loop through the regions and lines looking for fuzzy matches
	// on Y-Coordinates in existing line items
	for _, region := range ocrResponse.Regions {
		for _, line := range region.Lines {

			// Grab the Y-Coordinate to look for a match
			coordinates := strings.Split(line.BoundingBox, ",")
			yCoordinate, _ := strconv.Atoi(coordinates[1])

			// Loop through existing line items found via the code regex
			for k, lineItem := range lineItems {

				diff := lineItem.YCoordinate - yCoordinate
				if diff < 0 {
					diff = -diff
				}

				// If the Y-Coordinate is within a given threshold, the we have a hit!
				if diff < 10 {

					// Classify each word in the line and add it to the line item
					for _, word := range line.Words {

						if priceRegex.MatchString(word.Text) {

							// price, _ := strconv.ParseFloat(word.Text, 10)
							lineItems[k].Price = word.Text

						} else if !codeRegex.MatchString(word.Text) {
							lineItems[k].Description += word.Text
						}
					}
				}
			}
		}
	}

	// fmt.Printf("\n\n%+v\n\n", lineItems)

	return &httpResponse{ContentType: "application/json", Payload: lineItems}, nil
}
