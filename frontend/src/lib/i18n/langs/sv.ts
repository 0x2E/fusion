const lang = {
	// actions
	'common.submit': 'Skicka',
	'common.cancel': 'Avbryt',
	'common.confirm': 'Bekräfta',
	'common.add': 'Lägg till',
	'common.delete': 'Ta bort',
	'common.edit': 'Redigera',
	'common.save': 'Spara',
	'common.close': 'Stäng',
	'common.search': 'Sök',
	'common.login': 'Logga in',
	'common.logout': 'Logga ut',

	// common words
	'common.unread': 'Oläst',
	'common.bookmark': 'Bokmärke',
	'common.all': 'Alla',
	'common.feeds': 'Flöden',
	'common.group': 'Grupp',
	'common.groups': 'Grupper',
	'common.settings': 'Inställningar',
	'common.name': 'Namn',
	'common.password': 'Lösenord',
	'common.link': 'Länk',
	'common.advanced': 'Avancerat',
	'common.shortcuts': 'Tangentbordsgenvägar',
	'common.more': 'Mer',
	'common.current_page': 'Aktuell sida',

	// state
	'state.success': 'Lyckades',
	'state.error': 'Fel',
	'state.loading': 'Laddar',
	'state.no_data': 'Ingen data',
	'state.no_more_data': 'Ingen mer data',

	// feed
	'feed.refresh': 'Uppdatera flöde',
	'feed.refresh.all': 'Uppdatera alla flöden',
	'feed.refresh.all.confirm':
		'Är du säker på att du vill uppdatera alla flöden förutom de pausade?',
	'feed.refresh.all.run_in_background': 'Starta uppdatering i bakgrunden',
	'feed.refresh.resume': 'Återuppta uppdatering',
	'feed.refresh.suspend': 'Pausa uppdatering',
	'feed.delete.confirm': 'Är du säker på att du vill ta bort detta flöde?',
	'feed.banner.suspended': 'Detta flöde har pausats',
	'feed.banner.failed': 'Misslyckades med att uppdatera flödet. Fel: {error}',

	'feed.import.title': 'Lägg till flöden',
	'feed.import.manually': 'Manuellt',
	'feed.import.manually.link.description':
		'Antingen RSS-länken eller webbplatslänken. Servern kommer automatiskt att försöka hitta RSS-flödet. Befintligt flöde med samma länk kommer att ersättas.',
	'feed.import.manually.name.description': 'Valfritt. Lämna tomt för automatisk namngivning.',
	'feed.import.manually.no_valid_feed_error':
		'Inget giltigt flöde hittades. Kontrollera länken eller skicka en flödeslänk direkt.',
	'feed.import.manually.link_candidates.label': 'Välj en länk',
	'feed.import.opml': 'Importera OPML',
	'feed.import.opml.file.label': 'Välj en OPML-fil',
	'feed.import.opml.file.description':
		'Filen bör vara i {opml}-format. Du kan få en från din tidigare RSS-läsare.',
	'feed.import.opml.file_read_error': 'Misslyckades med att ladda filinnehåll',
	'feed.import.opml.how_it_works.title': 'Hur fungerar det?',
	'feed.import.opml.how_it_works.description.1':
		'Flöden kommer att importeras till motsvarande grupp. Om gruppen inte finns kommer den automatiskt att skapas.',
	'feed.import.opml.how_it_works.description.2':
		"Flerdimensionella grupper kommer att planas ut till en endimensionell struktur med namngivningskonvention som 'a/b/c'.",
	'feed.import.opml.how_it_works.description.3':
		'Befintligt flöde med samma länk kommer att ersättas.',

	// item
	'item.search.placeholder': 'Sök i titel och innehåll',
	'item.mark_all_as_read': 'Markera alla som lästa',
	'item.mark_as_read': 'Markera som läst',
	'item.mark_as_unread': 'Markera som oläst',
	'item.add_to_bookmark': 'Lägg till bokmärke',
	'item.remove_from_bookmark': 'Ta bort från bokmärken',
	'item.goto_feed': 'Gå till flöde',
	'item.visit_the_original': 'Besök originallänk',
	'item.share': 'dela',

	// settings
	'settings.appearance': 'Utseende',
	'settings.appearance.description': 'Dessa inställningar sparas i din webbläsare.',
	'settings.appearance.field.language.label': 'Språk',

	'settings.global_actions': 'Globala åtgärder',
	'settings.global_actions.refresh_all_feeds': 'Uppdatera alla flöden',
	'settings.global_actions.export_all_feeds': 'Exportera alla flöden',

	'settings.groups.description': 'Gruppens namn måste vara unikt.',
	'settings.groups.delete.confirm':
		'Är du säker på att du vill ta bort denna grupp? Alla dess flöden kommer att flyttas till standardgruppen',
	'settings.groups.delete.error.delete_the_default': 'Kan inte ta bort standardgruppen',

	// auth
	'auth.logout.confirm': 'Är du säker på att du vill logga ut?',
	'auth.logout.failed_message': 'Misslyckades med att logga ut. Försök igen.',

	// shortcuts
	'shortcuts.show_help': 'Visa tangentbordsgenvägar',
	'shortcuts.next_item': 'Nästa objekt',
	'shortcuts.prev_item': 'Föregående objekt',
	'shortcuts.toggle_unread': 'Växla läst/oläst',
	'shortcuts.toggle_bookmark': 'Växla bokmärke',
	'shortcuts.view_original': 'Visa originallänk',
	'shortcuts.next_feed': 'Nästa flöde',
	'shortcuts.prev_feed': 'Föregående flöde',
	'shortcuts.open_selected': 'Öppna markering',
	'shortcuts.goto_search_page': 'Gå till sökning',
	'shortcuts.goto_unread_page': 'Gå till olästa',
	'shortcuts.goto_bookmarks_page': 'Gå till bokmärken',
	'shortcuts.goto_all_items_page': 'Gå till alla objekt',
	'shortcuts.goto_feeds_page': 'Gå till flöden',
	'shortcuts.goto_settings_page': 'Gå till inställningar'
} as const;

export default lang;
