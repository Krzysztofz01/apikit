package content

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type HtmlContent interface {
	GetFirstElement(xpath string) (HtmlContentElement, bool, error)
	GetSingleElement(xpath string) (HtmlContentElement, bool, error)
	GetAllElements(xpath string) ([]HtmlContentElement, error)
	GetRawContent() (string, error)
}

type htmlContent struct {
	document *html.Node
}

func CreateHtmlContent(html string) (HtmlContent, error) {
	if len(html) == 0 {
		return nil, fmt.Errorf("content: invalid empty html source provided")
	}

	htmlReader := strings.NewReader(html)

	node, err := htmlquery.Parse(htmlReader)
	if err != nil {
		return nil, fmt.Errorf("content: failed to parse the html content: %w", err)
	}

	return &htmlContent{
		document: node,
	}, nil
}

func (h *htmlContent) GetAllElements(xpath string) ([]HtmlContentElement, error) {
	nodes, err := htmlquery.QueryAll(h.document, xpath)
	if err != nil {
		return nil, fmt.Errorf("content: failed to query for nodes: %w", err)
	}

	elements := make([]HtmlContentElement, 0, len(nodes))
	for _, node := range nodes {
		if element, err := createHtmlContentElement(node); err != nil {
			return nil, fmt.Errorf("content: failed to create the html element: %w", err)
		} else {
			elements = append(elements, element)
		}
	}

	return elements, nil
}

func (h *htmlContent) GetFirstElement(xpath string) (HtmlContentElement, bool, error) {
	a, _ := h.GetRawContent()
	_ = a

	node, err := htmlquery.Query(h.document, xpath)
	if err != nil {
		return nil, false, fmt.Errorf("content: failed to query for the first node: %w", err)
	}

	if node == nil {
		return nil, false, nil
	}

	if element, err := createHtmlContentElement(node); err != nil {
		return nil, true, fmt.Errorf("content: failed to create the html element: %w", err)
	} else {
		return element, true, nil
	}
}

func (h *htmlContent) GetRawContent() (string, error) {
	return htmlquery.OutputHTML(h.document, true), nil
}

func (h *htmlContent) GetSingleElement(xpath string) (HtmlContentElement, bool, error) {
	nodes, err := htmlquery.QueryAll(h.document, xpath)
	if err != nil {
		return nil, false, fmt.Errorf("content: failed to query for nodes: %w", err)
	}

	if len(nodes) != 1 {
		return nil, true, fmt.Errorf("content: multiple matching nodes found: %w", err)
	}

	if element, err := createHtmlContentElement(nodes[0]); err != nil {
		return nil, true, fmt.Errorf("content: failed to create the html element: %w", err)
	} else {
		return element, true, nil
	}
}

type HtmlContentValuePreprocess func(in string) (string, error)

type HtmlContentElement interface {
	GetInnerTextString(preprocess HtmlContentValuePreprocess) (string, error)
	GetInnerTextInt(preprocess HtmlContentValuePreprocess) (int, error)
	GetInnerTextFloat(preprocess HtmlContentValuePreprocess) (float64, error)
	GetAttributeValueString(name string, preprocess HtmlContentValuePreprocess) (string, error)
	GetAttributeValueInt(name string, preprocess HtmlContentValuePreprocess) (int, error)
	GetAttributeValueFloat(name string, preprocess HtmlContentValuePreprocess) (float64, error)
}

type htmlContentElement struct {
	node *html.Node
}

func createHtmlContentElement(node *html.Node) (HtmlContentElement, error) {
	if node == nil {
		return nil, fmt.Errorf("content: element node is nil")
	}

	return &htmlContentElement{
		node: node,
	}, nil
}

func (h *htmlContentElement) GetAttributeValue(name string, preprocess HtmlContentValuePreprocess) (string, error) {
	if len(name) == 0 {
		return "", fmt.Errorf("content: invalid attribute name provided")
	}

	value := htmlquery.SelectAttr(h.node, name)

	if preprocess != nil {
		return func() (preprocessValue string, err error) {
			defer func() {
				if panicErr := recover(); panicErr != nil {
					err = fmt.Errorf("content: attribute value preprocessing failed: %s", panicErr)
				}
			}()

			preprocessValue, err = preprocess(value)
			return
		}()
	}

	return value, nil
}

func (h *htmlContentElement) GetInnerText(preprocess HtmlContentValuePreprocess) (string, error) {
	value := htmlquery.OutputHTML(h.node, false)
	if preprocess != nil {
		return func() (preprocessValue string, err error) {
			defer func() {
				if panicErr := recover(); panicErr != nil {
					err = fmt.Errorf("content: inner text value preprocessing failed: %s", panicErr)
				}
			}()

			preprocessValue, err = preprocess(value)
			return
		}()
	}

	return value, nil
}

func (h *htmlContentElement) GetAttributeValueFloat(name string, preprocess HtmlContentValuePreprocess) (float64, error) {
	value, err := h.GetAttributeValue(name, preprocess)
	if err != nil {
		return 0, fmt.Errorf("content: failed to access attribute float value: %w", err)
	}

	if valueF, err := strconv.ParseFloat(value, 64); err != nil {
		return 0, fmt.Errorf("content: failed to parse the attribute value as float64: %w", err)
	} else {
		return valueF, nil
	}
}

func (h *htmlContentElement) GetAttributeValueInt(name string, preprocess HtmlContentValuePreprocess) (int, error) {
	value, err := h.GetAttributeValue(name, preprocess)
	if err != nil {
		return 0, fmt.Errorf("content: failed to access attribute int value: %w", err)
	}

	valueI64, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("content: failed to parse the attribute value as int: %w", err)
	}

	if valueI64 > math.MaxInt32 {
		return 0, fmt.Errorf("content: the target integer value is overflowing")
	}

	return int(valueI64), nil
}

func (h *htmlContentElement) GetAttributeValueString(name string, preprocess HtmlContentValuePreprocess) (string, error) {
	if value, err := h.GetAttributeValue(name, preprocess); err != nil {
		return "", fmt.Errorf("content: failed to access attribute string value: %w", err)
	} else {
		return value, nil
	}
}

func (h *htmlContentElement) GetInnerTextFloat(preprocess HtmlContentValuePreprocess) (float64, error) {
	value, err := h.GetInnerText(preprocess)
	if err != nil {
		return 0, fmt.Errorf("content: failed to access inner text float value: %w", err)
	}

	if valueF, err := strconv.ParseFloat(value, 64); err != nil {
		return 0, fmt.Errorf("content: failed to parse the inner text value as float64: %w", err)
	} else {
		return valueF, nil
	}
}

func (h *htmlContentElement) GetInnerTextInt(preprocess HtmlContentValuePreprocess) (int, error) {
	value, err := h.GetInnerText(preprocess)
	if err != nil {
		return 0, fmt.Errorf("content: failed to access inner text int value: %w", err)
	}

	valueI64, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("content: failed to parse the inner text value as int: %w", err)
	}

	if valueI64 > math.MaxInt32 {
		return 0, fmt.Errorf("content: the target integer value is overflowing")
	}

	return int(valueI64), nil
}

func (h *htmlContentElement) GetInnerTextString(preprocess HtmlContentValuePreprocess) (string, error) {
	if value, err := h.GetInnerText(preprocess); err != nil {
		return "", fmt.Errorf("content: failed to access inner text string value: %w", err)
	} else {
		return value, nil
	}
}
