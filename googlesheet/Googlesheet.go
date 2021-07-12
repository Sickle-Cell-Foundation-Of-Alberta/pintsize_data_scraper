package googlesheet

import (
	"context"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// grouped list of variables
type GoogleSheet struct {
	service       *sheets.Service
	spreadsheetId string
}

const (
	client_secret_path = "pintsize-client_secret.json"
)

// initialization of google sheets
func (googleServices *GoogleSheet) Init(spreadsheetId string) (err error) {
	googleServices.spreadsheetId = spreadsheetId
	googleServices.service, err = sheets.NewService(context.Background(), option.WithCredentialsFile(client_secret_path), option.WithScopes(sheets.SpreadsheetsScope))

	return err
}

// Read function for google sheets
func (googleServices *GoogleSheet) Read(readRange string) ([][]interface{}, error) {
	readValues, err := googleServices.service.Spreadsheets.Values.Get(googleServices.spreadsheetId, readRange).Do()

	if err != nil {
		return nil, err
	}
	return readValues.Values, err
}

//Write function for google sheets
func (googleServices *GoogleSheet) Write(writeRange string, values [][]interface{}) error {
	valueInputOption := "RAW"
	rb := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         values,
	}
	_, err := googleServices.service.Spreadsheets.Values.Append(googleServices.spreadsheetId, writeRange, rb).ValueInputOption(valueInputOption).Do()

	return err
}

//Update function for google sheets
func (googleServices *GoogleSheet) Update(updateRange string, updateValues [][]interface{}) error {
	valueInputOption := "RAW"
	rb := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         updateValues,
	}
	_, err := googleServices.service.Spreadsheets.Values.Update(googleServices.spreadsheetId, updateRange, rb).ValueInputOption(valueInputOption).Do()

	return err
}

///Clear function for google sheets
func (googleServices *GoogleSheet) Clear(clearRange string) error {
	rb := &sheets.ClearValuesRequest{}
	_, err := googleServices.service.Spreadsheets.Values.Clear(googleServices.spreadsheetId, clearRange, rb).Do()

	return err
}
