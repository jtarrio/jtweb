(function () {
    'use strict';

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
        'en': `<div class="comment">
    <jv-if cond="has_none_count"><h1>No comments</h1></jv-if>
    <jv-if cond="has_singular_count"><h1>1 comment</h1></jv-if>
    <jv-if cond="has_plural_count"><h1><jv>count</jv> comments</h1></jv-if>
    <jv-for items="comments" item="comment">
        <div class="commentByline">By <jv>comment.author</jv> on <a jv-href="comment.url" jv-name="comment.anchor"><jv date>comment.when</jv></a></div>
        <div class="commentText><jv html>comment.text</jv></div>
    </jv-for>
    <jv-if cond="can_add_comment">
        <form id="commentform">
            <h1>Would you like to write a comment?</h1>
            <div class="commentForm">
                <div>Your name or nickname: <input type="text" name="author"> (it will be published)</div>
                <div>Your comment:</div>
                <textarea name="text" id="jtComment"></textarea>
                <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
            </div>
            <div class="commentButtons">
                <input type="reset" value="Reset"><input type="submit" value="Submit"><input type="button" value="Preview" id="jtPreviewButton">
            </div>
        </form>
    </jv-if>
</div>`,
        'gl': `<div class="comment">
    <jv-if cond="has_none_count"><h1>Ningún comentario</h1></jv-if>
    <jv-if cond="has_singular_count"><h1>1 comentario</h1></jv-if>
    <jv-if cond="has_plural_count"><h1><jv>count</jv> comentarios</h1></jv-if>
    <jv-for items="comments" item="comment">
        <div class="commentByline">Por <jv>comment.author</jv> o <a jv-href="comment.url" jv-name="comment.anchor"><jv date>comment.when</jv></a></div>
        <div class="commentText><jv html>comment.text</jv></div>
    </jv-for>
    <jv-if cond="can_add_comment">
        <form id="commentform">
            <h1>Queres escribir un comentario?</h1>
            <div class="commentForm">
                <div>O teu nome ou sobrenome: <input type="text" name="author"> (hase publicar)</div>
                <div>O teu comentario:</div>
                <textarea name="text" id="jtComment"></textarea>
                <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
            </div>
            <div class="commentButtons">
                <input type="reset" value="Descartar"><input type="submit" value="Enviar"><input type="button" value="Previsualizar" id="jtPreviewButton">
            </div>
        </form>
    </jv-if>
</div>`,
        'es': `<div class="comment">
    <jv-if cond="has_none_count"><h1>Ningún comentario</h1></jv-if>
    <jv-if cond="has_singular_count"><h1>1 comentario</h1></jv-if>
    <jv-if cond="has_plural_count"><h1><jv>count</jv> comentarios</h1></jv-if>
    <jv-for items="comments" item="comment">
        <div class="commentByline">Por <jv>comment.author</jv> el <a jv-href="comment.url" jv-name="comment.anchor"><jv date>comment.when</jv></a></div>
        <div class="commentText><jv html>comment.text</jv></div>
    </jv-for>
    <jv-if cond="can_add_comment">
        <form id="commentform">
            <h1>¿Quieres escribir un comentario?</h1>
            <div class="commentForm">
                <div>Tu nombre o sobrenombre: <input type="text" name="author"> (se publicará)</div>
                <div>Tu comentario:</div>
                <textarea name="text" id="jtComment"></textarea>
                <div id="jtPreviewContainer"><div id="jtPreviewBox"></div></div>
            </div>
            <div class="commentButtons">
                <input type="reset" value="Descartar"><input type="submit" value="Enviar"><input type="button" value="Previsualizar" id="jtPreviewButton">
            </div>
        </form>
    </jv-if>
</div>`,
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
    function getTemplate() {
        let template = Templates[getLanguage()];
        if (!template)
            template = Templates['en'];
        return template;
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

    // <p>Some text <jv>varName</jv></p>
    // <p>Parse date <jv date>varName</jv></p>
    // <p>Insert as html <jv html>varName</jv></p>
    // <p>Nested vars <jv>varName.field1.field2</jv>
    // <a jv-href="varName">...content...</a>
    // <jv-if cond="varName">...content...</jv-if>
    // <jv-if not cond="varName">...content...</jv-if>
    // <jv-for items="varName" item="itemVarName">...content...</jv-for>
    // <jv-for items="varName" item="itemVarName" index="indexVarName">...content...</jv-for>
    function applyTemplate(element, map) {
        if (element.nodeName == 'JV') {
            replaceElement(element, map);
        }
        else if (element.nodeName == 'JV-IF') {
            doIf(element, map);
        }
        else if (element.nodeName == 'JV-FOR') {
            doFor(element, map);
        }
        else {
            replaceAttributes(element, map);
        }
    }
    function getMapValue(name, map) {
        let scope = map;
        if (name === null || name === undefined)
            return undefined;
        while (true) {
            if (name == '')
                return scope;
            if ('object' !== typeof scope)
                return undefined;
            let dot = name.indexOf('.');
            let index = dot == -1 ? name : name.substring(0, dot);
            name = dot == -1 ? '' : name.substring(dot + 1);
            let newScope = undefined;
            if (Array.isArray(scope)) {
                let indexNum = Number(index);
                if (!Number.isNaN(indexNum)) {
                    newScope = scope[indexNum];
                }
            }
            if (newScope === undefined)
                newScope = scope[index];
            scope = newScope;
        }
    }
    function applyTemplateToChildren(parent, map) {
        let child = parent.firstElementChild;
        while (child != null) {
            let next = child.nextElementSibling;
            applyTemplate(child, map);
            child = next;
        }
    }
    function replaceElement(element, map) {
        let replacement = getMapValue(element.textContent?.trim(), map);
        if (element.hasAttribute('html')) {
            element.outerHTML = String(replacement);
        }
        else if (element.hasAttribute('date')) {
            element.replaceWith(formatDate(String(replacement)));
        }
        else {
            element.replaceWith(String(replacement));
        }
    }
    function doIf(element, map) {
        let cond = Boolean(getMapValue(element.getAttribute('cond'), map));
        if (element.hasAttribute('not')) {
            cond = !cond;
        }
        if (cond) {
            applyTemplateToChildren(element, map);
            while (element.firstChild != null)
                element.before(element.firstChild);
        }
        element.remove();
    }
    function doFor(element, map) {
        let items = getMapValue(element.getAttribute('items'), map);
        let itemVarName = element.getAttribute('item');
        let indexVarName = element.getAttribute('index');
        for (let index in items) {
            let item = items[index];
            let childMap = { ...map };
            if (itemVarName)
                childMap[itemVarName] = item;
            if (indexVarName)
                childMap[indexVarName] = index;
            let child = element.firstElementChild;
            while (child != null) {
                let next = child.nextElementSibling;
                let clone = child.cloneNode(true);
                element.before(clone);
                applyTemplate(clone, childMap);
                child = next;
            }
        }
        element.remove();
    }
    function replaceAttributes(element, map) {
        let toDelete = [];
        for (let attr of element.attributes || []) {
            if (!attr.name.startsWith('jv-'))
                continue;
            let attrName = attr.name.substring(3);
            let varName = attr.value;
            let replacement = getMapValue(varName, map);
            toDelete.push(attr.name);
            if ('boolean' === typeof replacement) {
                if (replacement) {
                    element.setAttribute(attrName, '');
                }
            }
            else {
                element.setAttribute(attrName, String(replacement));
            }
        }
        toDelete.forEach(name => element.removeAttribute(name));
        applyTemplateToChildren(element, map);
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
        async list(postId) {
            return post('/list', { 'PostId': postId });
        }
        async add(newComment) {
            return post('/add', newComment);
        }
        async render(text) {
            return post('/render', { 'Text': text });
        }
    }
    var Sort;
    (function (Sort) {
        Sort[Sort["NewestFirst"] = 0] = "NewestFirst";
    })(Sort || (Sort = {}));
    async function post(url, data) {
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
        }
        else {
            baseUrl.pathname = pathname.substring(0, lastSlash) + '_';
        }
        return baseUrl.toString();
    })();

    const AnchorPrefix = 'comment_';
    class JtCommentsElement extends HTMLElement {
        api;
        postId;
        allTemplate;
        constructor() {
            super();
            this.api = new UserApi();
        }
        connectedCallback() {
            this.postId = this.getAttribute('post-id');
            this.allTemplate = this.getTemplate();
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
            if (!comments.Config.IsReadable || (!comments.Config.IsWritable && comments.List.length == 0)) {
                this.remove();
                return;
            }
            let numComments = comments.List.length;
            let renderedComments = comments.List.map(c => ({
                author: c.Author,
                when: c.When,
                url: new URL('#' + AnchorPrefix + c.Id, window.location.toString()).toString(),
                anchor: AnchorPrefix + c.Id,
                text: c.Text,
            }));
            let block = this.allTemplate.cloneNode(true);
            applyTemplate(block, {
                'has_none_count': (numComments == 0),
                'has_singular_count': (numComments == 1),
                'has_plural_count': (numComments > 1),
                'can_add_comment': comments.Config.IsWritable,
                'might_have_comments': comments.Config.IsReadable && (comments.Config.IsWritable || comments.List.length > 0),
                'count': String(numComments),
                'comments': renderedComments,
            });
            if (comments.Config.IsWritable) {
                this.attachFormEvents(block);
            }
            this.appendChild(block);
            if (window.location.hash.startsWith('#' + AnchorPrefix)) {
                let anchor = window.location.hash.substring(1);
                let element = document.querySelector(`[name=${anchor}]`);
                if (element)
                    element.scrollIntoView();
            }
        }
        attachFormEvents(form) {
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
            catch {
                msg = MessageType.ErrorPostingComment;
            }
            let p = document.createElement('p');
            p.classList.add("jtSubmitMessage");
            p.textContent = getMessage(msg);
            form.insertAdjacentElement("beforebegin", p);
            p.scrollIntoView();
        }
        getTemplate() {
            let template = this.getElementsByTagName('template')[0];
            if (template) {
                template.remove();
                return template.content;
            }
            template = document.createElement('template');
            template.innerHTML = getTemplate();
            return template.content;
        }
    }
    window.addEventListener('DOMContentLoaded', _ => customElements.define('jt-comments', JtCommentsElement));

})();
