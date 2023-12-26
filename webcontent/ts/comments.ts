import applyTemplate from "./templates";
import { getTemplate, formatDate } from "./languages";

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

class JtCommentsElement extends HTMLElement {
    _baseUrl: string;

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
        this._baseUrl = baseUrl.pathname += '_';
    }

    connectedCallback() {
        let postId = this.getAttribute('post-id');
        if (postId === null) return;
        fetch(this._baseUrl + '/list/' + postId).
            then(response => response.json()).
            then(data => this.render(data)).
            catch(console.log);
    }

    private render(comments: Comments) {
        if (!comments.IsAvailable) {
            this.remove();
            return;
        }

        let template = this.getTemplate('comments');
        let numComments = comments.List.length;
        let block = template.cloneNode(true);
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
        let template = this.getTemplate('comment');
        for (let comment of comments.List) {
            let row = template.cloneNode(true);
            applyTemplate(row as Element, {
                'author': comment.Author,
                'when': formatDate(comment.When),
                'url': new URL('#c' + comment.Id, window.location.toString()).toString(),
                'anchor': 'c' + comment.Id,
                'text': { html: comment.Text },
            });
            list.appendChild(row);
        }
    }

    private renderForm(elem: Element) {
        let template = this.getTemplate('commentform');
        elem.appendChild(template.cloneNode(true));
    }

    private getTemplate(id: string): DocumentFragment {
        let template = this.querySelector('#' + id) as HTMLTemplateElement;
        if (template) {
            template.remove();
            return template.content;
        }
        template = document.createElement('template');
        template.innerHTML = getTemplate(id);
        return template.content;
    }
}

customElements.define('jt-comments', JtCommentsElement);
