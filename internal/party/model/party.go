package model

type Party struct {
	ID          int64
	Name        string
	Description string
	Promises    string
	Image       string
	IsAllowed   bool
	BanReason   *string
}

type CreatePartyInput struct {
	Name        string
	Description string
	Promises    string
	Image       string
}

type UpdatePartyInput struct {
	ID          int64
	Name        string
	Description string
	Promises    string
	Image       string
	IsAllowed   bool
	BanReason   *string
}

type ListPartiesFilter struct {
	Limit  int
	Offset int
}
