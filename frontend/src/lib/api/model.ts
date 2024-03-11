export type Group = {
	id: number;
	name: string;
};

export type Feed = {
	id: number;
	name: string;
	link: string;
	failure: string;
	updated_at: Date;
	suspended: boolean;
	group: Group;
};

export type Item = {
	id: number;
	title: string;
	link: string;
	content: string;
	pub_date: Date;
	unread: boolean;
	bookmark: boolean;
	feed: { id: number; name: string };
};
