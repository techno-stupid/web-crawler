package main

import (
	"Crawler/services"
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"strconv"
)

const (
	baseURL          = "https://shop.adidas.jp"
	chromeDriverPath = "./driver/chromedriver.exe"
	port             = 4444
)

func main() {

	service, err := selenium.NewChromeDriverService(chromeDriverPath, port)
	if err != nil {
		log.Fatalf("something went wrong while starting the ChromeDriver server: %v", err)
	}
	defer service.Stop()

	wd := createWebDriver(port)
	defer wd.Quit()

	if err := wd.Get(baseURL); err != nil {
		log.Fatalf("Failed to load page: %v", err)
	}

	ids := services.RetrieveProductIDs(wd)
	fmt.Println(ids)

	allProducts := services.FetchAllProducts(wd, ids)

	if err := services.SaveProductsToExcel(allProducts); err != nil {
		log.Printf("Something went wrong while saving Xlxs %v", err)
	}
}

func createWebDriver(port int) selenium.WebDriver {
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}
	caps.AddChrome(
		chrome.Capabilities{
			Path: "",
			Args: []string{
				"--headless",
				"--disable-dev-shm-usage",
				"--disable-web-security",
				"--no-sandbox",
				"--log-level=3",
				"--disable-gpu",
				"--window-size=1920,1080",
			},
		},
	)
	wd, err := selenium.NewRemote(caps, "http://127.0.0.1:"+strconv.Itoa(port)+"/wd/hub")
	if err != nil {
		log.Fatalf("Failed to create WebDriver: %v", err)
	}
	fmt.Println("Selenium WebDriver created")
	return wd
}
