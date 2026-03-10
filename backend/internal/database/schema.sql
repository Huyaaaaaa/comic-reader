PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA synchronous = NORMAL;
PRAGMA busy_timeout = 5000;

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  auth_mode TEXT NOT NULL DEFAULT 'local',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
INSERT OR IGNORE INTO users (id, username, auth_mode) VALUES (1, 'local', 'local');

CREATE TABLE IF NOT EXISTS categories (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  display_order INTEGER NOT NULL DEFAULT 0
);
INSERT OR IGNORE INTO categories (id, name, display_order) VALUES
  (0, '未知', 0),
  (1, '單行本', 1),
  (2, '同人誌', 2),
  (3, '雜誌短篇/彩頁', 3),
  (4, 'CG', 4);

CREATE TABLE IF NOT EXISTS comics (
  id INTEGER PRIMARY KEY,
  title TEXT NOT NULL DEFAULT '',
  subtitle TEXT NOT NULL DEFAULT '',
  cover_url TEXT NOT NULL DEFAULT '',
  cover_local_rel_path TEXT NOT NULL DEFAULT '',
  rating REAL NOT NULL DEFAULT 0,
  rating_count INTEGER NOT NULL DEFAULT 0,
  favorites_remote INTEGER NOT NULL DEFAULT 0,
  category_id INTEGER,
  category_name TEXT NOT NULL DEFAULT '',
  source_created_at TEXT NOT NULL DEFAULT '',
  source_updated_at TEXT NOT NULL DEFAULT '',
  source_last_seen_at TEXT NOT NULL DEFAULT (datetime('now')),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (category_id) REFERENCES categories(id)
);
CREATE INDEX IF NOT EXISTS idx_comics_category ON comics(category_id);

CREATE TABLE IF NOT EXISTS authors (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  external_id INTEGER,
  name TEXT NOT NULL,
  normalized_name TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_authors_external_id ON authors(external_id);
CREATE INDEX IF NOT EXISTS idx_authors_normalized_name ON authors(normalized_name);

CREATE TABLE IF NOT EXISTS comic_authors (
  comic_id INTEGER NOT NULL,
  author_id INTEGER NOT NULL,
  position INTEGER NOT NULL DEFAULT 0,
  source_name TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (comic_id, author_id),
  FOREIGN KEY (comic_id) REFERENCES comics(id) ON DELETE CASCADE,
  FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_comic_authors_author ON comic_authors(author_id, comic_id);

CREATE TABLE IF NOT EXISTS tags (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS comic_tags (
  comic_id INTEGER NOT NULL,
  tag_id INTEGER NOT NULL,
  PRIMARY KEY (comic_id, tag_id),
  FOREIGN KEY (comic_id) REFERENCES comics(id) ON DELETE CASCADE,
  FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_comic_tags_tag_comic ON comic_tags(tag_id, comic_id);

CREATE TABLE IF NOT EXISTS comic_images (
  comic_id INTEGER NOT NULL,
  sort INTEGER NOT NULL,
  image_url TEXT NOT NULL,
  extension TEXT NOT NULL DEFAULT 'jpg',
  local_rel_path TEXT NOT NULL DEFAULT '',
  file_size INTEGER NOT NULL DEFAULT 0,
  file_sha256 TEXT,
  cached_at TEXT,
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (comic_id, sort),
  FOREIGN KEY (comic_id) REFERENCES comics(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_images_local_path ON comic_images(local_rel_path);

CREATE TABLE IF NOT EXISTS comic_order_index (
  comic_id INTEGER PRIMARY KEY,
  sort_key REAL NOT NULL,
  source TEXT NOT NULL DEFAULT 'remote',
  remote_page INTEGER,
  remote_pos INTEGER,
  order_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (comic_id) REFERENCES comics(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_order_sort_key ON comic_order_index(sort_key ASC, comic_id DESC);

CREATE TABLE IF NOT EXISTS comic_cache_state (
  comic_id INTEGER PRIMARY KEY,
  meta_level INTEGER NOT NULL DEFAULT 0 CHECK(meta_level IN (0,1,2)),
  cover_ready INTEGER NOT NULL DEFAULT 0 CHECK(cover_ready IN (0,1)),
  images_total INTEGER NOT NULL DEFAULT 0,
  images_local INTEGER NOT NULL DEFAULT 0,
  first_collected_at TEXT,
  fully_cached_at TEXT,
  last_error TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (comic_id) REFERENCES comics(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reading_history (
  user_id INTEGER NOT NULL,
  comic_id INTEGER NOT NULL,
  locator_json TEXT NOT NULL DEFAULT '{}',
  last_read_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (user_id, comic_id),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (comic_id) REFERENCES comics(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_favorites (
  user_id INTEGER NOT NULL,
  comic_id INTEGER NOT NULL,
  ensure_offline INTEGER NOT NULL DEFAULT 1 CHECK(ensure_offline IN (0,1)),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (user_id, comic_id),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (comic_id) REFERENCES comics(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS search_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL DEFAULT 1,
  keyword TEXT NOT NULL,
  searched_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_search_history_user_searched ON search_history(user_id, searched_at DESC);

CREATE TABLE IF NOT EXISTS download_tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  comic_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL DEFAULT 1,
  task_type TEXT NOT NULL DEFAULT 'images',
  trigger_source TEXT NOT NULL DEFAULT 'manual',
  status TEXT NOT NULL,
  priority INTEGER NOT NULL DEFAULT 0,
  total INTEGER NOT NULL DEFAULT 0,
  current INTEGER NOT NULL DEFAULT 0,
  retry_count INTEGER NOT NULL DEFAULT 0,
  lock_token TEXT,
  lock_acquired_at TEXT,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  finished_at TEXT,
  FOREIGN KEY (comic_id) REFERENCES comics(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_download_tasks_active ON download_tasks(user_id, comic_id, task_type) WHERE status NOT IN ('completed', 'canceled', 'failed');

CREATE TABLE IF NOT EXISTS sync_jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  job_type TEXT NOT NULL,
  status TEXT NOT NULL,
  priority INTEGER NOT NULL DEFAULT 0,
  params_json TEXT NOT NULL DEFAULT '{}',
  cursor_json TEXT NOT NULL DEFAULT '{}',
  progress REAL NOT NULL DEFAULT 0,
  last_error TEXT NOT NULL DEFAULT '',
  next_run_at TEXT,
  started_at TEXT,
  finished_at TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_sync_jobs_type_running ON sync_jobs(job_type) WHERE status IN ('queued', 'running');

CREATE TABLE IF NOT EXISTS app_settings (
  key TEXT PRIMARY KEY,
  value_json TEXT NOT NULL,
  value_type TEXT NOT NULL,
  default_value TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS update_records (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  update_type TEXT NOT NULL,
  channel TEXT NOT NULL DEFAULT 'stable',
  remote_version TEXT NOT NULL DEFAULT '',
  current_version TEXT NOT NULL DEFAULT '',
  has_update INTEGER NOT NULL DEFAULT 0,
  check_mode TEXT NOT NULL DEFAULT 'manual',
  result TEXT NOT NULL DEFAULT '',
  checked_at TEXT NOT NULL DEFAULT (datetime('now')),
  applied_at TEXT
);

CREATE TABLE IF NOT EXISTS import_export_jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  direction TEXT NOT NULL,
  scope TEXT NOT NULL,
  status TEXT NOT NULL,
  file_path TEXT NOT NULL DEFAULT '',
  options_json TEXT NOT NULL DEFAULT '{}',
  summary_json TEXT NOT NULL DEFAULT '{}',
  conflict_count INTEGER NOT NULL DEFAULT 0,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  finished_at TEXT
);

CREATE TABLE IF NOT EXISTS source_sites (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  base_url TEXT NOT NULL UNIQUE,
  navigator_url TEXT NOT NULL DEFAULT '',
  priority INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  status TEXT NOT NULL DEFAULT 'unknown',
  last_latency_ms INTEGER,
  last_checked_at TEXT,
  consecutive_failures INTEGER NOT NULL DEFAULT 0,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_source_sites_priority ON source_sites(enabled DESC, priority ASC, id ASC);

CREATE TABLE IF NOT EXISTS source_health_checks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source_id INTEGER NOT NULL,
  check_type TEXT NOT NULL DEFAULT 'heartbeat',
  status TEXT NOT NULL,
  latency_ms INTEGER,
  error_message TEXT NOT NULL DEFAULT '',
  checked_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (source_id) REFERENCES source_sites(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_source_health_source_checked ON source_health_checks(source_id, checked_at DESC);

CREATE TABLE IF NOT EXISTS catalog_snapshots (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source_id INTEGER,
  total_comics INTEGER NOT NULL DEFAULT 0,
  total_pages INTEGER NOT NULL DEFAULT 0,
  last_page_count INTEGER NOT NULL DEFAULT 0,
  captured_at TEXT NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (source_id) REFERENCES source_sites(id) ON DELETE SET NULL
);

INSERT OR IGNORE INTO app_settings (key, value_json, value_type, default_value, description) VALUES
  ('list_page_size', '100', 'number', '100', '列表每页显示数量'),
  ('reader_preload_first', '3', 'number', '3', '阅读页首屏预加载数量'),
  ('reader_preload_scroll', '5', 'number', '5', '阅读页滚动预取数量'),
  ('reader_mode', '"long_scroll"', 'string', '"long_scroll"', '阅读模式'),
  ('reader_animation', 'true', 'boolean', 'true', '是否启用翻页动画'),
  ('reader_fullscreen_last_mode', 'false', 'boolean', 'false', '上次全屏状态'),
  ('sync_head_scan_pages', '5', 'number', '5', '头部扫描页数'),
  ('l1_cache_mode', '"passive"', 'string', '"passive"', 'L1 缓存模式'),
  ('l1_passive_scope', '"list_and_detail"', 'string', '"list_and_detail"', 'L1 被动触发范围'),
  ('l1_active_count', '500', 'number', '500', 'L1 主动定量缓存数量'),
  ('l2_cache_mode', '"passive"', 'string', '"passive"', 'L2 缓存模式'),
  ('l2_active_count', '500', 'number', '500', 'L2 主动定量缓存数量'),
  ('l3_cache_mode', '"passive"', 'string', '"passive"', 'L3 缓存模式'),
  ('l3_active_count', '500', 'number', '500', 'L3 主动定量缓存数量'),
  ('cover_limit_mode', '"count"', 'string', '"count"', '封面限制模式'),
  ('cover_limit_value', '1000', 'number', '1000', '封面上限值'),
  ('cover_eviction_policy', '"LRU"', 'string', '"LRU"', '封面淘汰策略'),
  ('rate_limit_level', '"safe"', 'string', '"safe"', '防封禁档位'),
  ('content_update_mode', '"interval"', 'string', '"interval"', '内容更新检查模式'),
  ('content_update_interval', '30', 'number', '30', '内容更新检查间隔(分钟)'),
  ('app_update_mode', '"startup"', 'string', '"startup"', '应用更新检查模式'),
  ('app_update_auto_install', 'false', 'boolean', 'false', '是否自动安装应用更新'),
  ('source_heartbeat_interval_minutes', '60', 'number', '60', '源站心跳检测间隔(分钟)'),
  ('source_request_timeout_seconds', '15', 'number', '15', '单次源站请求超时时间(秒)'),
  ('source_request_retries', '3', 'number', '3', '单次源站请求重试次数'),
  ('source_failure_threshold', '3', 'number', '3', '连续失败判定不可用阈值'),
  ('last_sync_seq', '0', 'number', '0', '最后同步序列号');
