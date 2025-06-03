const lang = {
	// actions
	'common.submit': 'Submeter',
	'common.cancel': 'Cancelar',
	'common.confirm': 'Confirmar',
	'common.add': 'Adicionar',
	'common.delete': 'Eliminar',
	'common.edit': 'Editar',
	'common.save': 'Guardar',
	'common.close': 'Fechar',
	'common.search': 'Pesquisar',
	'common.login': 'Iniciar sessão',
	'common.logout': 'Terminar sessão',

	// common words
	'common.unread': 'Não lidos',
	'common.bookmark': 'Favoritos',
	'common.all': 'Todos',
	'common.feeds': 'Feeds',
	'common.group': 'Grupo',
	'common.groups': 'Grupos',
	'common.settings': 'Definições',
	'common.name': 'Nome',
	'common.password': 'Palavra-passe',
	'common.link': 'Ligação',
	'common.advanced': 'Avançado',
	'common.shortcuts': 'Atalhos de teclado',
	'common.more': 'Mais',
	'common.current_page': 'Página atual',

	// state
	'state.success': 'Sucesso',
	'state.error': 'Erro',
	'state.loading': 'A carregar',
	'state.no_data': 'Sem dados',
	'state.no_more_data': 'Não há mais dados',

	// feed
	'feed.refresh': 'Atualizar Feed',
	'feed.refresh.all': 'Atualizar Todos os Feeds',
	'feed.refresh.all.confirm':
		'Tem a certeza que pretende atualizar todos os feeds exceto os suspensos?',
	'feed.refresh.all.run_in_background': 'Iniciar atualização em segundo plano',
	'feed.refresh.resume': 'Retomar atualização',
	'feed.refresh.suspend': 'Suspender atualização',
	'feed.delete.confirm': 'Tem a certeza que pretende eliminar este feed?',
	'feed.banner.suspended': 'Este feed foi suspenso',
	'feed.banner.failed': 'Falha ao atualizar o feed. Erro: {error}',

	'feed.import.title': 'Adicionar Feeds',
	'feed.import.manually': 'Manualmente',
	'feed.import.manually.link.description':
		'Pode ser o link RSS ou o link do site. O servidor tentará localizar automaticamente o feed RSS. O feed existente com o mesmo link será substituído.',
	'feed.import.manually.name.description': 'Opcional. Deixe em branco para nomeação automática.',
	'feed.import.manually.no_valid_feed_error':
		'Nenhum feed válido foi encontrado. Verifique o link ou envie um link de feed diretamente.',
	'feed.import.manually.link_candidates.label': 'Selecione um link',
	'feed.import.opml': 'Importar OPML',
	'feed.import.opml.file.label': 'Escolha um ficheiro OPML',
	'feed.import.opml.file.description':
		'O ficheiro deve estar no formato {opml}. Pode obter um do seu leitor RSS anterior.',
	'feed.import.opml.file_read_error': 'Falha ao carregar o conteúdo do ficheiro',
	'feed.import.opml.how_it_works.title': 'Como funciona?',
	'feed.import.opml.how_it_works.description.1':
		'Os feeds serão importados para o grupo correspondente, que será criado automaticamente se não existir.',
	'feed.import.opml.how_it_works.description.2':
		"O grupo multidimensional será simplificado para uma estrutura unidimensional, usando uma convenção de nomenclatura como 'a/b/c'.",
	'feed.import.opml.how_it_works.description.3':
		'O feed existente com o mesmo link será substituído.',

	// item
	'item.search.placeholder': 'Pesquisar no título e conteúdo',
	'item.mark_all_as_read': 'Marcar tudo como lido',
	'item.mark_as_read': 'Marcar como lido',
	'item.mark_as_unread': 'Marcar como não lido',
	'item.add_to_bookmark': 'Adicionar aos favoritos',
	'item.remove_from_bookmark': 'Remover dos favoritos',
	'item.goto_feed': 'Ir para o feed',
	'item.visit_the_original': 'Visitar link original',
	'item.share': 'Partilhar',

	// settings
	'settings.appearance': 'Aparência',
	'settings.appearance.description': 'Estas definições são armazenadas no seu navegador.',
	'settings.appearance.field.language.label': 'Idioma',

	'settings.global_actions': 'Ações globais',
	'settings.global_actions.refresh_all_feeds': 'Atualizar todos os feeds',
	'settings.global_actions.export_all_feeds': 'Exportar todos os feeds',

	'settings.groups.description': 'O nome do grupo deve ser único.',
	'settings.groups.delete.confirm':
		'Tem a certeza que pretende eliminar este grupo? Todos os seus feeds serão movidos para o grupo predefinido',
	'settings.groups.delete.error.delete_the_default': 'Não é possível eliminar o grupo predefinido',

	// auth
	'auth.logout.confirm': 'Tem a certeza que pretende terminar a sessão?',
	'auth.logout.failed_message': 'Falha ao terminar a sessão. Por favor, tente novamente.',

	// shortcuts
	'shortcuts.show_help': 'Mostrar atalhos de teclado',
	'shortcuts.next_item': 'Próximo item',
	'shortcuts.prev_item': 'Item anterior',
	'shortcuts.toggle_unread': 'Alternar lido/não lido',
	'shortcuts.toggle_bookmark': 'Alternar favorito',
	'shortcuts.view_original': 'Ver link original',
	'shortcuts.next_feed': 'Próximo feed',
	'shortcuts.prev_feed': 'Feed anterior',
	'shortcuts.open_selected': 'Abrir seleção',
	'shortcuts.goto_search_page': 'Ir para pesquisa',
	'shortcuts.goto_unread_page': 'Ir para não lidos',
	'shortcuts.goto_bookmarks_page': 'Ir para favoritos',
	'shortcuts.goto_all_items_page': 'Ir para todos os itens',
	'shortcuts.goto_feeds_page': 'Ir para feeds',
	'shortcuts.goto_settings_page': 'Ir para definições'
} as const;

export default lang;
