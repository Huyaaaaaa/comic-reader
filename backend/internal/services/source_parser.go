package services

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/huyaaaaaa/hehuan-reader/internal/models"
)

type monsterParser struct{}

func newMonsterParser() *monsterParser {
	return &monsterParser{}
}

func (parser *monsterParser) ParseListPage(baseURL string, html string) (*models.RemoteListPage, error) {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	cipher := extractPageCipher(html)

	items := make([]models.RemoteComicListItem, 0, 120)
	document.Find(".gallery .image").Each(func(index int, selection *goquery.Selection) {
		href, ok := selection.Find(".image-inner a[href*='post.php?ID=']").First().Attr("href")
		if !ok {
			return
		}
		comicID := extractQueryInt64(href, "ID")
		if comicID == 0 {
			return
		}
		title := cleanRemoteText(cipher.decrypt(selection.Find(".image-info h5.title a").First().Text()))
		coverURL := strings.TrimSpace(firstNonEmptyAttr(selection.Find("img").First(), "data-src", "src"))
		rating, ratingCount := parseRatingMeta(selection.Find(".rating").First().Text())
		favorites := extractTrailingInt(selection.Find(".pull-right small").First().Text())
		items = append(items, models.RemoteComicListItem{
			ID:          comicID,
			Title:       title,
			CoverURL:    resolveMaybeURL(baseURL, coverURL),
			Rating:      rating,
			RatingCount: ratingCount,
			Favorites:   favorites,
		})
	})

	totalPages := 1
	document.Find(".pagination a[href*='page=']").Each(func(index int, selection *goquery.Selection) {
		if href, ok := selection.Attr("href"); ok {
			page := int(extractQueryInt64(href, "page"))
			if page > totalPages {
				totalPages = page
			}
		}
	})
	currentPage := 1
	activeText := strings.TrimSpace(document.Find(".pagination li.active span").First().Text())
	if value, err := strconv.Atoi(activeText); err == nil && value > 0 {
		currentPage = value
	}
	return &models.RemoteListPage{CurrentPage: currentPage, TotalPages: totalPages, Items: items}, nil
}

func (parser *monsterParser) ParseDetailPage(baseURL string, comicID int64, html string) (*models.RemoteComicDetailBundle, error) {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	cipher := extractPageCipher(html)

	bundle := &models.RemoteComicDetailBundle{ID: comicID}
	nameRoot := document.Find(".product-wrap .name").First()
	titleNodes := nameRoot.Find("h2.d")
	bundle.Title = cleanRemoteText(cipher.decrypt(titleNodes.First().Text()))
	if bundle.Title == "" {
		bundle.Title = cleanRemoteText(cipher.decrypt(document.Find("h1.page-header.d").First().Text()))
	}
	if titleNodes.Length() > 1 {
		bundle.Subtitle = cleanRemoteText(cipher.decrypt(titleNodes.Eq(1).Text()))
	}
	bundle.CoverURL = resolveMaybeURL(baseURL, strings.TrimSpace(firstNonEmptyAttr(document.Find(".product-image img").First(), "src", "data-src")))
	bundle.Rating, bundle.RatingCount = parseRatingMeta(nameRoot.Text())
	bundle.Authors = parseAuthors(cipher, nameRoot)
	bundle.CategoryID, bundle.CategoryName = parseCategory(cipher, document)
	bundle.Tags = parseTags(cipher, document)
	bundle.SourceCreatedAt, bundle.SourceUpdatedAt = parseDetailDates(cipher, document)
	if bundle.SourceCreatedAt == "" && bundle.Title == "" {
		return nil, fmt.Errorf("unexpected detail page format")
	}
	return bundle, nil
}

func (parser *monsterParser) ParseReaderPage(baseURL string, comicID int64, html string, bundle *models.RemoteComicDetailBundle) error {
	httpImage := extractReaderBaseImageURL(html)
	if httpImage == "" {
		return fmt.Errorf("reader base image url not found")
	}
	items, err := extractReaderImageItems(html)
	if err != nil {
		return err
	}
	images := make([]models.RemoteComicImage, 0, len(items))
	for _, item := range items {
		sortIndex, err := strconv.Atoi(strings.TrimSpace(item.Sort))
		if err != nil {
			continue
		}
		ext := strings.TrimSpace(item.Extension)
		if ext == "" {
			ext = "jpg"
		}
		imageURL := fmt.Sprintf("%s%s_w1500.%s", httpImage, item.NewFilename, ext)
		if version := strings.TrimSpace(item.Version); version != "" && version != "0" {
			imageURL += "?v=" + url.QueryEscape(version)
		}
		images = append(images, models.RemoteComicImage{
			Sort:      sortIndex,
			ImageURL:  resolveMaybeURL(baseURL, imageURL),
			Extension: ext,
		})
	}
	sort.Slice(images, func(left int, right int) bool {
		return images[left].Sort < images[right].Sort
	})
	bundle.Images = images
	return nil
}

type readerImageItem struct {
	Sort        string `json:"sort"`
	ComicID     string `json:"comic_id"`
	ExtPath     string `json:"ext_path_folder"`
	NewFilename string `json:"new_filename"`
	Extension   string `json:"extension"`
	Version     string `json:"version"`
}

var (
	pageCipherPattern    = regexp.MustCompile(`(?s)var\s+aei\s*=\s*'([^']+)'.*?var\s+aek\s*=\s*'([^']+)'.*?var\s+enc\s*=\s*(true|false)`)
	ratingPattern        = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*\(([0-9]+)\)`)
	trailingCountPattern = regexp.MustCompile(`([0-9]+)\s*$`)
	readerBasePattern    = regexp.MustCompile(`var\s+HTTP_IMAGE\s*=\s*"([^"]+)"`)
	readerListPattern    = regexp.MustCompile(`(?s)Original_Image_List\s*=\s*(\[.*?\]);`)
	countSuffixPattern   = regexp.MustCompile(`\([0-9]+\)\s*$`)
)

type pageCipher struct {
	enabled bool
	key     []byte
	iv      []byte
}

func extractPageCipher(html string) pageCipher {
	matches := pageCipherPattern.FindStringSubmatch(html)
	if len(matches) != 4 {
		return pageCipher{}
	}
	return pageCipher{
		enabled: strings.EqualFold(matches[3], "true"),
		iv:      []byte(matches[1]),
		key:     []byte(matches[2]),
	}
}

func (cipher pageCipher) decrypt(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || !cipher.enabled || len(cipher.key) != 16 || len(cipher.iv) != aes.BlockSize {
		return value
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(decoded) == 0 || len(decoded)%aes.BlockSize != 0 {
		return value
	}
	block, err := aes.NewCipher(cipher.key)
	if err != nil {
		return value
	}
	plain := make([]byte, len(decoded))
	mode := NewCBCDecrypter(block, cipher.iv)
	mode.CryptBlocks(plain, decoded)
	plain, err = pkcs7Unpad(plain)
	if err != nil {
		return value
	}
	return cleanRemoteText(string(plain))
}

func parseAuthors(cipher pageCipher, root *goquery.Selection) []models.RemoteAuthorRef {
	authors := make([]models.RemoteAuthorRef, 0, 4)
	seen := map[int64]struct{}{}
	root.Find("a[href*='author_id=']").Each(func(index int, selection *goquery.Selection) {
		href, ok := selection.Attr("href")
		if !ok {
			return
		}
		authorID := extractQueryInt64(href, "author_id")
		if authorID != 0 {
			if _, ok := seen[authorID]; ok {
				return
			}
			seen[authorID] = struct{}{}
		}
		name := cleanRemoteText(cipher.decrypt(selection.Text()))
		if name == "" {
			return
		}
		var externalID *int64
		if authorID != 0 {
			externalID = &authorID
		}
		authors = append(authors, models.RemoteAuthorRef{ExternalID: externalID, Name: name, Position: len(authors)})
	})
	return authors
}

func parseCategory(cipher pageCipher, document *goquery.Document) (*int64, string) {
	selection := document.Find("#category_list a[href*='category_id=']").First()
	if selection.Length() == 0 {
		return nil, ""
	}
	href, _ := selection.Attr("href")
	categoryID := extractQueryInt64(href, "category_id")
	name := cleanRemoteText(cipher.decrypt(selection.Text()))
	name = countSuffixPattern.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)
	if categoryID == 0 {
		return nil, name
	}
	return &categoryID, name
}

func parseTags(cipher pageCipher, document *goquery.Document) []models.RemoteTagRef {
	tags := make([]models.RemoteTagRef, 0, 12)
	seen := map[int64]struct{}{}
	document.Find("#more-information2 a[href*='tag_id=']").Each(func(index int, selection *goquery.Selection) {
		href, ok := selection.Attr("href")
		if !ok {
			return
		}
		tagID := extractQueryInt64(href, "tag_id")
		if tagID == 0 {
			return
		}
		if _, ok := seen[tagID]; ok {
			return
		}
		seen[tagID] = struct{}{}
		name := cleanRemoteText(cipher.decrypt(selection.Text()))
		name = countSuffixPattern.ReplaceAllString(name, "")
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		tags = append(tags, models.RemoteTagRef{ID: tagID, Name: name})
	})
	return tags
}

func parseDetailDates(cipher pageCipher, document *goquery.Document) (string, string) {
	createdAt := ""
	updatedAt := ""
	document.Find("#more-information2 strong").Each(func(index int, selection *goquery.Selection) {
		label := cleanRemoteText(cipher.decrypt(selection.Text()))
		value := cleanRemoteText(cipher.decrypt(selection.NextFiltered("p").First().Text()))
		switch label {
		case "創建日期", "创建日期":
			createdAt = value
		case "最後修改", "最后修改":
			updatedAt = value
		}
	})
	return createdAt, updatedAt
}

func parseRatingMeta(raw string) (float64, int) {
	matches := ratingPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(matches) != 3 {
		return 0, 0
	}
	rating, _ := strconv.ParseFloat(matches[1], 64)
	ratingCount, _ := strconv.Atoi(matches[2])
	return rating, ratingCount
}

func extractTrailingInt(raw string) int {
	matches := trailingCountPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(matches) != 2 {
		return 0
	}
	value, _ := strconv.Atoi(matches[1])
	return value
}

func extractReaderBaseImageURL(html string) string {
	matches := readerBasePattern.FindStringSubmatch(html)
	if len(matches) != 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

func extractReaderImageItems(html string) ([]readerImageItem, error) {
	matches := readerListPattern.FindStringSubmatch(html)
	if len(matches) != 2 {
		return nil, fmt.Errorf("reader image list not found")
	}
	items := []readerImageItem{}
	if err := json.Unmarshal([]byte(matches[1]), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func extractQueryInt64(raw string, key string) int64 {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	value := parsed.Query().Get(key)
	result, _ := strconv.ParseInt(value, 10, 64)
	return result
}

func resolveMaybeURL(baseURL string, raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err == nil && parsed.IsAbs() {
		return parsed.String()
	}
	base, err := url.Parse(normalizeURL(baseURL))
	if err != nil {
		return value
	}
	reference, err := url.Parse(value)
	if err != nil {
		return value
	}
	return base.ResolveReference(reference).String()
}

func firstNonEmptyAttr(selection *goquery.Selection, names ...string) string {
	for _, name := range names {
		if value, ok := selection.Attr(name); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func cleanRemoteText(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = strings.Join(strings.Fields(value), " ")
	return value
}

type cbcDecrypter struct {
	block cipher.Block
	iv    []byte
}

func NewCBCDecrypter(block cipher.Block, iv []byte) *cbcDecrypter {
	copiedIV := make([]byte, len(iv))
	copy(copiedIV, iv)
	return &cbcDecrypter{block: block, iv: copiedIV}
}

func (decrypter *cbcDecrypter) CryptBlocks(dst []byte, src []byte) {
	blockSize := decrypter.block.BlockSize()
	previous := make([]byte, blockSize)
	copy(previous, decrypter.iv)
	for offset := 0; offset < len(src); offset += blockSize {
		decrypter.block.Decrypt(dst[offset:offset+blockSize], src[offset:offset+blockSize])
		for index := 0; index < blockSize; index++ {
			dst[offset+index] ^= previous[index]
		}
		copy(previous, src[offset:offset+blockSize])
	}
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty plaintext")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > len(data) {
		return nil, fmt.Errorf("invalid padding")
	}
	for _, item := range data[len(data)-padding:] {
		if int(item) != padding {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return bytes.Clone(data[:len(data)-padding]), nil
}
