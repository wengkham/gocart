package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const shoes string = "Shoes"
const jeans string = "Jeans"
const smartphones string = "Smart Phones"
const ring string = "Ring"
const tv string = "TV"
const radio string = "Radio"

var checkoutMemTable = map[int64]map[string]map[string]string{}
var pricesMemTable = map[int64]map[string]string{}
var comboKeyMemTable = map[string]float32{}
var comboMemTable = [][]string{}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// Controllers
	r.GET("/", func(c *gin.Context) {
		products := getProducts()
		comboList := createComboDiscount()
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":     "Cart",
			"products":  products,
			"comboList": comboList,
		})
	})

	r.POST("/checkout", func(c *gin.Context) {
		items := c.PostFormArray("items")
		checkouts, prices := getCheckoutItems(items)
		id := createMemTable(checkouts, prices)

		if len(items) == 0 {
			c.Redirect(http.StatusMovedPermanently, "/")
			c.Abort()
		} else {
			c.HTML(http.StatusOK, "checkout.tmpl", gin.H{
				"title":     "Checkout",
				"checkouts": checkouts,
				"prices":    prices,
				"id":        id,
			})
		}
	})

	r.POST("/pay", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.PostForm("id"))
		prices, exist := pricesMemTable[int64(id)]

		if !exist {
			c.Redirect(http.StatusMovedPermanently, "/")
			c.Abort()
		} else {
			c.HTML(http.StatusOK, "pay.tmpl", gin.H{
				"title":  "Pay",
				"prices": prices,
			})
		}
	})

	r.POST("/done", func(c *gin.Context) {
		c.HTML(http.StatusOK, "done.tmpl", gin.H{
			"title": "Payment done",
		})
	})

	r.Run(":8081") // listen and serve on 0.0.0.0:8080
}

// Product store
func getProducts() map[string]float32 {
	products := map[string]float32{}
	products[shoes] = 10.00
	products[ring] = 20.00
	products[jeans] = 50.00
	products[smartphones] = 500.00
	products[tv] = 1500.00
	products[radio] = 300.00

	return products
}

//Discount store
func getDiscount() map[string]float32 {
	discounts := map[string]float32{}
	discounts[shoes] = 0.03
	discounts[ring] = 0.04
	discounts[jeans] = 0.10
	discounts[smartphones] = 0.20

	return discounts
}

// Retrieve combo discount if any
func getComboDiscount(items []string) float32 {
	comboCounter := map[int]int{}
	var comboDiscounts float32

	for ci, cv := range comboMemTable {
		cLen := len(cv)
		for _, v := range items {
			i := sort.Search(cLen, func(i int) bool { return v <= cv[i] })
			if i < cLen && cv[i] == v {
				comboCounter[ci]++
			}
		}
		if comboCounter[ci] == cLen {
			comboKey := strings.Join(comboMemTable[ci], "")
			discount := comboKeyMemTable[comboKey]
			if discount > comboDiscounts {
				comboDiscounts = discount
			}
		}
	}
	return comboDiscounts
}

// Save mem table
func setComboDiscount(comboList []combos) {
	for _, k := range comboList {
		combo := k.Combo
		sort.Strings(combo)
		comboKey := strings.Join(combo, "")
		comboKeyMemTable[comboKey] = k.Discount
		comboMemTable = append(comboMemTable, k.Combo)
	}
}

// Combo discount store
func createComboDiscount() []combos {
	comboList := []combos{}
	comboList = append(comboList, combos{Combo: []string{shoes, jeans}, Discount: 20})
	comboList = append(comboList, combos{Combo: []string{shoes, jeans, ring}, Discount: 25})
	comboList = append(comboList, combos{Combo: []string{radio, tv, smartphones, ring}, Discount: 150})
	setComboDiscount(comboList)

	return comboList
}

// Checkout
func getCheckoutItems(items []string) (map[string]map[string]string, map[string]string) {
	checkout := map[string]map[string]string{}
	prices := map[string]string{}

	products := getProducts()
	discounts := getDiscount()

	var totalComboDiscount, totalFinalPrice float32
	totalComboDiscount = getComboDiscount(items)

	for _, item := range items {
		price, p := products[item]
		if p {
			var discount, finalPrice float32
			percent, d := discounts[item]

			if d {
				discount = price * percent
			}

			finalPrice = price - discount
			totalFinalPrice += finalPrice

			checkout[item] = map[string]string{}
			checkout[item]["price"] = fmt.Sprintf("%.2f", price)
			checkout[item]["discount"] = fmt.Sprintf("%.2f", discount)
			checkout[item]["final_price"] = fmt.Sprintf("%.2f", finalPrice)
		}
	}

	prices["totalComboDiscount"] = fmt.Sprintf("%.2f", totalComboDiscount)
	prices["totalFinalPrice"] = fmt.Sprintf("%.2f", totalFinalPrice)
	prices["totalFinalComboPrice"] = fmt.Sprintf("%.2f", totalFinalPrice-totalComboDiscount)

	return checkout, prices
}

// Cart and price store
func createMemTable(checkouts map[string]map[string]string, prices map[string]string) int64 {
	now := time.Now()
	sec := now.Unix()

	checkoutMemTable[sec] = checkouts
	pricesMemTable[sec] = prices

	return sec
}

type combos struct {
	Combo    []string
	Discount float32
}
