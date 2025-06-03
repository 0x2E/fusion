const lang = {
	// actions
	'common.submit': 'Soumettre',
	'common.cancel': 'Annuler',
	'common.confirm': 'Confirmer',
	'common.add': 'Ajouter',
	'common.delete': 'Supprimer',
	'common.edit': 'Modifier',
	'common.save': 'Enregistrer',
	'common.close': 'Fermer',
	'common.search': 'Rechercher',
	'common.login': 'Se connecter',
	'common.logout': 'Se déconnecter',

	// common words
	'common.unread': 'Non lu',
	'common.bookmark': 'Favori',
	'common.all': 'Tous',
	'common.feeds': 'Flux',
	'common.group': 'Groupe',
	'common.groups': 'Groupes',
	'common.settings': 'Paramètres',
	'common.name': 'Nom',
	'common.password': 'Mot de passe',
	'common.link': 'Lien',
	'common.advanced': 'Avancé',
	'common.shortcuts': 'Raccourcis clavier',
	'common.more': 'Plus',
	'common.current_page': 'Page actuelle',

	// state
	'state.success': 'Succès',
	'state.error': 'Erreur',
	'state.loading': 'Chargement',
	'state.no_data': 'Aucune donnée',
	'state.no_more_data': 'Pas plus de données',

	// feed
	'feed.refresh': 'Actualiser le flux',
	'feed.refresh.all': 'Actualiser tous les flux',
	'feed.refresh.all.confirm':
		'Êtes-vous sûr de vouloir actualiser tous les flux sauf ceux suspendus?',
	'feed.refresh.all.run_in_background': "Démarrer l'actualisation en arrière-plan",
	'feed.refresh.resume': "Reprendre l'actualisation",
	'feed.refresh.suspend': "Suspendre l'actualisation",
	'feed.delete.confirm': 'Êtes-vous sûr de vouloir supprimer ce flux?',
	'feed.banner.suspended': 'Ce flux a été suspendu',
	'feed.banner.failed': "Échec de l'actualisation du flux. Erreur: {error}",

	'feed.import.title': 'Ajouter des flux',
	'feed.import.manually': 'Manuellement',
	'feed.import.manually.link.description':
		'Soit le lien RSS, soit le lien du site web. Le serveur tentera automatiquement de localiser le flux RSS. Le flux existant avec le même lien sera remplacé.',
	'feed.import.manually.name.description': 'Optionnel. Laissez vide pour un nommage automatique.',
	'feed.import.manually.no_valid_feed_error':
		"Aucun flux valide n'a été trouvé. Veuillez vérifier le lien ou soumettre directement un lien de flux.",
	'feed.import.manually.link_candidates.label': 'Sélectionner un lien',
	'feed.import.opml': 'Importer OPML',
	'feed.import.opml.file.label': 'Choisir un fichier OPML',
	'feed.import.opml.file.description':
		'Le fichier doit être au format {opml}. Vous pouvez en obtenir un de votre précédent lecteur RSS.',
	'feed.import.opml.file_read_error': 'Échec du chargement du contenu du fichier',
	'feed.import.opml.how_it_works.title': 'Comment ça marche?',
	'feed.import.opml.how_it_works.description.1':
		"Les flux seront importés dans le groupe correspondant, qui sera créé automatiquement s'il n'existe pas.",
	'feed.import.opml.how_it_works.description.2':
		"Le groupe multidimensionnel sera aplati en une structure unidimensionnelle, en utilisant une convention de nommage comme 'a/b/c'.",
	'feed.import.opml.how_it_works.description.3':
		'Le flux existant avec le même lien sera remplacé.',

	// item
	'item.search.placeholder': 'Rechercher dans le titre et le contenu',
	'item.mark_all_as_read': 'Marquer tout comme lu',
	'item.mark_as_read': 'Marquer comme lu',
	'item.mark_as_unread': 'Marquer comme non lu',
	'item.add_to_bookmark': 'Ajouter aux favoris',
	'item.remove_from_bookmark': 'Retirer des favoris',
	'item.goto_feed': 'Aller au flux',
	'item.visit_the_original': 'Visiter le lien original',
	'item.share': 'Partager',

	// settings
	'settings.appearance': 'Apparence',
	'settings.appearance.description': 'Ces paramètres sont stockés dans votre navigateur.',
	'settings.appearance.field.language.label': 'Langue',

	'settings.global_actions': 'Actions globales',
	'settings.global_actions.refresh_all_feeds': 'Actualiser tous les flux',
	'settings.global_actions.export_all_feeds': 'Exporter tous les flux',

	'settings.groups.description': 'Le nom du groupe doit être unique.',
	'settings.groups.delete.confirm':
		'Êtes-vous sûr de vouloir supprimer ce groupe? Tous ses flux seront déplacés vers le groupe par défaut',
	'settings.groups.delete.error.delete_the_default': 'Impossible de supprimer le groupe par défaut',

	// auth
	'auth.logout.confirm': 'Êtes-vous sûr de vouloir vous déconnecter?',
	'auth.logout.failed_message': 'Échec de la déconnexion. Veuillez réessayer.',

	// shortcuts
	'shortcuts.show_help': 'Afficher les raccourcis clavier',
	'shortcuts.next_item': 'Élément suivant',
	'shortcuts.prev_item': 'Élément précédent',
	'shortcuts.toggle_unread': 'Basculer lu/non lu',
	'shortcuts.toggle_bookmark': 'Basculer favori',
	'shortcuts.view_original': 'Voir le lien original',
	'shortcuts.next_feed': 'Flux suivant',
	'shortcuts.prev_feed': 'Flux précédent',
	'shortcuts.open_selected': 'Ouvrir la sélection',
	'shortcuts.goto_search_page': 'Aller à la recherche',
	'shortcuts.goto_unread_page': 'Aller aux non lus',
	'shortcuts.goto_bookmarks_page': 'Aller aux favoris',
	'shortcuts.goto_all_items_page': 'Aller à tous les éléments',
	'shortcuts.goto_feeds_page': 'Aller aux flux',
	'shortcuts.goto_settings_page': 'Aller aux paramètres'
} as const;

export default lang;
