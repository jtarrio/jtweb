import applyTemplate from "./templates";
import * as Lang from "./languages";
import * as Preview from "./preview";
import { Comments, UserApi } from "./api";

const AnchorPrefix = 'comment_';

class JtCommentsElement extends HTMLElement {
    private api: UserApi;
    private postId: string | null;
    private allTemplate: DocumentFragment;

    constructor() {
        super();
        this.api = new UserApi();
    }

    connectedCallback() {
        this.postId = this.getAttribute('post-id');
        this.allTemplate = this.getTemplate();
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
        let block = this.allTemplate.cloneNode(true) as Element;
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
            if (element) element.scrollIntoView();
        }
    }

    private attachFormEvents(form: Element) {
        form.querySelector('form')?.addEventListener('submit', e => {
            this.submitComment(e.target as HTMLFormElement);
            e.preventDefault();
        });
        let commentBox = form.querySelector('#jtComment');
        let previewButton = form.querySelector('#jtPreviewButton');
        let previewBox = form.querySelector('#jtPreviewBox');
        let containerBox = form.querySelector('#jtPreviewContainer')
        if (commentBox && previewBox) {
            Preview.setup({
                input: commentBox as HTMLTextAreaElement,
                output: previewBox as HTMLElement,
                toggle: previewButton as HTMLElement | null,
                container: containerBox as HTMLElement | null,
                api: this.api
            });
        }
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
        } catch {
            msg = Lang.MessageType.ErrorPostingComment;
        }
        let p = document.createElement('p');
        p.classList.add("jtSubmitMessage");
        p.textContent = Lang.getMessage(msg);
        form.insertAdjacentElement("beforebegin", p);
        p.scrollIntoView();
    }

    private getTemplate(): DocumentFragment {
        let template = this.getElementsByTagName('template')[0];
        if (template) {
            template.remove();
            return template.content;
        }
        template = document.createElement('template');
        template.innerHTML = Lang.getTemplate();
        return template.content;
    }
}

window.addEventListener('DOMContentLoaded', _ =>
    customElements.define('jt-comments', JtCommentsElement));
