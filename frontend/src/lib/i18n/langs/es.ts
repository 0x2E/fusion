const lang = {
	// actions
	'common.submit': 'Enviar',
	'common.cancel': 'Cancelar',
	'common.confirm': 'Confirmar',
	'common.add': 'Añadir',
	'common.delete': 'Eliminar',
	'common.edit': 'Editar',
	'common.save': 'Guardar',
	'common.close': 'Cerrar',
	'common.search': 'Buscar',
	'common.login': 'Iniciar sesión',
	'common.logout': 'Cerrar sesión',

	// common words
	'common.unread': 'No leído',
	'common.bookmark': 'Marcador',
	'common.all': 'Todos',
	'common.feeds': 'Feeds',
	'common.group': 'Grupo',
	'common.groups': 'Grupos',
	'common.settings': 'Configuración',
	'common.name': 'Nombre',
	'common.password': 'Contraseña',
	'common.link': 'Enlace',
	'common.advanced': 'Avanzado',
	'common.shortcuts': 'Atajos de teclado',
	'common.more': 'Más',
	'common.current_page': 'Página actual',

	// state
	'state.success': 'Éxito',
	'state.error': 'Error',
	'state.loading': 'Cargando',
	'state.no_data': 'Sin datos',
	'state.no_more_data': 'No hay más datos',

	// feed
	'feed.refresh': 'Actualizar Feed',
	'feed.refresh.all': 'Actualizar Todos los Feeds',
	'feed.refresh.all.confirm':
		'¿Estás seguro de que quieres actualizar todos los feeds excepto los suspendidos?',
	'feed.refresh.all.run_in_background': 'Iniciar actualización en segundo plano',
	'feed.refresh.resume': 'Reanudar actualización',
	'feed.refresh.suspend': 'Suspender actualización',
	'feed.delete.confirm': '¿Estás seguro de que quieres eliminar este feed?',
	'feed.banner.suspended': 'Este feed ha sido suspendido',
	'feed.banner.failed': 'Error al actualizar el feed. Error: {error}',

	'feed.import.title': 'Añadir Feeds',
	'feed.import.manually': 'Manualmente',
	'feed.import.manually.link.description':
		'El enlace RSS o el enlace del sitio web. El servidor intentará localizar automáticamente el feed RSS. El feed existente con el mismo enlace será reemplazado.',
	'feed.import.manually.name.description':
		'Opcional. Dejar en blanco para nombrar automáticamente.',
	'feed.import.manually.no_valid_feed_error':
		'No se encontró ningún feed válido. Por favor, verifica el enlace o envía un enlace de feed directamente.',
	'feed.import.manually.link_candidates.label': 'Seleccionar un enlace',
	'feed.import.opml': 'Importar OPML',
	'feed.import.opml.file.label': 'Seleccionar un archivo OPML',
	'feed.import.opml.file.description':
		'El archivo debe estar en formato {opml}. Puedes obtener uno de tu lector RSS anterior.',
	'feed.import.opml.file_read_error': 'Error al cargar el contenido del archivo',
	'feed.import.opml.how_it_works.title': '¿Cómo funciona?',
	'feed.import.opml.how_it_works.description.1':
		'Los feeds se importarán al grupo correspondiente, que se creará automáticamente si no existe.',
	'feed.import.opml.how_it_works.description.2':
		"Los grupos multidimensionales se aplanarán a una estructura unidimensional, utilizando una convención de nomenclatura como 'a/b/c'.",
	'feed.import.opml.how_it_works.description.3':
		'El feed existente con el mismo enlace será reemplazado.',

	// item
	'item.search.placeholder': 'Buscar en título y contenido',
	'item.mark_all_as_read': 'Marcar todo como leído',
	'item.mark_as_read': 'Marcar como leído',
	'item.mark_as_unread': 'Marcar como no leído',
	'item.add_to_bookmark': 'Añadir a marcadores',
	'item.remove_from_bookmark': 'Eliminar de marcadores',
	'item.goto_feed': 'Ir al feed',
	'item.visit_the_original': 'Visitar enlace original',
	'item.share': 'Compartir',

	// settings
	'settings.appearance': 'Apariencia',
	'settings.appearance.description': 'Esta configuración se guarda en tu navegador.',
	'settings.appearance.field.language.label': 'Idioma',

	'settings.global_actions': 'Acciones globales',
	'settings.global_actions.refresh_all_feeds': 'Actualizar todos los feeds',
	'settings.global_actions.export_all_feeds': 'Exportar todos los feeds',

	'settings.groups.description': 'El nombre del grupo debe ser único.',
	'settings.groups.delete.confirm':
		'¿Estás seguro de que quieres eliminar este grupo? Todos sus feeds se moverán al grupo predeterminado',
	'settings.groups.delete.error.delete_the_default': 'No se puede eliminar el grupo predeterminado',

	// auth
	'auth.logout.confirm': '¿Estás seguro de que quieres cerrar sesión?',
	'auth.logout.failed_message': 'Error al cerrar sesión. Por favor, inténtalo de nuevo.',

	// shortcuts
	'shortcuts.show_help': 'Mostrar atajos de teclado',
	'shortcuts.next_item': 'Siguiente elemento',
	'shortcuts.prev_item': 'Elemento anterior',
	'shortcuts.toggle_unread': 'Alternar leído/no leído',
	'shortcuts.toggle_bookmark': 'Alternar marcador',
	'shortcuts.view_original': 'Ver original',
	'shortcuts.next_feed': 'Siguiente feed',
	'shortcuts.prev_feed': 'Feed anterior',
	'shortcuts.open_selected': 'Abrir selección',
	'shortcuts.goto_search_page': 'Ir a búsqueda',
	'shortcuts.goto_unread_page': 'Ir a no leídos',
	'shortcuts.goto_bookmarks_page': 'Ir a marcadores',
	'shortcuts.goto_all_items_page': 'Ir a todos los elementos',
	'shortcuts.goto_feeds_page': 'Ir a feeds',
	'shortcuts.goto_settings_page': 'Ir a configuración'
} as const;

export default lang;
