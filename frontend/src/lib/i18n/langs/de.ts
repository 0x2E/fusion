const lang = {
	// actions
	'common.submit': 'Absenden',
	'common.cancel': 'Abbrechen',
	'common.confirm': 'Bestätigen',
	'common.add': 'Hinzufügen',
	'common.delete': 'Löschen',
	'common.edit': 'Bearbeiten',
	'common.save': 'Speichern',
	'common.close': 'Schließen',
	'common.search': 'Suchen',
	'common.login': 'Anmelden',
	'common.logout': 'Abmelden',

	// common words
	'common.unread': 'Ungelesen',
	'common.bookmark': 'Lesezeichen',
	'common.all': 'Alle',
	'common.feeds': 'Feeds',
	'common.group': 'Gruppe',
	'common.groups': 'Gruppen',
	'common.settings': 'Einstellungen',
	'common.name': 'Name',
	'common.password': 'Passwort',
	'common.link': 'Link',
	'common.advanced': 'Erweitert',
	'common.shortcuts': 'Tastaturkürzel',
	'common.more': 'Mehr',
	'common.current_page': 'Aktuelle Seite',

	// state
	'state.success': 'Erfolg',
	'state.error': 'Fehler',
	'state.loading': 'Wird geladen',
	'state.no_data': 'Keine Daten',
	'state.no_more_data': 'Keine weiteren Daten',

	// feed
	'feed.refresh': 'Feed aktualisieren',
	'feed.refresh.all': 'Alle Feeds aktualisieren',
	'feed.refresh.all.confirm':
		'Sind Sie sicher, dass Sie alle Feeds außer den ausgesetzten aktualisieren möchten?',
	'feed.refresh.all.run_in_background': 'Aktualisierung im Hintergrund starten',
	'feed.refresh.resume': 'Aktualisierung fortsetzen',
	'feed.refresh.suspend': 'Aktualisierung aussetzen',
	'feed.delete.confirm': 'Sind Sie sicher, dass Sie diesen Feed löschen möchten?',
	'feed.banner.suspended': 'Dieser Feed wurde ausgesetzt',
	'feed.banner.failed': 'Fehler beim Aktualisieren des Feeds. Fehler: {error}',

	'feed.import.title': 'Feeds hinzufügen',
	'feed.import.manually': 'Manuell',
	'feed.import.manually.link.description':
		'Entweder der RSS-Link oder der Website-Link. Der Server wird automatisch versuchen, den RSS-Feed zu lokalisieren. Der bestehende Feed mit demselben Link wird überschrieben.',
	'feed.import.manually.name.description': 'Optional. Leer lassen für automatische Benennung.',
	'feed.import.manually.no_valid_feed_error':
		'Es wurde kein gültiger Feed gefunden. Bitte überprüfen Sie den Link oder reichen Sie direkt einen Feed-Link ein.',
	'feed.import.manually.link_candidates.label': 'Wählen Sie einen Link',
	'feed.import.opml': 'OPML importieren',
	'feed.import.opml.file.label': 'Wählen Sie eine OPML-Datei',
	'feed.import.opml.file.description':
		'Die Datei sollte im {opml}-Format sein. Sie können eine aus Ihrem vorherigen RSS-Reader erhalten.',
	'feed.import.opml.file_read_error': 'Fehler beim Laden des Dateiinhalts',
	'feed.import.opml.how_it_works.title': 'Wie funktioniert es?',
	'feed.import.opml.how_it_works.description.1':
		'Feeds werden in die entsprechende Gruppe importiert, die automatisch erstellt wird, wenn sie nicht existiert.',
	'feed.import.opml.how_it_works.description.2':
		"Mehrdimensionale Gruppen werden in eine eindimensionale Struktur umgewandelt, unter Verwendung einer Namenskonvention wie 'a/b/c'.",
	'feed.import.opml.how_it_works.description.3':
		'Der bestehende Feed mit demselben Link wird überschrieben.',

	// item
	'item.search.placeholder': 'Suche in Titel und Inhalt',
	'item.mark_all_as_read': 'Alle als gelesen markieren',
	'item.mark_as_read': 'Als gelesen markieren',
	'item.mark_as_unread': 'Als ungelesen markieren',
	'item.add_to_bookmark': 'Zu Lesezeichen hinzufügen',
	'item.remove_from_bookmark': 'Aus Lesezeichen entfernen',
	'item.goto_feed': 'Zum Feed gehen',
	'item.visit_the_original': 'Originallink besuchen',
	'item.share': 'Teilen',

	// settings
	'settings.appearance': 'Erscheinungsbild',
	'settings.appearance.description': 'Diese Einstellungen werden in Ihrem Browser gespeichert.',
	'settings.appearance.field.language.label': 'Sprache',

	'settings.global_actions': 'Globale Aktionen',
	'settings.global_actions.refresh_all_feeds': 'Alle Feeds aktualisieren',
	'settings.global_actions.export_all_feeds': 'Alle Feeds exportieren',

	'settings.groups.description': 'Der Gruppenname sollte eindeutig sein.',
	'settings.groups.delete.confirm':
		'Sind Sie sicher, dass Sie diese Gruppe löschen möchten? Alle ihre Feeds werden in die Standardgruppe verschoben',
	'settings.groups.delete.error.delete_the_default':
		'Die Standardgruppe kann nicht gelöscht werden',

	// auth
	'auth.logout.confirm': 'Sind Sie sicher, dass Sie sich abmelden möchten?',
	'auth.logout.failed_message': 'Abmeldung fehlgeschlagen. Bitte versuchen Sie es erneut.',

	// shortcuts
	'shortcuts.show_help': 'Tastaturkürzel anzeigen',
	'shortcuts.next_item': 'Nächster Eintrag',
	'shortcuts.prev_item': 'Vorheriger Eintrag',
	'shortcuts.toggle_unread': 'Gelesen/Ungelesen umschalten',
	'shortcuts.toggle_bookmark': 'Lesezeichen umschalten',
	'shortcuts.view_original': 'Originallink ansehen',
	'shortcuts.next_feed': 'Nächster Feed',
	'shortcuts.prev_feed': 'Vorheriger Feed',
	'shortcuts.open_selected': 'Auswahl öffnen',
	'shortcuts.goto_search_page': 'Zur Suche gehen',
	'shortcuts.goto_unread_page': 'Zu Ungelesenen gehen',
	'shortcuts.goto_bookmarks_page': 'Zu Lesezeichen gehen',
	'shortcuts.goto_all_items_page': 'Zu allen Einträgen gehen',
	'shortcuts.goto_feeds_page': 'Zu Feeds gehen',
	'shortcuts.goto_settings_page': 'Zu Einstellungen gehen'
} as const;

export default lang;
