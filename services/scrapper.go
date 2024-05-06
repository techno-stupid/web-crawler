package services

import (
	"Crawler/models"
	"fmt"
	"github.com/tebeka/selenium"
	"log"
	"strings"
	"time"
)

const (
	baseURL      = "https://shop.adidas.jp"
	productCount = 2
	totalPages   = 1
)

func RetrieveProductIDs(driver selenium.WebDriver) []string {
	var ids []string

	for page := 1; page <= totalPages; page++ {
		url := fmt.Sprintf("%v/item/?order=11&gender=mens&limit=100&category=wear&page=%v", baseURL, page)
		fmt.Printf("Fetching Product Page: %v \n", url)
		if err := driver.Get(url); err != nil {
			log.Printf("Failed to load page %s: %v", page, err)
		}

		listItems, err := driver.FindElements(selenium.ByCSSSelector, ".itemCardArea-cards")
		if err != nil {
			log.Printf("Failed to find list items: %v", err)
		}

		for index, item := range listItems {
			scrollIndex := index
			if index > 0 && index < 10 {
				scrollIndex = index * 10
				err := scroller(driver, fmt.Sprintf(".itemCardArea-cards:nth-child(%v)", scrollIndex))
				if err != nil {
					log.Printf("Target Dom not availabe yet: %v", fmt.Sprintf(".itemCardArea-cards:nth-child(%v)", scrollIndex))
				}
			}
			elem, err := item.FindElement(selenium.ByCSSSelector, ".image_link")
			if err != nil {
				log.Printf("[%v] Failed to find element for attribute: %v", index, err)
				continue
			}
			id, err := elem.GetAttribute("data-ga-eec-product-id")
			if err != nil {
				log.Printf("[%v] Failed to get element attribute: %v", index, err)
				continue
			}
			ids = append(ids, id)
		}
	}

	fmt.Println(fmt.Sprintf("Total %v Product Found", len(ids)))
	return ids
}
func scroller(driver selenium.WebDriver, selector string) error {
	el, err := driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return fmt.Errorf("dom scrolling error: %v", err)
	}
	if _, err := driver.ExecuteScript("arguments[0].scrollIntoView(true);", []interface{}{el}); err != nil {
		log.Printf("Failed to scroll element into view: %v", err)
	}
	time.Sleep(1)
	return nil
}

func FetchAllProducts(driver selenium.WebDriver, ids []string) []models.Product {
	var products []models.Product
	for key, productID := range ids {
		if key >= productCount {
			break
		}
		product := fetchProductInfo(driver, productID)
		products = append(products, product)

		log.Printf("Product info for ID %s Fetched successfully", productID)
	}
	return products
}

func fetchProductInfo(driver selenium.WebDriver, productID string) models.Product {
	var product models.Product
	url := baseURL + "/products/" + productID + "/"

	log.Printf("Fetching %s:", productID)
	if err := driver.Get(url); err != nil {
		log.Printf("Failed to load page for product ID %s: %v", productID, err)
	}
	err := scroller(driver, ".coordinateItems .carouselListitem")
	if err != nil {
		_ = scroller(driver, ".js-articlePromotion")
	}

	product = getProductInfo(driver)
	product.ID = productID
	product.SizeChart = parseSizeChartHTML(driver)
	product.Coordinates = getCoordinatedProductInfo(driver)
	product.MetaData = ExtractProductMetaData(driver)
	product.Tags = ExtractProductTags(driver)
	return product
}

func getProductInfo(driver selenium.WebDriver) models.Product {
	var productInfo models.Product

	productInfo.Category = ExtractElementText(driver, ".categoryName")
	productInfo.Name = ExtractElementText(driver, ".itemTitle")
	productInfo.Price = ExtractElementText(driver, ".price-value")

	sizesElements, err := driver.FindElements(selenium.ByCSSSelector, ".sizeSelectorListItemButton")
	if err != nil {
		log.Printf("Failed to find size elements: %v", err)
	} else {
		for _, sizeElement := range sizesElements {
			size, err := sizeElement.Text()
			if err != nil {
				log.Printf("Failed to get size text: %v", err)
				continue
			}
			productInfo.Sizes = append(productInfo.Sizes, size)
		}
	}

	// Fetching sub-heading
	subheadingElem, err := driver.FindElement(selenium.ByCSSSelector, ".itemFeature")
	if err != nil {
		log.Printf("DescriptionTitle not available")
	} else {
		productInfo.DescriptionTitle, _ = subheadingElem.Text()
	}

	// Fetching main text
	mainTextElem, err := driver.FindElement(selenium.ByCSSSelector, ".commentItem-mainText")
	if err != nil {
		log.Printf("DescriptionMainText not available")
	} else {
		productInfo.DescriptionMainText, _ = mainTextElem.Text()
	}

	return productInfo
}

func getCoordinatedProductInfo(driver selenium.WebDriver) []models.CoordinatedProductInfo {
	var coordinatedProducts []models.CoordinatedProductInfo

	// Find all carousel list items
	carouselListItems, err := driver.FindElements(selenium.ByCSSSelector, ".coordinateItems .carouselListitem")
	if err != nil {
		log.Printf("Failed to find carousel list items: %v", err)
		return coordinatedProducts
	}

	for _, item := range carouselListItems {
		if err := item.Click(); err != nil {
			continue
		}

		time.Sleep(1)
		coordinatedProduct := models.CoordinatedProductInfo{
			Name:           ExtractElementText(driver, ".coordinate_item_container .title"),
			Price:          ExtractElementText(driver, ".coordinate_item_container .price-value"),
			ProductNumber:  ExtractAttributeValue(driver, ".coordinate_item_tile", "data-articleid"),
			ImageURL:       baseURL + ExtractAttributeValue(driver, ".coordinate_image_body.test-img", "models"),
			ProductPageURL: baseURL + ExtractAttributeValue(driver, ".coordinate_item_container .test-link_a", "href"),
		}
		coordinatedProducts = append(coordinatedProducts, coordinatedProduct)
	}

	return coordinatedProducts
}

func parseSizeChartHTML(driver selenium.WebDriver) models.SizeChart {
	var sizeChart models.SizeChart

	// Find all table rows in the size chart
	rows, err := driver.FindElements(selenium.ByCSSSelector, ".sizeChartTRow")
	if err != nil {
		log.Printf("Failed to find size chart rows: %v", err)
		return sizeChart
	}

	// Extract column headers
	columnHeaders, err := driver.FindElements(selenium.ByCSSSelector, ".sizeChartTHeaderCell")
	if err != nil {
		log.Printf("Failed to find size chart column headers: %v", err)
		return sizeChart
	}

	var headerRow []string
	for _, header := range columnHeaders {
		text, err := header.Text()
		if err != nil {
			log.Printf("Failed to get column header text: %v", err)
			continue
		}
		headerRow = append(headerRow, text)
	}
	sizeChart.Measurements = append(sizeChart.Measurements, headerRow)

	// Iterate over each row
	for _, row := range rows {
		// Find all table cells in the row
		cells, err := row.FindElements(selenium.ByCSSSelector, ".sizeChartTCell")
		if err != nil {
			log.Printf("Failed to find size chart cells: %v", err)
			continue
		}

		var measurements []string

		// Iterate over each cell
		for _, cell := range cells {
			// Get the text content of the cell
			text, err := cell.Text()
			if err != nil {
				log.Printf("Failed to get size chart cell text: %v", err)
				continue
			}

			// Append the text content to the measurements slice
			measurements = append(measurements, text)
		}

		// Check if the row contains measurements
		if len(measurements) > 0 {
			sizeChart.Measurements = append(sizeChart.Measurements, measurements)
		}
	}

	return sizeChart
}

func ExtractProductMetaData(driver selenium.WebDriver) models.MetaData {
	var productMetaData models.MetaData

	// Extract overall rating
	overallRatingElem, err := driver.FindElement(selenium.ByCSSSelector, ".OverallRatingElement")
	if err != nil {
		fmt.Printf("Product metadata not available")
		return productMetaData
	}
	overallRatingText, _ := overallRatingElem.Text()
	productMetaData.AverageRating = overallRatingText

	// Extract number of reviews
	numReviewsElem, _ := driver.FindElement(selenium.ByCSSSelector, ".NumReviewsElement")
	numReviewsText, _ := numReviewsElem.Text()
	productMetaData.ReviewerCount = numReviewsText

	// Extract recommended rate
	recommendedRateElem, _ := driver.FindElement(selenium.ByCSSSelector, ".RecommendedRateElement")
	recommendedRateText, _ := recommendedRateElem.Text()
	productMetaData.RecommendetionRate = recommendedRateText

	// Extract item ratings
	var itemRatings []models.ItemRating
	itemRatingElems, _ := driver.FindElements(selenium.ByCSSSelector, ".ItemRatingElement")
	for _, itemRatingElem := range itemRatingElems {
		labelElem, _ := itemRatingElem.FindElement(selenium.ByCSSSelector, ".LabelElement")
		label, _ := labelElem.Text()

		ratingElem, _ := itemRatingElem.FindElement(selenium.ByCSSSelector, ".RatingElement")
		rating, _ := ratingElem.Text()

		itemRatings = append(itemRatings, models.ItemRating{Label: label, Rating: rating})
	}
	productMetaData.ItemRatings = itemRatings

	// Extract user reviews
	var userReviews []models.Review
	reviewElems, _ := driver.FindElements(selenium.ByCSSSelector, ".ReviewElement")
	for _, reviewElem := range reviewElems {
		review := models.Review{}

		dateElem, _ := reviewElem.FindElement(selenium.ByCSSSelector, ".DateElement")
		date, _ := dateElem.Text()
		review.Date = date

		titleElem, _ := reviewElem.FindElement(selenium.ByCSSSelector, ".TitleElement")
		title, _ := titleElem.Text()
		review.Title = title

		descElem, _ := reviewElem.FindElement(selenium.ByCSSSelector, ".DescriptionElement")
		desc, _ := descElem.Text()
		review.Description = desc

		// Extract review rating
		ratingElem, _ := reviewElem.FindElement(selenium.ByCSSSelector, ".RatingElement")
		ratingText, _ := ratingElem.Text()
		review.Rating = ratingText

		AttributeId, _ := reviewElem.GetAttribute("id")
		reviewerID := parseReviewerID(AttributeId)
		review.ReviewerID = reviewerID

		userReviews = append(userReviews, review)
	}
	productMetaData.UserReviews = userReviews

	return productMetaData
}

func ExtractProductTags(driver selenium.WebDriver) []string {
	var tags []string
	itemTagElems, _ := driver.FindElements(selenium.ByCSSSelector, ".productTags .categoryLink .inner a")
	for _, itemTagElem := range itemTagElems {
		tag, _ := itemTagElem.Text()
		tags = append(tags, tag)
	}
	return tags
}

func ExtractAttributeValue(driver selenium.WebDriver, selector, attribute string) string {
	elem, err := driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		log.Printf("Failed to find element for attribute: %v", err)
		return ""
	}
	attr, err := elem.GetAttribute(attribute)
	if err != nil {
		log.Printf("Failed to get element attribute: %v", err)
		return ""
	}
	return attr
}

func ExtractElementText(driver selenium.WebDriver, selector string) string {
	elem, err := driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		log.Printf("Failed to find element for text: %v", err)
		return ""
	}
	text, err := elem.Text()
	if err != nil {
		log.Printf("Failed to get element text: %v", err)
		return ""
	}
	return text
}
func parseReviewerID(id string) string {
	fragments := strings.Split(id, "_")
	return fragments[len(fragments)-1]
}
