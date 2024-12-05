package graph

import "scrapping_service/internal/scrapping/external"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Scrapping external.Scrapping
}
