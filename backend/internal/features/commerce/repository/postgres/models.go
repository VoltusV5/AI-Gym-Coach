package commerce_postgres_repository

import "time"

type OrderInput struct {
	ProgramID        string `json:"program_id"`
	ProgramVersionID string `json:"program_version_id"`
}
type Order struct {
	ID               string    `json:"id"`
	Status           string    `json:"status"`
	AmountMinor      int64     `json:"amount_minor"`
	Currency         string    `json:"currency"`
	CheckoutURL      *string   `json:"checkout_url"`
	Entitlement      any       `json:"entitlement"`
	ProgramID        string    `json:"program_id"`
	ProgramVersionID string    `json:"program_version_id"`
	TrainerID        int       `json:"trainer_id"`
	UserID           int       `json:"-"`
	CommissionBps    int64     `json:"-"`
	CommissionMinor  int64     `json:"-"`
	AuthorMinor      int64     `json:"-"`
	PaymentID        string    `json:"-"`
	CreatedAt        time.Time `json:"created_at"`
}
type Event struct {
	EventID     string `json:"provider_event_id"`
	PaymentID   string `json:"provider_payment_id"`
	OrderID     string `json:"order_id"`
	Type        string `json:"event_type"`
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}
type Commission struct {
	Bps         int64     `json:"commission_bps"`
	EffectiveAt time.Time `json:"effective_at"`
	Version     int64     `json:"version"`
}
type PayoutInput struct {
	TrainerID   int    `json:"trainer_id"`
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}
type Payout struct {
	PaidAt           *time.Time `json:"paid_at"`
	ID               string     `json:"id"`
	TrainerID        int        `json:"trainer_id"`
	AmountMinor      int64      `json:"amount_minor"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	ProviderPayoutID *string    `json:"provider_payout_id"`
}
type Decision struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}
type Page[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor"`
}
