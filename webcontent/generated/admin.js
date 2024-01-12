(function () {
    'use strict';

    class AdminPage {
        root;
        constructor(root) {
            this.root = root;
            this.wireEvent('input[name=ApplyFilter]', 'click', _ => this.loadList(0));
            let list = this.getListTableElement();
            if (list) {
                this.wireEventFromRoot(list, 'thead input[type=checkbox]', 'change', e => this.toggleSelectAll(e));
            }
        }
        wireEventFromRoot(root, selector, event, handler) {
            let element = root.querySelector(selector);
            if (element)
                element.addEventListener(event, handler);
        }
        wireEvent(selector, event, handler) {
            this.wireEventFromRoot(this.root, selector, event, handler);
        }
        getListNameElement() {
            return this.root.querySelector('#listName');
        }
        getListTableElement() {
            return this.root.querySelector('#list');
        }
        getListLinksElement() {
            return this.root.querySelector('#listLinks');
        }
        getItemsPerPage() {
            let form = this.root.querySelector('#filters');
            if (form) {
                let items = form.querySelector('[name=ItemsPerPage');
                if (items)
                    return Number(items.value);
            }
            return 20;
        }
        createRow(columns) {
            let rowHtml = `<td><input type="checkbox"></td>`;
            for (let i = 0; i < columns; ++i) {
                rowHtml += `<td></td>`;
            }
            let row = document.createElement('tr');
            row.innerHTML = rowHtml;
            return row;
        }
        async loadList(start) {
            let filter = this.getFilter();
            let items = await this.getItems(filter, this.getItemsPerPage(), start);
            let listName = this.getListNameElement();
            if (listName)
                listName.textContent = this.getFilterTitle(filter);
            let listTable = this.getListTableElement()?.querySelector('tbody');
            if (listTable) {
                while (listTable.firstChild)
                    listTable.firstChild.remove();
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
                while (listLinks.firstChild)
                    listLinks.firstChild.remove();
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
        toggleSelectAll(e) {
            let boxes = this.getListTableElement()?.querySelectorAll('tbody input[type=checkbox]');
            for (let box of boxes) {
                box.checked = e.target.checked;
            }
        }
        gatherSelectedItems() {
            let items = [];
            let list = this.getListTableElement();
            if (!list)
                return items;
            let rows = list.querySelectorAll('tbody tr');
            for (let row of rows) {
                let checkbox = row.querySelector('input[type=checkbox]');
                if (checkbox.checked) {
                    items.push(row['item']);
                }
            }
            return items;
        }
    }

    var Sort;
    (function (Sort) {
        Sort[Sort["NewestFirst"] = 0] = "NewestFirst";
    })(Sort || (Sort = {}));
    class AdminApi {
        constructor() {
            this.apiUrl = findApiUrl();
        }
        apiUrl;
        async findComments(filter, sort, limit, start) {
            let params = {
                'Filter': filter,
                'Sort': sort,
                'Limit': limit,
                'Start': start,
            };
            return post(this.apiUrl + '/findComments', params);
        }
        async deleteComments(ids) {
            let params = {
                'Ids': Object.fromEntries(ids)
            };
            await post(this.apiUrl + '/deleteComments', params);
        }
        async findPosts(filter, sort, limit, start) {
            let params = {
                'Filter': filter,
                'Sort': sort,
                'Limit': limit,
                'Start': start,
            };
            return post(this.apiUrl + '/findPosts', params);
        }
        async bulkSetVisible(ids, visible) {
            let params = {
                'Ids': Object.fromEntries(ids),
                'Visible': visible
            };
            await post(this.apiUrl + '/bulkSetVisible', params);
        }
        async bulkUpdatePostConfigs(postIds, writable, readable) {
            let params = {
                'PostIds': postIds,
                'Config': {
                    'IsWritable': writable,
                    'IsReadable': readable,
                },
            };
            await post(this.apiUrl + '/bulkUpdatePostConfigs', params);
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

    function doAdminComments(rootElement) {
        new AdminComments(rootElement);
    }
    class AdminComments extends AdminPage {
        constructor(root) {
            super(root);
            this.api = new AdminApi();
            this.wireEvent('input[name=MakeVisible]', 'click', _ => this.changeVisible(true));
            this.wireEvent('input[name=MakeNonVisible]', 'click', _ => this.changeVisible(false));
            this.wireEvent('input[name=Delete]', 'click', _ => this.deleteComments());
            this.loadList(0);
        }
        api;
        getFilter() {
            let out = { Visible: null };
            let form = this.root.querySelector('#filters');
            if (!form)
                throw "Could not find filters box";
            let visible = form.querySelector('[name=Visible]').value;
            if (visible == 'true')
                out.Visible = true;
            if (visible == 'false')
                out.Visible = false;
            return out;
        }
        getFilterTitle(filter) {
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
        getItems(filter, itemsPerPage, start) {
            return this.api.findComments(filter, Sort.NewestFirst, itemsPerPage, start);
        }
        getRowContents(item) {
            return [
                item.Visible ? 'Yes' : 'No',
                item.PostId,
                item.Author,
                item.When,
                item.Text
            ];
        }
        async changeVisible(visible) {
            let ids = this.gatherSelectedIds();
            if (ids.size == 0)
                return;
            await this.api.bulkSetVisible(ids, visible);
            this.loadList(0);
        }
        async deleteComments() {
            let ids = this.gatherSelectedIds();
            if (ids.size == 0)
                return;
            await this.api.deleteComments(ids);
            this.loadList(0);
        }
        gatherSelectedIds() {
            let items = this.gatherSelectedItems();
            let ids = new Map();
            for (let item of items) {
                let postId = item.PostId;
                let commentId = item.CommentId;
                if (!ids.has(postId)) {
                    ids.set(postId, [commentId]);
                }
                else {
                    ids.get(postId).push(commentId);
                }
            }
            return ids;
        }
    }

    function doAdminPosts(rootElement) {
        new AdminPosts(rootElement);
    }
    class AdminPosts extends AdminPage {
        constructor(root) {
            super(root);
            this.api = new AdminApi();
            this.wireEvent('input[name=MakeOpen]', 'click', _ => this.changeState(true, true));
            this.wireEvent('input[name=MakeClosed]', 'click', _ => this.changeState(false, true));
            this.wireEvent('input[name=MakeDisabled]', 'click', _ => this.changeState(false, false));
            this.loadList(0);
        }
        api;
        getFilter() {
            let out = { CommentsReadable: null, CommentsWritable: null };
            let form = this.root.querySelector('#filters');
            if (!form)
                throw "Could not find filters box";
            let state = form.querySelector('[name=State]').value;
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
        getFilterTitle(filter) {
            if (filter.CommentsWritable === true) {
                return 'Latest posts with open comments';
            }
            else if (filter.CommentsReadable === true) {
                return 'Latest posts with closed comments';
            }
            else if (filter.CommentsReadable === false) {
                return 'Latest posts with disabled comments';
            }
            else {
                return 'Latest posts';
            }
        }
        getItems(filter, itemsPerPage, start) {
            return this.api.findPosts(filter, Sort.NewestFirst, itemsPerPage, start);
        }
        getRowContents(item) {
            return [
                item.Config.IsReadable ? item.Config.IsWritable ? 'open' : 'closed' : 'disabled',
                item.PostId
            ];
        }
        async changeState(writable, readable) {
            let ids = this.gatherSelectedIds();
            await this.api.bulkUpdatePostConfigs(ids, writable, readable);
            this.loadList(0);
        }
        gatherSelectedIds() {
            let items = this.gatherSelectedItems();
            let ids = [];
            for (let item of items) {
                ids.push(item.PostId);
            }
            return ids;
        }
    }

    function initTabs(tabTitle, tabPage) {
        new Tabs(tabTitle, tabPage);
    }
    class Tabs {
        tabTitle;
        tabPage;
        constructor(tabTitle, tabPage) {
            this.tabTitle = tabTitle;
            this.tabPage = tabPage;
            this.titles = this.findElements(this.tabTitle);
            this.pages = this.findElements(this.tabPage);
            for (let title of this.titles) {
                title.addEventListener('click', e => this.switchTab(e));
            }
        }
        titles;
        pages;
        getNum(elem, name) {
            let id = elem.id;
            if (!id.startsWith(name + '-'))
                return null;
            return Number(id.substring(name.length + 1));
        }
        findElements(name) {
            let elems = document.querySelectorAll(`[id|=${name}]`);
            let out = [];
            for (let elem of elems) {
                let num = this.getNum(elem, name);
                if (num != null) {
                    let tabElem = elem;
                    tabElem.tabNum = num;
                    out.push(tabElem);
                }
            }
            return out;
        }
        switchTab(e) {
            let elem = e.target;
            if (elem.tabNum !== undefined)
                this.selectTab(elem.tabNum);
        }
        selectTab(num) {
            for (let title of this.titles) {
                if (title.tabNum == num) {
                    title.classList.add('tabSelected');
                }
                else {
                    title.classList.remove('tabSelected');
                }
            }
            for (let page of this.pages) {
                if (page.tabNum == num) {
                    page.classList.add('tabOpen');
                }
                else {
                    page.classList.remove('tabOpen');
                }
            }
        }
    }

    window.addEventListener('load', _ => doAdmin());
    function doAdmin() {
        initTabs('tabTitle', 'tabPage');
        doAdminComments(document.getElementById('tabPage-0'));
        doAdminPosts(document.getElementById('tabPage-1'));
    }

})();
