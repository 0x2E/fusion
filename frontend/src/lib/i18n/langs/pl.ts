const lang = {
	// actions
	'common.submit': 'Wyslij',
	'common.cancel': 'Anuluj',
	'common.confirm': 'Potwierdź',
	'common.add': 'Dodaj',
	'common.delete': 'Usuń',
	'common.edit': 'Edytuj',
	'common.save': 'Zapisz',
	'common.close': 'Zamknij',
	'common.search': 'Szukaj',
	'common.login': 'Zaloguj się',
	'common.logout': 'Wyloguj się',

	// common words
	'common.unread': 'Nieprzeczytane',
	'common.bookmark': 'Zakładki',
	'common.all': 'Wszystkie',
	'common.feeds': 'Kanały',
	'common.group': 'Grupa',
	'common.groups': 'Grupy',
	'common.settings': 'Ustawienia',
	'common.name': 'Login',
	'common.password': 'Hasło',
	'common.link': 'Link',
	'common.advanced': 'Zaawansowane',
	'common.shortcuts': 'Skróty klawiaturowe',
	'common.more': 'Więcej',
	'common.current_page': 'Bieżąca strona',

	// state
	'state.success': 'Sukces',
	'state.error': 'Błąd',
	'state.loading': 'Wczytywanie',
	'state.no_data': 'Brak danych',
	'state.no_more_data': 'Nie ma więcej danych',

	// feed
	'feed.refresh': 'Odśwież kanał',
	'feed.refresh.all': 'Odśwież wszystkie kanały',
	'feed.refresh.all.confirm':
		'Czy na pewno chcesz odświeżyć wszystkie kanały, z wyjątkiem zawieszonych?',
	'feed.refresh.all.run_in_background': 'Rozpocznij odświeżanie w tle',
	'feed.refresh.resume': 'Wznów odświeżanie',
	'feed.refresh.suspend': 'Zatzymaj odświeżanie',
	'feed.delete.confirm': 'Czy na pewno chcesz usunąc ten kanał?',
	'feed.banner.suspended': 'Odświeżanie tego kanału zostało zawieszone',
	'feed.banner.failed': 'Nie udało się odświeżyć kanału. Błąd: {error}',

	'feed.import.title': 'Dodaj kanały',
	'feed.import.manually': 'Ręcznie',
	'feed.import.manually.link.description':
		'Link do strony lub kanału RSS. Server automatiycznie spróbuje zlokalizować RSS. Istniejący kanał z tym samym linkiem zostanie nadpisany.',
	'feed.import.manually.name.description':
		'Opcjopnalne. Zostaw puste, aby przydzielić nazwę automatycznie.',
	'feed.import.manually.no_valid_feed_error':
		'Nie znaleziono poprawnych kanałów. Sprawdź poprawność linku, lub prześlij bezpośredni link do kanału.',
	'feed.import.manually.link_candidates.label': 'Wybierz link',
	'feed.import.opml': 'Zaimportuj OPML',
	'feed.import.opml.file.label': 'Wybierz plik OPML',
	'feed.import.opml.file.description':
		'Plik powinien być w formacie {opml}. Możesz wyeskportować go z poprzedniego czytnika RSS.',
	'feed.import.opml.file_read_error': 'Nie udało się wczytać pliku',
	'feed.import.opml.how_it_works.title': 'Jak to działa?',
	'feed.import.opml.how_it_works.description.1':
		'Kanały zostaną przypisane do odpoiwiedniej grupy, która zostanie stworzona automatycznie o ile nie istnieje.',
	'feed.import.opml.how_it_works.description.2':
		"Wielowymiarowa grupa zostanie zamieniona na jednowymiarową, użuywając konwencji 'a/b/c'.",
	'feed.import.opml.how_it_works.description.3':
		'Istniejące kanały z tym samym linkiem zostaną nadpisane.',

	// item
	'item.search.placeholder': 'Szukaj w tytule i treści',
	'item.mark_all_as_read': 'Oznacz wszystkie jako przeczytane',
	'item.mark_as_read': 'Oznacz jako przeczytane',
	'item.mark_as_unread': 'Oznacz jako nieprzeczytane',
	'item.add_to_bookmark': 'Dodaj do zakładek',
	'item.remove_from_bookmark': 'Usuń z zakładek',
	'item.goto_feed': 'Idź do kanału',
	'item.visit_the_original': 'Odwiedź link źródłowy',
	'item.share': 'Udostępnij',

	// settings
	'settings.appearance': 'Wygląd',
	'settings.appearance.description': 'Te ustawienia są przechowywane w Twojej przeglądarce.',
	'settings.appearance.field.language.label': 'Język',

	'settings.global_actions': 'Akcje globalne',
	'settings.global_actions.refresh_all_feeds': 'Odśwież wszystkie kanały',
	'settings.global_actions.export_all_feeds': 'Eksportuj wszystkie kanały',

	'settings.groups.description': 'Nazwa grupy powinna być unikalna',
	'settings.groups.delete.confirm':
		'Czy na pewno chcesz usunąć tę grupę? Wszystkie kanały, które zawiera, zostaną przeniesione do grupy domyślnej.',
	'settings.groups.delete.error.delete_the_default': 'Nie można usunąć domyślnej grupy',

	// auth
	'auth.logout.confirm': 'Czy na pewno chcesz się wylogować?',
	'auth.logout.failed_message': 'Logowanie nie powiodło się. Spróbój ponownie.',

	// shortcuts
	'shortcuts.show_help': 'Pokaż skróty klawiaturowe',
	'shortcuts.next_item': 'Następna pozycja',
	'shortcuts.prev_item': 'Poprzednia pozycja',
	'shortcuts.toggle_unread': 'Oznacz jako przeczytany/nieprzeczytany',
	'shortcuts.toggle_bookmark': 'Dodaj lub usuń z zakładek',
	'shortcuts.view_original': 'Zobacz stronę źródłową',
	'shortcuts.next_feed': 'Następny kanał',
	'shortcuts.prev_feed': 'Poprzedni kanał',
	'shortcuts.open_selected': 'Otwórz zaznaczenie',
	'shortcuts.goto_search_page': 'Szukaj…',
	'shortcuts.goto_unread_page': 'Idź do nieprzeczytanych',
	'shortcuts.goto_bookmarks_page': 'Idź do zakładek',
	'shortcuts.goto_all_items_page': 'Idź do wszystkich pozycji',
	'shortcuts.goto_feeds_page': 'Idź do kanałów',
	'shortcuts.goto_settings_page': 'Idź do ustawień'
} as const;

export default lang;
