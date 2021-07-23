package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sickle-Cell-Foundation-Of-Alberta/pintsize_data_scraper/googlesheet"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

// Implementation of Struct. It is used to group related data to form a single unit.
type Data struct {
	uniqueID          string
	Blood_Branch      string
	Country           string
	Donor_Centre      string
	Province_location string
	Address           string
	Date              string
}

func main() {
	// Checks and ensures that cahced filed along with any previous data has been erased and intialized.
	fileReinitialize()

	// Consutrct of the data files
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

	// List that will  store the respectful data found.
	dataList := make([]Data, 0)

	// Start of extracting of data
	c.OnHTML(`ul.cbs_wss_booking_clinic_select_locations`, func(e *colly.HTMLElement) {
		country := "Canada"
		data := Data{
			Country: country,
		}

		// Iterate over every li components to construct the relevant info/data
		// Correspond each data value with it's correct div
		e.ForEach("li", func(_ int, el *colly.HTMLElement) {
			// data.uniqueID = strconv.Itoa(i)
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

			bloodCity := strings.Split(provinceLocation, ", ")
			data.Blood_Branch = bloodCity[0]

			address := el.ChildText("div.address1")
			if address == "" {
				log.Println("No province found", e.Request)
			} else {
				data.Address = address
			}

			// Iterates for each option compoenent it's values to return it's corresponding data
			// Correspond each data value to it's respectful variables
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

	// List of strings that contains different cities to query over and scrap the relevant data needed.
	cityList := [4]string{"Edmonton", "Calgary", "Red%20Deer", "Lethbridge"}
	for _, element := range cityList {
		c.Visit("https://myaccount.blood.ca/en/donate/select-clinic?apt-slc=" + element)
	}

	// Json Enconder
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	// Dump json to the standard output
	enc.Encode(dataList)

	// Json conversion to CSV
	jsonConversion()

}

// Checks and ensures that cahced filed along with any previous data has been erased and intialized.
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

// Json to CSV conversion function
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

	// Google Services to allow us to initialize to the google shet and load the relevant data obtained
	var googleServices googlesheet.GoogleSheet
	if err := googleServices.Init(spreadsheetId); err != nil {
		log.Fatalf("Unable to init google sheet api: %v\nCredential missing?", err)
	}

	// Range within the google sheet to allow us to check it's values and reintialize it by clearing it's data
	readRange := "Cbs_Donoation_Data!A:G"
	values, err := googleServices.Read(readRange)
	// log.Println((len(values)))
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	} else if len(values) >= 0 {
		clearRange := "Cbs_Donoation_Data!A:Z"
		err = googleServices.Clear(clearRange)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Initialization of the format the google sheet should be
	var writeValues [][]interface{}
	sheet_row := []interface{}{"iD", "City", "Country", "Donor Centre", "Location", "Address", "Next Availability Date"}
	writeValues = append(writeValues, sheet_row)
	err = googleServices.Write("Cbs_Donoation_Data", writeValues)
	if err != nil {
		log.Fatal(err)
	}

	// Iterating over each json data to obtain the data and append it to the Row List of string
	for index, element := range jsonData {
		initial := 0
		if index < 10 {
			element.uniqueID = strconv.Itoa(initial) + strconv.Itoa(initial) + strconv.Itoa(index)
		} else if index >= 10 && index < 100 {
			element.uniqueID = strconv.Itoa(initial) + strconv.Itoa(index)
		} else {
			element.uniqueID = strconv.Itoa(index)
		}
		var row []string
		row = append(row, element.uniqueID)
		row = append(row, element.Blood_Branch)
		row = append(row, element.Country)
		row = append(row, element.Donor_Centre)
		row = append(row, element.Province_location)
		row = append(row, element.Address)

		const layout = "2006-01-02"

		t, err := time.Parse(layout, element.Date)
		if err != nil {
			log.Fatal(err)
		}
		element.Date = t.Weekday().String() + " " + element.Date
		row = append(row, element.Date)
		time.Sleep(4 * time.Second)

		// var values []interface{}
		// values = append(values, element.Country, element.Donor_Centre, element.Province_location, element.Address, element.Date)

		// Writing of the data into the google sheet.
		var writeValues [][]interface{}
		sheet_row := []interface{}{element.uniqueID, element.Blood_Branch, element.Country, element.Donor_Centre, element.Province_location, element.Address, element.Date}
		writeValues = append(writeValues, sheet_row)
		err = googleServices.Write("Cbs_Donoation_Data", writeValues)
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
