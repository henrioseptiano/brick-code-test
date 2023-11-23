package repository

import (
	"brick-code-test/model"
	"encoding/csv"
	"errors"
	"os"

	"gorm.io/gorm"
)

type ProductRepository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *ProductRepository{
	return &ProductRepository{
		DB:db,
	}
}

func (pr *ProductRepository) InsertProduct(product model.Product) error{ 
	if err := pr.DB.Create(&product).Error; err != nil {
		return err
	}
	return nil
}

func (pr *ProductRepository) ListProducts() ([]model.Product, error){
	var products []model.Product 
	if err := pr.DB.Order("rating DESC").Limit(100).Find(&products).Error; err != nil {
		return products, err
	}
	return products,nil
}

func (pr *ProductRepository) SaveToCSV(products []model.Product, filename string) error{ 
	if _, err := os.Stat(filename); err == nil {
        // If the file exists, attempt to remove it
        err := os.Remove(filename)
        if err != nil {
            return errors.New("Cannot Save to CSV")
        }
    }

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Writing headers to CSV
	headers := []string{"Name", "Description", "ImageURL", "Price", "Rating", "MerchantName"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Writing product data to CSV
	for _, product := range products {
		row := []string{product.Name, product.Description, product.ImageUrl, product.Price, product.Rating, product.MerchantName}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}