package models

import "github.com/dekunt/uadmin"

// Category model ...
type Category struct {
	uadmin.Model
	Name string `uadmin:"required"`
	Icon string `uadmin:"image"`
}
