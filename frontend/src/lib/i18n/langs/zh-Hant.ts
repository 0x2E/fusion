const lang = {
	// actions
	'common.submit': '提交',
	'common.cancel': '取消',
	'common.confirm': '確認',
	'common.add': '新增',
	'common.delete': '刪除',
	'common.edit': '編輯',
	'common.save': '儲存',
	'common.close': '關閉',
	'common.search': '搜尋',
	'common.login': '登入',
	'common.logout': '登出',

	// common words
	'common.unread': '未讀',
	'common.bookmark': '書籤',
	'common.all': '全部',
	'common.feeds': '訂閱源',
	'common.group': '群組',
	'common.groups': '群組',
	'common.settings': '設定',
	'common.name': '名稱',
	'common.password': '密碼',
	'common.link': '連結',
	'common.advanced': '進階',
	'common.shortcuts': '鍵盤快捷鍵',
	'common.more': '更多',
	'common.current_page': '當前頁面',

	// state
	'state.success': '成功',
	'state.error': '錯誤',
	'state.loading': '載入中',
	'state.no_data': '無資料',
	'state.no_more_data': '沒有更多資料',

	// feed
	'feed.refresh': '重新整理訂閱源',
	'feed.refresh.all': '重新整理所有訂閱源',
	'feed.refresh.all.confirm': '您確定要重新整理除了已暫停的所有訂閱源嗎？',
	'feed.refresh.all.run_in_background': '在背景開始重新整理',
	'feed.refresh.resume': '恢復重新整理',
	'feed.refresh.suspend': '暫停重新整理',
	'feed.delete.confirm': '您確定要刪除此訂閱源嗎？',
	'feed.banner.suspended': '此訂閱源已被暫停',
	'feed.banner.failed': '無法重新整理訂閱源。錯誤：{error}',

	'feed.import.title': '新增訂閱源',
	'feed.import.manually': '手動新增',
	'feed.import.manually.link.description':
		'可輸入 RSS 連結或網站連結。伺服器將自動嘗試定位 RSS 訂閱源。具有相同連結的現有訂閱源將被覆蓋。',
	'feed.import.manually.name.description': '選填。留空將自動命名。',
	'feed.import.manually.no_valid_feed_error':
		'找不到有效的訂閱源。請檢查連結，或直接提交訂閱源連結。',
	'feed.import.manually.link_candidates.label': '選擇一個連結',
	'feed.import.opml': '匯入 OPML',
	'feed.import.opml.file.label': '選擇 OPML 檔案',
	'feed.import.opml.file.description': '檔案應為 {opml} 格式。您可以從先前的 RSS 閱讀器取得。',
	'feed.import.opml.file_read_error': '無法載入檔案內容',
	'feed.import.opml.how_it_works.title': '運作方式？',
	'feed.import.opml.how_it_works.description.1':
		'訂閱源將被匯入至相應的群組，如果該群組不存在，系統將自動建立。',
	'feed.import.opml.how_it_works.description.2':
		"多維群組將被扁平化為一維結構，使用類似 'a/b/c' 的命名慣例。",
	'feed.import.opml.how_it_works.description.3': '具有相同連結的現有訂閱源將被覆蓋。',

	// item
	'item.search.placeholder': '搜尋標題和內容',
	'item.mark_all_as_read': '標記全部為已讀',
	'item.mark_as_read': '標記為已讀',
	'item.mark_as_unread': '標記為未讀',
	'item.add_to_bookmark': '加入書籤',
	'item.remove_from_bookmark': '從書籤中移除',
	'item.goto_feed': '前往訂閱源',
	'item.visit_the_original': '訪問原始連結',
	'item.share': '分享',

	// settings
	'settings.appearance': '外觀',
	'settings.appearance.description': '這些設定將儲存在您的瀏覽器中。',
	'settings.appearance.field.language.label': '語言',

	'settings.global_actions': '全域操作',
	'settings.global_actions.refresh_all_feeds': '重新整理所有訂閱源',
	'settings.global_actions.export_all_feeds': '匯出所有訂閱源',

	'settings.groups.description': '群組名稱必須是唯一的。',
	'settings.groups.delete.confirm': '您確定要刪除此群組嗎？所有訂閱源將被移動到預設群組',
	'settings.groups.delete.error.delete_the_default': '無法刪除預設群組',

	// auth
	'auth.logout.confirm': '您確定要登出嗎？',
	'auth.logout.failed_message': '登出失敗。請再試一次。',

	// shortcuts
	'shortcuts.show_help': '顯示鍵盤快捷鍵',
	'shortcuts.next_item': '下一項',
	'shortcuts.prev_item': '上一項',
	'shortcuts.toggle_unread': '切換已讀/未讀',
	'shortcuts.toggle_bookmark': '切換書籤',
	'shortcuts.view_original': '查看原始連結',
	'shortcuts.next_feed': '下一個訂閱源',
	'shortcuts.prev_feed': '上一個訂閱源',
	'shortcuts.open_selected': '開啟選擇項',
	'shortcuts.goto_search_page': '前往搜尋',
	'shortcuts.goto_unread_page': '前往未讀',
	'shortcuts.goto_bookmarks_page': '前往書籤',
	'shortcuts.goto_all_items_page': '前往所有項目',
	'shortcuts.goto_feeds_page': '前往訂閱源',
	'shortcuts.goto_settings_page': '前往設定'
} as const;

export default lang;
