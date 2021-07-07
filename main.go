package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/Sickle-Cell-Foundation-Of-Alberta/pintsize_data_scraper/googlesheet"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

// Implementation of Struct. It is used to group related data to form a single unit.
type Data struct {
	Country           string
	Donor_Centre      string
	Province_location string
	Address           string
	Date              string
}

func main() {
	fileReinitialize()

	fName := "./data/cbs_data.json"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	// Creates a new Collector instance with default configuration
	c := colly.NewCollector(
		colly.AllowedDomains("myaccount.blood.ca", "myaccount.blood.ca/en/donate"),

		// CacheDir specifies the location where GET requests are cached as files
		colly.CacheDir("./bloodServices_cache"),
	)

	dataList := make([]Data, 0)

	// Start of extracting of data
	c.OnHTML(`ul.cbs_wss_booking_clinic_select_locations`, func(e *colly.HTMLElement) {
		country := "Canada"
		data := Data{
			Country: country,
		}
		// Iterate over li components to construct the relevant info/data
		e.ForEach("li", func(_ int, el *colly.HTMLElement) {
			donorCentre := el.ChildText("div.title h3")
			if donorCentre == "" {
				log.Println("No province found", e.Request)
			} else {
				data.Donor_Centre = donorCentre
			}

			provinceLocation := el.ChildText("div.address2")
			if provinceLocation == "" {
				log.Println("No province found", e.Request)
			} else {
				data.Province_location = provinceLocation
			}

			address := el.ChildText("div.address1")
			if address == "" {
				log.Println("No province found", e.Request)
			} else {
				data.Address = address
			}

			el.ForEach("option", func(_ int, eh *colly.HTMLElement) {
				eh.ForEach("value", func(_ int, el *colly.HTMLElement) {
				})
				donorDates := eh.Attr("value")
				if donorDates == "" {
					log.Println("No province found", eh.Request)
				} else {
					data.Date = donorDates
				}
				dataList = append(dataList, data)
			})

		})
	})

	cityList := [4]string{"Edmonton", "Calgary", "Red%20Deer", "Lethbridge"}
	for _, element := range cityList {
		c.Visit("https://myaccount.blood.ca/en/donate/select-clinic?apt-slc=" + element)
	}

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	// Dump json to the standard output
	enc.Encode(dataList)

	jsonConversion()

}

func fileReinitialize() {
	err := os.RemoveAll("./data")
	err1 := os.RemoveAll("./bloodServices_cache")
	if err1 != nil {
		log.Fatal(err)
	}
	err = os.Mkdir("data", 0755)
	if err != nil {
		log.Fatal(err)
	}

}

func jsonConversion() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	secretKey := os.Getenv("spreadsheet_Id")

	jsonDataFromFile, err := ioutil.ReadFile("./data/cbs_data.json")

	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal JSON data
	var jsonData []Data
	err = json.Unmarshal([]byte(jsonDataFromFile), &jsonData)

	if err != nil {
		fmt.Println(err)
	}

	csvFile, err := os.Create("./data/cbs_data.csv")

	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)

	spreadsheetId := secretKey

	var googleServices googlesheet.GoogleSheet
	if err := googleServices.Init(spreadsheetId); err != nil {
		log.Fatalf("Unable to init google sheet api: %v\nCredential missing?", err)
	}

	readRange := "Sheet1!A:E"
	values, err := googleServices.Read(readRange)
	log.Println((len(values)))
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	} else if len(values) >= 0 {
		clearRange := "Sheet1!A:Z"
		err = googleServices.Clear(clearRange)
		if err != nil {
			log.Fatal(err)
		}
	}

	var writeValues [][]interface{}
	sheet_row := []interface{}{"id", "Country", "Donor Centre", "City", "Address", "Next Availability Date"}
	writeValues = append(writeValues, sheet_row)
	err = googleServices.Write("Sheet1", writeValues)
	if err != nil {
		log.Fatal(err)
	}

	for index, element := range jsonData {
		log.Println(index)
		var row []string
		row = append(row, element.Country)
		row = append(row, element.Donor_Centre)
		row = append(row, element.Province_location)
		row = append(row, element.Address)
		row = append(row, element.Date)

		var values []interface{}
		values = append(values, element.Country, element.Donor_Centre, element.Province_location, element.Address, element.Date)
		time.Sleep(4 * time.Second)

		var writeValues [][]interface{}
		sheet_row := []interface{}{index, element.Country, element.Donor_Centre, element.Province_location, element.Address, element.Date}
		writeValues = append(writeValues, sheet_row)
		err = googleServices.Write("Sheet1", writeValues)
		if err != nil {
			log.Fatal(err)
		}

		if err != nil {
			log.Fatal(err)
		}

		writer.Write(row)
	}

	writer.Flush()

}
