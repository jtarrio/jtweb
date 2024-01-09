import {AdminApi, Filter, Sort} from './api';

var api = new AdminApi();
window.addEventListener('load', _ => doAdmin());

const COMMENTS_PER_PAGE = 20;

function doAdmin() {
    wireEvents();
    loadList(0);
}

function wireEvents() {
    wireEvent('#cmtList thead input[type=checkbox]', 'change', toggleSelectAll);
    wireEvent('input[name=ApplyFilter]', 'click', _ => loadList(0));
    wireEvent('input[name=MakeVisible]', 'click', makeVisible);
    wireEvent('input[name=MakeNonVisible]', 'click', makeNonVisible);
}

function wireEvent(selector: string, event: string, handler: EventListenerOrEventListenerObject) {
    let element = document.querySelector(selector);
    if (element) element.addEventListener(event, handler);
}

async function loadList(start: number) {
    let filter = getFilter();
    let comments = await api.find(filter, Sort.NewestFirst, COMMENTS_PER_PAGE, start);
    let listName = document.getElementById('cmtListName');
    if (listName) listName.textContent = getFilterTitle(filter);
    let listTable = document.querySelector('#cmtList tbody');
    if (listTable) {
        while (listTable.firstChild) listTable.firstChild.remove();
        for (let cmt of comments.List) {
            let row = document.createElement('tr');
            row['comment'] = cmt;
            // select visible post author when text
            row.innerHTML = `<td><input type="checkbox"></td><td></td><td></td><td></td><td></td><td></td>`;
            let cells = row.querySelectorAll('td');
            cells[1].textContent = cmt.Comment.Visible ? 'Yes' : 'No';
            cells[2].textContent = cmt.PostId;
            cells[3].textContent = cmt.Comment.Author;
            cells[4].textContent = cmt.Comment.When;
            cells[5].textContent = cmt.Comment.Text;
            listTable.appendChild(row);
        }
    }
    let listLinks = document.getElementById('cmtListLinks');
    if (listLinks) {
        while (listLinks.firstChild) listLinks.firstChild.remove();
        if (start > 0) {
            let link = document.createElement('a');
            link.href="javascript:0";
            link.textContent = "Previous";
            link.addEventListener('click', _ => loadList(start - COMMENTS_PER_PAGE));
            listLinks.appendChild(link);
            if (comments.More) {
                listLinks.insertAdjacentText('beforeend', ' ');
            }
        }
        if (comments.More) {
            let link = document.createElement('a');
            link.href="javascript:0";
            link.textContent = "Next";
            link.addEventListener('click', _ => loadList(start + COMMENTS_PER_PAGE));
            listLinks.appendChild(link);
        }
    }

}

function toggleSelectAll(e: Event) {
    let boxes = document.querySelectorAll('#cmtList tbody input[type=checkbox]') as NodeListOf<HTMLInputElement>;
    for (let box of boxes) {
        box.checked = (e.target! as HTMLInputElement).checked;
    }
}

function gatherSelectedIds(): Map<string, string[]> {
    let ids = new Map<string, string[]>();
    let rows = document.querySelectorAll('#cmtList tbody tr') as NodeListOf<HTMLTableRowElement>;
    for (let row of rows) {
        let checkbox = row.querySelector('input[type=checkbox]') as HTMLInputElement | null;
        if (checkbox?.checked) {
            let comment = row['comment'];
            let postId = comment.PostId;
            let commentId = comment.Comment.Id;
            if (!ids.has(postId)) ids.set(postId, []);
            ids.get(postId)!.push(commentId);
        }
    }
    return ids;
}

async function makeVisible() {
    let ids = gatherSelectedIds();
    if (ids.size == 0) return;
    await api.setVisible(ids, true);
    loadList(0)
}

async function makeNonVisible() {
    let ids = gatherSelectedIds();
    if (ids.size == 0) return;
    await api.setVisible(ids, false);
    loadList(0)
}

function getFilter(): Filter {
    let out : Filter = {Visible: null};
    let form = document.getElementById('filters');
    if (!form) throw "Could not find filters box";
    let visible = (form.querySelector('[name=Visible]') as HTMLSelectElement).value;
    if (visible == 'true') out.Visible = true;
    if (visible == 'false') out.Visible = false;
    return out;
}

function getFilterTitle(filter: Filter): string {
    if (filter.Visible === null) {
        return 'Latest comments';
    } else if (filter.Visible) {
        return 'Latest visible comments';
    } else {
        return 'Latest non-visible comments';
    }
}