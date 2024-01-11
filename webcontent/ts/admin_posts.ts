import { AdminPage } from './adminpage';
import { AdminApi, FoundPost, FoundPosts, PostFilter, Sort } from './api';

export function doAdminPosts(rootElement: HTMLElement) {
    new AdminPosts(rootElement);
}

class AdminPosts extends AdminPage<FoundPost, PostFilter> {
    constructor(root: HTMLElement) {
        super(root);
        this.api = new AdminApi();
        this.wireEvent('input[name=ApplyPostFilter]', 'click', _ => this.loadList(0));
        this.loadList(0);
    }

    private api: AdminApi;

    protected getFilter(): PostFilter {
        let out: PostFilter = { CommentsReadable: null, CommentsWritable: null };
        let form = this.root.querySelector('#postFilters');
        if (!form) throw "Could not find filters box";
        let state = (form.querySelector('[name=State]') as HTMLSelectElement).value;
        if (state == 'open') {
            out.CommentsReadable = true;
            out.CommentsWritable = true;
        }
        if (state == 'closed') {
            out.CommentsReadable = true;
            out.CommentsWritable = false;
        }
        if (state == 'disabled') {
            out.CommentsReadable = false;
        }
        return out;
    }

    protected getFilterTitle(filter: PostFilter): string {
        if (filter.CommentsWritable === true) {
            return 'Latest posts with open comments';
        } else if (filter.CommentsReadable === true) {
            return 'Latest posts with closed comments';
        } else if (filter.CommentsReadable === false) {
            return 'Latest posts with disabled comments';
        } else {
            return 'Latest posts';
        }
    }

    protected getItemsPerPage(): number {
        let form = this.root.querySelector('#postFilters');
        if (form) {
            let items = form.querySelector('[name=ItemsPerPage') as HTMLSelectElement | null;
            if (items) return Number(items.value);
        }
        return 20;
    }

    protected getItems(filter: PostFilter, itemsPerPage: number, start: number): Promise<FoundPosts> {
        return this.api.findPosts(filter, Sort.NewestFirst, itemsPerPage, start);
    }

    protected getRowContents(item: FoundPost): string[] {
        return [
            item.Config.IsReadable ? item.Config.IsWritable ? 'open' : 'closed' : 'disabled',
            item.PostId
        ];
    }

    protected getListNameElement(): HTMLElement | null {
        return this.root.querySelector('#postListName');
    }

    protected getListTableElement(): HTMLElement | null {
        return this.root.querySelector('#postList');
    }

    protected getListLinksElement(): HTMLElement | null {
        return this.root.querySelector('#postListLinks');
    }
}

