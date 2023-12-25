type Comments = {
    PostId: string,
    IsAvailable: boolean,
    IsWritable: boolean,
    List: [{
        Id: string,
        Author: string,
        When: string,
        Text: string
    }]
};

(() => {
    const defaultTemplates = {
        'comments':
            `<h1 id="singular_count">1 comment</h1>
             <h1 id="plural_count"><span id="count"></span> comments</h1>
             <div id="comments"></div>
             <div id="entryform">
                 (Entry form)
             </div>
             <div id="noentryform">
                 (No entry form)
             </div>`,
        'comment':
            `<p>By <span id="author"></span> on <span id="when"></span></p>
             <p id="text"></p>`,
    };

    let scripts = document.getElementsByTagName('script');
    let baseUrl = new URL(scripts[scripts.length - 1].attributes['src'].value, window.location.toString());
    const suffix = "comments.js";
    if (baseUrl.pathname.endsWith(suffix)) {
        baseUrl.pathname = baseUrl.pathname.substring(0, baseUrl.pathname.length - suffix.length);
    }
    if (!baseUrl.pathname.endsWith('/')) {
        baseUrl.pathname += '/';
    }
    baseUrl.pathname += '_';

    customElements.define('jt-comments', class extends HTMLElement {
        constructor() {
            super();
        }

        connectedCallback() {
            let postId = this.getAttribute('post-id');
            if (postId === null) return;
            fetch(baseUrl + '/list/' + postId).then(response => response.json()).then(data => this._render(data)).catch(console.log);
        }

        _render(comments: Comments) {
            let template = this._getTemplate('comments');
            let block = template.content.cloneNode(true) as ParentNode;
            let numComments = comments.List.length;
            this._keepOneChildWithId(block, numComments == 1, 'singular_count', 'plural_count');
            this._keepOneChildWithId(block, comments.IsWritable, 'entryform', 'noentryform');
            this._replaceContentWithId(block, 'count', String(comments.List.length));
            let list = block.querySelector('#comments');
            if (list) this._renderComments(list, comments);
            this.appendChild(block);
        }

        _renderComments(list: Element, comments: Comments) {
            let template = this._getTemplate('comment');
            for (let comment of comments.List) {
                let row = template.content.cloneNode(true) as ParentNode;
                this._replaceContentWithId(row, 'author', comment.Author);
                this._replaceContentWithId(row, 'when', new Date(comment.When).toLocaleString());
                this._replaceHtmlWithId(row, 'text', comment.Text);
                list.appendChild(row);
            }
        }

        _getTemplate(id: string) {
            let template = this.querySelector('#' + id) as HTMLTemplateElement;
            if (template) {
                template.remove();
                return template;
            }
            template = document.createElement('template');
            template.innerHTML = defaultTemplates[id];
            return template;
        }

        _keepOneChildWithId(parent: ParentNode, selector: boolean, idTrue: string, idFalse: string) {
            if (selector) {
                parent.querySelector('#' + idFalse)?.remove();
                parent.querySelector('#' + idTrue)?.removeAttribute('id');
            } else {
                parent.querySelector('#' + idTrue)?.remove();
                parent.querySelector('#' + idFalse)?.removeAttribute('id');
            }
        }

        _replaceContentWithId(parent: ParentNode, id: string, content: string) {
            for (let child of parent.querySelectorAll('#' + id)) {
                child.removeAttribute('id');
                child.textContent = content;    
            }
        }

        _replaceHtmlWithId(parent: ParentNode, id: string, content: string) {
            for (let child of parent.querySelectorAll('#' + id)) {
                child.removeAttribute('id');
                child.innerHTML = content;    
            }
        }

    });
})()