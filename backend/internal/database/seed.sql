INSERT OR IGNORE INTO authors (id, external_id, name, normalized_name) VALUES
  (1, 101, '示例作者一', 'shilizuozheyi'),
  (2, 102, '示例作者二', 'shilizuozheer');

INSERT OR IGNORE INTO tags (id, name) VALUES
  (1, '示例标签'),
  (2, '离线阅读'),
  (3, '长图流');

INSERT OR IGNORE INTO comics (id, title, subtitle, cover_url, cover_local_rel_path, rating, rating_count, favorites_remote, category_id, category_name) VALUES
  (10001, '示例漫画：雨后书屋', '用于开发链路联调的示例数据', '', '', 8.6, 112, 56, 1, '單行本'),
  (10002, '示例漫画：夏夜电波', '演示详情与阅读页结构', '', '', 8.2, 98, 44, 2, '同人誌'),
  (10003, '示例漫画：白昼余温', '用于缓存和阅读链路占位', '', '', 7.9, 76, 31, 3, '雜誌短篇/彩頁');

INSERT OR IGNORE INTO comic_authors (comic_id, author_id, position, source_name) VALUES
  (10001, 1, 0, '示例作者一'),
  (10002, 2, 0, '示例作者二'),
  (10003, 1, 0, '示例作者一'),
  (10003, 2, 1, '示例作者二');

INSERT OR IGNORE INTO comic_tags (comic_id, tag_id) VALUES
  (10001, 1),
  (10001, 2),
  (10002, 1),
  (10002, 3),
  (10003, 2),
  (10003, 3);

INSERT OR IGNORE INTO comic_order_index (comic_id, sort_key, source, remote_page, remote_pos) VALUES
  (10001, 1.0, 'bootstrap', 1, 1),
  (10002, 2.0, 'bootstrap', 1, 2),
  (10003, 3.0, 'bootstrap', 1, 3);

INSERT OR IGNORE INTO comic_cache_state (comic_id, meta_level, cover_ready, images_total, images_local) VALUES
  (10001, 2, 1, 5, 2),
  (10002, 2, 0, 4, 0),
  (10003, 2, 1, 6, 6);

INSERT OR IGNORE INTO comic_images (comic_id, sort, image_url, extension, local_rel_path, file_size, cached_at) VALUES
  (10001, 0, '', 'jpg', 'demo/10001/000.jpg', 102400, datetime('now')),
  (10001, 1, '', 'jpg', 'demo/10001/001.jpg', 102400, datetime('now')),
  (10001, 2, '', 'jpg', '', 0, NULL),
  (10001, 3, '', 'jpg', '', 0, NULL),
  (10001, 4, '', 'jpg', '', 0, NULL),
  (10002, 0, '', 'jpg', '', 0, NULL),
  (10002, 1, '', 'jpg', '', 0, NULL),
  (10002, 2, '', 'jpg', '', 0, NULL),
  (10002, 3, '', 'jpg', '', 0, NULL),
  (10003, 0, '', 'jpg', 'demo/10003/000.jpg', 102400, datetime('now')),
  (10003, 1, '', 'jpg', 'demo/10003/001.jpg', 102400, datetime('now')),
  (10003, 2, '', 'jpg', 'demo/10003/002.jpg', 102400, datetime('now')),
  (10003, 3, '', 'jpg', 'demo/10003/003.jpg', 102400, datetime('now')),
  (10003, 4, '', 'jpg', 'demo/10003/004.jpg', 102400, datetime('now')),
  (10003, 5, '', 'jpg', 'demo/10003/005.jpg', 102400, datetime('now'));

INSERT OR IGNORE INTO catalog_snapshots (id, source_id, total_comics, total_pages, last_page_count, captured_at) VALUES
  (1, NULL, 3, 1, 3, datetime('now'));
