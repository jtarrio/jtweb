export function setup(params: {
    toggle: HTMLElement,
    input: HTMLTextAreaElement,
    output: HTMLElement,
    container: HTMLElement | null,
    apiUrl: string
}) {
    new Preview(params.apiUrl, params.input, params.output, params.container || params.output, params.toggle);
}

class Preview {
    constructor(
        private apiUrl: string,
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
        let content = JSON.stringify({ 'Text': text });
        let data = await fetch(this.apiUrl + '/render', { method: 'POST', body: content });
        if (data.status != 200) return;
        let result = await data.json();
        this.lastPreview = text;
        this.output.innerHTML = result['Text'];
    }
}