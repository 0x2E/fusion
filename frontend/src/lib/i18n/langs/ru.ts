const lang = {
	// actions
	'common.submit': 'Отправить',
	'common.cancel': 'Отмена',
	'common.confirm': 'Подтвердить',
	'common.add': 'Добавить',
	'common.delete': 'Удалить',
	'common.edit': 'Редактировать',
	'common.save': 'Сохранить',
	'common.close': 'Закрыть',
	'common.search': 'Поиск',
	'common.login': 'Войти',
	'common.logout': 'Выйти',

	// common words
	'common.unread': 'Непрочитанные',
	'common.bookmark': 'Закладка',
	'common.all': 'Все',
	'common.feeds': 'Ленты',
	'common.group': 'Группа',
	'common.groups': 'Группы',
	'common.settings': 'Настройки',
	'common.name': 'Имя',
	'common.password': 'Пароль',
	'common.link': 'Ссылка',
	'common.advanced': 'Дополнительно',
	'common.shortcuts': 'Горячие клавиши',
	'common.more': 'Ещё',
	'common.current_page': 'Текущая страница',

	// state
	'state.success': 'Успешно',
	'state.error': 'Ошибка',
	'state.loading': 'Загрузка',
	'state.no_data': 'Нет данных',
	'state.no_more_data': 'Больше нет данных',

	// feed
	'feed.refresh': 'Обновить ленту',
	'feed.refresh.all': 'Обновить все ленты',
	'feed.refresh.all.confirm': 'Вы уверены, что хотите обновить все ленты, кроме приостановленных?',
	'feed.refresh.all.run_in_background': 'Начать обновление в фоновом режиме',
	'feed.refresh.resume': 'Возобновить обновление',
	'feed.refresh.suspend': 'Приостановить обновление',
	'feed.delete.confirm': 'Вы уверены, что хотите удалить эту ленту?',
	'feed.banner.suspended': 'Эта лента приостановлена',
	'feed.banner.failed': 'Не удалось обновить ленту. Ошибка: {error}',

	'feed.import.title': 'Добавить ленты',
	'feed.import.manually': 'Вручную',
	'feed.import.manually.link.description':
		'Ссылка RSS или ссылка на сайт. Сервер автоматически попытается найти RSS-ленту. Существующая лента с такой же ссылкой будет перезаписана.',
	'feed.import.manually.name.description':
		'Опционально. Оставьте пустым для автоматического именования.',
	'feed.import.manually.no_valid_feed_error':
		'Действительный канал не найден. Пожалуйста, проверьте ссылку или отправьте ссылку на канал напрямую.',
	'feed.import.manually.link_candidates.label': 'Выберите ссылку',
	'feed.import.opml': 'Импорт OPML',
	'feed.import.opml.file.label': 'Выберите файл OPML',
	'feed.import.opml.file.description':
		'Файл должен быть в формате {opml}. Вы можете получить его из предыдущего RSS-читателя.',
	'feed.import.opml.file_read_error': 'Не удалось загрузить содержимое файла',
	'feed.import.opml.how_it_works.title': 'Как это работает?',
	'feed.import.opml.how_it_works.description.1':
		'Ленты будут импортированы в соответствующую группу, которая будет создана автоматически, если ее не существует.',
	'feed.import.opml.how_it_works.description.2':
		"Многомерные группы будут преобразованы в одномерную структуру с использованием соглашения об именовании, например 'a/b/c'.",
	'feed.import.opml.how_it_works.description.3':
		'Существующая лента с такой же ссылкой будет перезаписана.',

	// item
	'item.search.placeholder': 'Поиск в заголовке и содержимом',
	'item.mark_all_as_read': 'Отметить все как прочитанные',
	'item.mark_as_read': 'Отметить как прочитанное',
	'item.mark_as_unread': 'Отметить как непрочитанное',
	'item.add_to_bookmark': 'Добавить в закладки',
	'item.remove_from_bookmark': 'Удалить из закладок',
	'item.goto_feed': 'Перейти к ленте',
	'item.visit_the_original': 'Посетить оригинальную ссылку',
	'item.share': 'Предоставить общий доступ',

	// settings
	'settings.appearance': 'Внешний вид',
	'settings.appearance.description': 'Эти настройки сохраняются в вашем браузере.',
	'settings.appearance.field.language.label': 'Язык',

	'settings.global_actions': 'Глобальные действия',
	'settings.global_actions.refresh_all_feeds': 'Обновить все ленты',
	'settings.global_actions.export_all_feeds': 'Экспортировать все ленты',

	'settings.groups.description': 'Имя группы должно быть уникальным.',
	'settings.groups.delete.confirm':
		'Вы уверены, что хотите удалить эту группу? Все ее ленты будут перемещены в группу по умолчанию',
	'settings.groups.delete.error.delete_the_default': 'Невозможно удалить группу по умолчанию',

	// auth
	'auth.logout.confirm': 'Вы уверены, что хотите выйти?',
	'auth.logout.failed_message': 'Не удалось выйти. Пожалуйста, попробуйте еще раз.',

	// shortcuts
	'shortcuts.show_help': 'Показать горячие клавиши',
	'shortcuts.next_item': 'Следующий элемент',
	'shortcuts.prev_item': 'Предыдущий элемент',
	'shortcuts.toggle_unread': 'Переключить прочитано/непрочитано',
	'shortcuts.toggle_bookmark': 'Переключить закладку',
	'shortcuts.view_original': 'Просмотреть оригинальную ссылку',
	'shortcuts.next_feed': 'Следующая лента',
	'shortcuts.prev_feed': 'Предыдущая лента',
	'shortcuts.open_selected': 'Открыть выбранное',
	'shortcuts.goto_search_page': 'Перейти к поиску',
	'shortcuts.goto_unread_page': 'Перейти к непрочитанным',
	'shortcuts.goto_bookmarks_page': 'Перейти к закладкам',
	'shortcuts.goto_all_items_page': 'Перейти ко всем элементам',
	'shortcuts.goto_feeds_page': 'Перейти к лентам',
	'shortcuts.goto_settings_page': 'Перейти к настройкам'
} as const;

export default lang;
