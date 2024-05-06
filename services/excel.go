package services

import (
	"Crawler/models"
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"strings"
	"time"
)

func SaveProductsToExcel(products []models.Product) error {
	fileName := fmt.Sprintf("output/scraped_products_%s.xlsx", time.Now().Format("2006-01-02"))
	fileExcel, err := xlsx.OpenFile(fileName)
	if err != nil {
		fileExcel = xlsx.NewFile()
	}

	var sheet *xlsx.Sheet
	sheet, _ = fileExcel.AddSheet("Products for Men")

	if sheet.MaxRow == 0 {
		row := sheet.AddRow()
		headers := []string{"ID", "Product Title", "Category", "Price", "Sizes", "Description Title", "Description Text", "Coordinated Products", "SizeChart", "Average Rating", "Review Count", "Recommended Rate", "KWs", "Item Ratings", "User Reviews"}
		for _, header := range headers {
			cell := row.AddCell()
			cell.SetString(header)
		}
	}

	for _, product := range products {
		row := sheet.AddRow()
		coordinatesJSON, err := json.Marshal(product.Coordinates)
		if err != nil {
			fmt.Println("Coordinates decoding error:", err)
		}
		itemRatingsJSON, err := json.Marshal(product.MetaData.ItemRatings)
		if err != nil {
			fmt.Println("ItemRatings decoding error:", err)
		}
		userReviewsJSON, err := json.Marshal(product.MetaData.UserReviews)
		if err != nil {
			fmt.Println("UserReviews decoding error:", err)
		}

		row.AddCell().SetString(product.ID)
		row.AddCell().SetString(product.Name)
		row.AddCell().SetString(product.Category)
		row.AddCell().SetString(product.Price)
		row.AddCell().SetString(strings.Join(product.Sizes, ","))
		row.AddCell().SetString(product.DescriptionTitle)
		row.AddCell().SetString(product.DescriptionMainText)
		row.AddCell().SetString(string(coordinatesJSON))
		row.AddCell().SetString(fmt.Sprintf("%v", product.SizeChart))
		row.AddCell().SetString(product.MetaData.AverageRating)
		row.AddCell().SetString(product.MetaData.ReviewerCount)
		row.AddCell().SetString(product.MetaData.RecommendetionRate)
		row.AddCell().SetString(strings.Join(product.Tags, ","))
		row.AddCell().SetString(string(itemRatingsJSON))
		row.AddCell().SetString(string(userReviewsJSON))

	}
	err = fileExcel.Save(fileName)
	if err != nil {
		fmt.Println("Error saving Excel file:", err)
		return err
	}
	fmt.Println("Excel dump completed.")
	return nil
}
