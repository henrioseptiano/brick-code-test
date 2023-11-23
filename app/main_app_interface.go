package app

import "brick-code-test/model"

type IMainRepository interface {
	InsertProduct(product model.Product) error
	ListProducts()([]model.Product, error)
	SaveToCSV(products []model.Product, filename string) error
}