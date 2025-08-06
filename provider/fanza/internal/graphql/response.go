package graphql

import (
	"time"
)

type PPVContent struct {
	ID                  string        `json:"id"`
	Floor               string        `json:"floor"`
	Title               string        `json:"title"`
	IsExclusiveDelivery bool          `json:"isExclusiveDelivery"`
	ReleaseStatus       string        `json:"releaseStatus"`
	Description         string        `json:"description"`
	Notices             []interface{} `json:"notices"`
	IsNoIndex           bool          `json:"isNoIndex"`
	IsAllowForeign      bool          `json:"isAllowForeign"`
	Announcements       []interface{} `json:"announcements"`
	FeatureArticles     []struct {
		Link struct {
			URL      string `json:"url"`
			Text     string `json:"text"`
			Typename string `json:"__typename"`
		} `json:"link"`
		Typename string `json:"__typename"`
	} `json:"featureArticles"`
	PackageImage struct {
		LargeURL  string `json:"largeUrl"`
		MediumURL string `json:"mediumUrl"`
		Typename  string `json:"__typename"`
	} `json:"packageImage"`
	SampleImages []struct {
		Number        int    `json:"number"`
		ImageURL      string `json:"imageUrl"`
		LargeImageURL string `json:"largeImageUrl"`
		Typename      string `json:"__typename"`
	} `json:"sampleImages"`
	Products []struct {
		ID           string `json:"id"`
		Priority     int    `json:"priority"`
		DeliveryUnit struct {
			ID                      string `json:"id"`
			Priority                int    `json:"priority"`
			StreamMaxQualityGroup   string `json:"streamMaxQualityGroup"`
			DownloadMaxQualityGroup string `json:"downloadMaxQualityGroup"`
			Typename                string `json:"__typename"`
		} `json:"deliveryUnit"`
		PriceInclusiveTax int         `json:"priceInclusiveTax"`
		Sale              interface{} `json:"sale"`
		ExpireDays        interface{} `json:"expireDays"`
		LicenseType       string      `json:"licenseType"`
		ShopName          string      `json:"shopName"`
		AvailableCoupon   struct {
			Name             string `json:"name"`
			ExpirationPolicy struct {
				ExpirationDays int    `json:"expirationDays"`
				Typename       string `json:"__typename"`
			} `json:"expirationPolicy"`
			ExpirationAt    interface{} `json:"expirationAt"`
			DiscountedPrice int         `json:"discountedPrice"`
			MinPayment      int         `json:"minPayment"`
			DestinationURL  string      `json:"destinationUrl"`
			Typename        string      `json:"__typename"`
		} `json:"availableCoupon"`
		Typename string `json:"__typename"`
	} `json:"products"`
	MostPopularContentImage struct {
		Typename      string `json:"__typename"`
		LargeImageURL string `json:"largeImageUrl"`
		ImageURL      string `json:"imageUrl"`
	} `json:"mostPopularContentImage"`
	PriceSummary struct {
		LowestSalePrice int         `json:"lowestSalePrice"`
		LowestPrice     int         `json:"lowestPrice"`
		Campaign        interface{} `json:"campaign"`
		Typename        string      `json:"__typename"`
	} `json:"priceSummary"`
	WeeklyRanking  interface{} `json:"weeklyRanking"`
	MonthlyRanking interface{} `json:"monthlyRanking"`
	WishlistCount  int         `json:"wishlistCount"`
	Sample2DMovie  struct {
		FileID   string `json:"fileID"`
		Typename string `json:"__typename"`
	} `json:"sample2DMovie"`
	SampleMovie struct {
		Has2D    bool   `json:"has2d"`
		HasVr    bool   `json:"hasVr"`
		Typename string `json:"__typename"`
	} `json:"sampleMovie"`
	DeliveryStartDate time.Time `json:"deliveryStartDate"`
	MakerReleasedAt   time.Time `json:"makerReleasedAt"`
	Duration          int       `json:"duration"`
	Actresses         []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		NameRuby string `json:"nameRuby"`
		ImageURL string `json:"imageUrl"`
		Typename string `json:"__typename"`
	} `json:"actresses"`
	AmateurActress struct {
		ID              string        `json:"id"`
		Name            string        `json:"name"`
		ImageURL        string        `json:"imageUrl"`
		Age             interface{}   `json:"age"`
		Waist           interface{}   `json:"waist"`
		Bust            interface{}   `json:"bust"`
		BustCup         interface{}   `json:"bustCup"`
		Height          interface{}   `json:"height"`
		Hip             interface{}   `json:"hip"`
		RelatedContents []interface{} `json:"relatedContents"`
		Typename        string        `json:"__typename"`
	} `json:"amateurActress"`
	Histrions []interface{} `json:"histrions"`
	Directors []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Typename string `json:"__typename"`
	} `json:"directors"`
	Series struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Typename string `json:"__typename"`
	} `json:"series"`
	Maker struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Typename string `json:"__typename"`
	} `json:"maker"`
	Label struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Typename string `json:"__typename"`
	} `json:"label"`
	Genres []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Typename string `json:"__typename"`
	} `json:"genres"`
	ContentType    string   `json:"contentType"`
	RelatedWords   []string `json:"relatedWords"`
	MakerContentID string   `json:"makerContentId"`
	PlayableInfo   struct {
		PlayableDevices []struct {
			DeviceDeliveryUnits []struct {
				ID                      string `json:"id"`
				DeviceDeliveryQualities []struct {
					IsDownloadable bool   `json:"isDownloadable"`
					IsStreamable   bool   `json:"isStreamable"`
					Typename       string `json:"__typename"`
				} `json:"deviceDeliveryQualities"`
				Typename string `json:"__typename"`
			} `json:"deviceDeliveryUnits"`
			Device   string `json:"device"`
			Name     string `json:"name"`
			Priority int    `json:"priority"`
			Typename string `json:"__typename"`
		} `json:"playableDevices"`
		DeviceGroups []struct {
			ID      string `json:"id"`
			Devices []struct {
				DeviceDeliveryUnits []struct {
					DeviceDeliveryQualities []struct {
						IsStreamable   bool   `json:"isStreamable"`
						IsDownloadable bool   `json:"isDownloadable"`
						Typename       string `json:"__typename"`
					} `json:"deviceDeliveryQualities"`
					Typename string `json:"__typename"`
				} `json:"deviceDeliveryUnits"`
				Typename string `json:"__typename"`
			} `json:"devices"`
			Typename string `json:"__typename"`
		} `json:"deviceGroups"`
		VrViewingType interface{} `json:"vrViewingType"`
		Typename      string      `json:"__typename"`
	} `json:"playableInfo"`
	Typename string `json:"__typename"`
}

type ReviewSummary struct {
	Average          float64 `json:"average"`
	Total            int     `json:"total"`
	WithCommentTotal int     `json:"withCommentTotal"`
	Distributions    []struct {
		Total            int    `json:"total"`
		WithCommentTotal int    `json:"withCommentTotal"`
		Rating           int    `json:"rating"`
		Typename         string `json:"__typename"`
	} `json:"distributions"`
	Typename string `json:"__typename"`
}

type ContentPageDataResponse struct {
	PPVContent    PPVContent    `json:"ppvContent"`
	ReviewSummary ReviewSummary `json:"reviewSummary"`
	Typename      string        `json:"__typename"`
}

type UserReviewsResponse struct {
	Reviews struct {
		Items []struct {
			ID           string    `json:"id"`
			Title        string    `json:"title"`
			Rating       int       `json:"rating"`
			ReviewerID   string    `json:"reviewerId"`
			Nickname     string    `json:"nickname"`
			IsPurchased  bool      `json:"isPurchased"`
			Comment      string    `json:"comment"`
			HelpfulCount int       `json:"helpfulCount"`
			Service      string    `json:"service"`
			IsExposure   bool      `json:"isExposure"`
			PublishDate  time.Time `json:"publishDate"`
			Typename     string    `json:"__typename"`
		} `json:"items"`
		Typename string `json:"__typename"`
	} `json:"reviews"`
}
