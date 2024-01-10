export type Comments = {
    PostId: string,
    IsAvailable: boolean,
    IsWritable: boolean,
    List: Comment[],
};

export type Comment = {
    Id: string,
    Visible: string,
    Author: string,
    When: string,
    Text: string,
};

export type NewComment = {
    PostId: string,
    Author: string,
    Text: string,
}

export type Filter = {
    Visible: boolean | null,
}

export enum Sort {
    NewestFirst,
}

export type FoundComments = {
    List: {
        PostId: string,
        Comment: Comment,
    }[],
    More: boolean,
}

export class UserApi {
    constructor() {
        this.apiUrl = findApiUrl();
    }

    private apiUrl: string;

    async list(postId: string): Promise<Comments> {
        return get(this.apiUrl + '/list/' + postId);
    }

    async add(newComment: NewComment): Promise<Comment> {
        return post(this.apiUrl + '/add', newComment);
    }

    async render(text: string): Promise<string> {
        return post(this.apiUrl + '/render', { 'Text': text });
    }
}

export class AdminApi {
    constructor() {
        this.apiUrl = findApiUrl();
    }

    private apiUrl: string;

    async find(filter: Filter, sort: Sort, limit: number, start: number): Promise<FoundComments> {
        let params = {
            'Filter': filter,
            'Sort': sort,
            'Limit': limit,
            'Start': start,
        };
        return post(this.apiUrl + '/find', params);
    }

    async setVisible(ids: Map<string, string[]>, visible: boolean) {
        let params = {
            'Ids': Object.fromEntries(ids),
            'Visible': visible
        };
        await post(this.apiUrl + '/setVisible', params);
    }
}

async function get<R>(url: string): Promise<R> {
    let response = await fetch(url, { method: 'GET', mode: 'cors' });
    if (response.status != 200) {
        throw `Error ${response.status}: ${await response.text()}`;
    }
    return response.json();
}

async function post<R, M>(url: string, data: M): Promise<R> {
    let response = await fetch(url, { method: 'POST', mode: 'cors', body: JSON.stringify(data) });
    if (response.status != 200) {
        throw `Error ${response.status}: ${await response.text()}`;
    }
    return response.json();
}

function findApiUrl(): string {
    let scripts = document.getElementsByTagName('script');
    let baseUrl = new URL(scripts[scripts.length - 1].attributes['src'].value, window.location.toString());
    let pathname = baseUrl.pathname;
    let lastSlash = pathname.lastIndexOf('/');
    if (lastSlash === undefined) {
        baseUrl.pathname = '/_';
    } else {
        baseUrl.pathname = pathname.substring(0, lastSlash) + '_';
    }
    return baseUrl.toString();
}
