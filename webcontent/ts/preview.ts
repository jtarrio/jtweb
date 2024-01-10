import { UserApi } from "./api";

export function setup(params: {
    input: HTMLTextAreaElement,
    output: HTMLElement,
    toggle: HTMLElement | null,
    container: HTMLElement | null,
    api: UserApi
}) {
    new Preview(params.api, params.input, params.output, params.container || params.output, params.toggle);
}

function findForm(element: HTMLElement): HTMLFormElement | null {
    let current: Element | null = element;
    while (current !== null && current.tagName != 'FORM') current = current.parentElement;
    return current as HTMLFormElement | null;
}

class Preview {
    constructor(
        private api: UserApi,
        private input: HTMLTextAreaElement,
        private output: HTMLElement,
        private container: HTMLElement,
        toggle: HTMLElement | null) {
        this.form = findForm(this.input);
        this.previewFn = _ => this.launchPreview();
        this.resetPreviewFn = _ => this.resetPreview();
        this.timeout = undefined;
        this.lastPreview = ['', ''];
        if (toggle === null) {
            this.togglePreview();
        } else {
            toggle.addEventListener('click', _ => this.togglePreview());
        }
    }

    private static PreviewInterval = 1000;

    private form: HTMLFormElement | null;
    private previewFn: (e: Event) => void;
    private resetPreviewFn: (e: Event) => void;
    private timeout: number | undefined;
    private lastPreview: [string, string];

    visible() {
        return this.container.classList.contains('jtPreview');
    }

    togglePreview() {
        if (this.visible()) {
            this.input.removeEventListener('input', this.previewFn);
            this.form?.removeEventListener('reset', this.resetPreviewFn);
            while (true) {
                let child = this.output.firstChild;
                if (!child) break;
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

    private launchPreview() {
        if (this.timeout !== undefined) return;
        this.timeout = setTimeout(() => this.doPreview(), Preview.PreviewInterval);
    }

    private async doPreview() {
        if (!this.visible()) return;
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

    private async resetPreview() {
        this.output.innerHTML = '';
    }
}