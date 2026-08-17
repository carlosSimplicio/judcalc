package domain

type Service struct {
	ID            int64    `json:"id"`
	AreaID        int64    `json:"area_id"`
	Name          string   `json:"name"`
	AmountCents   *int64   `json:"amount_cents"`
	PercentageMin *float64 `json:"percentage_min"`
	PercentageMax *float64 `json:"percentage_max"`
}
