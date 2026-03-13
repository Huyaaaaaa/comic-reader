package repository

import (
	"comic-viewer-claude/internal/model"
)

// AuthorBrief 作者摘要
type AuthorBrief struct {
	AuthorID   int    `json:"author_id"`
	AuthorName string `json:"author_name"`
	ComicCount int    `json:"comic_count"`
}

// GetDistinctAuthors 获取去重的作者列表（带作品数）
func (r *Repository) GetDistinctAuthors() ([]AuthorBrief, error) {
	var authors []AuthorBrief
	err := r.db.Model(&model.ComicAuthor{}).
		Select("author_id, author_name, COUNT(DISTINCT comic_id) as comic_count").
		Group("author_id, author_name").
		Order("comic_count DESC").
		Find(&authors).Error
	return authors, err
}

// GetComicsByAuthorID 获取同作者的漫画列表
func (r *Repository) GetComicsByAuthorID(authorID int, page, pageSize int) ([]model.Comic, int64, error) {
	var comicIDs []int
	r.db.Model(&model.ComicAuthor{}).
		Where("author_id = ?", authorID).
		Pluck("comic_id", &comicIDs)

	var total int64
	r.db.Model(&model.Comic{}).Where("id IN ?", comicIDs).Count(&total)

	var comics []model.Comic
	err := r.db.Where("id IN ?", comicIDs).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&comics).Error
	return comics, total, err
}

// CategoryBrief 分类摘要
type CategoryBrief struct {
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
	ComicCount   int    `json:"comic_count"`
}

// GetCategories 获取分类列表（带作品数）
func (r *Repository) GetCategories() ([]CategoryBrief, error) {
	var categories []CategoryBrief
	err := r.db.Model(&model.Comic{}).
		Select("category_id, category_name, COUNT(*) as comic_count").
		Where("category_name != ''").
		Group("category_id, category_name").
		Order("comic_count DESC").
		Find(&categories).Error
	return categories, err
}
