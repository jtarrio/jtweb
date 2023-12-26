import applyTemplate from "./templates";
import * as Lang from "./languages";

type Comments = {
    PostId: string,
    IsAvailable: boolean,
    IsWritable: boolean,
    List: Comment[],
};

type Comment = {
    Id: string,
    Visible: string,
    Author: string,
    When: string,
    Text: string,
};

function rndName() {
    let out: string[] = [];
    for (let i = 0; i < 8; ++i) {
        out.push(String.fromCharCode(97 + Math.random() * 26));
    }
    return out.join('');
}

class JtCommentsElement extends HTMLElement {
    apiUrl: string;
    postId: string | null;
    allTemplate: DocumentFragment;
    commentTemplate: DocumentFragment;
    formTemplate: DocumentFragment;

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

    private async refresh() {
        while (this.firstChild != null) {
            this.removeChild(this.firstChild);
        }
        if (this.postId === null) return;
        let response = await fetch(this.apiUrl + '/list/' + this.postId);
        if (response.status == 200) {
            this.render(await response.json());
        }
    }

    private render(comments: Comments) {
        if (!comments.IsAvailable) {
            this.remove();
            return;
        }

        let numComments = comments.List.length;
        let block = this.allTemplate.cloneNode(true);
        applyTemplate(block as Element, {
            'singular_count': (numComments == 1),
            'plural_count': (numComments != 1),
            'count': String(numComments),
            'comments': (c: Element) => { this.renderComments(c, comments); },
            'newcomment': comments.IsWritable ? (c: Element) => { this.renderForm(c); } : false,
        });
        this.appendChild(block);
    }

    private renderComments(list: Element, comments: Comments) {
        for (let comment of comments.List) {
            let row = this.commentTemplate.cloneNode(true);
            applyTemplate(row as Element, {
                'author': comment.Author,
                'when': Lang.formatDate(comment.When),
                'url': new URL('#c' + comment.Id, window.location.toString()).toString(),
                'anchor': 'c' + comment.Id,
                'text': { html: comment.Text },
            });
            list.appendChild(row);
        }
    }

    private renderForm(elem: Element) {
        let form = this.formTemplate.cloneNode(true) as Element;
        form.querySelector('form')?.addEventListener('submit', e => {
            this.submitComment(e.target as HTMLFormElement);
            e.preventDefault();
        });
        elem.appendChild(form);
    }

    private async submitComment(form: HTMLFormElement) {
        let formData = new FormData(form);
        let data = JSON.stringify({
            'PostId': this.postId,
            'Author': formData.get('author'),
            'Text': formData.get('text'),
        });
        let response = await fetch(this.apiUrl + '/add', { method: "POST", body: data });
        let msg = Lang.MessageType.ErrorPostingComment;
        if (response.status == 200) {
            form.reset();
            let comment = await response.json();
            if (comment.Visible) {
                this.refresh();
                return;
            }
            msg = Lang.MessageType.CommentPostedAsDraft;
        }
        let p = document.createElement('p');
        p.textContent = Lang.getMessage(Lang.MessageType.CommentPostedAsDraft);
        form.insertAdjacentElement("beforebegin", p);
    }

    private getTemplate(id?: string): DocumentFragment {
        let name = 'jt-comments' + (id ? '-' + id : '');
        let template = document.getElementById(name) as HTMLTemplateElement;
        if (template) return template.content;
        template = document.createElement('template');
        template.innerHTML = Lang.getTemplate(id ? id : 'main');
        return template.content;
    }
}

customElements.define('jt-comments', JtCommentsElement);
