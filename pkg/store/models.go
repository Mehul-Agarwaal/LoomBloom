package store

import "time"

type Organization struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	GSTNumber        string    `json:"gst_number"`
	OwnerName        string    `json:"owner_name"`
	OwnerEmail       string    `json:"owner_email"`
	Phone            string    `json:"phone"`
	PhoneVerified    bool      `json:"phone_verified"`
	SubscriptionPlan string    `json:"subscription_plan"`
	PasswordHash     string    `json:"-"`
	CreatedAt        time.Time `json:"created_at"`
}

type Machine struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	MachineNumber  string    `json:"machine_number"`
	Size           string    `json:"size"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type Worker struct {
	ID                string    `json:"id"`
	OrganizationID    string    `json:"organization_id"`
	Name              string    `json:"name"`
	Phone             string    `json:"phone"`
	AverageProduction float64   `json:"average_production"`
	ClosePerShift     int       `json:"close_per_shift"`
	NumberOfShifts    int       `json:"number_of_shifts"`
	CreatedAt         time.Time `json:"created_at"`
}

type StockItem struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Quality        string    `json:"quality"`
	StoredOn       string    `json:"stored_on"`
	Quantity       float64   `json:"quantity"`
	Unit           string    `json:"unit"`
	CreatedAt      time.Time `json:"created_at"`
}

type ProductionEntry struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	MachineID      string    `json:"machine_id"`
	WorkerID       string    `json:"worker_id"`
	EntryDate      time.Time `json:"entry_date"`
	Shift          string    `json:"shift"`
	Quantity       float64   `json:"quantity"`
	Unit           string    `json:"unit"`
	CreatedAt      time.Time `json:"created_at"`
}

type Session struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Token          string    `json:"token"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

