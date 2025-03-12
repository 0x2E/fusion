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
	req_proxy: string;
	group: Group;
};

export type Item = {
	id: number;
	title: string;
	link: string;
	content: string;
	unread: boolean;
	bookmark: boolean;
	pub_date: Date;
	updated_at: Date;
	feed: Pick<Feed, 'id' | 'name' | 'link'>;
};
