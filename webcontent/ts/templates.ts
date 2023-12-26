export type TmplReplaceFn = (element: Element) => void
export type TmplContent = { text?: string, html?: string, visible?: boolean };
export type TmplReplacer = string | boolean | TmplContent | TmplReplaceFn;
export type TmplMap = { [key: string]: TmplReplacer }

export default function applyTemplate(element: Element, map: TmplMap) {
    let elements = findPlaceholders(element);
    for (let elem of elements) {
        if (elem.nodeName == 'JTVAR') {
            replacePlaceholder(elem, map);
        } else if (elem.hasAttribute('jtvar')) {
            fillElement(elem, map);
        } else {
            fillAttributes(elem, map);
        }
    }
}

function findPlaceholders(root: Element) {
    return new class implements Iterable<Element> {
        [Symbol.iterator](): Iterator<Element> {
            return new class implements Iterator<Element> {
                iter: IterableIterator<[number, Element]>;
                constructor() {
                    this.iter = root.querySelectorAll('*').entries();
                }
                next(): IteratorResult<Element> {
                    while (true) {
                        let next = this.iter.next();
                        if (next.done) return { done: true, value: undefined };
                        const elem = next.value[1];
                        if (elem.nodeName == 'JTVAR' || elem.hasAttribute('jtvar'))
                            return { done: false, value: elem };
                        for (let attr of elem.attributes) {
                            if (attr.value.startsWith('jtvar ')) return { done: false, value: elem };
                        }
                    }
                }
            }
        }
    };
}

function replacePlaceholder(elem: Element, map: TmplMap) {
    for (let attr of elem.attributes) {
        let replacer = map[attr.name];
        if (replacer === undefined) continue;
        if ("function" === typeof replacer) {
            replacer(elem);
        } else if ("string" === typeof replacer) {
            elem.insertAdjacentText("afterend", replacer);
        } else if ("object" === typeof replacer) {
            if (replacer.html !== undefined) {
                elem.insertAdjacentHTML("afterend", replacer.html);
            } else if (replacer.text !== undefined) {
                elem.insertAdjacentText("afterend", replacer.text);
            }
        }
        break;
    }
    elem.remove();
}

function fillElement(elem: Element, map: TmplMap) {
    let name = elem.getAttribute('jtvar');
    if (!name) return;
    let replacer = map[name];
    if (replacer === undefined) {
        elem.removeAttribute('jtvar');
        return;
    }
    if ("function" === typeof replacer) {
        replacer(elem);
    } else if ("string" === typeof replacer) {
        elem.textContent = replacer;
    } else if ("boolean" === typeof replacer) {
        if (!replacer) {
            elem.remove();
            return;
        }
    } else if ("object" === typeof replacer) {
        if (replacer.html !== undefined) {
            elem.innerHTML = replacer.html;
        } else if (replacer.text !== undefined) {
            elem.textContent = replacer.text;
        } else if (replacer.visible !== undefined) {
            if (!replacer.visible) {
                elem.remove();
                return;
            }
        }
    }
    elem.removeAttribute('jtvar');
}

function fillAttributes(elem: Element, map: TmplMap) {
    for (let attr of elem.attributes) {
        if (!attr.value.startsWith('jtvar ')) continue;
        let name = attr.value.substring(6).trim();
        if (!name) continue;
        let replacer = map[name];
        if (replacer === undefined) {
            elem.removeAttribute(attr.name);
        } else if ("string" === typeof replacer) {
            elem.setAttribute(attr.name, replacer);
        } else if ("object" === typeof replacer && replacer.text !== undefined) {
            elem.setAttribute(attr.name, replacer.text);
        }
    }
}
