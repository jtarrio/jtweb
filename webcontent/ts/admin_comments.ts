import { AdminPage } from './adminpage';
import { AdminApi, CommentFilter, RawComment, FoundComments, Sort } from './api';

export function doAdminComments(rootElement: HTMLElement) {
    new AdminComments(rootElement);
}

class AdminComments extends AdminPage<RawComment, CommentFilter> {
    constructor(root: HTMLElement) {
        super(root);
        this.api = new AdminApi();
        this.wireEvent('input[name=ApplyCommentFilter]', 'click', _ => this.loadList(0));
        this.wireEvent('input[name=MakeVisible]', 'click', _ => this.makeVisible());
        this.wireEvent('input[name=MakeNonVisible]', 'click', _ => this.makeNonVisible());
        this.loadList(0);
    }

    api: AdminApi;

    protected getFilter(): CommentFilter {
        let out: CommentFilter = { Visible: null };
        let form = this.root.querySelector('#cmtFilters');
        if (!form) throw "Could not find filters box";
        let visible = (form.querySelector('[name=Visible]') as HTMLSelectElement).value;
        if (visible == 'true') out.Visible = true;
        if (visible == 'false') out.Visible = false;
        return out;
    }

    protected getFilterTitle(filter: CommentFilter): string {
        if (filter.Visible === null) {
            return 'Latest comments';
        } else if (filter.Visible) {
            return 'Latest visible comments';
        } else {
            return 'Latest non-visible comments';
        }
    }

    protected getItemsPerPage(): number {
        let form = this.root.querySelector('#cmtFilters');
        if (form) {
            let items = form.querySelector('[name=ItemsPerPage') as HTMLSelectElement | null;
            if (items) return Number(items.value);
        }
        return 20;
    }

    protected getItems(filter: CommentFilter, itemsPerPage: number, start: number): Promise<FoundComments> {
        return this.api.findComments(filter, Sort.NewestFirst, itemsPerPage, start);
    }

    protected getRowContents(item: RawComment): string[] {
        return [
            item.Visible ? 'Yes' : 'No',
            item.PostId,
            item.Author,
            item.When,
            item.Text
        ];
    }

    protected getListNameElement(): HTMLElement | null {
        return this.root.querySelector('#cmtListName');
    }

    protected getListTableElement(): HTMLElement | null {
        return this.root.querySelector('#cmtList');
    }

    protected getListLinksElement(): HTMLElement | null {
        return this.root.querySelector('#cmtListLinks');
    }

    async makeVisible() {
        let ids = this.gatherSelectedIds();
        if (ids.size == 0) return;
        await this.api.setVisible(ids, true);
        this.loadList(0)
    }

    async makeNonVisible() {
        let ids = this.gatherSelectedIds();
        if (ids.size == 0) return;
        await this.api.setVisible(ids, false);
        this.loadList(0)
    }

    gatherSelectedIds(): Map<string, string[]> {
        let items = this.gatherSelectedItems();
        let ids = new Map<string, string[]>();
        for (let item of items) {
            let postId = item.PostId;
            let commentId = item.CommentId;
            if (!ids.has(postId)) {
                ids.set(postId, [commentId]);
            } else {
                ids.get(postId)!.push(commentId);
            }
        }
        return ids;
    }
}
