const lang = {
	// actions
	'common.submit': 'Enviar',
	'common.cancel': 'Cancel·lar',
	'common.confirm': 'Confirmar',
	'common.add': 'Afegir',
	'common.delete': 'Eliminar',
	'common.edit': 'Editar',
	'common.save': 'Guardar',
	'common.close': 'Tancar',
	'common.search': 'Cerca',
	'common.login': 'Iniciar la sessió',
	'common.logout': 'Tancar la sessió',

	// common words
	'common.unread': 'No llegits',
	'common.bookmark': 'Marcadors',
	'common.all': 'Tots',
	'common.feeds': 'Canals',
	'common.group': 'Grup',
	'common.groups': 'Grups',
	'common.settings': 'Configuració',
	'common.name': 'Nom',
	'common.password': 'Contrasenya',
	'common.link': 'Enllaç',
	'common.advanced': 'Avançat',
	'common.shortcuts': 'Dreceres del teclat',
	'common.more': 'Més',
	'common.current_page': 'Pàgina actual',

	// state
	'state.success': 'Èxit',
	'state.error': 'Error',
	'state.loading': 'Carregant',
	'state.no_data': 'Sense dades',
	'state.no_more_data': 'No hi ha més dades',

	// feed
	'feed.refresh': 'Actualitzar el canal',
	'feed.refresh.all': 'Actualitzar tots els canals',
	'feed.refresh.all.confirm':
		'Estàs segur que vols actualitzar tots els canals excepte els suspesos?',
	'feed.refresh.all.run_in_background': "Iniciar l'actualització en segon pla",
	'feed.refresh.resume': "Reprendre l'actualització",
	'feed.refresh.suspend': "Suspendre l'actualització",
	'feed.delete.confirm': 'Estàs segur que vols eliminar aquest canal?',
	'feed.banner.suspended': 'Aquest canal ha sigut suspès',
	'feed.banner.failed': 'Error en actualitzar el canal. Error: {error}',

	'feed.import.title': 'Afegir canals',
	'feed.import.manually': 'Manualment',
	'feed.import.manually.link.description':
		"L'enllaç RSS o l'enllaç del lloc web. El servidor intentarà localitzar automàticament el canal RSS. Els canals existents amb el mateix enllaç es substituiran.",
	'feed.import.manually.name.description': 'Opcional. Deixar en blanc per anomenar automàticament.',
	'feed.import.manually.no_valid_feed_error':
		"No s'ha trobat cap canal vàlid. Si us plau, verifica l'enllaç o fes servir un enllaç de canal directament.",
	'feed.import.manually.link_candidates.label': 'Seleccionar un enllaç',
	'feed.import.opml': 'Importar OPML',
	'feed.import.opml.file.label': 'Seleccionar un fitxer OPML',
	'feed.import.opml.file.description':
		"El fitxer ha d'estar en format {opml}. Pots obtenir un del teu lector RSS anterior.",
	'feed.import.opml.file_read_error': 'Error en carregar el contingut del fitxer',
	'feed.import.opml.how_it_works.title': 'Com funciona?',
	'feed.import.opml.how_it_works.description.1':
		"Els canals s'importaran al grup corresponent, que es crearà automàticament si no existeix.",
	'feed.import.opml.how_it_works.description.2':
		"Els grups multidimensionals s'aplanaran a una estructura unidimensional, utilitzant una convenció de nomenclatura com 'a/b/c'.",
	'feed.import.opml.how_it_works.description.3':
		'Els canals existents amb el mateix enllaç es substituiran.',

	// item
	'item.search.placeholder': 'Cercar al títol i contingut',
	'item.mark_all_as_read': 'Marcar tots com a llegits',
	'item.mark_as_read': 'Marcar com a llegit',
	'item.mark_as_unread': 'Marcar com a no llegit',
	'item.add_to_bookmark': 'Afegir als marcadors',
	'item.remove_from_bookmark': 'Treure dels marcadors',
	'item.goto_feed': 'Anar al canal',
	'item.visit_the_original': "Visitar l'enllaç original",
	'item.share': 'Compatir',

	// settings
	'settings.appearance': 'Aparença',
	'settings.appearance.description': "Aquesta configuració s'ha guardat al teu navegador.",
	'settings.appearance.field.language.label': 'Idioma',

	'settings.global_actions': 'Accions globals',
	'settings.global_actions.refresh_all_feeds': 'Actualitzar tots els canals',
	'settings.global_actions.export_all_feeds': 'Exportar tots els canals',

	'settings.groups.description': 'El nom del grup ha de ser únic.',
	'settings.groups.delete.confirm':
		'Estàs segur que vols eliminar aquest grup? Tots els seus canals es mouran al grup predeterminat',
	'settings.groups.delete.error.delete_the_default': 'No es pot eliminar el grup predeterminat',

	// auth
	'auth.logout.confirm': 'Estàs segur que vols tancar la sessió?',
	'auth.logout.failed_message': 'Error en tancar sessió. Si us plau, torna-ho a intentar.',

	// shortcuts
	'shortcuts.show_help': 'Mostrar les dreceres del teclat',
	'shortcuts.next_item': 'Element següent',
	'shortcuts.prev_item': 'Element anterior',
	'shortcuts.toggle_unread': 'Marcar com a (no) llegit',
	'shortcuts.toggle_bookmark': 'Afegir o treure dels marcadors',
	'shortcuts.view_original': "Veure l'original",
	'shortcuts.next_feed': 'Canal següent',
	'shortcuts.prev_feed': 'Canal anterior',
	'shortcuts.open_selected': 'Obrir el seleccionat',
	'shortcuts.goto_search_page': 'Anar a la cerca',
	'shortcuts.goto_unread_page': 'Anar als no llegits',
	'shortcuts.goto_bookmarks_page': 'Anar als marcadors',
	'shortcuts.goto_all_items_page': 'Anar a tots',
	'shortcuts.goto_feeds_page': 'Anar als canals',
	'shortcuts.goto_settings_page': 'Anar a la configuració'
} as const;

export default lang;
