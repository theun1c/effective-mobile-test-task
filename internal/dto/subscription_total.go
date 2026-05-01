package dto

type SubscriptionTotalQuery struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	UserID      *string `json:"user_id,omitempty"`
	ServiceName *string `json:"service_name,omitempty"`
}

type SubscriptionTotalResponse struct {
	TotalCost int `json:"total_cost"`
}
