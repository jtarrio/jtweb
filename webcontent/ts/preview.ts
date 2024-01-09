import { UserApi } from "./api";

export function setup(params: {
    toggle: HTMLElement,
    input: HTMLTextAreaElement,
    output: HTMLElement,
    container: HTMLElement | null,
    api: UserApi
}) {
    new Preview(params.api, params.input, params.output, params.container || params.output, params.toggle);
}

class Preview {
    constructor(
        private api: UserApi,
        private input: HTMLTextAreaElement,
        private output: HTMLElement,
        private container: HTMLElement,
        toggle: HTMLElement) {
        this.previewFn = _ => this.launchPreview();
        this.timeout = undefined;
        this.lastPreview = '';
        toggle.addEventListener('click', _ => this.togglePreview());
    }

    private static PreviewInterval = 1000;

    private previewFn: (e: Event) => void;
    private timeout: number | undefined;
    private lastPreview: string;

    visible() {
        return this.container.classList.contains('jtPreview');
    }

    togglePreview() {
        if (this.visible()) {
            this.input.removeEventListener('input', this.previewFn);
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
        this.doPreview();
    }

    launchPreview() {
        if (this.timeout !== undefined) return;
        this.timeout = setTimeout(() => this.doPreview(), Preview.PreviewInterval);
    }

    async doPreview() {
        if (!this.visible()) return;
        let text = this.input.value;
        if (this.lastPreview == text) return;
        this.timeout = undefined;
        let result = this.api.render(text);
        this.lastPreview = text;
        this.output.innerHTML = result['Text'];
    }
}