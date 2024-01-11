export function initTabs(tabTitle: string, tabPage: string) {
    new Tabs(tabTitle, tabPage);
}

interface TabElement extends Element {
    tabNum: number;
}

class Tabs {
    constructor(private tabTitle: string, private tabPage: string) {
        this.titles = this.findElements(this.tabTitle);
        this.pages = this.findElements(this.tabPage);
        for (let title of this.titles) {
            title.addEventListener('click', e => this.switchTab(e));
        }
    }

    private titles: TabElement[];
    private pages: TabElement[];

    private getNum(elem: Element, name: string): number | null {
        let id = elem.id;
        if (!id.startsWith(name + '-')) return null;
        return Number(id.substring(name.length + 1));
    }

    private findElements(name: string): TabElement[] {
        let elems = document.querySelectorAll(`[id|=${name}]`);
        let out: TabElement[] = [];
        for (let elem of elems) {
            let num = this.getNum(elem, name);
            if (num != null) {
                let tabElem = elem as TabElement;
                tabElem.tabNum = num;
                out.push(tabElem);
            }
        }
        return out;
    }

    private switchTab(e: Event) {
        let elem = e.target as TabElement;
        if (elem.tabNum !== undefined) this.selectTab(elem.tabNum);
    }

    private selectTab(num: number) {
        for (let title of this.titles) {
            if (title.tabNum == num) {
                title.classList.add('tabSelected');
            } else {
                title.classList.remove('tabSelected');
            }
        }
        for (let page of this.pages) {
            if (page.tabNum == num) {
                page.classList.add('tabOpen');
            } else {
                page.classList.remove('tabOpen');
            }
        }
    }
}