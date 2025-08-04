package modelmediaasia

type actorInfoResponse struct {
	Data struct {
		ID                int    `json:"id"`
		Name              string `json:"name"`
		NameCn            string `json:"name_cn"`
		Avatar            string `json:"avatar"`
		Gender            string `json:"gender"`
		HeightFt          int    `json:"height_ft"`
		HeightIn          int    `json:"height_in"`
		WeightLbs         int    `json:"weight_lbs"`
		MeasurementsChest string `json:"measurements_chest"`
		MeasurementsWaist int    `json:"measurements_waist"`
		MeasurementsHips  int    `json:"measurements_hips"`
		// API return empty for the following fields.
		Cover       string `json:"cover"`
		MobileCover string `json:"mobile_cover"`
		SocialMedia string `json:"socialmedia"`
		PeriodViews int    `json:"period_views"`
		BirthDay    string `json:"birth_day"`
		BirthPlace  string `json:"birth_place"`
		Description string `json:"description"`
		Tooltips    string `json:"tooltips"`
		HeightCm    int    `json:"height_cm"`
		WeightKg    int    `json:"weight_kg"`
		Photos      []struct {
			Image string `json:"image"`
		} `json:"photos"`
	} `json:"data"`
}

type movieInfoResponse struct {
	Data struct {
		ID            int    `json:"id"`
		SerialNumber  string `json:"serial_number"`
		Title         string `json:"title"`
		TitleCn       string `json:"title_cn"`
		Description   string `json:"description"`
		DescriptionCn string `json:"description_cn"`
		Trailer       string `json:"trailer"`
		Duration      int    `json:"duration"`
		Cover         string `json:"cover"`
		PreviewVideo  string `json:"preview_video"`
		PublishedAt   int64  `json:"published_at"`
		Models        []struct {
			ID                int    `json:"id"`
			Name              string `json:"name"`
			NameCn            string `json:"name_cn"`
			Avatar            string `json:"avatar"`
			Gender            string `json:"gender"`
			HeightFt          int    `json:"height_ft"`
			HeightIn          int    `json:"height_in"`
			WeightLbs         int    `json:"weight_lbs"`
			MeasurementsChest string `json:"measurements_chest"`
			MeasurementsWaist int    `json:"measurements_waist"`
			MeasurementsHips  int    `json:"measurements_hips"`
			// API return empty for the following fields.
			Cover       string      `json:"cover"`
			MobileCover string      `json:"mobile_cover"`
			BirthDay    string      `json:"birth_day"`
			BirthPlace  string      `json:"birth_place"`
			HeightCm    int         `json:"height_cm"`
			WeightKg    int         `json:"weight_kg"`
			Videos      interface{} `json:"videos"`
			Photos      interface{} `json:"photos"`
		} `json:"models"`
		Tags []struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			NameCn string `json:"name_cn"`
		} `json:"tags"`
	} `json:"data"`
}

type searchResponse struct {
	Data struct {
		Videos []struct {
			ID            int    `json:"id"`
			SerialNumber  string `json:"serial_number"`
			Title         string `json:"title"`
			TitleCn       string `json:"title_cn"`
			Description   string `json:"description"`
			DescriptionCn string `json:"description_cn"`
			Cover         string `json:"cover"`
			PublishedAt   int64  `json:"published_at"`
		} `json:"videos"`
		Models []struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			NameCn string `json:"name_cn"`
			Avatar string `json:"avatar"`
		} `json:"models"`
	} `json:"data"`
}
