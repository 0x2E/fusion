const lang = {
	// actions
	'common.submit': 'Submit',
	'common.cancel': 'Cancel',
	'common.confirm': 'Confirm',
	'common.add': 'Add',
	'common.delete': 'Delete',
	'common.edit': 'Edit',
	'common.save': 'Save',
	'common.close': 'Close',
	'common.search': 'Search',
	'common.login': 'Log in',
	'common.logout': 'Log out',

	// common words
	'common.unread': 'Unread',
	'common.bookmark': 'Bookmark',
	'common.all': 'All',
	'common.feeds': 'Feeds',
	'common.group': 'Group',
	'common.groups': 'Groups',
	'common.settings': 'Settings',
	'common.name': 'Name',
	'common.password': 'Password',
	'common.link': 'Link',
	'common.advanced': 'Advanced',
	'common.shortcuts': 'Keyboard shortcuts',
	'common.more': 'More',
	'common.current_page': 'Current page',

	// state
	'state.success': 'Success',
	'state.error': 'Error',
	'state.loading': 'Loading',
	'state.no_data': 'No data',
	'state.no_more_data': 'No more data',

	// feed
	'feed.refresh': 'Refresh Feed',
	'feed.refresh.all': 'Refresh All Feeds',
	'feed.refresh.all.confirm':
		'Are you sure you want to refresh all feeds except the suspended ones?',
	'feed.refresh.all.run_in_background': 'Start refreshing in the background',
	'feed.refresh.resume': 'Resume refreshing',
	'feed.refresh.suspend': 'Suspend refreshing',
	'feed.delete.confirm': 'Are you sure you want to delete this feed?',
	'feed.banner.suspended': 'This feed has been suspended',
	'feed.banner.failed': 'Failed to refresh the feed. Error: {error}',

	'feed.import.title': 'Add Feeds',
	'feed.import.manually': 'Manually',
	'feed.import.manually.link.description':
		'Either the RSS link or the website link. The server will automatically attempt to locate the RSS feed. The existing feed with the same link will be overridden.',
	'feed.import.manually.name.description': 'Optional. Leave blank for automatic naming.',
	'feed.import.manually.no_valid_feed_error':
		'No valid feed was found. Please check the link, or submit a feed link directly.',
	'feed.import.manually.link_candidates.label': 'Select a link',
	'feed.import.opml': 'Import OPML',
	'feed.import.opml.file.label': 'Pick a OPML file',
	'feed.import.opml.file.description':
		'The file should be {opml} format. You can get one from your previous RSS reader.',
	'feed.import.opml.file_read_error': 'Failed to load file content',
	'feed.import.opml.how_it_works.title': 'How it works?',
	'feed.import.opml.how_it_works.description.1':
		'Feeds will be imported into the corresponding group, which will be created automatically if it does not exist.',
	'feed.import.opml.how_it_works.description.2':
		"Multidimensional group will be flattened to a one-dimensional structure, using a naming convention like 'a/b/c'.",
	'feed.import.opml.how_it_works.description.3':
		'The existing feed with the same link will be overridden.',

	// item
	'item.search.placeholder': 'Search in title and content',
	'item.mark_all_as_read': 'Mark all as read',
	'item.mark_as_read': 'Mark as read',
	'item.mark_as_unread': 'Mark as unread',
	'item.add_to_bookmark': 'Add to bookmark',
	'item.remove_from_bookmark': 'Remove from bookmark',
	'item.goto_feed': 'Go to feed',
	'item.visit_the_original': 'Visit original link',
	'item.share': 'Share',

	// settings
	'settings.appearance': 'Appearance',
	'settings.appearance.description': 'These settings are stored in your browser.',
	'settings.appearance.field.language.label': 'Language',

	'settings.global_actions': 'Global actions',
	'settings.global_actions.refresh_all_feeds': 'Refresh all feeds',
	'settings.global_actions.export_all_feeds': 'Export all feeds',

	'settings.groups.description': "Group's name should be unique.",
	'settings.groups.delete.confirm':
		'Are you sure you want to delete this group? All its feeds will be moved to the default group',
	'settings.groups.delete.error.delete_the_default': 'Cannot delete default group',

	// auth
	'auth.logout.confirm': 'Are you sure you want to log out?',
	'auth.logout.failed_message': 'Log out failed. Please try again.',

	// shortcuts
	'shortcuts.show_help': 'Show keyboard shortcuts',
	'shortcuts.next_item': 'Next item',
	'shortcuts.prev_item': 'Previous item',
	'shortcuts.toggle_unread': 'Toggle read/unread',
	'shortcuts.toggle_bookmark': 'Toggle bookmark',
	'shortcuts.view_original': 'View original',
	'shortcuts.next_feed': 'Next feed',
	'shortcuts.prev_feed': 'Previous feed',
	'shortcuts.open_selected': 'Open selection',
	'shortcuts.goto_search_page': 'Go to search',
	'shortcuts.goto_unread_page': 'Go to unread',
	'shortcuts.goto_bookmarks_page': 'Go to bookmarks',
	'shortcuts.goto_all_items_page': 'Go to all items',
	'shortcuts.goto_feeds_page': 'Go to feeds',
	'shortcuts.goto_settings_page': 'Go to settings'
} as const;

export default lang;
