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
                <input type="submit" value="Submit"><input type="reset" value="Reset">
            </form>
        `,
        },
        'gl': {
            'main': `
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
                <input type="submit" value="Enviar"><input type="reset" value="Descartar">
            </form>
        `,
        },
        'es': {
            'main': `
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
            <input type="submit" value="Enviar"><input type="reset" value="Descartar">
        </form>
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
                    'julio', 'agosto', 'setiembre', 'octubre', 'noviembre', 'diciembre'][d.getMonth() - 1] +
                    ' de ' + d.getFullYear() +
                    ' a las ' + String(d.getHours()).padStart(2, '0') +
                    ':' + String(d.getMinutes()).padStart(2, '0'));
            case 'gl':
                return (d.getDate() +
                    ' de ' + ['xaneiro', 'febreiro', 'marzo', 'abril', 'maio', 'xuño',
                    'xullo', 'agosto', 'setembro', 'outubro', 'novembro', 'decembro'][d.getMonth() - 1] +
                    ' de ' + d.getFullYear() +
                    ' ás ' + String(d.getHours()).padStart(2, '0') +
                    ':' + String(d.getMinutes()).padStart(2, '0'));
            case 'en':
            default:
                return (['January', 'February', 'March', 'April', 'May', 'June',
                    'July', 'August', 'September', 'October', 'November', 'December'][d.getMonth() - 1] +
                    ' ' + d.getDate() +
                    ', ' + d.getFullYear() +
                    ' at ' + String(d.getHours()).padStart(2, '0') +
                    ':' + String(d.getMinutes()).padStart(2, '0'));
        }
    }

    class JtCommentsElement extends HTMLElement {
        apiUrl;
        postId;
        allTemplate;
        commentTemplate;
        formTemplate;
        constructor() {
            super();
            let scripts = document.getElementsByTagName('script');
            let baseUrl = new URL(scripts[scripts.length - 1].attributes['src'].value, window.location.toString());
            const suffix = "comments.js";
            if (baseUrl.pathname.endsWith(suffix)) {
                baseUrl.pathname = baseUrl.pathname.substring(0, baseUrl.pathname.length - suffix.length);
            }
            if (!baseUrl.pathname.endsWith('/')) {
                baseUrl.pathname += '/';
            }
            this.apiUrl = baseUrl.pathname += '_';
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
            let response = await fetch(this.apiUrl + '/list/' + this.postId);
            if (response.status == 200) {
                this.render(await response.json());
            }
        }
        render(comments) {
            if (!comments.IsAvailable) {
                this.remove();
                return;
            }
            let numComments = comments.List.length;
            let block = this.allTemplate.cloneNode(true);
            applyTemplate(block, {
                'singular_count': (numComments == 1),
                'plural_count': (numComments != 1),
                'count': String(numComments),
                'comments': (c) => { this.renderComments(c, comments); },
                'newcomment': comments.IsWritable ? (c) => { this.renderForm(c); } : false,
            });
            this.appendChild(block);
        }
        renderComments(list, comments) {
            for (let comment of comments.List) {
                let row = this.commentTemplate.cloneNode(true);
                applyTemplate(row, {
                    'author': comment.Author,
                    'when': formatDate(comment.When),
                    'url': new URL('#c' + comment.Id, window.location.toString()).toString(),
                    'anchor': 'c' + comment.Id,
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
            elem.appendChild(form);
        }
        async submitComment(form) {
            let formData = new FormData(form);
            let data = JSON.stringify({
                'PostId': this.postId,
                'Author': formData.get('author'),
                'Text': formData.get('text'),
            });
            let response = await fetch(this.apiUrl + '/add', { method: "POST", body: data });
            MessageType.ErrorPostingComment;
            if (response.status == 200) {
                form.reset();
                let comment = await response.json();
                if (comment.Visible) {
                    this.refresh();
                    return;
                }
                MessageType.CommentPostedAsDraft;
            }
            let p = document.createElement('p');
            p.textContent = getMessage(MessageType.CommentPostedAsDraft);
            form.insertAdjacentElement("beforebegin", p);
        }
        getTemplate(id) {
            let name = 'jt-comments' + (id ? '-' + id : '');
            let template = document.getElementById(name);
            if (template)
                return template.content;
            template = document.createElement('template');
            template.innerHTML = getTemplate(id ? id : 'main');
            return template.content;
        }
    }
    customElements.define('jt-comments', JtCommentsElement);

})();
