import applyTemplate from "./templates";
import * as Lang from "./languages";
import * as Preview from "./preview";
import { Comments, UserApi } from "./api";

class JtCommentsElement extends HTMLElement {
    private api: UserApi;
    private postId: string | null;
    private allTemplate: DocumentFragment;
    private commentTemplate: DocumentFragment;
    private formTemplate: DocumentFragment;

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

    private async refresh() {
        while (this.firstChild != null) {
            this.removeChild(this.firstChild);
        }
        if (this.postId === null) return;
        this.render(await this.api.list(this.postId));
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
        let commentBox = form.querySelector('#jtComment');
        let previewButton = form.querySelector('#jtPreviewButton');
        let previewBox = form.querySelector('#jtPreviewBox');
        let containerBox = form.querySelector('#jtPreviewContainer')
        if (commentBox && previewButton && previewBox) {
            Preview.setup({
                toggle: previewButton as HTMLElement,
                input: commentBox as HTMLTextAreaElement,
                output: previewBox as HTMLElement,
                container: containerBox as HTMLElement | null,
                api: this.api
            });
        }
        elem.appendChild(form);
    }

    private async submitComment(form: HTMLFormElement) {
        let msg: Lang.MessageType;
        let formData = new FormData(form);
        try {
            let comment = await this.api.add({
                PostId: this.postId!,
                Author: formData.get('author')! as string,
                Text: formData.get('text')! as string,
            });
            form.reset();
            if (comment.Visible) {
                this.refresh();
                return;
            }
            msg = Lang.MessageType.CommentPostedAsDraft;
        } catch (_) {
            msg = Lang.MessageType.ErrorPostingComment;
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
