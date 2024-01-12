(function () {
    'use strict';

    function applyTemplate(element, map) {
        let elements = findPlaceholders(element);
        for (let elem of elements) {
            if (elem.nodeName == 'JTVAR') {
                replacePlaceholder(elem, map);
            }
            else if (elem.hasAttribute('jtvar')) {
                fillElement(elem, map);
            }
            else {
                fillAttributes(elem, map);
            }
        }
    }
    function findPlaceholders(root) {
        return new class {
            [Symbol.iterator]() {
                return new class {
                    iter;
                    constructor() {
                        this.iter = root.querySelectorAll('*').entries();
                    }
                    next() {
                        while (true) {
                            let next = this.iter.next();
                            if (next.done)
                                return { done: true, value: undefined };
                            const elem = next.value[1];
                            if (elem.nodeName == 'JTVAR' || elem.hasAttribute('jtvar'))
                                return { done: false, value: elem };
                            for (let attr of elem.attributes) {
                                if (attr.value.startsWith('jtvar '))
                                    return { done: false, value: elem };
                            }
                        }
                    }
                };
            }
        };
    }
    function replacePlaceholder(elem, map) {
        for (let attr of elem.attributes) {
            let replacer = map[attr.name];
            if (replacer === undefined)
                continue;
            if ("function" === typeof replacer) {
                replacer(elem);
            }
            else if ("string" === typeof replacer) {
                elem.insertAdjacentText("afterend", replacer);
            }
            else if ("object" === typeof replacer) {
                if (replacer.html !== undefined) {
                    elem.insertAdjacentHTML("afterend", replacer.html);
                }
                else if (replacer.text !== undefined) {
                    elem.insertAdjacentText("afterend", replacer.text);
                }
            }
            break;
        }
        elem.remove();
    }
    function fillElement(elem, map) {
        let name = elem.getAttribute('jtvar');
        if (!name)
            return;
        let replacer = map[name];
        if (replacer === undefined) {
            elem.removeAttribute('jtvar');
            return;
        }
        if ("function" === typeof replacer) {
            replacer(elem);
        }
        else if ("string" === typeof replacer) {
            elem.textContent = replacer;
        }
        else if ("boolean" === typeof replacer) {
            if (!replacer) {
                elem.remove();
                return;
            }
        }
        else if ("object" === typeof replacer) {
            if (replacer.html !== undefined) {
                elem.innerHTML = replacer.html;
            }
            else if (replacer.text !== undefined) {
                elem.textContent = replacer.text;
            }
            else if (replacer.visible !== undefined) {
                if (!replacer.visible) {
                    elem.remove();
                    return;
                }
            }
        }
        elem.removeAttribute('jtvar');
    }
    function fillAttributes(elem, map) {
        for (let attr of elem.attributes) {
            if (!attr.value.startsWith('jtvar '))
                continue;
            let name = attr.value.substring(6).trim();
            if (!name)
                continue;
            let replacer = map[name];
            if (replacer === undefined) {
                elem.removeAttribute(attr.name);
            }
            else if ("string" === typeof replacer) {
                elem.setAttribute(attr.name, replacer);
            }
            else if ("object" === typeof replacer && replacer.text !== undefined) {
                elem.setAttribute(attr.name, replacer.text);
            }
        }
    }

    var MessageType;
    (function (MessageType) {
        MessageType[MessageType["ErrorPostingComment"] = 0] = "ErrorPostingComment";
        MessageType[MessageType["CommentPostedAsDraft"] = 1] = "CommentPostedAsDraft";
    })(MessageType || (MessageType = {}));
    const Messages = {
        'en': {
            [MessageType.ErrorPostingComment]: 'There was an error while submitting the comment.',
            [MessageType.CommentPostedAsDraft]: 'Your comment was submitted and will become visible when it is approved.',
        },
        'es': {
            [MessageType.ErrorPostingComment]: 'Hubo un error enviando el comentario.',
            [MessageType.CommentPostedAsDraft]: 'Se ha recibido tu comentario y será publicado cuando se apruebe.',
        },
        'gl': {
            [MessageType.ErrorPostingComment]: 'Houbo un erro ao enviar o comentario.',
            [MessageType.CommentPostedAsDraft]: 'Recibiuse o teu comentario e vai ser publicado cando se aprobe.',
        }
    };
    const Templates = {
        'en': {
            'main': `
            <h1 jtvar="none_count">No comments</h1>
            <h1 jtvar="singular_count">1 comment</h1>
            <h1 jtvar="plural_count"><jtvar count></jtvar> comments</h1>
            <div jtvar="comments"></div>
            <div jtvar="newcomment"></div>
        `,
            'entry': `
            <p>By <jtvar author></jtvar> on <a href="jtvar url" name="jtvar anchor"><jtvar when></jtvar></a></p>
            <p jtvar="text"></p>
        `,
            'form': `
            <form id="commentform">
                <p>Your name: <input type="text" name="author"></p>
                <p>Comment: <textarea name="text" rows="10" cols="50"></textarea></p>
                <input type="submit" value="Submit"><input type="reset" value="Reset"><input type="button" value="Preview" id="jtPreviewButton">
            </form>
            <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
        `,
        },
        'gl': {
            'main': `
            <h1 jtvar="none_count">Ningún comentario</h1>
            <h1 jtvar="singular_count">1 comentario</h1>
            <h1 jtvar="plural_count"><jtvar count></jtvar> comentarios</h1>
            <div jtvar="comments"></div>
            <div jtvar="newcomment"></div>
        `,
            'entry': `
            <p>Por <jtvar author></jtvar> o <a href="jtvar url" name="jtvar anchor"><jtvar when></jtvar></a></p>
            <p jtvar="text"></p>
        `,
            'form': `
            <form id="commentform">
                <p>O teu nome: <input type="text" name="author"></p>
                <p>Comentario: <textarea name="text" rows="10" cols="50"></textarea></p>
                <input type="submit" value="Enviar"><input type="reset" value="Descartar"><input type="button" value="Previsualizar" id="jtPreviewButton">
            </form>
            <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
        `,
        },
        'es': {
            'main': `
            <h1 jtvar="none_count">Ningún comentario</h1>
            <h1 jtvar="singular_count">1 comentario</h1>
            <h1 jtvar="plural_count"><jtvar count></jtvar> comentarios</h1>
            <div jtvar="comments"></div>
            <div jtvar="newcomment"></div>
        `,
            'entry': `
            <p>Por <jtvar author></jtvar> el <a href="jtvar url" name="jtvar anchor"><jtvar when></jtvar></a></p>
            <p jtvar="text"></p>
        `,
            'form': `
        <form id="commentform">
            <p>Tu nombre: <input type="text" name="author"></p>
            <p>Comentario: <textarea name="text" rows="10" cols="50"></textarea></p>
            <input type="submit" value="Enviar"><input type="reset" value="Descartar"><input type="button" value="Previsualizar" id="jtPreviewButton">
        </form>
        <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
        `,
        },
    };

    function getLanguage() {
        let elem = document.body;
        while (elem != null && elem.lang == '') {
            elem = elem.parentElement;
        }
        let lang = elem ? elem.lang : '';
        let underline = lang.indexOf('_');
        if (underline > 0)
            return lang.substring(0, underline);
        return lang;
    }
    function getTemplate(name) {
        let templates = Templates[getLanguage()];
        if (!templates)
            templates = Templates['en'];
        return templates[name];
    }
    function getMessage(msg) {
        let msgs = Messages[getLanguage()];
        if (!msgs)
            msgs = Messages['en'];
        let out = msgs[msg];
        return out === undefined ? '[unknown message: ' + MessageType[msg] + ']' : out;
    }
    function formatDate(date) {
        let d = new Date(date);
        switch (getLanguage()) {
            case 'es':
                return (d.getDate() +
                    ' de ' + ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio',
                    'julio', 'agosto', 'setiembre', 'octubre', 'noviembre', 'diciembre'][d.getMonth()] +
                    ' de ' + d.getFullYear() +
                    ' a las ' + String(d.getHours()).padStart(2, '0') +
                    ':' + String(d.getMinutes()).padStart(2, '0'));
            case 'gl':
                return (d.getDate() +
                    ' de ' + ['xaneiro', 'febreiro', 'marzo', 'abril', 'maio', 'xuño',
                    'xullo', 'agosto', 'setembro', 'outubro', 'novembro', 'decembro'][d.getMonth()] +
                    ' de ' + d.getFullYear() +
                    ' ás ' + String(d.getHours()).padStart(2, '0') +
                    ':' + String(d.getMinutes()).padStart(2, '0'));
            case 'en':
            default:
                return (['January', 'February', 'March', 'April', 'May', 'June',
                    'July', 'August', 'September', 'October', 'November', 'December'][d.getMonth()] +
                    ' ' + d.getDate() +
                    ', ' + d.getFullYear() +
                    ' at ' + String(d.getHours()).padStart(2, '0') +
                    ':' + String(d.getMinutes()).padStart(2, '0'));
        }
    }

    function setup(params) {
        new Preview(params.api, params.input, params.output, params.container || params.output, params.toggle);
    }
    function findForm(element) {
        let current = element;
        while (current !== null && current.tagName != 'FORM')
            current = current.parentElement;
        return current;
    }
    class Preview {
        api;
        input;
        output;
        container;
        constructor(api, input, output, container, toggle) {
            this.api = api;
            this.input = input;
            this.output = output;
            this.container = container;
            this.form = findForm(this.input);
            this.previewFn = _ => this.launchPreview();
            this.resetPreviewFn = _ => this.resetPreview();
            this.timeout = undefined;
            this.lastPreview = ['', ''];
            if (toggle === null) {
                this.togglePreview();
            }
            else {
                toggle.addEventListener('click', _ => this.togglePreview());
            }
        }
        static PreviewInterval = 1000;
        form;
        previewFn;
        resetPreviewFn;
        timeout;
        lastPreview;
        visible() {
            return this.container.classList.contains('jtPreview');
        }
        togglePreview() {
            if (this.visible()) {
                this.input.removeEventListener('input', this.previewFn);
                this.form?.removeEventListener('reset', this.resetPreviewFn);
                while (true) {
                    let child = this.output.firstChild;
                    if (!child)
                        break;
                    child.remove();
                }
                this.container.classList.remove('jtPreview');
                return;
            }
            this.container.classList.add('jtPreview');
            this.input.addEventListener('input', this.previewFn);
            this.form?.addEventListener('reset', this.resetPreviewFn);
            this.doPreview();
        }
        launchPreview() {
            if (this.timeout !== undefined)
                return;
            this.timeout = setTimeout(() => this.doPreview(), Preview.PreviewInterval);
        }
        async doPreview() {
            if (!this.visible())
                return;
            let text = this.input.value;
            this.timeout = undefined;
            let preview = this.lastPreview[1];
            if (this.lastPreview[0] != text) {
                let result = await this.api.render(text);
                preview = result['Text'];
            }
            this.output.innerHTML = preview;
            this.lastPreview = [text, preview];
        }
        async resetPreview() {
            this.output.innerHTML = '';
        }
    }

    class UserApi {
        constructor() {
            this.apiUrl = findApiUrl();
        }
        apiUrl;
        async list(postId) {
            return post(this.apiUrl + '/list', { 'PostId': postId });
        }
        async add(newComment) {
            return post(this.apiUrl + '/add', newComment);
        }
        async render(text) {
            return post(this.apiUrl + '/render', { 'Text': text });
        }
    }
    var Sort;
    (function (Sort) {
        Sort[Sort["NewestFirst"] = 0] = "NewestFirst";
    })(Sort || (Sort = {}));
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

    const AnchorPrefix = 'comment_';
    class JtCommentsElement extends HTMLElement {
        api;
        postId;
        allTemplate;
        commentTemplate;
        formTemplate;
        constructor() {
            super();
            this.api = new UserApi();
        }
        connectedCallback() {
            this.postId = this.getAttribute('post-id');
            this.allTemplate = this.getTemplate();
            this.commentTemplate = this.getTemplate('entry');
            this.formTemplate = this.getTemplate('form');
            this.refresh();
        }
        async refresh() {
            while (this.firstChild != null) {
                this.removeChild(this.firstChild);
            }
            if (this.postId === null)
                return;
            this.render(await this.api.list(this.postId));
        }
        render(comments) {
            if (!comments.Config.IsReadable) {
                this.remove();
                return;
            }
            let numComments = comments.List.length;
            let block = this.allTemplate.cloneNode(true);
            applyTemplate(block, {
                'none_count': (numComments == 0),
                'singular_count': (numComments == 1),
                'plural_count': (numComments > 1),
                'count': String(numComments),
                'comments': (c) => { this.renderComments(c, comments); },
                'newcomment': comments.Config.IsWritable ? (c) => { this.renderForm(c); } : false,
            });
            this.appendChild(block);
            if (window.location.hash.startsWith('#' + AnchorPrefix)) {
                let anchor = window.location.hash.substring(1);
                let element = document.querySelector(`[name=${anchor}]`);
                if (element)
                    element.scrollIntoView();
            }
        }
        renderComments(list, comments) {
            for (let comment of comments.List) {
                let row = this.commentTemplate.cloneNode(true);
                applyTemplate(row, {
                    'author': comment.Author,
                    'when': formatDate(comment.When),
                    'url': new URL('#' + AnchorPrefix + comment.Id, window.location.toString()).toString(),
                    'anchor': AnchorPrefix + comment.Id,
                    'text': { html: comment.Text },
                });
                list.appendChild(row);
            }
        }
        renderForm(elem) {
            let form = this.formTemplate.cloneNode(true);
            form.querySelector('form')?.addEventListener('submit', e => {
                this.submitComment(e.target);
                e.preventDefault();
            });
            let commentBox = form.querySelector('#jtComment');
            let previewButton = form.querySelector('#jtPreviewButton');
            let previewBox = form.querySelector('#jtPreviewBox');
            let containerBox = form.querySelector('#jtPreviewContainer');
            if (commentBox && previewBox) {
                setup({
                    input: commentBox,
                    output: previewBox,
                    toggle: previewButton,
                    container: containerBox,
                    api: this.api
                });
            }
            elem.appendChild(form);
        }
        async submitComment(form) {
            let msg;
            let formData = new FormData(form);
            try {
                let comment = await this.api.add({
                    PostId: this.postId,
                    Author: formData.get('author'),
                    Text: formData.get('text'),
                });
                form.reset();
                if (comment.Visible) {
                    this.refresh();
                    return;
                }
                msg = MessageType.CommentPostedAsDraft;
            }
            catch (_) {
                msg = MessageType.ErrorPostingComment;
            }
            let p = document.createElement('p');
            p.classList.add("jtSubmitMessage");
            p.textContent = getMessage(MessageType.CommentPostedAsDraft);
            form.insertAdjacentElement("beforebegin", p);
            p.scrollIntoView();
        }
        getTemplate(id) {
            let name = 'jt-comments' + (id ? '-' + id : '');
            let template = document.getElementById(name);
            if (template) {
                template.remove();
                return template.content;
            }
            template = document.createElement('template');
            template.innerHTML = getTemplate(id ? id : 'main');
            return template.content;
        }
    }
    customElements.define('jt-comments', JtCommentsElement);

})();
