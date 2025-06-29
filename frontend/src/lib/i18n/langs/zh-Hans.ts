const lang = {
	// actions
	'common.submit': '提交',
	'common.cancel': '取消',
	'common.confirm': '确认',
	'common.add': '添加',
	'common.delete': '删除',
	'common.edit': '编辑',
	'common.save': '保存',
	'common.close': '关闭',
	'common.search': '搜索',
	'common.login': '登录',
	'common.logout': '退出登录',

	// common words
	'common.unread': '未读',
	'common.bookmark': '书签',
	'common.all': '全部',
	'common.feeds': '订阅源',
	'common.group': '分组',
	'common.groups': '分组',
	'common.settings': '设置',
	'common.name': '名称',
	'common.password': '密码',
	'common.link': '链接',
	'common.advanced': '高级',
	'common.shortcuts': '键盘快捷键',
	'common.more': '更多',
	'common.current_page': '当前页面',

	// state
	'state.success': '成功',
	'state.error': '错误',
	'state.loading': '加载中',
	'state.no_data': '暂无数据',
	'state.no_more_data': '暂无更多数据',

	// feed
	'feed.refresh': '刷新订阅源',
	'feed.refresh.all': '刷新所有订阅源',
	'feed.refresh.all.confirm': '确定要刷新除已暂停外的所有订阅源吗？',
	'feed.refresh.all.run_in_background': '在后台开始刷新',
	'feed.refresh.resume': '恢复刷新',
	'feed.refresh.suspend': '暂停刷新',
	'feed.delete.confirm': '确定要删除此订阅源吗？',
	'feed.banner.suspended': '此订阅源已暂停刷新',
	'feed.banner.failed': '刷新订阅源时失败。错误：{error}',

	'feed.import.title': '添加订阅源',
	'feed.import.manually': '手动添加',
	'feed.import.manually.link.description':
		'可以是订阅源链接或网站链接。服务器将自动尝试定位订阅源。相同链接的现有订阅源将被覆盖。',
	'feed.import.manually.name.description': '可选。留空将自动命名。',
	'feed.import.manually.no_valid_feed_error':
		'找不到有效的订阅源。请检查链接，或直接提交订阅源链接。',
	'feed.import.manually.link_candidates.label': '选择一个链接',
	'feed.import.opml': '导入 OPML',
	'feed.import.opml.file.label': '选择 OPML 文件',
	'feed.import.opml.file.description':
		'文件应为 {opml} 格式。您可以从之前的 RSS 阅读器获取此类文件。',
	'feed.import.opml.file_read_error': '加载文件内容失败',
	'feed.import.opml.how_it_works.title': '工作原理？',
	'feed.import.opml.how_it_works.description.1':
		'订阅源将被导入到相应的分组中，如果该分组不存在，将自动创建。',
	'feed.import.opml.how_it_works.description.2':
		'多维分组将被扁平化为一维结构，使用如 "a/b/c" 的命名约定。',
	'feed.import.opml.how_it_works.description.3': '相同链接的现有订阅源将被覆盖。',

	// item
	'item.search.placeholder': '搜索标题和内容',
	'item.mark_all_as_read': '标记所有为已读',
	'item.mark_as_read': '标记为已读',
	'item.mark_as_unread': '标记为未读',
	'item.add_to_bookmark': '添加到书签',
	'item.remove_from_bookmark': '从书签中移除',
	'item.goto_feed': '前往订阅源',
	'item.visit_the_original': '访问原始链接',
	'item.share': '分享',

	// settings
	'settings.appearance': '外观',
	'settings.appearance.description': '这些设置存储在您的浏览器中。',
	'settings.appearance.field.language.label': '语言',

	'settings.global_actions': '全局操作',
	'settings.global_actions.refresh_all_feeds': '刷新所有订阅源',
	'settings.global_actions.export_all_feeds': '导出所有订阅源',

	'settings.groups.description': '分组名称必须唯一。',
	'settings.groups.delete.confirm': '确定要删除此分组吗？其中的所有订阅源将被移至默认分组',
	'settings.groups.delete.error.delete_the_default': '无法删除默认分组',

	// auth
	'auth.logout.confirm': '确定要退出登录吗？',
	'auth.logout.failed_message': '退出登录失败。请重试。',

	// shortcuts
	'shortcuts.show_help': '显示键盘快捷键',
	'shortcuts.next_item': '下一项',
	'shortcuts.prev_item': '上一项',
	'shortcuts.toggle_unread': '切换已读/未读',
	'shortcuts.toggle_bookmark': '切换书签',
	'shortcuts.view_original': '查看原始链接',
	'shortcuts.next_feed': '下一个订阅源',
	'shortcuts.prev_feed': '上一个订阅源',
	'shortcuts.open_selected': '打开选择项',
	'shortcuts.goto_search_page': '前往搜索',
	'shortcuts.goto_unread_page': '前往未读',
	'shortcuts.goto_bookmarks_page': '前往书签',
	'shortcuts.goto_all_items_page': '前往所有项目',
	'shortcuts.goto_feeds_page': '前往订阅源',
	'shortcuts.goto_settings_page': '前往设置'
} as const;

export default lang;
