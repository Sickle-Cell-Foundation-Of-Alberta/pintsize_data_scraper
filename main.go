package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gocolly/colly"
)

// #Integrate Quickstart.go into main.go
// Implementation of Struct. It is used to group related data to form a single unit.
type Data struct {
	Country           string
	Donor_Centre      string
	Province_location string
	Address           string
	Date              string
}

func main() {
	fName := "cbs_data.json"
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

	dataList := make([]Data, 0, 200)

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
				log.Println("No province found", e.Request.URL)
			} else {
				data.Donor_Centre = donorCentre
			}

			provinceLocation := el.ChildText("div.address2")
			if provinceLocation == "" {
				log.Println("No province found", e.Request.URL)
			} else {
				data.Province_location = provinceLocation
			}

			address := el.ChildText("div.address1")
			if address == "" {
				log.Println("No province found", e.Request.URL)
			} else {
				data.Address = address
			}

			el.ForEach("option", func(_ int, eh *colly.HTMLElement) {
				eh.ForEach("value", func(_ int, ek *colly.HTMLElement) {
				})
				donorDates := eh.Attr("value")
				if donorDates == "" {
					log.Println("No province found", eh.Request.URL)
				} else {
					data.Date = donorDates
				}
				dataList = append(dataList, data)
			})

		})
	})

	cityList := [4]string{"Edmonton", "Calgary", "Red%20Deer", "Lethbridge"}
	for _, element := range cityList {
		// log.Println("https://myaccount.blood.ca/en/donate/select-clinic?apt-slc=" + element)
		c.Visit("https://myaccount.blood.ca/en/donate/select-clinic?apt-slc=" + element)
	}

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	// Dump json to the standard output
	enc.Encode(dataList)

	jsonConversion()
}

func jsonConversion() {
	jsonDataFromFile, err := ioutil.ReadFile("./cbs_data.json")

	if err != nil {
		fmt.Println(err)
	}

	// Unmarshal JSON data
	var jsonData []Data
	err = json.Unmarshal([]byte(jsonDataFromFile), &jsonData)

	if err != nil {
		fmt.Println(err)
	}

	csvFile, err := os.Create("./cbs_data.csv")

	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)

	for _, element := range jsonData {
		var row []string
		row = append(row, element.Country)
		row = append(row, element.Donor_Centre)
		row = append(row, element.Province_location)
		row = append(row, element.Address)
		row = append(row, element.Date)
		// log.Println(row)
		writer.Write(row)
	}
	writer.Flush()

}
