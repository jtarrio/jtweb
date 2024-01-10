(function () {
    'use strict';

    var Sort;
    (function (Sort) {
        Sort[Sort["NewestFirst"] = 0] = "NewestFirst";
    })(Sort || (Sort = {}));
    class AdminApi {
        constructor() {
            this.apiUrl = findApiUrl();
        }
        apiUrl;
        async find(filter, sort, limit, start) {
            let params = {
                'Filter': filter,
                'Sort': sort,
                'Limit': limit,
                'Start': start,
            };
            return post(this.apiUrl + '/find', params);
        }
        async setVisible(ids, visible) {
            let params = {
                'Ids': Object.fromEntries(ids),
                'Visible': visible
            };
            await post(this.apiUrl + '/setVisible', params);
        }
    }
    async function post(url, data) {
        let response = await fetch(url, { method: 'POST', mode: 'cors', body: JSON.stringify(data) });
        if (response.status != 200) {
            throw `Error ${response.status}: ${await response.text()}`;
        }
        return response.json();
    }
    function findApiUrl() {
        let scripts = document.getElementsByTagName('script');
        let baseUrl = new URL(scripts[scripts.length - 1].attributes['src'].value, window.location.toString());
        let pathname = baseUrl.pathname;
        let lastSlash = pathname.lastIndexOf('/');
        if (lastSlash === undefined) {
            baseUrl.pathname = '/_';
        }
        else {
            baseUrl.pathname = pathname.substring(0, lastSlash) + '_';
        }
        return baseUrl.toString();
    }

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
    function wireEvent(selector, event, handler) {
        let element = document.querySelector(selector);
        if (element)
            element.addEventListener(event, handler);
    }
    async function loadList(start) {
        let filter = getFilter();
        let comments = await api.find(filter, Sort.NewestFirst, COMMENTS_PER_PAGE, start);
        let listName = document.getElementById('cmtListName');
        if (listName)
            listName.textContent = getFilterTitle(filter);
        let listTable = document.querySelector('#cmtList tbody');
        if (listTable) {
            while (listTable.firstChild)
                listTable.firstChild.remove();
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
            while (listLinks.firstChild)
                listLinks.firstChild.remove();
            if (start > 0) {
                let link = document.createElement('a');
                link.href = "javascript:0";
                link.textContent = "Previous";
                link.addEventListener('click', _ => loadList(start - COMMENTS_PER_PAGE));
                listLinks.appendChild(link);
                if (comments.More) {
                    listLinks.insertAdjacentText('beforeend', ' ');
                }
            }
            if (comments.More) {
                let link = document.createElement('a');
                link.href = "javascript:0";
                link.textContent = "Next";
                link.addEventListener('click', _ => loadList(start + COMMENTS_PER_PAGE));
                listLinks.appendChild(link);
            }
        }
    }
    function toggleSelectAll(e) {
        let boxes = document.querySelectorAll('#cmtList tbody input[type=checkbox]');
        for (let box of boxes) {
            box.checked = e.target.checked;
        }
    }
    function gatherSelectedIds() {
        let ids = new Map();
        let rows = document.querySelectorAll('#cmtList tbody tr');
        for (let row of rows) {
            let checkbox = row.querySelector('input[type=checkbox]');
            if (checkbox?.checked) {
                let comment = row['comment'];
                let postId = comment.PostId;
                let commentId = comment.Comment.Id;
                if (!ids.has(postId))
                    ids.set(postId, []);
                ids.get(postId).push(commentId);
            }
        }
        return ids;
    }
    async function makeVisible() {
        let ids = gatherSelectedIds();
        if (ids.size == 0)
            return;
        await api.setVisible(ids, true);
        loadList(0);
    }
    async function makeNonVisible() {
        let ids = gatherSelectedIds();
        if (ids.size == 0)
            return;
        await api.setVisible(ids, false);
        loadList(0);
    }
    function getFilter() {
        let out = { Visible: null };
        let form = document.getElementById('filters');
        if (!form)
            throw "Could not find filters box";
        let visible = form.querySelector('[name=Visible]').value;
        if (visible == 'true')
            out.Visible = true;
        if (visible == 'false')
            out.Visible = false;
        return out;
    }
    function getFilterTitle(filter) {
        if (filter.Visible === null) {
            return 'Latest comments';
        }
        else if (filter.Visible) {
            return 'Latest visible comments';
        }
        else {
            return 'Latest non-visible comments';
        }
    }

})();
