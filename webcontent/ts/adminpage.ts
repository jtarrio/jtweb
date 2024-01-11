interface ItemList<T> {
    List: T[],
    More: boolean,
}

export abstract class AdminPage<Type, Filter> {
    constructor(protected root: HTMLElement) {
        let list = this.getListTableElement();
        if (list) {
        this.wireEventFromRoot(list, 'thead input[type=checkbox]', 'change', e => this.toggleSelectAll(e));
        }
    }

    protected wireEventFromRoot(root: HTMLElement, selector: string, event: string, handler: EventListenerOrEventListenerObject) {
        let element = root.querySelector(selector);
        if (element) element.addEventListener(event, handler);
    }

    protected wireEvent(selector: string, event: string, handler: EventListenerOrEventListenerObject) {
        this.wireEventFromRoot(this.root, selector, event, handler);
    }

    protected abstract getFilter() : Filter;
    protected abstract getFilterTitle(filter: Filter): string;
    protected abstract getItemsPerPage(): number;
    protected abstract getItems(filter: Filter, itemsPerPage: number, start: number) : Promise<ItemList<Type>>;
    protected abstract getRowContents(item: Type): string[];
    protected abstract getListNameElement(): HTMLElement | null;
    protected abstract getListTableElement(): HTMLElement | null;
    protected abstract getListLinksElement(): HTMLElement | null;

    protected createRow(columns: number) : HTMLTableRowElement {
        let rowHtml =  `<td><input type="checkbox"></td>`;
        for (let i = 0; i < columns; ++i) {
            rowHtml += `<td></td>`;
        }
        let row = document.createElement('tr');
        row.innerHTML = rowHtml;
        return row;
    }

    async loadList(start: number) {
        let filter = this.getFilter();
        let items = await this.getItems(filter, this.getItemsPerPage(), start);
        let listName = this.getListNameElement();
        if (listName) listName.textContent = this.getFilterTitle(filter);
        let listTable = this.getListTableElement()?.querySelector('tbody');
        if (listTable) {
            while (listTable.firstChild) listTable.firstChild.remove();
            for (let item of items.List) {
                let row = document.createElement('tr');
                row['item'] = item;

                let td = document.createElement('td');
                td.innerHTML = `<input type="checkbox">`;
                row.appendChild(td);
                for (let col of this.getRowContents(item)) {
                    let td = document.createElement('td');
                    td.textContent = col;
                    row.appendChild(td);
                }
                listTable.appendChild(row);
            }
        }
        let listLinks = this.getListLinksElement();
        if (listLinks) {
            while (listLinks.firstChild) listLinks.firstChild.remove();
            if (items.More) {
                let link = document.createElement('a');
                link.href = "javascript:0";
                link.textContent = "Next";
                link.addEventListener('click', _ => this.loadList(start + this.getItemsPerPage()));
                listLinks.appendChild(link);
                if (items.More) {
                    listLinks.insertAdjacentText('beforeend', ' ');
                }
            }
            if (start > 0) {
                let link = document.createElement('a');
                link.href = "javascript:0";
                link.textContent = "Previous";
                link.addEventListener('click', _ => this.loadList(start - this.getItemsPerPage()));
                listLinks.appendChild(link);
            }
        }
    }

    toggleSelectAll(e: Event) {
        let boxes = this.getListTableElement()?.querySelectorAll('tbody input[type=checkbox]') as NodeListOf<HTMLInputElement>;
        for (let box of boxes) {
            box.checked = (e.target! as HTMLInputElement).checked;
        }    
    }

    gatherSelectedItems(): Type[] {
        let items: Type[] = [];
        let list = this.getListTableElement();
        if (!list) return items;
        let rows = list.querySelectorAll('tbody tr') as NodeListOf<HTMLTableRowElement>;
        for (let row of rows) {
            let checkbox = row.querySelector('input[type=checkbox]') as HTMLInputElement;
            if (checkbox.checked) {
                items.push(row['item']);
            }
        }
        return items;
    }
}
