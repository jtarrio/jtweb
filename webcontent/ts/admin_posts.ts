import { AdminPage } from './adminpage';
import { AdminApi, FoundPost, FoundPosts, PostFilter, Sort } from './api';

export function doAdminPosts(rootElement: HTMLElement) {
    new AdminPosts(rootElement);
}

class AdminPosts extends AdminPage<FoundPost, PostFilter> {
    constructor(root: HTMLElement) {
        super(root);
        this.api = new AdminApi();
        this.wireEvent('input[name=MakeOpen]', 'click', _ => this.changeState(true, true));
        this.wireEvent('input[name=MakeClosed]', 'click', _ => this.changeState(false, true));
        this.wireEvent('input[name=MakeDisabled]', 'click', _ => this.changeState(false, false));
        this.loadList(0);
    }

    private api: AdminApi;

    protected getFilter(): PostFilter {
        let out: PostFilter = { CommentsReadable: null, CommentsWritable: null };
        let form = this.root.querySelector('#filters');
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

    protected getItems(filter: PostFilter, itemsPerPage: number, start: number): Promise<FoundPosts> {
        return this.api.findPosts(filter, Sort.NewestFirst, itemsPerPage, start);
    }

    protected getRowContents(item: FoundPost): string[] {
        return [
            item.Config.IsReadable ? item.Config.IsWritable ? 'open' : 'closed' : 'disabled',
            item.PostId
        ];
    }

    private async changeState(writable: boolean, readable: boolean) {
        let ids = this.gatherSelectedIds();
        await this.api.bulkUpdatePostConfigs(ids, writable, readable);
        this.loadList(0);
    }

    private gatherSelectedIds(): string[] {
        let items = this.gatherSelectedItems();
        let ids: string[] = [];
        for (let item of items) {
            ids.push(item.PostId);
        }
        return ids;
    }

}

