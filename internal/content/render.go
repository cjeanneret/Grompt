package content

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/net/html"
)

func Render(data []byte, format Format) (fyne.CanvasObject, error) {
	switch format {
	case FormatMarkdown:
		return renderMarkdown(string(data)), nil
	case FormatHTML:
		return renderHTML(string(data))
	default:
		return nil, fmt.Errorf("unsupported render format: %s", format)
	}
}

func renderMarkdown(markdown string) *widget.RichText {
	richText := widget.NewRichTextFromMarkdown(markdown)
	richText.Wrapping = fyne.TextWrapWord
	ApplyTypography(richText)
	return richText
}

func renderHTML(rawHTML string) (fyne.CanvasObject, error) {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	ctx := renderContext{segments: make([]widget.RichTextSegment, 0, 32)}
	renderNode(doc, textStyle{}, &ctx)

	richText := widget.NewRichText(ctx.segments...)
	richText.Wrapping = fyne.TextWrapWord
	ApplyTypography(richText)
	return richText, nil
}

type textStyle struct {
	bold      bool
	italic    bool
	monospace bool
}

type renderContext struct {
	segments []widget.RichTextSegment
}

func renderNode(node *html.Node, style textStyle, ctx *renderContext) {
	switch node.Type {
	case html.TextNode:
		appendText(ctx, normalizeWhitespace(node.Data), style)
		return
	case html.ElementNode:
		renderElement(node, style, ctx)
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		renderNode(child, style, ctx)
	}
}

func renderElement(node *html.Node, style textStyle, ctx *renderContext) {
	switch node.Data {
	case "br":
		appendRawText(ctx, "\n", style)
		return
	case "p", "div", "section", "article":
		appendRawText(ctx, "\n", style)
		renderChildren(node, style, ctx)
		appendRawText(ctx, "\n", style)
		return
	case "h1", "h2", "h3":
		headerStyle := style
		headerStyle.bold = true
		appendRawText(ctx, "\n", headerStyle)
		renderChildren(node, headerStyle, ctx)
		appendRawText(ctx, "\n", headerStyle)
		return
	case "strong", "b":
		next := style
		next.bold = true
		renderChildren(node, next, ctx)
		return
	case "em", "i":
		next := style
		next.italic = true
		renderChildren(node, next, ctx)
		return
	case "code", "pre":
		next := style
		next.monospace = true
		renderChildren(node, next, ctx)
		return
	case "a":
		appendLink(node, style, ctx)
		return
	case "ul":
		appendRawText(ctx, "\n", style)
		renderList(node, false, style, ctx)
		appendRawText(ctx, "\n", style)
		return
	case "ol":
		appendRawText(ctx, "\n", style)
		renderList(node, true, style, ctx)
		appendRawText(ctx, "\n", style)
		return
	}

	renderChildren(node, style, ctx)
}

func renderChildren(node *html.Node, style textStyle, ctx *renderContext) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		renderNode(child, style, ctx)
	}
}

func renderList(node *html.Node, ordered bool, style textStyle, ctx *renderContext) {
	index := 1
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode || child.Data != "li" {
			continue
		}

		prefix := "- "
		if ordered {
			prefix = strconv.Itoa(index) + ". "
		}
		appendRawText(ctx, prefix, style)
		renderChildren(child, style, ctx)
		appendRawText(ctx, "\n", style)
		index++
	}
}

func appendLink(node *html.Node, style textStyle, ctx *renderContext) {
	var href string
	for _, attr := range node.Attr {
		if attr.Key == "href" {
			href = attr.Val
			break
		}
	}

	text := strings.TrimSpace(extractText(node))
	if text == "" {
		text = href
	}
	if text == "" {
		return
	}

	if href == "" {
		appendText(ctx, text, style)
		return
	}

	parsedURL, err := url.Parse(href)
	if err != nil {
		appendText(ctx, text, style)
		return
	}

	ctx.segments = append(ctx.segments, &widget.HyperlinkSegment{
		Text: text,
		URL:  parsedURL,
	})
}

func appendText(ctx *renderContext, text string, style textStyle) {
	if text == "" {
		return
	}
	appendRawText(ctx, text, style)
}

func appendRawText(ctx *renderContext, text string, style textStyle) {
	if text == "" {
		return
	}

	ctx.segments = append(ctx.segments, &widget.TextSegment{
		Text: text,
		Style: widget.RichTextStyle{
			Inline:   true,
			SizeName: ThemeSizeContentBody,
			TextStyle: fyne.TextStyle{
				Bold:      style.bold,
				Italic:    style.italic,
				Monospace: style.monospace,
			},
		},
	})
}

func normalizeWhitespace(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}

func extractText(node *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(normalizeWhitespace(n.Data))
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return b.String()
}
