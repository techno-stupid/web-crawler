package models

type Product struct {
	ID                  string
	Name                string
	Category            string
	Price               string
	Sizes               []string
	SizeChart           SizeChart
	MetaData            MetaData
	DescriptionTitle    string
	DescriptionMainText string
	Coordinates         []CoordinatedProductInfo
	Tags                []string
}

type SizeChart struct {
	CategoryNames []string
	Measurements  [][]string
}

type CoordinatedProductInfo struct {
	Name           string
	Price          string
	ProductNumber  string
	ImageURL       string
	ProductPageURL string
}

type MetaData struct {
	AverageRating      string
	ReviewerCount      string
	RecommendetionRate string
	ItemRatings        []ItemRating
	UserReviews        []Review
}

type Review struct {
	Title       string
	Description string
	Date        string
	Rating      string
	ReviewerID  string
}

type ItemRating struct {
	Label  string
	Rating string
}
