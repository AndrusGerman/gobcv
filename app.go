package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	htmlStr, err := getHtml("https://www.bcv.org.ve/")
	if err != nil {
		log.Fatal(err)
	}
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		log.Fatal(err)
	}

	var bodyNode = findByTag(doc, "body")
	var findEuroContainer = findByTag(findById(bodyNode, "euro"), "strong") // example value 144,37320000
	var euroContent = findEuroContainer.FirstChild.Data

	var findDolarContainer = findByTag(findById(bodyNode, "dolar"), "strong") // example value 168,34059493

	var dolarContent = findDolarContainer.FirstChild.Data

	type Currency struct {
		Currency string
		Value    float64
	}

	var euroValue = strings.ReplaceAll(strings.TrimSpace(euroContent), ",", ".")
	var dolarValue = strings.ReplaceAll(strings.TrimSpace(dolarContent), ",", ".")

	euroValueFloat, err := strconv.ParseFloat(euroValue, 64)
	if err != nil {
		log.Fatal(err)
	}

	dolarValueFloat, err := strconv.ParseFloat(dolarValue, 64)
	if err != nil {
		log.Fatal(err)
	}

	var currencies = []Currency{
		{Currency: "euro", Value: euroValueFloat},
		{Currency: "dolar", Value: dolarValueFloat},
	}

	fmt.Println(currencies)
}

func findById(node *html.Node, id string) *html.Node {
	// Verificar si el nodo actual tiene el ID buscado
	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if attr.Key == "id" && attr.Val == id {
				return node
			}
		}
	}

	// Buscar recursivamente en todos los nodos hijos
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findById(child, id); result != nil {
			return result
		}
	}

	return nil
}

func findByTag(node *html.Node, tagName string) *html.Node {
	// Verificar si el nodo actual es del tipo de elemento buscado
	if node.Type == html.ElementNode && node.Data == tagName {
		return node
	}

	// Buscar recursivamente en todos los nodos hijos
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findByTag(child, tagName); result != nil {
			return result
		}
	}

	return nil
}

func whatIsThis(node *html.Node) {
	fmt.Println(node.Data, node.Type, node.Namespace)

}

func getHtml(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return string(body), nil
}
