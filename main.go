package main

import (
	"WalletsCrawler/config"
	. "WalletsCrawler/models"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	//filterAllJsonFilesInFolder("./", "filtered_data.xlsx")
	walletAddress := "3eJPAtEjw5cMzhNLsiXx2Y3yZoi1sQ5eXs4Ntiru6FxC"
	gmgnUrl := "https://gmgn.ai/_next/data/lvk7vUyrbtyJdU3QUPhjP/sol/address/%s.json?chain=sol&address=%s"
	http.Get(fmt.Sprintf(gmgnUrl, walletAddress, walletAddress))
}

func filterAllJsonFilesInFolder(folderPath, outputFileName string) {
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	f := excelize.NewFile()
	sheetName := "FilteredData"
	f.NewSheet(sheetName)
	f.SetCellValue(sheetName, "A1", "Wallet")
	f.SetCellValue(sheetName, "B1", "Winrate")
	f.SetCellValue(sheetName, "C1", "Bought USD")
	f.SetCellValue(sheetName, "D1", "Sold USD")
	f.SetCellValue(sheetName, "E1", "Bought Count")
	f.SetCellValue(sheetName, "F1", "Sold Count")
	f.SetCellValue(sheetName, "G1", "PL USD")
	f.SetCellValue(sheetName, "H1", "Kind")

	row := 2
	seenWallets := make(map[string]bool)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(folderPath, file.Name())
			filteredTraders := filterTradersFromJsonFile(filePath)
			for _, trader := range filteredTraders {
				if !seenWallets[trader.Signer] {
					seenWallets[trader.Signer] = true

					f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), trader.Signer)
					f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("%.2f", trader.WinRate))
					f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), convertToCurrency(trader.BoughtUsd))
					f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), convertToCurrency(trader.SoldUsd))
					f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), trader.BoughtCount)
					f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), trader.SoldCount)
					f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), convertToCurrency(trader.PlUsd))
					f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), trader.Kind)
					row++
				}
			}
		}
	}

	if err := f.SaveAs(outputFileName); err != nil {
		log.Fatalf("Failed to save Excel file: %v", err)
	}
}

func convertToCurrency(amount string) string {
	converted, _ := strconv.ParseFloat(amount, 64)
	return fmt.Sprintf("$ %.2f", converted)
}

func filterTradersFromJsonFile(filePath string) []FilteredTrader {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)

	var traders []Trader
	err = json.Unmarshal(byteValue, &traders)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON %s: %v", filePath, err)
	}

	var filteredTraders []FilteredTrader
	for _, item := range traders {
		if item.Attributes.BoughtCount <= 5 && item.Attributes.BoughtToken != "0.0" {
			plUsd, _ := strconv.ParseFloat(item.Attributes.PlUsd, 64)
			boughtUsd, _ := strconv.ParseFloat(item.Attributes.BoughtUsd, 64)
			winRate := plUsd / boughtUsd * 100
			if winRate >= 250 {
				filteredTraders = append(filteredTraders, FilteredTrader{
					Signer:      item.Attributes.Signer,
					WinRate:     winRate,
					BoughtUsd:   item.Attributes.BoughtUsd,
					SoldUsd:     item.Attributes.SoldUsd,
					BoughtCount: item.Attributes.BoughtCount,
					SoldCount:   item.Attributes.SoldCount,
					PlUsd:       item.Attributes.PlUsd,
					Kind:        item.Attributes.Kind,
				})
			}
		}
	}
	return filteredTraders
}

func prepareJsonFile() {
	// Open the file
	file, err := os.Open("data/Coins.txt")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close() // Make sure to close the file when done

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	var allTraders []Trader

	// Iterate through each line
	indexCoin := 0
	for scanner.Scan() {
		line := scanner.Text() // Get the line as a string
		for i := 1; i <= 10; i++ {
			requestUrl := fmt.Sprintf(line, i)
			log.Printf("URL: %s", requestUrl)
			req, err := http.NewRequest("GET", requestUrl, nil)
			if err != nil {
				log.Printf("Error creating request for %s: %v", requestUrl, err)
				continue
			}

			req.Header.Set("User-Agent", config.USER_AGENT)
			req.Header.Set("Accept", config.ACCEPT)
			req.Header.Set("Cookie", config.PHOTON_COOKIE)

			client := &http.Client{}
			response, err := client.Do(req)
			if err != nil {
				log.Printf("Error making request to %s: %v", requestUrl, err)
				continue
			}
			defer response.Body.Close()

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Printf("Error reading response body: %v", err)
				continue
			}

			var jsonResponse TopTraderResponse
			err = json.Unmarshal(body, &jsonResponse)
			if err != nil {
				log.Printf("Error unmarshalling JSON: %v", err)
				log.Printf("Response body: %s", body)
				continue
			}

			allTraders = append(allTraders, jsonResponse.Data...)
		}
		saveResponsesToJsonFile(allTraders, fmt.Sprintf("%d.json", indexCoin))
		indexCoin++
	}

	// Check for any errors that may have occurred during scanning
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
}

func saveResponsesToJsonFile(traders []Trader, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(traders, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling JSON data: %v", err)
	}

	_, err = file.Write(jsonData)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}
}

func getFloatFromInterface(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}
