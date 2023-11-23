package usecase

import (
	"brick-code-test/app"
	"brick-code-test/model"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
)

type ProductUsecase struct{
	ProductRepository app.IMainRepository
	DoneChan chan struct{}
}

func New(productRepo app.IMainRepository, doneChan chan struct{}) *ProductUsecase{
	return &ProductUsecase{
		ProductRepository: productRepo,
		DoneChan: doneChan,
	}
}

func (ps *ProductUsecase) RunScheduledTasks() {
	// Run ScrapeProducts concurrently every 10 minutes
	go func() {
		for {
			ps.ScrapeAndSave()
			time.Sleep(10 * time.Minute)
		}
	}()

	// Run SaveToCSV concurrently every 15 minutes
	go func() {
		for {
			ps.SaveToCSV("products.csv")
			time.Sleep(15 * time.Minute)
		}
	}()

	// Block main goroutine to keep the application running
	select {
	    case <-ps.DoneChan:
			return
	}
}

func (ps *ProductUsecase) ScrapeAndSave() {
	var wg sync.WaitGroup
	ch := make(chan []model.Product)
	errCh := make(chan error)
	fmt.Println("Running Web Scrapper")
	fmt.Println("=====================")
	// Run ScrapeProducts concurrently and Set Worker Pools to 5
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			products, err := ps.ScrapeProducts("Handphone", 5) // Example values, modify according to your use case
			if err != nil {
				errCh <- err
				return
			}
			ch <- products
		}()
	}

	// Wait for all scraping goroutines to finish
	go func() {
		wg.Wait()
		close(ch)
		close(errCh)
	}()

	// Receive results from goroutines
	for {
		select {
		case products, ok := <-ch:
			if !ok {
				return
			}
			// Delay CSV saving for 10 seconds
			time.Sleep(10 * time.Second)

			fmt.Println("Web Scrapping Done. Saving to CSV")
	        fmt.Println("=====================")
			// Save to CSV after a delay
			err := ps.SaveToCSV("files/products.csv") 
			if err != nil {
				fmt.Println("Error saving to CSV:", err)
			}
			fmt.Println("Products Has been Saved to CSV")
			_ = products // Placeholder for processing product
		case err := <-errCh:
			fmt.Println("Error scraping products:", err)
			ps.DoneChan <- struct{}{}
			// Handle the error appropriately
		}
	}

}

func (ps *ProductUsecase) ScrapeProducts(searchTerm string, pagesCount int) ([]model.Product, error) {
	
	// Logic for scraping product details
	var products []model.Product

	// Implement scraping logic with Goquery
	products = ps.scrapeProductProcess(searchTerm, pagesCount)

	// Implement saving products to database 
	for _, product := range products {
		timeNow := time.Now()
		productId, _ := uuid.NewUUID()
		product.ProductID = productId.String()
		product.CreatedDate = timeNow
		product.UpdatedDate = &timeNow
		err := ps.ProductRepository.InsertProduct(product)
		if err != nil {
			return nil, err
		}
	}

	return products, nil
}

func (ps *ProductUsecase) SaveToCSV(fileName string) error{
	listProducts, err := ps.ProductRepository.ListProducts()
	if err != nil {
		return err
	}

	err = ps.ProductRepository.SaveToCSV(listProducts, fileName)
	if err != nil {
		return  err
	}
	return nil
}

func (ps *ProductUsecase) scrapeProductProcess(searchTerm string, pagesCount int) []model.Product{
	products := make([]model.Product, 0)

	for i := 1; i <= pagesCount; i++ {
		var url string 
		if i == 1 {
			url = fmt.Sprint("https://www.tokopedia.com/search?navsource=&srp_component_id=02.01.00.00&srp_page_id=&srp_page_title=&st=&q="+ strings.ReplaceAll(searchTerm, " ", "+"))
		}else {
			counterToInt := strconv.Itoa(i)
			url = fmt.Sprint("https://www.tokopedia.com/search?navsource=&page="+counterToInt+"&srp_component_id=02.01.00.00&srp_page_id=&srp_page_title=&st=&q="+ strings.ReplaceAll(searchTerm, " ", "+"))
		}
		client := &http.Client{}
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println("Error making Get:", err)
			return products
		}

		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.9999.999 Safari/537.36")
		response, err := client.Do(request)
		if err != nil {
			log.Println("Error making Request:", err)
			response.Body.Close()
			return products
		}
		defer response.Body.Close()
			
		if response.StatusCode != http.StatusOK {
			log.Fatalf("Unexpected status Error : %d", response.StatusCode)
			return products
		}
		
		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			log.Println("Error when read:", err)
			return products
		}
		// some of the url link detail are not valid, since the link are modified from GTM. not uniformed well
		detailProduct := make([]string, 0)
		doc.Find(".pcv3__info-content").Each(func(i int, s *goquery.Selection) {
			linkUrl, _ := s.Attr("href")
			clients := &http.Client{}
			requestDetailProduct, err := http.NewRequest("GET", linkUrl, nil)
			if err != nil {
				//log.Println("Error making Get:", err)
				detailProduct = append(detailProduct, "")
				return
			}
			
			requestDetailProduct.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.9999.999 Safari/537.36")
			responseDetailProduct, err := clients.Do(requestDetailProduct)
			if err != nil {
				//log.Println("Error making Request:", err)
				detailProduct = append(detailProduct, "")
				return
			}
			defer responseDetailProduct.Body.Close()

			
			if responseDetailProduct.StatusCode != http.StatusOK {
				//log.Fatalf("Unexpected status Error : %d", responseDetailProduct.StatusCode)
				detailProduct = append(detailProduct, "")
				return
			}
		
			docDetailProduct, err := goquery.NewDocumentFromReader(responseDetailProduct.Body)
			if err != nil {
				//log.Println("Error when read:", err)
				detailProduct = append(detailProduct, "")
				return
			}

			desc := docDetailProduct.Find("div[data-testid='lblPDPDescriptionProduk']").Text()
			name := docDetailProduct.Find("h1[data-testid='lblPDPDetailProductName']").Text()
			price := docDetailProduct.Find(".price").Text()
			imageUrl, _ :=  docDetailProduct.Find(".css-1c345mg").Attr("src")
			product := model.Product{
				Name: name,
				Description: desc,
				ImageUrl: imageUrl,
				Price: price,
			}

			detailProduct = append(detailProduct, name)
			products = append(products, product)
		})
	
		// we need to get rating and merchant name. cannot get it from detail product page since it is dynamic content
		ratings := make([]string, 0)
		doc.Find(".prd_rating-average-text").Each(func(i int, s *goquery.Selection) {
			if detailProduct[i] == "" {
				return
			}
			ratings = append(ratings, s.Text())
		})

		merchantNames := make([]string, 0)
		doc.Find(".prd_link-shop-name").Each(func(i int, s *goquery.Selection) {
			if detailProduct[i] == "" {
				return
			}
			merchantNames = append(merchantNames, s.Text())
		})

		counter := 0
		if i > 1 {
			counter = (len(products) - len(ratings)) - 1
		}

		
		for j := counter; j < len(products); j++ {
			if i == 1 && (j < len(ratings) && j < len(merchantNames)) {
				products[j].Rating = ratings[j]
				products[j].MerchantName = merchantNames[j]
			}else if i > 1{
				counterRatings := 0
				products[j].Rating = ratings[counterRatings]
				products[j].MerchantName = merchantNames[counterRatings]
				counterRatings++
			}
		}
	}	
	return products
}