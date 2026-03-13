package parser

import (
	"comic-viewer-claude/internal/crypto"
	"comic-viewer-claude/internal/model"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParseListPage 解析列表页
func ParseListPage(html string) ([]model.ComicListItem, int, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, 0, fmt.Errorf("解析HTML失败: %w", err)
	}

	// 提取密钥
	iv, key, err := crypto.ExtractKeys(html)
	if err != nil {
		// 密钥提取失败，使用空密钥（降级处理）
		iv, key = "", ""
	}

	var items []model.ComicListItem

	// 解析漫画列表
	doc.Find("div.image[id^='ads-group-']").Each(func(i int, s *goquery.Selection) {
		item := model.ComicListItem{}

		// ID 和封面
		link := s.Find("div.image-inner a.apo")
		if href, exists := link.Attr("href"); exists {
			re := regexp.MustCompile(`ID=(\d+)`)
			if matches := re.FindStringSubmatch(href); len(matches) > 1 {
				item.ID, _ = strconv.Atoi(matches[1])
			}
		}

		// 封面URL
		img := s.Find("img.lazyload")
		if src, exists := img.Attr("data-src"); exists {
			item.CoverURL = src
		}

	// 标题（加密）
		titleLink := s.Find("h5.title a.d")
		if titleLink.Length() > 0 {
			encTitle := strings.TrimSpace(titleLink.Text())
			if key != "" && iv != "" {
				item.Title, _ = crypto.Decrypt(encTitle, key, iv)
			} else {
				item.Title = encTitle
			}
		}

		// 收藏数
		favSpan := s.Find("div.pull-right small")
		if favSpan.Length() > 0 {
			favText := strings.TrimSpace(favSpan.Text())
			re := regexp.MustCompile(`(\d+)`)
			if matches := re.FindStringSubmatch(favText); len(matches) > 1 {
				item.Favorites, _ = strconv.Atoi(matches[1])
			}
		}

		// 评分
		ratingSpan := s.Find("div.rating span[style]")
		if ratingSpan.Length() > 0 {
			ratingText := strings.TrimSpace(ratingSpan.Text())
			re := regexp.MustCompile(`([\d.]+)\((\d+)\)`)
			if matches := re.FindStringSubmatch(ratingText); len(matches) > 2 {
				item.Rating, _ = strconv.ParseFloat(matches[1], 64)
				item.RatingCount, _ = strconv.Atoi(matches[2])
			}
		}

		// 作者
		authorLink := s.Find("a[href*='author_id=']")
		if authorLink.Length() > 0 {
			authorText := strings.TrimSpace(authorLink.Text())
			if key != "" && iv != "" && authorText != "" {
				item.Author, _ = crypto.Decrypt(authorText, key, iv)
			} else {
				item.Author = authorText
			}

			if href, exists := authorLink.Attr("href"); exists {
				re := regexp.MustCompile(`author_id=(\d+)`)
				if matches := re.FindStringSubmatch(href); len(matches) > 1 {
					item.AuthorID, _ = strconv.Atoi(matches[1])
				}
			}
		}

		if item.ID > 0 {
			items = append(items, item)
		}
	})

	// 解析总页数
	totalPages := 1
	lastPageLink := doc.Find("ul.pagination li:last-child a")
	if href, exists := lastPageLink.Attr("href"); exists {
		re := regexp.MustCompile(`page=(\d+)`)
		if matches := re.FindStringSubmatch(href); len(matches) > 1 {
			totalPages, _ = strconv.Atoi(matches[1])
		}
	}

	return items, totalPages, nil
}

// ParseDetailPage 解析详情页
func ParseDetailPage(html string, comicID int) (*model.ComicDetail, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %w", err)
	}

	// 提取密钥
	iv, key, err := crypto.ExtractKeys(html)
	if err != nil {
		iv, key = "", ""
	}

	detail := &model.ComicDetail{ID: comicID}

	// 封面
	coverImg := doc.Find("div.product-image img")
	if src, exists := coverImg.Attr("src"); exists {
		detail.CoverURL = src
	}

	// 标题（加密）
	titles := doc.Find("div.name h2.d")
	if titles.Length() > 0 {
		titleText := strings.TrimSpace(titles.Eq(0).Text())
		if key != "" && iv != "" && titleText != "" {
			detail.Title, _ = crypto.Decrypt(titleText, key, iv)
		} else {
			detail.Title = titleText
		}
	}
	if titles.Length() > 1 {
	subtitleText := strings.TrimSpace(titles.Eq(1).Text())
		if key != "" && iv != "" && subtitleText != "" {
			detail.Subtitle, _ = crypto.Decrypt(subtitleText, key, iv)
		} else {
			detail.Subtitle = subtitleText
		}
	}

	// 作者（支持多个）
	seenAuthors := make(map[string]bool)
	doc.Find("a.apo.btn.d[href*='author_id']").Each(func(i int, s *goquery.Selection) {
		rawAuthorName := strings.TrimSpace(s.Text())
		authorName := rawAuthorName
		if key != "" && iv != "" && rawAuthorName != "" {
			authorName, _ = crypto.Decrypt(rawAuthorName, key, iv)
		}

		var authorID int
		if href, exists := s.Attr("href"); exists {
			re := regexp.MustCompile(`author_id=(\d+)`)
			if matches := re.FindStringSubmatch(href); len(matches) > 1 {
				authorID, _ = strconv.Atoi(matches[1])
			}
		}

		authorKey := fmt.Sprintf("%d_%s", authorID, authorName)
		if authorName != "" && !seenAuthors[authorKey] {
			detail.Authors = append(detail.Authors, model.AuthorInfo{
				AuthorID:   authorID,
				AuthorName: authorName,
			})
			seenAuthors[authorKey] = true
		}
	})

	// 设置主作者
	if len(detail.Authors) > 0 {
		var authorNames []string
		for _, author := range detail.Authors {
			authorNames = append(authorNames, author.AuthorName)
			if author.AuthorID > 0 && detail.AuthorID == 0 {
				detail.AuthorID = author.AuthorID
			}
		}
		detail.Author = strings.Join(authorNames, " / ")
	}

	// 评分
	ratingSpan := doc.Find("span.fa.fa-2x span")
	if ratingSpan.Length() > 0 {
		ratingText := strings.TrimSpace(ratingSpan.Text())
		re := regexp.MustCompile(`([\d.]+)\((\d+)\)`)
		if matches := re.FindStringSubmatch(ratingText); len(matches) > 2 {
			detail.Rating, _ = strconv.ParseFloat(matches[1], 64)
			detail.RatingCount, _ = strconv.Atoi(matches[2])
		}
	}

	// 标签 - 从 comic.addTag/delTag 提取
	tagRe := regexp.MustCompile(`comic\.(?:add|del)Tag\(\s*\d+\s*,\s*(\d+)\s*,\s*'([^']+)'`)
	tagMatches := tagRe.FindAllStringSubmatch(html, -1)
	seenTags := make(map[int]bool)
	for _, match := range tagMatches {
		if len(match) > 2 {
			tagID, _ := strconv.Atoi(match[1])
			tagName := match[2]
			if !seenTags[tagID] {
				detail.Tags = append(detail.Tags, model.TagInfo{
					TagID:   tagID,
					TagName: tagName,
				})
				seenTags[tagID] = true
			}
		}
	}

	// 分类
	catRe := regexp.MustCompile(`comic\.addCategory\(\s*\d+\s*,\s*(\d+)\s*,\s*'([^']+)'`)
	if catMatch := catRe.FindStringSubmatch(html); len(catMatch) > 2 {
		detail.CategoryID, _ = strconv.Atoi(catMatch[1])
		detail.CategoryName = catMatch[2]
	}

	// 日期
	dateRe := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})`)
	dates := dateRe.FindAllString(html, -1)
	if len(dates) > 0 {
		detail.CreatedAt = dates[0]
	}
	if len(dates) > 1 {
		detail.UpdatedAt = dates[1]
	}

	// 阅读链接
	readLink := doc.Find("a[href*='readOnline2.php']")
	if href, exists := readLink.Attr("href"); exists {
		detail.ReaderURL = href
	}

	return detail, nil
}

// ParseReaderPage 解析阅读页
func ParseReaderPage(html string, comicID int) (string, []model.ImageInfo, error) {
	// 提取 HTTP_IMAGE
	httpImageRe := regexp.MustCompile(`var\s+HTTP_IMAGE\s*=\s*"([^"]+)"`)
	httpImageMatch := httpImageRe.FindStringSubmatch(html)
	if len(httpImageMatch) < 2 {
		return "", nil, fmt.Errorf("无法提取 HTTP_IMAGE")
	}
	httpImage := httpImageMatch[1]

	// 提取 Original_Image_List JSON
	imgListRe := regexp.MustCompile(`Original_Image_List\s*=\s*(\[.*?\]);`)
	imgListMatch := imgListRe.FindStringSubmatch(html)
	if len(imgListMatch) < 2 {
		return "", nil, fmt.Errorf("无法提取 Original_Image_List")
	}

	// 解析JSON
	var imageData []map[string]interface{}
	if err := json.Unmarshal([]byte(imgListMatch[1]), &imageData); err != nil {
		return "", nil, fmt.Errorf("解析图片列表失败: %w", err)
	}

	var images []model.ImageInfo
	for _, item := range imageData {
		sort, _ := item["sort"].(float64)
		newFilename, _ := item["new_filename"].(string)
		extension, _ := item["extension"].(string)

		img := model.ImageInfo{
		Sort:      int(sort),
			ComicID:   comicID,
			Filename:  newFilename,
			Extension: extension,
			URL:       fmt.Sprintf("%s%s_w900.%s", httpImage, newFilename, extension),
		}
		images = append(images, img)
	}

	return httpImage, images, nil
}

// ParseTagsFromPage 从页面解析标签列表
func ParseTagsFromPage(html string) ([]model.TagInfo, error) {
	var tags []model.TagInfo
	seenTags := make(map[int]bool)

	// 从 comic.addTag/delTag 提取明文标签
	tagRe := regexp.MustCompile(`comic\.(?:add|del)Tag\(\s*\d+\s*,\s*(\d+)\s*,\s*'([^']+)'`)
	tagMatches := tagRe.FindAllStringSubmatch(html, -1)
	for _, match := range tagMatches {
		if len(match) > 2 {
			tagID, _ := strconv.Atoi(match[1])
			tagName := match[2]
			if !seenTags[tagID] {
				tags = append(tags, model.TagInfo{
					TagID:   tagID,
					TagName: tagName,
				})
				seenTags[tagID] = true
			}
		}
	}

	// 如果没有找到，尝试从链接提取（加密文本）
	if len(tags) == 0 {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			return tags, nil
		}

		iv, key, err := crypto.ExtractKeys(html)
		if err != nil {
			return tags, nil
		}

		doc.Find("a[href*='tag_id=']").Each(func(i int, s *goquery.Selection) {
			if href, exists := s.Attr("href"); exists {
				re := regexp.MustCompile(`tag_id=(\d+)`)
				if matches := re.FindStringSubmatch(href); len(matches) > 1 {
					tagID, _ := strconv.Atoi(matches[1])
					if !seenTags[tagID] {
				encName := strings.TrimSpace(s.Text())
						tagName, _ := crypto.Decrypt(encName, key, iv)
						// 去掉可能的尾部数字（投票数）
						tagName = regexp.MustCompile(`\(\d+\)$`).ReplaceAllString(tagName, "")
			tagName = strings.TrimSpace(tagName)
						if tagName != "" {
							tags = append(tags, model.TagInfo{
								TagID:   tagID,
								TagName: tagName,
							})
							seenTags[tagID] = true
						}
					}
				}
			}
		})
	}

	return tags, nil
}
