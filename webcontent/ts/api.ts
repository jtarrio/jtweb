export type Comments = {
    PostId: string,
    Config: {
        IsReadable: boolean,
        IsWritable: boolean,
    },
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

export class UserApi {
    async list(postId: string): Promise<Comments> {
        return post('/list', { 'PostId': postId });
    }

    async add(newComment: NewComment): Promise<Comment> {
        return post('/add', newComment);
    }

    async render(text: string): Promise<string> {
        return post('/render', { 'Text': text });
    }
}

export type CommentFilter = {
    Visible: boolean | null,
}

export enum Sort {
    NewestFirst,
}

export type RawComment = {
    PostId: string,
    CommentId: string,
    Visible: boolean,
    Author: string,
    When: string,
    Text: string,
}

export type FoundComments = {
    List: RawComment[],
    More: boolean,
}

export type PostFilter = {
    CommentsReadable: boolean | null,
    CommentsWritable: boolean | null,
}

export type FoundPosts = {
    List: FoundPost[],
    More: boolean,
}

export type FoundPost = {
    PostId: string,
    Config: CommentConfig,
}

export type CommentConfig = {
    IsReadable: boolean,
    IsWritable: boolean,
}

export class AdminApi {
    async findComments(filter: CommentFilter, sort: Sort, limit: number, start: number): Promise<FoundComments> {
        let params = {
            'Filter': filter,
            'Sort': sort,
            'Limit': limit,
            'Start': start,
        };
        return post('/findComments', params);
    }

    async deleteComments(ids: Map<string, string[]>) {
        let params = {
            'Ids': Object.fromEntries(ids)
        };
        await post('/deleteComments', params);
    }

    async findPosts(filter: PostFilter, sort: Sort, limit: number, start: number): Promise<FoundPosts> {
        let params = {
            'Filter': filter,
            'Sort': sort,
            'Limit': limit,
            'Start': start,
        };
        return post('/findPosts', params);
    }

    async bulkSetVisible(ids: Map<string, string[]>, visible: boolean) {
        let params = {
            'Ids': Object.fromEntries(ids),
            'Visible': visible
        };
        await post('/bulkSetVisible', params);
    }

    async bulkUpdatePostConfigs(postIds: string[], writable: boolean, readable: boolean) {
        let params = {
            'PostIds': postIds,
            'Config': {
                'IsWritable': writable,
                'IsReadable': readable,
            },
        }
        await post('/bulkUpdatePostConfigs', params);
    }
}

async function post<R, M>(url: string, data: M): Promise<R> {
    let response = await fetch(apiUrl + url, { method: 'POST', mode: 'cors', body: JSON.stringify(data) });
    if (response.status != 200) {
        throw `Error ${response.status}: ${await response.text()}`;
    }
    return response.json();
}

const apiUrl = (() => {
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
})();
